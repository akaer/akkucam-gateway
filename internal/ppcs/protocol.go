package ppcs

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Exported protocol identifiers. They are part of the package API so tests and
// protocol users can build packets without duplicating magic values.
const (
	LanSearchPort = 32108
	MsgLanSearch  = 0x30
	MsgPunchPkt   = 0x41
	MsgP2PReady   = 0x42
	MsgAlive      = 0xe0
	MsgAliveAck   = 0xe1
	MsgClose      = 0xf0
)

// Header is the four-byte PPPP framing header.
type Header struct {
	Type   byte
	Length uint16
}

// Punch is the decoded LAN discovery punch payload.
type Punch struct {
	Prefix    string
	Serial    uint32
	CheckCode string
}

// DID formats the punch payload as the canonical device identifier.
func (p Punch) DID() string {
	return fmt.Sprintf("%s-%06d-%s", p.Prefix, p.Serial, p.CheckCode)
}

// Packet wraps payload in the PPPP framing used by the base station.
func Packet(messageType byte, payload []byte) []byte {
	packet := make([]byte, ppcsHeaderSize+len(payload))
	packet[0] = ppcsMagic
	packet[1] = messageType
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(payload)))
	copy(packet[ppcsHeaderSize:], payload)
	return packet
}

// ParseHeader validates and decodes the PPPP framing header.
func ParseHeader(packet []byte) (Header, error) {
	if len(packet) < ppcsHeaderSize {
		return Header{}, fmt.Errorf("%w: shorter than four bytes", ErrInvalidPacket)
	}
	if packet[0] != ppcsMagic {
		return Header{}, fmt.Errorf("%w: unexpected magic 0x%02x", ErrInvalidPacket, packet[0])
	}
	header := Header{Type: packet[1], Length: binary.BigEndian.Uint16(packet[2:4])}
	if int(header.Length) > len(packet)-ppcsHeaderSize {
		return header, fmt.Errorf("%w: payload length exceeds datagram", ErrInvalidPacket)
	}
	return header, nil
}

// ParsePunch decodes a complete PUNCH packet.
func ParsePunch(packet []byte) (Punch, error) {
	header, err := ParseHeader(packet)
	if err != nil {
		return Punch{}, err
	}
	if header.Type != MsgPunchPkt || header.Length < punchBodyLength || len(packet) < punchMinLength {
		return Punch{}, fmt.Errorf("%w: not a complete PUNCH packet", ErrInvalidPacket)
	}
	body := packet[ppcsHeaderSize:]
	return Punch{
		Prefix:    cString(body[punchPrefixOffset : punchPrefixOffset+punchPrefixLength]),
		Serial:    binary.BigEndian.Uint32(body[punchSerialOffset : punchSerialOffset+punchSerialLength]),
		CheckCode: cString(body[punchCheckOffset : punchCheckOffset+punchCheckLength]),
	}, nil
}

func cString(value []byte) string {
	if index := bytes.IndexByte(value, 0); index >= 0 {
		value = value[:index]
	}
	return string(value)
}
