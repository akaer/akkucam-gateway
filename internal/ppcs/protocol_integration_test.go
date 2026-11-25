package ppcs

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

const (
	testDID      = "CMAG-000001-ABCDE"
	testPassword = "correct-code"
)

type fakeHub struct {
	conn       *net.UDPConn
	client     *net.UDPAddr
	sessionKey []byte
	serverErr  chan error
	mu         sync.Mutex
}

func newFakeHub(t *testing.T) *fakeHub {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	return &fakeHub{conn: conn, sessionKey: []byte("0123456789abcdef"), serverErr: make(chan error, 1)}
}

func (h *fakeHub) target() *net.UDPAddr { return h.conn.LocalAddr().(*net.UDPAddr) }
func (h *fakeHub) close()               { _ = h.conn.Close() }

func specPacket(messageType byte, payload []byte) []byte {
	packet := make([]byte, 4+len(payload))
	packet[0], packet[1] = 0xf1, messageType
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(payload)))
	copy(packet[4:], payload)
	return packet
}

func specPunch() []byte {
	payload := make([]byte, 20)
	copy(payload[0:8], "CMAG")
	binary.BigEndian.PutUint32(payload[8:12], 1)
	copy(payload[12:18], "ABCDE")
	return specPacket(0x41, payload)
}

func specDRW(channel byte, sequence uint16, app []byte) []byte {
	payload := make([]byte, 4+len(app))
	payload[0], payload[1] = 0xd1, channel
	binary.BigEndian.PutUint16(payload[2:4], sequence)
	copy(payload[4:], app)
	return specPacket(0xd0, payload)
}

func specAuthApp(authType uint16, data []byte) []byte {
	body := make([]byte, 16+len(data))
	binary.LittleEndian.PutUint16(body[0:2], authType)
	binary.LittleEndian.PutUint16(body[2:4], uint16(len(data)))
	copy(body[16:], data)
	app := make([]byte, 4+len(body))
	app[0], app[1], app[2], app[3] = byte(len(body)), byte(len(body)>>8), byte(len(body)>>16), 1
	copy(app[4:], body)
	return app
}

func specAES(key, input []byte, decrypt bool) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	output := make([]byte, 16)
	if decrypt {
		block.Decrypt(output, input)
	} else {
		block.Encrypt(output, input)
	}
	return output
}

func specPasswordKey(password string) []byte {
	key := make([]byte, 16)
	copy(key, password)
	return key
}

func (h *fakeHub) read(deadline time.Duration) ([]byte, *net.UDPAddr, error) {
	buffer := make([]byte, 65536)
	_ = h.conn.SetReadDeadline(time.Now().Add(deadline))
	n, remote, err := h.conn.ReadFromUDP(buffer)
	return append([]byte(nil), buffer[:n]...), remote, err
}

func (h *fakeHub) write(packet []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.conn.WriteToUDP(packet, h.client)
	return err
}

func (h *fakeHub) establishAndAuthenticate(expectedPassword string) error {
	search, client, err := h.read(2 * time.Second)
	if err != nil {
		return err
	}
	if !bytes.Equal(search, []byte{0xf1, 0x30, 0x00, 0x00}) {
		return errors.New("client did not send spec LAN_SEARCH")
	}
	h.client = client
	punch := specPunch()
	if _, err := h.conn.WriteToUDP(punch, client); err != nil {
		return err
	}
	echo, _, err := h.read(time.Second)
	if err != nil || !bytes.Equal(echo, punch) {
		return errors.New("client did not echo PUNCH byte-for-byte")
	}
	if err := h.write(specPacket(0x42, punch[4:])); err != nil {
		return err
	}

	challenge := []byte("challenge-16byte")
	if err := h.write(specDRW(0, 0, specAuthApp(1, challenge))); err != nil {
		return err
	}

	var authResponse []byte
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		packet, _, readErr := h.read(300 * time.Millisecond)
		if readErr != nil {
			continue
		}
		if len(packet) >= 44 && packet[0] == 0xf1 && packet[1] == 0xd0 && packet[5] == 0 && packet[11] == 1 && binary.LittleEndian.Uint16(packet[12:14]) == 2 {
			authResponse = packet[28:44]
			break
		}
	}
	if authResponse == nil {
		return errors.New("client did not send auth response")
	}
	expected := specAES(specPasswordKey(expectedPassword), challenge, false)
	if !bytes.Equal(authResponse, expected) {
		return h.write(specDRW(0, 1, specAuthApp(4, nil)))
	}
	encryptedKey := specAES(specPasswordKey(expectedPassword), h.sessionKey, false)
	return h.write(specDRW(0, 1, specAuthApp(3, encryptedKey)))
}

