package ppcs

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Live is an authenticated PPCS media session. It de-obfuscates and
// reassembles incoming DRW frames and writes Annex-B H.264 access units
// unchanged to the configured output.
type Live struct {
	session        *localSession
	key            []byte
	xorCtl         byte
	rx             map[byte]*channelReceiver
	cameraIndex    int
	controlMu      sync.Mutex
	batteryHandler func(channel, capacity int, charging bool)
}

// OpenLive starts a media session for one camera channel and quality mode.
func OpenLive(ctx context.Context, ip, did, password string, cameraIndex, quality int) (*Live, error) {
	return openLiveAt(ctx, ip, did, password, cameraIndex, quality, &net.UDPAddr{IP: net.IPv4bcast, Port: LanSearchPort})
}

func openLiveAt(ctx context.Context, ip, did, password string, cameraIndex, quality int, searchTarget *net.UDPAddr) (*Live, error) {
	session, key, _, err := openAuthenticatedAt(ctx, ip, did, password, searchTarget)
	if err != nil {
		return nil, err
	}
	live := &Live{
		session:     session,
		key:         key,
		xorCtl:      controlInitialXor,
		rx:          make(map[byte]*channelReceiver),
		cameraIndex: cameraIndex,
	}

	indexPayload := make([]byte, controlIndexSize)
	indexPayload[controlIndexCamera] = byte(cameraIndex)
	indexPayload[controlIndexQuality] = byte(quality)
	video, err := live.encodeControl(controlVideoType, indexPayload)
	if err != nil {
		live.Close()
		return nil, err
	}
	audio, err := live.encodeControl(controlAudioType, indexPayload)
	if err != nil {
		live.Close()
		return nil, err
	}
	if err := session.send(buildDRW(controlChannel, session.sequence, append(video, audio...))); err != nil {
		live.Close()
		return nil, err
	}
	session.sequence++
	_ = live.requestBattery()
	return live, nil
}

// SetBatteryHandler registers a callback for battery status frames.
func (l *Live) SetBatteryHandler(handler func(channel, capacity int, charging bool)) {
	l.batteryHandler = handler
}

// Close tears down the underlying UDP session.
func (l *Live) Close() {
	if l != nil && l.session != nil {
		l.session.close()
		l.session = nil
	}
}

// Run receives, acknowledges and reassembles PPCS media frames. Annex-B H.264
// access units are written unchanged to output.
func (l *Live) Run(ctx context.Context, output io.Writer) error {
	defer l.Close()
	keepAlive := time.NewTicker(liveKeepAlive)
	defer keepAlive.Stop()
	batteryTicker := time.NewTicker(liveBatteryPeriod)
	defer batteryTicker.Stop()
	buffer := make([]byte, liveReadBuffer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-keepAlive.C:
			_ = l.session.send(Packet(MsgAlive, nil))
		case <-batteryTicker.C:
			_ = l.requestBattery()
		default:
		}

		_ = l.session.conn.SetReadDeadline(shortDeadline(ctx, liveReadPoll))
		n, remote, err := l.session.conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr interface{ Timeout() bool }
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			return err
		}
		if remote.String() != l.session.peer.String() || n < ppcsHeaderSize {
			continue
		}
		header, err := ParseHeader(buffer[:n])
		if err != nil {
			continue
		}
		switch header.Type {
		case MsgAlive:
			_ = l.session.send(Packet(MsgAliveAck, nil))
		case MsgClose:
			return ErrPeerClosed
		case msgDRW:
			channel, sequence, chunk, err := parseDRW(buffer[:n])
			if err != nil {
				continue
			}
			_ = l.session.send(buildDRWAck(channel, []uint16{sequence}))
			receiver := l.rx[channel]
			if receiver == nil {
				receiver = &channelReceiver{assembler: &appAssembler{}}
				l.rx[channel] = receiver
			}
			for _, ordered := range receiver.add(sequence, chunk) {
				for _, frame := range receiver.assembler.feed(ordered) {
					if l.handleControl(frame) {
						continue
					}
					payload, video, err := l.decodeMedia(frame)
					if err == nil && video && len(payload) > 0 {
						if _, err := output.Write(payload); err != nil {
							return err
						}
					}
				}
			}
		}
	}
}

func (l *Live) requestBattery() error {
	payload := make([]byte, controlHeaderSize)
	payload[0] = byte(l.cameraIndex)
	return l.sendControl(controlBattery, payload)
}

func (l *Live) sendControl(controlType uint16, payload []byte) error {
	l.controlMu.Lock()
	defer l.controlMu.Unlock()
	frame, err := l.encodeControl(controlType, payload)
	if err != nil {
		return err
	}
	sequence := l.session.sequence
	l.session.sequence++
	return l.session.send(buildDRW(controlChannel, sequence, frame))
}

func (l *Live) handleControl(frame []byte) bool {
	controlType, payload, ok := l.decodeControl(frame)
	if !ok {
		return false
	}
	if controlType == controlBattery && len(payload) >= 3 && l.batteryHandler != nil {
		l.batteryHandler(int(payload[0]), int(payload[1]), payload[2] == batteryChargingValue)
	}
	return true
}

