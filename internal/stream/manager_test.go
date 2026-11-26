package stream

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

type shortWriter struct{ maximum int }

func (w shortWriter) Write(data []byte) (int, error) {
	if len(data) > w.maximum {
		return w.maximum, errors.New("short write")
	}
	return len(data), nil
}

func TestCountingWriterCountsSuccessfullyWrittenBytesOnly(t *testing.T) {
	// Arrange
	var total atomic.Int64
	writer := &countingWriter{writer: shortWriter{maximum: 3}, bytes: &total}

	// Act
	n, err := writer.Write([]byte{1, 2, 3, 4, 5})

	// Assert
	if err == nil {
		t.Fatal("expected short write error")
	}
	if n != 3 {
		t.Fatalf("written=%d, want 3", n)
	}
	if total.Load() != 3 {
		t.Fatalf("total=%d, want 3", total.Load())
	}
}

func TestStatusUnknownStream(t *testing.T) {
	// Arrange
	manager := New(t.TempDir())

	// Act
	status := manager.Status("missing")

	// Assert
	if status.Running {
		t.Fatal("missing stream reported as running")
	}
}

func TestStatusReturnsDefensiveBatteryCopy(t *testing.T) {
	// Arrange
	manager := New(t.TempDir())
	run := &running{started: time.Now().Add(-2 * time.Second), channel: 1, done: make(chan struct{}), battery: &Battery{Capacity: 81, Charging: true}}
	run.bitrateBytes.Store(98765)
	manager.streams["device"] = run

	// Act
	status := manager.Status("device")

	// Assert
	if !status.Running || status.Channel != 1 || status.BitrateBytesPerSecond != 98765 || status.Battery == nil || status.Battery.Capacity != 81 {
		t.Fatalf("status=%+v", status)
	}
	status.Battery.Capacity = 1
	if manager.streams["device"].battery.Capacity != 81 {
		t.Fatal("Status exposed mutable internal battery state")
	}
}

func TestMeterBitrateReportsOneSecondDelta(t *testing.T) {
	// Arrange
	run := &running{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go meterBitrate(ctx, run, done)
	run.totalBytes.Store(4096)
	run.totalFrames.Store(12)

	// Act
	time.Sleep(1100 * time.Millisecond)
	close(done)
	cancel()

	// Assert
	if got := run.bitrateBytes.Load(); got != 4096 {
		t.Fatalf("bitrate=%d, want 4096", got)
	}
	if got := run.framesPerSec.Load(); got != 12 {
		t.Fatalf("fps=%d, want 12", got)
	}
}

func TestStopIsIdempotentForUnknownStream(t *testing.T) {
	// Arrange
	manager := New(t.TempDir())

	// Act
	manager.Stop("unknown")
	manager.Stop("unknown")

	// Assert
	if manager.ActiveCount() != 0 {
		t.Fatalf("active streams=%d, want 0", manager.ActiveCount())
	}
}

func TestCountingWriterPassesCompletePayload(t *testing.T) {
	// Arrange
	var target bytes.Buffer
	var total atomic.Int64
	writer := &countingWriter{writer: &target, bytes: &total}
	payload := []byte{0, 0, 0, 1, 0x65}

	// Act
	n, err := writer.Write(payload)

	// Assert
	if err != nil {
		t.Fatal(err)
	}
	if n != len(payload) {
		t.Fatalf("written=%d, want %d", n, len(payload))
	}
	if !bytes.Equal(target.Bytes(), payload) {
		t.Fatal("payload changed")
	}
}

type fakeLive struct {
	closed         chan struct{}
	started        chan struct{}
	channel        int
	batteryHandler func(int, int, bool)
}

func (f *fakeLive) SetBatteryHandler(handler func(int, int, bool)) { f.batteryHandler = handler }
func (f *fakeLive) Close() {
	select {
	case <-f.closed:
	default:
		close(f.closed)
	}
}
func (f *fakeLive) Run(ctx context.Context, output io.Writer) error {
	if f.batteryHandler != nil {
		f.batteryHandler(f.channel, 64, false)
	}
	_, _ = output.Write([]byte{0, 0, 0, 1, 0x65})
	close(f.started)
	<-ctx.Done()
	return ctx.Err()
}

func testManagerWithFakeLive(t *testing.T) (*Manager, chan *fakeLive) {
	t.Helper()
	manager := New(t.TempDir())
	opened := make(chan *fakeLive, 4)
	manager.openLive = func(_ context.Context, _, _, _ string, channel, _ int) (liveSource, error) {
		live := &fakeLive{closed: make(chan struct{}), started: make(chan struct{}), channel: channel}
		opened <- live
		return live, nil
	}
	manager.command = func(ctx context.Context, _ string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "cat >/dev/null")
	}
	return manager, opened
}

