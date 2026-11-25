//go:build hardware

package ppcs

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"
	"time"
)

// hardwareConfig reads the hardware test configuration from the environment.
// It skips the test when the required variables are absent.
func hardwareConfig(t *testing.T) (ip, did, code string, channel int) {
	t.Helper()
	ip = os.Getenv("AKKUCAM_TEST_IP")
	did = os.Getenv("AKKUCAM_TEST_DID")
	code = os.Getenv("AKKUCAM_TEST_SECURITY_CODE")
	if ip == "" || did == "" || code == "" {
		t.Skip("AKKUCAM_TEST_IP, AKKUCAM_TEST_DID and AKKUCAM_TEST_SECURITY_CODE are required")
	}
	channel = 0
	if value := os.Getenv("AKKUCAM_TEST_CHANNEL"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 || parsed > 3 {
			t.Fatalf("invalid AKKUCAM_TEST_CHANNEL %q", value)
		}
		channel = parsed
	}
	return ip, did, code, channel
}

// TestHardwareLiveVideo opens a real session and verifies that the stream
// carries Annex-B H.264 video. It uses its own session, separate from the
// battery metadata test, so a failure points at exactly one behavior.
func TestHardwareLiveVideo(t *testing.T) {
	// Arrange
	ip, did, code, channel := hardwareConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	live, err := OpenLive(ctx, ip, did, code, channel, 1)
	if err != nil {
		t.Fatal(err)
	}
	video := make(chan []byte, 1)
	done := make(chan error, 1)
	go func() { done <- live.Run(ctx, channelWriter{target: video}) }()

	// Act
	select {
	case payload := <-video:
		// Assert
		if !bytes.Contains(payload, []byte{0, 0, 0, 1}) {
			t.Fatalf("not Annex-B H264: %x", payload[:min(len(payload), 32)])
		}
	case err := <-done:
		t.Fatalf("session ended before video: %v", err)
	case <-time.After(25 * time.Second):
		t.Fatal("no video within 25 seconds")
	}

	cancel()
	<-done
}

// TestHardwareBatteryMetadata opens a real session and verifies that battery
// capacity arrives within a plausible range. It uses its own session so the
// assertion is independent of video delivery.
func TestHardwareBatteryMetadata(t *testing.T) {
	// Arrange
	ip, did, code, channel := hardwareConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	live, err := OpenLive(ctx, ip, did, code, channel, 1)
	if err != nil {
		t.Fatal(err)
	}
	battery := make(chan int, 1)
	live.SetBatteryHandler(func(gotChannel, capacity int, _ bool) {
		if gotChannel == channel {
			battery <- capacity
		}
	})
	done := make(chan error, 1)
	go func() { done <- live.Run(ctx, channelWriter{target: make(chan []byte, 1)}) }()

	// Act
	select {
	case capacity := <-battery:
		// Assert
		if capacity < 0 || capacity > 100 {
			t.Fatalf("battery capacity=%d", capacity)
		}
	case err := <-done:
		t.Fatalf("session ended before battery metadata: %v", err)
	case <-time.After(25 * time.Second):
		t.Fatal("no battery metadata within 25 seconds")
	}

	cancel()
	<-done
}
