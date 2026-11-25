package ppcs

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestParseHeaderRejectsMalformedPackets(t *testing.T) {
	// Arrange
	cases := []struct {
		name   string
		packet []byte
	}{
		{"nil", nil},
		{"truncated", []byte{0xf1, 0x30, 0x00}},
		{"wrong magic", []byte{0x00, 0x30, 0x00, 0x00}},
		{"length mismatch", []byte{0xf1, 0x30, 0x00, 0x02, 0x01}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			_, err := ParseHeader(tc.packet)

			// Assert
			if err == nil {
				t.Errorf("packet %x was accepted", tc.packet)
			}
		})
	}
}

func TestParsePunchRejectsWrongType(t *testing.T) {
	// Arrange
	packet := append([]byte(nil), specPunch()...)
	packet[1] = 0x42

	// Act
	_, err := ParsePunch(packet)

	// Assert
	if err == nil {
		t.Errorf("packet %x was accepted", packet)
	}
}

func TestParsePunchRejectsTruncation(t *testing.T) {
	// Arrange
	packet := specPunch()[:21]

	// Act
	_, err := ParsePunch(packet)

	// Assert
	if err == nil {
		t.Errorf("packet %x was accepted", packet)
	}
}

func TestParseMediaFragmentRejectsLengthMismatch(t *testing.T) {
	// Arrange
	packet := []byte{0xf1, 0xd0, 0x00, 0x08, 0xd1, 0x01, 0x00, 0x01, 1, 2, 3}

	// Act
	_, ok := ParseMediaFragment(packet)

	// Assert
	if ok {
		t.Fatal("accepted a fragment whose declared length does not match the datagram")
	}
}

func TestParseMediaFragmentRejectsDRWACK(t *testing.T) {
	// Arrange
	packet := []byte{0xf1, 0xd1, 0x00, 0x04, 0xd1, 0x01, 0x00, 0x01}

	// Act
	_, ok := ParseMediaFragment(packet)

	// Assert
	if ok {
		t.Fatal("accepted a DRW_ACK as media")
	}
}

func TestChannelReceiverOrdersPackets(t *testing.T) {
	// Arrange
	receiver := &channelReceiver{assembler: &appAssembler{}}

	// Act
	first := receiver.add(0, []byte("a"))
	second := receiver.add(2, []byte("c"))
	third := receiver.add(1, []byte("b"))

	// Assert
	if len(first) != 1 || !bytes.Equal(first[0], []byte("a")) {
		t.Fatalf("first packet=%q", first)
	}
	if len(second) != 0 {
		t.Fatalf("out-of-order packet emitted early: %q", second)
	}
	if len(third) != 2 || string(third[0]) != "b" || string(third[1]) != "c" {
		t.Fatalf("ordered output=%q", third)
	}
}

func TestChannelReceiverHandlesSequenceWrap(t *testing.T) {
	// Arrange
	receiver := &channelReceiver{assembler: &appAssembler{}}

	// Act
	first := receiver.add(0xffff, []byte("a"))
	second := receiver.add(0, []byte("b"))

	// Assert
	if len(first) != 1 || !bytes.Equal(first[0], []byte("a")) {
		t.Fatalf("first packet=%q", first)
	}
	if len(second) != 1 || !bytes.Equal(second[0], []byte("b")) {
		t.Fatalf("wrapped packet=%q", second)
	}
}

func TestAppAssemblerReassemblesSplitFrames(t *testing.T) {
	// Arrange
	frame := []byte{0x10, 0x00, 0x00, 0x01}
	frame = append(frame, make([]byte, 16)...)
	assembler := &appAssembler{}

	// Act
	first := assembler.feed(frame[:7])
	second := assembler.feed(frame[7:])

	// Assert
	if len(first) != 0 {
		t.Fatalf("emitted incomplete frame: %x", first)
	}
	if len(second) != 1 || !bytes.Equal(second[0], frame) {
		t.Fatalf("reassembled=%x", second)
	}
}

func TestAppAssemblerResynchronizesAfterCorruption(t *testing.T) {
	// Arrange
	frame := []byte{0x10, 0x00, 0x00, 0x01}
	frame = append(frame, make([]byte, 16)...)
	corruptThenValid := append([]byte{0xff, 0xff, 0xff}, frame...)
	assembler := &appAssembler{}

	// Act
	got := assembler.feed(corruptThenValid)

	// Assert
	if len(got) != 1 || !bytes.Equal(got[0], frame) {
		t.Fatalf("assembler did not resynchronize: %x", got)
	}
}

func TestH264DetectionAcceptsSPS(t *testing.T) {
	// Arrange
	payload := []byte{0, 0, 0, 1, 0x67, 1, 2}

	// Act
	detected := looksLikeH264(payload)

	// Assert
	if !detected {
		t.Fatal("SPS was not recognized")
	}
}

func TestH264DetectionRejectsSEIOnly(t *testing.T) {
	// Arrange
	payload := []byte{0, 0, 0, 1, 0x06, 1, 2}

	// Act
	detected := looksLikeH264(payload)

	// Assert
	if detected {
		t.Fatal("SEI-only payload was incorrectly classified as a video access unit")
	}
}

func TestSecurityCodeLengthIsValidatedBeforeNetworkAccess(t *testing.T) {
	// Arrange
	target := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cases := []struct {
		name string
		code string
	}{
		{"empty", ""},
		{"too long", "12345678901234567"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			_, _, _, err := openAuthenticatedAt(ctx, "127.0.0.1", testDID, tc.code, target)

			// Assert
			if err == nil || err.Error() != "security code must contain 1 to 16 characters" {
				t.Fatalf("code length %d error=%v", len(tc.code), err)
			}
		})
	}
}
