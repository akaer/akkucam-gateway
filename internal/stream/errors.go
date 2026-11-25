package stream

import "errors"

// Sentinel errors for the stream manager.
var (
	ErrDIDRequired = errors.New("device DID is required")
	ErrFFmpegStart = errors.New("start ffmpeg")
)
