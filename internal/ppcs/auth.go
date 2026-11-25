package ppcs

import (
	"context"
	"crypto/aes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// AuthResult describes a successful local authentication round trip.
type AuthResult struct {
	IP            string   `json:"ip"`
	Port          int      `json:"port"`
	DID           string   `json:"did"`
	Authenticated bool     `json:"authenticated"`
	RoundTrip     string   `json:"roundTrip"`
	Messages      []string `json:"messages"`
}

type localSession struct {
	conn     *net.UDPConn
	peer     *net.UDPAddr
	did      string
	sequence uint16
	writeMu  sync.Mutex
}

// Authenticate proves the security code against the base station and closes
// the session immediately. The code stays in memory only.
func Authenticate(ctx context.Context, wantedIP, wantedDID, password string) (AuthResult, error) {
	started := time.Now()
	session, key, messages, err := openAuthenticated(ctx, wantedIP, wantedDID, password)
	if err != nil {
		return AuthResult{}, err
	}
	defer session.close()
	_ = key
	messages = append(messages, "CLOSE")
	return AuthResult{
		IP:            session.peer.IP.String(),
		Port:          session.peer.Port,
		DID:           session.did,
		Authenticated: true,
		RoundTrip:     time.Since(started).Round(time.Millisecond).String(),
		Messages:      messages,
	}, nil
}

func openAuthenticated(ctx context.Context, wantedIP, wantedDID, password string) (*localSession, []byte, []string, error) {
	return openAuthenticatedAt(ctx, wantedIP, wantedDID, password, &net.UDPAddr{IP: net.IPv4bcast, Port: LanSearchPort})
}

func openAuthenticatedAt(ctx context.Context, wantedIP, wantedDID, password string, searchTarget *net.UDPAddr) (*localSession, []byte, []string, error) {
	if password == "" || len(password) > authPasswordMax {
		return nil, nil, nil, ErrInvalidSecurityCode
	}
	session, messages, err := openLocalAt(ctx, wantedIP, wantedDID, searchTarget)
	if err != nil {
		return nil, nil, messages, err
	}

	buffer := make([]byte, authBufferSize)
	for {
		if err := ctx.Err(); err != nil {
			session.close()
			return nil, nil, messages, fmt.Errorf("%w: %w", ErrAuthTimeout, err)
		}
		packet, err := session.read(ctx, buffer)
		if err != nil {
			session.close()
			return nil, nil, messages, err
		}
		header, err := ParseHeader(packet)
		if err != nil {
			continue
		}
		switch header.Type {
		case MsgP2PReady:
			continue
		case MsgAlive:
			_ = session.send(Packet(MsgAliveAck, nil))
		case msgDRW:
			channel, sequence, appData, err := parseDRW(packet)
			if err != nil {
				continue
			}
			_ = session.send(buildDRWAck(channel, []uint16{sequence}))
			if channel != authAckChannel {
				continue
			}
			streamType, body, err := parseAppFrame(appData)
			if err != nil || streamType != authStream || len(body) < authControlSize {
				continue
			}
			authType := binary.LittleEndian.Uint16(body[authTypeOffset : authTypeOffset+authTypeSize])
			dataSize := int(binary.LittleEndian.Uint16(body[authSizeOffset : authSizeOffset+authSizeSize]))
			switch authType {
			case authChallenge:
				if dataSize != authChallengeLen || len(body) < authControlSize+authBlockSize {
					session.close()
					return nil, nil, messages, ErrAuthChallenge
				}
				response, err := aesEncrypt(passwordKey(password), body[authControlSize:authControlSize+authBlockSize])
				if err != nil {
					session.close()
					return nil, nil, messages, err
				}
				if err := session.sendAuth(authResponse, response); err != nil {
					session.close()
					return nil, nil, messages, err
				}
				messages = append(messages, "AUTH_CHALLENGE", "AUTH_RESPONSE")
			case authSuccess:
				if dataSize != authChallengeLen || len(body) < authControlSize+authBlockSize {
					session.close()
					return nil, nil, messages, ErrAuthKey
				}
				key, err := aesDecrypt(passwordKey(password), body[authControlSize:authControlSize+authBlockSize])
				if err != nil {
					session.close()
					return nil, nil, messages, err
				}
				messages = append(messages, "AUTH_SUCCESS")
				return session, key, messages, nil
			case authFailed:
				session.close()
				return nil, nil, messages, ErrAuthRejected
			}
		}
	}
}

func openLocal(ctx context.Context, wantedIP, wantedDID string) (*localSession, []string, error) {
	return openLocalAt(ctx, wantedIP, wantedDID, &net.UDPAddr{IP: net.IPv4bcast, Port: LanSearchPort})
}

func openLocalAt(ctx context.Context, wantedIP, wantedDID string, target *net.UDPAddr) (*localSession, []string, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, nil, err
	}
	if err := enableBroadcast(conn); err != nil {
		conn.Close()
		return nil, nil, err
	}

	wantedDID = strings.ToUpper(strings.TrimSpace(wantedDID))
	search := Packet(MsgLanSearch, nil)
	buffer := make([]byte, searchBufferSize)
	var punchPacket []byte
	var punch Punch
	var peer *net.UDPAddr
	messages := make([]string, 0, 8)

	for punchPacket == nil {
		if err := ctx.Err(); err != nil {
			conn.Close()
			return nil, messages, fmt.Errorf("%w: %w", ErrLanTimeout, err)
		}
		_, _ = conn.WriteToUDP(search, target)
		_ = conn.SetReadDeadline(shortDeadline(ctx, punchReadPoll))
		n, remote, readErr := conn.ReadFromUDP(buffer)
		if readErr != nil {
			var netErr net.Error
			if errors.As(readErr, &netErr) && netErr.Timeout() {
				continue
			}
			conn.Close()
			return nil, messages, readErr
		}
		candidate, parseErr := ParsePunch(buffer[:n])
		if parseErr != nil || (wantedIP != "" && remote.IP.String() != wantedIP) || (wantedDID != "" && candidate.DID() != wantedDID) {
			continue
		}
		punch, peer = candidate, remote
		punchPacket = append([]byte(nil), buffer[:n]...)
		messages = append(messages, "PUNCH_PKT")
	}

	session := &localSession{conn: conn, peer: peer, did: punch.DID()}
	if err := session.send(punchPacket); err != nil {
		session.close()
		return nil, messages, err
	}
	messages = append(messages, "PUNCH_ECHO")

	for {
		packet, err := session.read(ctx, buffer)
		if err != nil {
			session.close()
			return nil, messages, fmt.Errorf("%w: %w", ErrP2PReadyTimeout, err)
		}
		header, err := ParseHeader(packet)
		if err == nil && header.Type == MsgP2PReady {
			messages = append(messages, "P2P_RDY")
			return session, messages, nil
		}
	}
}

