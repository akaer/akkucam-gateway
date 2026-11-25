package ppcs

import "errors"

// Sentinel errors make protocol failures distinguishable without matching on
// human-readable strings. Callers should use errors.Is.
var (
	ErrInvalidPacket       = errors.New("invalid PPPP packet")
	ErrInvalidDRW          = errors.New("invalid DRW packet")
	ErrInvalidAppFrame     = errors.New("invalid app frame")
	ErrInvalidSecurityCode = errors.New("security code must contain 1 to 16 characters")
	ErrAuthRejected        = errors.New("security code rejected by basis station")
	ErrAuthChallenge       = errors.New("invalid authentication challenge")
	ErrAuthKey             = errors.New("invalid encrypted session key")
	ErrAuthTimeout         = errors.New("authentication timeout")
	ErrLanTimeout          = errors.New("LAN PUNCH timeout")
	ErrP2PReadyTimeout     = errors.New("P2P_RDY timeout")
	ErrAESBlock            = errors.New("AES input must be exactly 16 bytes")
	ErrPeerClosed          = errors.New("basis station closed the session")
	ErrMotionNotApplied    = errors.New("basis station did not apply motion setting")
	ErrChannelRange        = errors.New("channel must be between 0 and 15")
	ErrCameraListTruncated = errors.New("camera list truncated")
	ErrCameraListCount     = errors.New("implausible camera count")
	ErrCameraParameters    = errors.New("camera parameters truncated")
	ErrHubInfoTruncated    = errors.New("hub device info truncated")
	ErrMediaSize           = errors.New("invalid media size")
)