func TestLocalSessionRejectsWrongSecurityCode(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.establishAndAuthenticate(testPassword) }()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Act
	_, _, _, err := openAuthenticatedAt(ctx, "127.0.0.1", testDID, "wrong-code", hub.target())

	// Assert
	if err == nil || err.Error() != "security code rejected by basis station" {
		t.Fatalf("error=%v", err)
	}
	if serverErr := <-hub.serverErr; serverErr != nil {
		t.Fatal(serverErr)
	}
}

func TestLocalSessionAuthenticationIntegration(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.establishAndAuthenticate(testPassword) }()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Act
	session, key, messages, err := openAuthenticatedAt(ctx, "127.0.0.1", testDID, testPassword, hub.target())

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	defer session.close()
	if !bytes.Equal(key, hub.sessionKey) {
		t.Fatalf("session key = %x", key)
	}
	if got := messages[len(messages)-1]; got != "AUTH_SUCCESS" {
		t.Fatalf("last message=%q", got)
	}
	if err := <-hub.serverErr; err != nil {
		t.Fatal(err)
	}
}

func specControlApp(key []byte, controlType uint16, payload []byte) []byte {
	head := make([]byte, 16)
	binary.LittleEndian.PutUint16(head[0:2], controlType)
	binary.LittleEndian.PutUint16(head[2:4], uint16(len(payload)))
	encryptedHead := specAES(key, head, false)
	bodyLength := 16 + len(payload)
	app := make([]byte, 4+bodyLength)
	app[0], app[1], app[2], app[3] = byte(bodyLength), byte(bodyLength>>8), byte(bodyLength>>16), 4
	copy(app[4:20], encryptedHead)
	copy(app[20:], payload)
	return app
}

func specVideoApp(key, h264 []byte) []byte {
	head := make([]byte, 16)
	binary.LittleEndian.PutUint16(head[0:2], 3)
	binary.LittleEndian.PutUint32(head[8:12], uint32(len(h264)))
	encryptedHead := specAES(key, head, false)
	bodyLength := 16 + len(h264)
	app := make([]byte, 4+bodyLength)
	app[0], app[1], app[2], app[3] = byte(bodyLength), byte(bodyLength>>8), byte(bodyLength>>16), 2
	copy(app[4:20], encryptedHead)
	copy(app[20:], h264)
	return app
}

func decodeSpecControl(packet, key []byte) (uint16, []byte, bool) {
	if len(packet) < 28 || packet[0] != 0xf1 || packet[1] != 0xd0 {
		return 0, nil, false
	}
	app := packet[8:]
	if len(app) < 20 || app[3] != 4 {
		return 0, nil, false
	}
	length := int(app[0]) | int(app[1])<<8 | int(app[2])<<16
	if 4+length > len(app) {
		return 0, nil, false
	}
	head := specAES(key, app[4:20], true)
	size := int(binary.LittleEndian.Uint16(head[2:4]))
	if 20+size > len(app) {
		return 0, nil, false
	}
	payload := append([]byte(nil), app[20:20+size]...)
	for i := range payload {
		payload[i] ^= head[4]
	}
	return binary.LittleEndian.Uint16(head[0:2]), payload, true
}