func TestStartReplacesExistingStreamForSameTarget(t *testing.T) {
	// Arrange
	manager, opened := testManagerWithFakeLive(t)
	firstResult, err := manager.Start(context.Background(), "192.0.2.10", "CMAG-000001-ABCDE", "code", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	first := <-opened
	if !manager.Status(firstResult.ID).Running {
		t.Fatal("first stream is not running")
	}

	// Act
	secondResult, err := manager.Start(context.Background(), "192.0.2.10", "CMAG-000001-ABCDE", "code", 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	second := <-opened
	<-second.started

	// Assert
	select {
	case <-first.closed:
	case <-time.After(time.Second):
		t.Fatal("replaced stream was not closed")
	}
	status := manager.Status(secondResult.ID)
	if !status.Running || status.Channel != 1 || status.Battery == nil || status.Battery.Capacity != 64 {
		t.Fatalf("replacement status=%+v", status)
	}
}

func TestStartKeepsHLSDirectoryInsideRoot(t *testing.T) {
	// Arrange
	manager, opened := testManagerWithFakeLive(t)

	// Act
	result, err := manager.Start(context.Background(), "192.0.2.10", "../../CMAG-000001-ABCDE/../..", "code", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	live := <-opened
	<-live.started
	defer manager.Stop(result.ID)

	// Assert
	if !filepath.IsLocal(result.ID) {
		t.Fatalf("stream ID %q is not local", result.ID)
	}
	if _, err := os.Stat(filepath.Join(manager.Root(), result.ID)); err != nil {
		t.Fatalf("HLS directory missing below root: %v", err)
	}
}

func TestStopClosesStream(t *testing.T) {
	// Arrange
	manager, opened := testManagerWithFakeLive(t)
	result, err := manager.Start(context.Background(), "192.0.2.10", "CMAG-000001-ABCDE", "code", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	live := <-opened
	<-live.started

	// Act
	manager.Stop(result.ID)

	// Assert
	select {
	case <-live.closed:
	case <-time.After(time.Second):
		t.Fatal("stopped stream was not closed")
	}
	if manager.Status(result.ID).Running {
		t.Fatal("stream still running after stop")
	}
}

func TestDifferentChannelsRunConcurrently(t *testing.T) {
	// Arrange
	manager, opened := testManagerWithFakeLive(t)
	firstResult, err := manager.Start(context.Background(), "192.0.2.10", "CMAG-000001-ABCDE", "code", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	first := <-opened
	<-first.started
	secondResult, err := manager.Start(context.Background(), "192.0.2.10", "CMAG-000001-ABCDE", "code", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	second := <-opened
	<-second.started

	// Act
	firstRunning := manager.Status(firstResult.ID).Running
	secondRunning := manager.Status(secondResult.ID).Running

	// Assert
	if firstResult.ID == secondResult.ID {
		t.Fatalf("channel stream IDs collide: %s", firstResult.ID)
	}
	if !firstRunning || !secondRunning {
		t.Fatalf("first=%v second=%v, want both running", firstRunning, secondRunning)
	}
	select {
	case <-first.closed:
		t.Fatal("starting channel 2 closed channel 1")
	default:
	}
}
