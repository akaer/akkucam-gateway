package stream

import "time"

// Stream manager tuning constants.
const (
	connectTimeout    = 12 * time.Second
	stopTimeout       = 2 * time.Second
	bitrateInterval   = time.Second
	hlsDirectoryMode  = 0o750
	ffmpegFrameRate   = 25
	ffmpegHLSSeconds  = 1
	ffmpegHLSListSize = 5
	ffmpegLogLevel    = "warning"
	ffmpegInputFormat = "h264"
	ffmpegInputPipe   = "pipe:0"
	ffmpegHLSFlags    = "delete_segments+append_list+omit_endlist+independent_segments"
)