func (h *fakeHub) serveLive() error {
	if err := h.establishAndAuthenticate(testPassword); err != nil {
		return err
	}
	wanted := map[uint16]bool{1: false, 3: false, 40448: false}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		packet, _, err := h.read(400 * time.Millisecond)
		if err != nil {
			continue
		}
		if len(packet) < 8 || packet[1] != 0xd0 {
			continue
		}
		apps := packet[8:]
		for len(apps) >= 4 {
			length := int(apps[0]) | int(apps[1])<<8 | int(apps[2])<<16
			if length < 16 || 4+length > len(apps) {
				break
			}
			ctrl, payload, ok := decodeSpecControl(specDRW(0, 0, apps[:4+length]), h.sessionKey)
			if ok {
				wanted[ctrl] = true
				if (ctrl == 1 || ctrl == 3) && (len(payload) != 8 || payload[0] != 1 || payload[4] != 4) {
					return errors.New("start command did not carry requested channel and quality")
				}
				if ctrl == 40448 && (len(payload) != 16 || payload[0] != 1) {
					return errors.New("battery request did not target selected channel")
				}
			}
			apps = apps[4+length:]
		}
		if wanted[1] && wanted[3] && wanted[40448] {
			break
		}
	}
	if !wanted[1] || !wanted[3] || !wanted[40448] {
		return errors.New("missing required live control commands")
	}

	battery := make([]byte, 16)
	battery[0], battery[1], battery[2] = 1, 73, 1
	if err := h.write(specDRW(0, 2, specControlApp(h.sessionKey, 40448, battery))); err != nil {
		return err
	}
	h264 := []byte{0, 0, 0, 1, 0x67, 0x4d, 0x40, 0x1f, 0, 0, 0, 1, 0x65, 1, 2, 3, 4}
	return h.write(specDRW(1, 0, specVideoApp(h.sessionKey, h264)))
}

func TestLiveProtocolVideoFromSpec(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.serveLive() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	live, err := openLiveAt(ctx, "127.0.0.1", testDID, testPassword, 1, 4, hub.target())
	if err != nil {
		t.Fatal(err)
	}
	video := make(chan []byte, 1)
	runDone := make(chan error, 1)
	go func() { runDone <- live.Run(ctx, channelWriter{target: video}) }()

	// Act
	var h264 []byte
	select {
	case h264 = <-video:
	case <-time.After(3 * time.Second):
		t.Fatal("no H264 output received")
	}

	// Assert
	if !bytes.Contains(h264, []byte{0, 0, 0, 1, 0x65}) {
		t.Fatalf("output is not Annex-B H264: %x", h264)
	}
	cancel()
	<-runDone
	if err := <-hub.serverErr; err != nil {
		t.Fatal(err)
	}
}

func TestLiveProtocolBatteryFromSpec(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.serveLive() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	live, err := openLiveAt(ctx, "127.0.0.1", testDID, testPassword, 1, 4, hub.target())
	if err != nil {
		t.Fatal(err)
	}
	type batteryReading struct {
		channel  int
		capacity int
		charging bool
	}
	battery := make(chan batteryReading, 1)
	live.SetBatteryHandler(func(channel, capacity int, charging bool) {
		battery <- batteryReading{channel: channel, capacity: capacity, charging: charging}
	})
	runDone := make(chan error, 1)
	go func() { runDone <- live.Run(ctx, io.Discard) }()

	// Act
	var reading batteryReading
	select {
	case reading = <-battery:
	case <-time.After(3 * time.Second):
		t.Fatal("battery metadata not delivered")
	}

	// Assert
	if reading.channel != 1 || reading.capacity != 73 || !reading.charging {
		t.Fatalf("battery=%+v", reading)
	}
	cancel()
	<-runDone
	if err := <-hub.serverErr; err != nil {
		t.Fatal(err)
	}
}

func (h *fakeHub) serveCameraList() error {
	if err := h.establishAndAuthenticate(testPassword); err != nil {
		return err
	}
	deadline := time.Now().Add(4 * time.Second)
	sequence := uint16(2)
	parametersSent := 0
	listSent, hubSent := false, false
	for time.Now().Before(deadline) {
		packet, _, err := h.read(300 * time.Millisecond)
		if err != nil {
			continue
		}
		controlType, payload, ok := decodeSpecControl(packet, h.sessionKey)
		if !ok {
			continue
		}
		var responseType uint16
		var response []byte
		switch controlType {
		case getCameraListControl:
			responseType, response, listSent = getCameraListControl, specCameraListPayload(), true
		case getDeviceInfoControl:
			responseType, response, hubSent = deviceInfoResponseControl, specHubInfoPayload(), true
		case getCameraParametersControl:
			channel := byte(0)
			if len(payload) > 0 {
				channel = payload[0]
			}
			response = make([]byte, 256)
			response[0], response[1], response[4], response[5] = channel, 1, 1, 73
			responseType = getCameraParametersControl
			parametersSent++
		default:
			continue
		}
		if err := h.write(specDRW(0, sequence, specControlApp(h.sessionKey, responseType, response))); err != nil {
			return err
		}
		sequence++
		if listSent && hubSent && parametersSent >= 2 {
			return nil
		}
	}
	return errors.New("client did not complete inventory requests")
}