func (l *Live) decodeControl(frame []byte) (uint16, []byte, bool) {
	streamType, body, err := parseAppFrame(frame)
	if err != nil || streamType != controlStream || len(body) < controlHeaderSize {
		return 0, nil, false
	}
	block, err := aes.NewCipher(l.key)
	if err != nil {
		return 0, nil, true
	}
	head := make([]byte, controlHeaderSize)
	block.Decrypt(head, body[:controlHeaderSize])
	controlType := binary.LittleEndian.Uint16(head[controlTypeOffset : controlTypeOffset+controlTypeSize])
	dataSize := int(binary.LittleEndian.Uint16(head[controlSizeOffset : controlSizeOffset+controlSizeSize]))
	if dataSize > len(body)-controlHeaderSize {
		return controlType, nil, true
	}
	payload := append([]byte(nil), body[controlHeaderSize:controlHeaderSize+dataSize]...)
	for index := range payload {
		payload[index] ^= head[controlXorOffset]
	}
	return controlType, payload, true
}

func (l *Live) encodeControl(controlType uint16, payload []byte) ([]byte, error) {
	bodyLength := controlHeaderSize + len(payload)
	frame := make([]byte, appFrameHeaderSize+bodyLength)
	frame[0], frame[1], frame[2], frame[appStreamOffset] = byte(bodyLength), byte(bodyLength>>8), byte(bodyLength>>16), controlStream
	head := frame[appFrameHeaderSize : appFrameHeaderSize+controlHeaderSize]
	binary.LittleEndian.PutUint16(head[controlTypeOffset:controlTypeOffset+controlTypeSize], controlType)
	binary.LittleEndian.PutUint16(head[controlSizeOffset:controlSizeOffset+controlSizeSize], uint16(len(payload)))
	if len(payload) > 0 {
		head[controlXorOffset] = l.xorCtl
		copy(frame[controlPayloadOffset:], payload)
		for index := controlPayloadOffset; index < len(frame); index++ {
			frame[index] ^= l.xorCtl
		}
		l.xorCtl = frame[0] ^ l.xorCtl
	}
	encrypted, err := aesEncrypt(l.key, head)
	if err != nil {
		return nil, err
	}
	copy(head, encrypted)
	return frame, nil
}

func (l *Live) decodeMedia(frame []byte) ([]byte, bool, error) {
	streamType, body, err := parseAppFrame(frame)
	if err != nil || streamType == authStream || streamType == controlStream || len(body) < mediaHeaderSize {
		return nil, false, err
	}
	block, err := aes.NewCipher(l.key)
	if err != nil {
		return nil, false, err
	}
	head := make([]byte, mediaHeaderSize)
	block.Decrypt(head, body[:mediaHeaderSize])
	dataSize := int(binary.LittleEndian.Uint32(head[mediaSizeOffset : mediaSizeOffset+mediaSizeLength]))
	if dataSize < 0 || dataSize > len(body)-mediaHeaderSize {
		return nil, false, fmt.Errorf("%w: %d", ErrMediaSize, dataSize)
	}
	payload := append([]byte(nil), body[mediaHeaderSize:mediaHeaderSize+dataSize]...)
	for index := range payload {
		payload[index] ^= head[controlXorOffset]
	}
	return payload, looksLikeH264(payload), nil
}

func looksLikeH264(payload []byte) bool {
	for _, marker := range h264Markers {
		if bytes.Contains(payload, marker) {
			return true
		}
	}
	return false
}

// channelReceiver orders DRW fragments by sequence number and tolerates
// 16-bit wraparound.
type channelReceiver struct {
	started   bool
	expected  uint16
	pending   map[uint16][]byte
	assembler *appAssembler
}

func (r *channelReceiver) add(sequence uint16, data []byte) [][]byte {
	if !r.started {
		r.started = true
		r.expected = sequence
		r.pending = make(map[uint16][]byte)
	}
	r.pending[sequence] = append([]byte(nil), data...)
	var output [][]byte
	for {
		chunk, ok := r.pending[r.expected]
		if !ok {
			break
		}
		delete(r.pending, r.expected)
		output = append(output, chunk)
		r.expected++
	}
	return output
}

// appAssembler reassembles 24-bit length-prefixed application frames from a
// byte stream and resynchronizes by dropping one byte on implausible prefixes.
type appAssembler struct{ buffer []byte }

func (a *appAssembler) feed(data []byte) [][]byte {
	a.buffer = append(a.buffer, data...)
	var frames [][]byte
	for len(a.buffer) >= appFrameHeaderSize {
		length := int(a.buffer[0]) | int(a.buffer[1])<<8 | int(a.buffer[2])<<16
		streamType := a.buffer[appStreamOffset]
		if length < appFrameMinLength || length > appFrameMaxLength || streamType < appStreamMin || streamType > appStreamMax {
			a.buffer = a.buffer[resyncStep:]
			continue
		}
		total := appFrameHeaderSize + length
		if len(a.buffer) < total {
			break
		}
		frames = append(frames, append([]byte(nil), a.buffer[:total]...))
		a.buffer = a.buffer[total:]
	}
	return frames
}