func (s *localSession) read(ctx context.Context, buffer []byte) ([]byte, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		_ = s.conn.SetReadDeadline(shortDeadline(ctx, sessionReadPoll))
		n, remote, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			return nil, err
		}
		if remote.String() == s.peer.String() {
			return append([]byte(nil), buffer[:n]...), nil
		}
	}
}

func (s *localSession) send(packet []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.conn.WriteToUDP(packet, s.peer)
	return err
}

func (s *localSession) sendAuth(authType uint16, payload []byte) error {
	body := make([]byte, authControlSize+len(payload))
	binary.LittleEndian.PutUint16(body[authTypeOffset:authTypeOffset+authTypeSize], authType)
	binary.LittleEndian.PutUint16(body[authSizeOffset:authSizeOffset+authSizeSize], uint16(len(payload)))
	copy(body[authControlSize:], payload)
	app := make([]byte, appFrameHeaderSize+len(body))
	length := len(body)
	app[0], app[1], app[2], app[appStreamOffset] = byte(length), byte(length>>8), byte(length>>16), authStream
	copy(app[appFrameHeaderSize:], body)
	packet := buildDRW(authAckChannel, s.sequence, app)
	s.sequence++
	return s.send(packet)
}

func (s *localSession) close() {
	if s.conn == nil {
		return
	}
	_ = s.send(Packet(MsgClose, nil))
	_ = s.conn.Close()
	s.conn = nil
}