func specHubInfoPayload() []byte {
	payload := make([]byte, 88)
	copy(payload[0:16], "SMH-T1R")
	copy(payload[16:32], "ExampleVendor")
	payload[32], payload[33], payload[34], payload[35] = 12, 0, 0, 1
	// total/free are decimal megabytes (MB). Example: 32 GB total, 8 GB free.
	binary.LittleEndian.PutUint32(payload[40:44], 32000)
	binary.LittleEndian.PutUint32(payload[44:48], 8000)
	copy(payload[56:72], "1.2.3")
	return payload
}

func TestQueryChannelsIntegrationFromSpec(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.serveCameraList() }()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Act
	channels, err := queryChannelsAt(ctx, "127.0.0.1", testDID, testPassword, hub.target())

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 2 || channels[1].Index != 1 || channels[1].Name != "Garten" {
		t.Fatalf("channels=%+v", channels)
	}
	if err := <-hub.serverErr; err != nil {
		t.Fatal(err)
	}
}

func TestQueryChannelsMotionParametersIntegrationFromSpec(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.serveCameraList() }()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Act
	channels, err := queryChannelsAt(ctx, "127.0.0.1", testDID, testPassword, hub.target())

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) == 0 {
		t.Fatal("no channels returned")
	}
	if channels[0].MotionEnabled == nil || !*channels[0].MotionEnabled {
		t.Fatalf("motion enabled=%v", channels[0].MotionEnabled)
	}
	if channels[0].PIRSensitivityDay == nil || *channels[0].PIRSensitivityDay != 73 {
		t.Fatalf("pir sensitivity=%v", channels[0].PIRSensitivityDay)
	}
	if err := <-hub.serverErr; err != nil {
		t.Fatal(err)
	}
}

func (h *fakeHub) serveMotionToggle() error {
	if err := h.establishAndAuthenticate(testPassword); err != nil {
		return err
	}
	sequence := uint16(2)
	mode := byte(0)
	setSeen := false
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		packet, _, err := h.read(300 * time.Millisecond)
		if err != nil {
			continue
		}
		controlType, payload, ok := decodeSpecControl(packet, h.sessionKey)
		if !ok {
			continue
		}
		switch controlType {
		case getCameraParametersControl:
			response := make([]byte, 256)
			response[0], response[1], response[4], response[5] = 1, 1, mode, 65
			if err := h.write(specDRW(0, sequence, specControlApp(h.sessionKey, getCameraParametersControl, response))); err != nil {
				return err
			}
			sequence++
			if setSeen {
				return nil
			}
		case setCameraParametersControl:
			if len(payload) != 256 || payload[0] != 1 || payload[4] != 1 {
				return errors.New("motion write did not preserve full parameter block and enable PIR mode")
			}
			mode, setSeen = payload[4], true
		}
	}
	return errors.New("motion toggle sequence incomplete")
}

func TestSetMotionReadModifyWriteVerifyIntegration(t *testing.T) {
	// Arrange
	hub := newFakeHub(t)
	defer hub.close()
	go func() { hub.serverErr <- hub.serveMotionToggle() }()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Act
	channel, err := setMotionAt(ctx, "127.0.0.1", testDID, testPassword, 1, true, hub.target())

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if channel.MotionEnabled == nil || !*channel.MotionEnabled || channel.PIRMode == nil || *channel.PIRMode != 1 {
		t.Fatalf("channel=%+v", channel)
	}
	if err := <-hub.serverErr; err != nil {
		t.Fatal(err)
	}
}

type channelWriter struct{ target chan<- []byte }

func (w channelWriter) Write(data []byte) (int, error) {
	w.target <- append([]byte(nil), data...)
	return len(data), nil
}
