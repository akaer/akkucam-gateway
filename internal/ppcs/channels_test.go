package ppcs

import (
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"
)

func specCameraListPayload() []byte {
	payload := make([]byte, 4+2*60)
	binary.LittleEndian.PutUint32(payload[0:4], 2)
	payload[4] = 0
	copy(payload[5:37], "Einfahrt")
	copy(payload[37:53], "SAPP-TAS")
	binary.LittleEndian.PutUint32(payload[53:57], 9216)
	binary.LittleEndian.PutUint32(payload[57:61], 328192)
	payload[61] = 1
	offset := 64
	payload[offset] = 1
	copy(payload[offset+1:offset+33], "Garten")
	copy(payload[offset+33:offset+49], "RVDP-TAS")
	return payload
}

func TestParseCameraListFromSpecification(t *testing.T) {
	// Arrange
	payload := specCameraListPayload()

	// Act
	channels, err := parseCameraList(payload)

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 2 {
		t.Fatalf("channels=%d", len(channels))
	}
	if channels[0].Index != 0 || channels[0].Name != "Einfahrt" || channels[0].Model != "SAPP-TAS" || channels[0].Updateable == nil || !*channels[0].Updateable {
		t.Fatalf("channel 0=%+v", channels[0])
	}
	if channels[0].FWVersion != "0.0.36.0" || channels[0].MCUVersion != "0.5.2.0" {
		t.Fatalf("formatted versions=%s/%s", channels[0].FWVersion, channels[0].MCUVersion)
	}
	if channels[1].Index != 1 || channels[1].Name != "Garten" || channels[1].Model != "RVDP-TAS" {
		t.Fatalf("channel 1=%+v", channels[1])
	}
}

func TestUnavailableCameraMetadataIsOmitted(t *testing.T) {
	// Arrange
	channel := CameraChannel{Index: 1}

	// Act
	encoded, err := json.Marshal(channel)

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	for _, unavailable := range []string{"name", "model", "fwVersion", "mcuVersion", "mac", "motionEnabled", "updateable"} {
		if strings.Contains(string(encoded), `"`+unavailable+`"`) {
			t.Fatalf("unavailable field %q present in %s", unavailable, encoded)
		}
	}
}

func TestParseHubInfo(t *testing.T) {
	// Arrange
	payload := specHubInfoPayload()

	// Act
	info, err := parseHubInfo(payload)

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if info.Model != "SMH-T1R" || info.Vendor != "ExampleVendor" || info.Firmware != "1.0.0.12" || info.MCUVersion != "1.2.3" {
		t.Fatalf("hub info=%+v", info)
	}
	if info.SDTotalBytes == nil || *info.SDTotalBytes != 32_000_000_000 || info.SDFreeBytes == nil || *info.SDFreeBytes != 8_000_000_000 {
		t.Fatalf("sd bytes=%+v", info)
	}
}

func TestParseCameraParametersRejectsTruncation(t *testing.T) {
	// Arrange
	payload := make([]byte, 255)

	// Act
	_, err := parseCameraParameters(payload)

	// Assert
	if err == nil {
		t.Fatal("accepted 255-byte parameters")
	}
}

func TestParseCameraParametersAcceptsFullBlock(t *testing.T) {
	// Arrange
	payload := make([]byte, 300)

	// Act
	parsed, err := parseCameraParameters(payload)

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 256 {
		t.Fatalf("len=%d, want 256", len(parsed))
	}
}

func TestParseCameraListRejectsInvalidCounts(t *testing.T) {
	// Arrange
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, 17)

	// Act
	_, err := parseCameraList(payload)

	// Assert
	if err == nil {
		t.Fatal("accepted implausible count")
	}
}

func TestParseCameraListRejectsTruncation(t *testing.T) {
	// Arrange
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, 1)

	// Act
	_, err := parseCameraList(payload)

	// Assert
	if err == nil {
		t.Fatal("accepted truncated camera entry")
	}
}
