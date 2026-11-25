package ppcs

import "testing"

func TestParseMediaFragment(t *testing.T) {
	// Arrange
	packet := []byte{0xf1, 0xd0, 0x00, 0x09, 0xd1, 0x01, 0x12, 0x34, 0x00, 0x00, 0x00, 0x01, 0x67}

	// Act
	fragment, ok := ParseMediaFragment(packet)

	// Assert
	if !ok {
		t.Fatal("expected valid fragment")
	}
	if fragment.Sequence != 0x1234 {
		t.Fatalf("sequence = %x", fragment.Sequence)
	}
	if len(fragment.Data) != 5 || fragment.Data[4] != 0x67 {
		t.Fatalf("unexpected media data: %x", fragment.Data)
	}
}
