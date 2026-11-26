package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"akkucam-gateway/internal/ppcs"
)

// unsafeID removes characters that are unsafe in file paths and URLs.
var unsafeID = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// Manager owns the active HLS streams keyed by sanitized DID and channel.
type Manager struct {
	root     string
	mu       sync.Mutex
	streams  map[string]*running
	openLive func(context.Context, string, string, string, int, int) (liveSource, error)
	command  func(context.Context, string) *exec.Cmd
}

// liveSource is the protocol boundary used by the manager. Tests replace it
// with a fake source.
type liveSource interface {
	Run(context.Context, io.Writer) error
	Close()
	SetBatteryHandler(func(channel, capacity int, charging bool))
}

// running holds the mutable state of one active stream.
type running struct {
	cancel       context.CancelFunc
	done         chan struct{}
	started      time.Time
	channel      int
	totalBytes   atomic.Int64
	totalFrames  atomic.Int64
	bitrateBytes atomic.Int64
	framesPerSec atomic.Int64
	metadataMu   sync.RWMutex
	battery      *Battery
}

// Battery is the last verified battery reading for a channel.
type Battery struct {
	Capacity int  `json:"capacity"`
	Charging bool `json:"charging"`
}

// Status is the API representation of a stream.
type Status struct {
	ID                    string   `json:"id"`
	Running               bool     `json:"running"`
	Channel               int      `json:"channel"`
	BitrateBytesPerSecond int64    `json:"bitrateBytesPerSecond"`
	FramesPerSecond       int64    `json:"framesPerSecond,omitempty"`
	Battery               *Battery `json:"battery,omitempty"`
	UptimeSeconds         int64    `json:"uptimeSeconds,omitempty"`
}

// Result is returned when a stream starts.
type Result struct {
	ID          string `json:"id"`
	PlaylistURL string `json:"playlistUrl"`
}

// New creates a stream manager rooted at the given HLS directory.
func New(root string) *Manager {
	return &Manager{
		root:    root,
		streams: make(map[string]*running),
		openLive: func(ctx context.Context, ip, did, password string, cameraIndex, quality int) (liveSource, error) {
			return ppcs.OpenLive(ctx, ip, did, password, cameraIndex, quality)
		},
		command: ffmpegCommand,
	}
}

// Root returns the HLS output directory.
func (m *Manager) Root() string { return m.root }

// ActiveCount returns the number of running streams.
func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.streams)
}

// Start opens a live session and pipes Annex-B H.264 into FFmpeg. Starting a
// stream for the same DID and channel replaces the existing one.
func (m *Manager) Start(parent context.Context, ip, did, password string, cameraIndex, quality int) (Result, error) {
	id := unsafeID.ReplaceAllString(did+"-ch"+strconv.Itoa(cameraIndex+1), "_")
	// unsafeID keeps the ID free of separators; the check makes that provable for
	// static analysis so the HLS directory cannot leave m.root.
	if !filepath.IsLocal(id) {
		return Result{}, ErrDIDRequired
	}
	m.Stop(id)

	ctx, cancel := context.WithCancel(context.Background())
	connectCtx, connectCancel := context.WithTimeout(parent, connectTimeout)
	live, err := m.openLive(connectCtx, ip, did, password, cameraIndex, quality)
	connectCancel()
	if err != nil {
		cancel()
		return Result{}, err
	}

	directory := filepath.Join(m.root, id)
	if err := os.RemoveAll(directory); err != nil {
		live.Close()
		cancel()
		return Result{}, err
	}
	if err := os.MkdirAll(directory, hlsDirectoryMode); err != nil {
		live.Close()
		cancel()
		return Result{}, err
	}

	playlist := filepath.Join(directory, "index.m3u8")
	command := m.command(ctx, playlist)
	stdin, err := command.StdinPipe()
	if err != nil {
		live.Close()
		cancel()
		return Result{}, err
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		live.Close()
		cancel()
		return Result{}, fmt.Errorf("%w: %w", ErrFFmpegStart, err)
	}

	run := &running{cancel: cancel, done: make(chan struct{}), started: time.Now(), channel: cameraIndex}
	live.SetBatteryHandler(func(channel, capacity int, charging bool) {
		if channel != cameraIndex {
			return
		}
		run.metadataMu.Lock()
		run.battery = &Battery{Capacity: capacity, Charging: charging}
		run.metadataMu.Unlock()
	})
	m.mu.Lock()
	m.streams[id] = run
	m.mu.Unlock()

	go func() {
		defer close(run.done)
		defer live.Close()
		counter := &countingWriter{writer: stdin, bytes: &run.totalBytes, frames: &run.totalFrames}
		meterDone := make(chan struct{})
		go meterBitrate(ctx, run, meterDone)
		err := live.Run(ctx, counter)
		close(meterDone)
		_ = stdin.Close()
		waitErr := command.Wait()
		if err != nil && ctx.Err() == nil && !errors.Is(err, context.Canceled) {
			log.Printf("live stream %s ended: %v", id, err)
		}
		if waitErr != nil && ctx.Err() == nil {
			log.Printf("ffmpeg stream %s ended: %v", id, waitErr)
		}
		m.mu.Lock()
		if m.streams[id] == run {
			delete(m.streams, id)
		}
		m.mu.Unlock()
	}()

	return Result{ID: id, PlaylistURL: "/hls/" + id + "/index.m3u8"}, nil
}

