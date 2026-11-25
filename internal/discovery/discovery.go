// Package discovery finds AkkuCam base stations on the local network using the
// WAI and PPPP broadcast protocols. It never contacts a cloud service.
package discovery

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// waiRequest is the fixed WAI0 discovery request observed on UDP/7200.
var waiRequest = []byte{
	0x57, 0x41, 0x49, 0x30, 0x00, 0x20, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xa8, 0x9e, 0xb6, 0xcf, 0x00, 0x00, 0x00, 0x00,
}

// Device is one discovered base station.
type Device struct {
	IP       string `json:"ip"`
	DID      string `json:"did,omitempty"`
	Model    string `json:"model,omitempty"`
	PPCSPort int    `json:"ppcsPort,omitempty"`
	MAC      string `json:"mac,omitempty"`
	Local    bool   `json:"local"`
}

// WAI performs a WAI0 broadcast discovery and collects replies until the
// context deadline.
func WAI(ctx context.Context) ([]Device, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: waiPort})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := enableBroadcast(conn); err != nil {
		return nil, err
	}
	if _, err := conn.WriteToUDP(waiRequest, &net.UDPAddr{IP: net.IPv4bcast, Port: waiPort}); err != nil {
		return nil, err
	}

	return collect(ctx, conn, parseWAI)
}

// PPCS performs a PPPP LAN search and collects PUNCH replies until the context
// deadline. The probe is repeated every 100 ms.
func PPCS(ctx context.Context) ([]Device, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := enableBroadcast(conn); err != nil {
		return nil, err
	}

	probe := []byte{ppcsMagic, ppcsLanSearch, 0x00, 0x00}
	var sendErr error
	var sendMu sync.Mutex
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(discoveryProbeInterval)
		defer ticker.Stop()
		for {
			if _, err := conn.WriteToUDP(probe, &net.UDPAddr{IP: net.IPv4bcast, Port: ppcsPort}); err != nil {
				sendMu.Lock()
				sendErr = err
				sendMu.Unlock()
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	devices, readErr := collect(ctx, conn, parsePPCS)
	<-done
	sendMu.Lock()
	err = sendErr
	sendMu.Unlock()
	if err != nil {
		return devices, err
	}
	return devices, readErr
}

type parser func([]byte, *net.UDPAddr) (Device, bool)

func collect(ctx context.Context, conn *net.UDPConn, parse parser) ([]Device, error) {
	devices := make(map[string]Device)
	buffer := make([]byte, discoveryBufferSize)
	for {
		deadline := time.Now().Add(discoveryReadWindow)
		if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
			deadline = ctxDeadline
		}
		_ = conn.SetReadDeadline(deadline)
		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				select {
				case <-ctx.Done():
					return values(devices), nil
				default:
					continue
				}
			}
			return values(devices), err
		}
		if device, ok := parse(buffer[:n], remote); ok {
			devices[device.IP] = device
		}
	}
}

func parseWAI(data []byte, remote *net.UDPAddr) (Device, bool) {
	if len(data) < waiMinLength || !bytes.Equal(data[:waiMagicLength], []byte(waiMagic)) {
		return Device{}, false
	}
	return Device{
		IP:    remote.IP.String(),
		DID:   cString(data[waiDIDOffset : waiDIDOffset+waiDIDLength]),
		Model: cString(data[waiModelOffset : waiModelOffset+waiModelLength]),
		Local: true,
	}, true
}

func parsePPCS(data []byte, remote *net.UDPAddr) (Device, bool) {
	if len(data) < ppcsMinLength || data[0] != ppcsMagic || data[1] != ppcsPunch {
		return Device{}, false
	}
	payloadLength := int(binary.BigEndian.Uint16(data[2:ppcsHeaderSize]))
	if payloadLength > len(data)-ppcsHeaderSize {
		return Device{}, false
	}
	prefix := cString(data[ppcsPrefixOffset : ppcsPrefixOffset+ppcsPrefixLength])
	serial := binary.BigEndian.Uint32(data[ppcsSerialOffset : ppcsSerialOffset+ppcsSerialLength])
	checkCode := cString(data[ppcsCheckOffset : ppcsCheckOffset+ppcsCheckLength])
	return Device{
		IP:       remote.IP.String(),
		DID:      fmt.Sprintf("%s-%06d-%s", prefix, serial, checkCode),
		PPCSPort: remote.Port,
		Local:    true,
	}, true
}

func cString(data []byte) string {
	if index := bytes.IndexByte(data, 0); index >= 0 {
		data = data[:index]
	}
	return strings.TrimSpace(string(data))
}

func values(devices map[string]Device) []Device {
	result := make([]Device, 0, len(devices))
	for _, device := range devices {
		result = append(result, device)
	}
	return result
}
