package gateway

import "time"

// Runtime defaults and request timeouts. Override the listen address and web
// root with AKKUCAM_LISTEN and AKKUCAM_WEB_ROOT.
const (
	DefaultListen     = ":8080"
	DefaultWebRoot    = "/app/web"
	DefaultHLSRoot    = "/tmp/akkucam-hls"
	ReadHeaderTimeout = 5 * time.Second
	DiscoveryTimeout  = 4 * time.Second
	ProbeTimeout      = 7 * time.Second
	AuthTimeout       = 7 * time.Second
	AuthAttempts      = 3
	AuthRetryDelay    = 250 * time.Millisecond
	SessionTimeout    = 12 * time.Second
	MaxCameraIndex    = 3
	MaxQuality        = 4
	DiscoverySources  = 2
)
