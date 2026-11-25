package ppcs

import "encoding/binary"

// MediaFragment is the payload transported in a PPCS DRW packet.
type MediaFragment struct {
	Sequence uint16
	Data     []byte
}

// ParseMediaFragment removes the observed PPCS f1d0/d101 UDP framing.
func ParseMediaFragment(packet []byte) (MediaFragment, bool) {
	if len(packet) < drwMinLength || packet[0] != ppcsMagic || packet[1] != msgDRW || packet[4] != drwMarker || packet[5] != mediaFragmentChannel {
		return MediaFragment{}, false
	}
	declared := int(binary.BigEndian.Uint16(packet[2:4]))
	if declared != len(packet)-ppcsHeaderSize {
		return MediaFragment{}, false
	}
	return MediaFragment{
		Sequence: binary.BigEndian.Uint16(packet[drwSequenceOffset : drwSequenceOffset+drwSequenceSize]),
		Data:     packet[drwPayloadOffset:],
	}, true
}
