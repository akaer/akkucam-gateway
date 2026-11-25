package discovery

import (
	"net"
	"testing"
)

func TestParsePPCSFullDID(t *testing.T) {
	// Arrange
	packet := []byte{
		0xf1, 0x41, 0x00, 0x14,
		'C', 'M', 'A', 'G', 0, 0, 0, 0,
		0x00, 0x00, 0x47, 0x72,
		'L', 'Y', 'H', 'F', 'C', 0, 0, 0,
	}
	remote := &net.UDPAddr{IP: net.ParseIP("192.168.178.75"), Port: 17523}

	// Act
	device, ok := parsePPCS(packet, remote)

	// Assert
	if !ok {
		t.Fatal("packet was not recognized")
	}
	if got, want := device.DID, "CMAG-018290-LYHFC"; got != want {
		t.Fatalf("DID = %q, want %q", got, want)
	}
}

func TestParseWAIContract(t *testing.T) {
	// Arrange
	packet := make([]byte, 124)
	copy(packet[0:4], "WAI1")
	copy(packet[16:33], "CMAG-000001-ABCDE")
	copy(packet[48:64], "SMH-T1R")
	remote := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 7200}

	// Act
	device, ok := parseWAI(packet, remote)

	// Assert
	if !ok {
		t.Fatal("WAI1 response was not recognized")
	}
	if device.DID != "CMAG-000001-ABCDE" || device.Model != "SMH-T1R" || device.IP != "192.0.2.10" {
		t.Fatalf("device=%+v", device)
	}
}

func TestParseWAIRejectsUnrelatedPacket(t *testing.T) {
	// Arrange
	remote := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 7200}
	packet := []byte("WAI0")

	// Act
	_, ok := parseWAI(packet, remote)

	// Assert
	if ok {
		t.Fatal("accepted WAI0 request as response")
	}
}

func TestParsePPCSRejectsUnrelatedPacket(t *testing.T) {
	// Arrange
	remote := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 7200}
	packet := []byte{0xf1, 0x42, 0, 0}

	// Act
	_, ok := parsePPCS(packet, remote)

	// Assert
	if ok {
		t.Fatal("accepted P2P_RDY as discovery result")
	}
}