func shortDeadline(ctx context.Context, duration time.Duration) time.Time {
	deadline := time.Now().Add(duration)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		return ctxDeadline
	}
	return deadline
}

func buildDRW(channel byte, sequence uint16, app []byte) []byte {
	payload := make([]byte, drwPayloadDataOff+len(app))
	payload[drwPayloadMarkerOff] = drwMarker
	payload[drwPayloadChannelOff] = channel
	binary.BigEndian.PutUint16(payload[drwPayloadSequenceOff:drwPayloadSequenceOff+drwSequenceSize], sequence)
	copy(payload[drwPayloadDataOff:], app)
	return Packet(msgDRW, payload)
}

func parseDRW(packet []byte) (byte, uint16, []byte, error) {
	header, err := ParseHeader(packet)
	if err != nil || header.Type != msgDRW || len(packet) < drwMinLength || packet[drwPayloadMarkerOff+ppcsHeaderSize] != drwMarker {
		return 0, 0, nil, ErrInvalidDRW
	}
	end := ppcsHeaderSize + int(header.Length)
	if end > len(packet) {
		return 0, 0, nil, fmt.Errorf("%w: truncated", ErrInvalidDRW)
	}
	return packet[drwChannelOffset], binary.BigEndian.Uint16(packet[drwSequenceOffset : drwSequenceOffset+drwSequenceSize]), packet[drwPayloadOffset:end], nil
}

func buildDRWAck(channel byte, sequences []uint16) []byte {
	payload := make([]byte, drwAckHeaderSize+drwSequenceSize*len(sequences))
	payload[drwPayloadMarkerOff] = drwMarker
	payload[drwPayloadChannelOff] = channel
	binary.BigEndian.PutUint16(payload[drwAckCountOff:drwAckCountOff+drwSequenceSize], uint16(len(sequences)))
	for index, sequence := range sequences {
		offset := drwAckSequenceOff + drwSequenceSize*index
		binary.BigEndian.PutUint16(payload[offset:offset+drwSequenceSize], sequence)
	}
	return Packet(msgDRWAck, payload)
}

func parseAppFrame(frame []byte) (byte, []byte, error) {
	if len(frame) < appFrameHeaderSize {
		return 0, nil, fmt.Errorf("%w: app frame too short", ErrInvalidAppFrame)
	}
	length := int(frame[0]) | int(frame[1])<<8 | int(frame[2])<<16
	if length < appFrameMinLength || appFrameHeaderSize+length > len(frame) {
		return 0, nil, fmt.Errorf("%w: invalid length", ErrInvalidAppFrame)
	}
	streamType := frame[appStreamOffset]
	if streamType < appStreamMin || streamType > appStreamMax {
		return 0, nil, fmt.Errorf("%w: invalid stream type %d", ErrInvalidAppFrame, streamType)
	}
	return streamType, frame[appFrameHeaderSize : appFrameHeaderSize+length], nil
}

func passwordKey(password string) []byte {
	key := make([]byte, authBlockSize)
	copy(key, password)
	return key
}

func aesEncrypt(key, input []byte) ([]byte, error) {
	if len(input) != aes.BlockSize {
		return nil, ErrAESBlock
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	output := make([]byte, aes.BlockSize)
	block.Encrypt(output, input)
	return output, nil
}

func aesDecrypt(key, input []byte) ([]byte, error) {
	if len(input) != aes.BlockSize {
		return nil, ErrAESBlock
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	output := make([]byte, aes.BlockSize)
	block.Decrypt(output, input)
	return output, nil
}