// ffmpegCommand builds the FFmpeg remuxer. Video is copied without re-encoding
// (ADR-0006); timestamps are regenerated at the negotiated frame rate.
func ffmpegCommand(ctx context.Context, playlist string) *exec.Cmd {
	filter := fmt.Sprintf("setts=pts=N/(%d*TB_OUT):dts=N/(%d*TB_OUT)", ffmpegFrameRate, ffmpegFrameRate)
	//nolint:gosec // G204: fixed argument list, playlist path from validated stream ID (ADR-0033)
	return exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", ffmpegLogLevel,
		"-fflags", "+genpts",
		"-r", strconv.Itoa(ffmpegFrameRate), "-f", ffmpegInputFormat, "-i", ffmpegInputPipe,
		"-an", "-c:v", "copy", "-bsf:v", filter,
		"-avoid_negative_ts", "make_zero",
		"-f", "hls", "-hls_time", strconv.Itoa(ffmpegHLSSeconds), "-hls_list_size", strconv.Itoa(ffmpegHLSListSize),
		"-hls_flags", ffmpegHLSFlags,
		playlist,
	)
}

// Status returns a defensive copy of the stream state. Unknown IDs report
// Running=false instead of an error.
func (m *Manager) Status(id string) Status {
	id = unsafeID.ReplaceAllString(id, "_")
	m.mu.Lock()
	run := m.streams[id]
	m.mu.Unlock()
	if run == nil {
		return Status{ID: id, Running: false}
	}
	run.metadataMu.RLock()
	battery := run.battery
	if battery != nil {
		batteryCopy := *battery
		battery = &batteryCopy
	}
	run.metadataMu.RUnlock()
	return Status{ID: id, Running: true, Channel: run.channel, BitrateBytesPerSecond: run.bitrateBytes.Load(), FramesPerSecond: run.framesPerSec.Load(), Battery: battery, UptimeSeconds: int64(time.Since(run.started).Seconds())}
}

// countingWriter counts bytes and complete writes that reach FFmpeg.
type countingWriter struct {
	writer io.Writer
	bytes  *atomic.Int64
	frames *atomic.Int64
}

// Write forwards data and updates the byte and frame counters.
func (w *countingWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if w.bytes != nil {
		w.bytes.Add(int64(n))
	}
	if err == nil && n == len(data) && w.frames != nil {
		w.frames.Add(1)
	}
	return n, err
}

// meterBitrate computes one-second deltas for the live metrics.
func meterBitrate(ctx context.Context, run *running, done <-chan struct{}) {
	ticker := time.NewTicker(bitrateInterval)
	defer ticker.Stop()
	previous := int64(0)
	previousFrames := int64(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			current := run.totalBytes.Load()
			run.bitrateBytes.Store(current - previous)
			previous = current
			frames := run.totalFrames.Load()
			run.framesPerSec.Store(frames - previousFrames)
			previousFrames = frames
		}
	}
}

// Stop cancels and waits for a stream. Stopping an unknown ID is a no-op.
func (m *Manager) Stop(id string) {
	id = unsafeID.ReplaceAllString(id, "_")
	m.mu.Lock()
	run := m.streams[id]
	if run != nil {
		delete(m.streams, id)
	}
	m.mu.Unlock()
	if run != nil {
		run.cancel()
		select {
		case <-run.done:
		case <-time.After(stopTimeout):
		}
	}
}
