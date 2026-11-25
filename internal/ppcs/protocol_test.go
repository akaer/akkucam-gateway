package ppcs

import "testing"

func TestParsePunch(t *testing.T) {
	// Arrange
	packet := []byte{
		0xf1, 0x41, 0x00, 0x14,
		'C', 'M', 'A', 'G', 0, 0, 0, 0,
		0x00, 0x00, 0x47, 0x72,
		'L', 'Y', 'H', 'F', 'C', 0, 0, 0,
	}

	// Act
	punch, err := ParsePunch(packet)

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if got, want := punch.DID(), "CMAG-018290-LYHFC"; got != want {
		t.Fatalf("DID = %q, want %q", got, want)
	}
}
