package ppcs

import "time"

// PPPP framing constants. The protocol is byte-oriented and big-endian unless
// a comment says otherwise.
const (
	ppcsMagic      = 0xf1
	ppcsHeaderSize = 4
	ppcsMaxPacket  = 65536

	msgDRW    = 0xd0
	msgDRWAck = 0xd1

	punchMinLength    = 22
	punchBodyLength   = 18
	punchPrefixOffset = 0
	punchPrefixLength = 8
	punchSerialOffset = 8
	punchSerialLength = 4
	punchCheckOffset  = 12
	punchCheckLength  = 6
)

// DRW / application framing constants.
const (
	drwMarker         = 0xd1
	drwChannelOffset  = 5
	drwSequenceOffset = 6
	drwPayloadOffset  = 8
	drwMinLength      = 8

	drwPayloadMarkerOff   = 0
	drwPayloadChannelOff  = 1
	drwPayloadSequenceOff = 2
	drwPayloadDataOff     = 4
	drwAckCountOff        = 2
	drwAckSequenceOff     = 4
	drwSequenceSize       = 2
	drwAckHeaderSize      = 4
	mediaFragmentChannel  = 1

	appFrameHeaderSize = 4
	appFrameMinLength  = 16
	appFrameMaxLength  = 16 * 1024 * 1024
	appStreamMin       = 1
	appStreamMax       = 4
	appStreamOffset    = 3
)

// Authentication and control-session constants.
const (
	authStream       = 1
	authChallenge    = 1
	authResponse     = 2
	authSuccess      = 3
	authFailed       = 4
	authBlockSize    = 16
	authPasswordMax  = 16
	authBufferSize   = 4096
	authAckChannel   = 0
	authControlSize  = 16
	authChallengeLen = 16
	authTypeOffset   = 0
	authSizeOffset   = 2
	authTypeSize     = 2
	authSizeSize     = 2

	searchBufferSize = 2048
	punchReadPoll    = 250 * time.Millisecond
	sessionReadPoll  = 500 * time.Millisecond
)

// Live-media and control-stream constants.
const (
	controlStream        = 4
	controlChannel       = 0
	controlHeaderSize    = 16
	controlPayloadOffset = 20
	controlTypeOffset    = 0
	controlTypeSize      = 2
	controlSizeOffset    = 2
	controlSizeSize      = 2
	controlXorOffset     = 4
	controlInitialXor    = 2
	controlBattery       = 40448
	controlIndexSize     = 8
	controlIndexCamera   = 0
	controlIndexQuality  = 4
	controlVideoType     = 1
	controlAudioType     = 3
	batteryChargingValue = 1

	mediaHeaderSize   = 16
	mediaSizeOffset   = 8
	mediaSizeLength   = 4
	liveReadBuffer    = 65536
	liveKeepAlive     = 4 * time.Second
	liveBatteryPeriod = 30 * time.Second
	liveReadPoll      = 500 * time.Millisecond
	resyncStep        = 1
)

// h264Markers are Annex-B start codes followed by common NAL unit types.
var h264Markers = [][]byte{
	{0, 0, 0, 1, 0x67},
	{0, 0, 0, 1, 0x65},
	{0, 0, 0, 1, 0x41},
	{0, 0, 0, 1, 0x61},
}

// IO-Control identifiers and fixed field sizes.
const (
	getDeviceInfoControl       = 5
	deviceInfoResponseControl  = 6
	getCameraListControl       = 40962
	setCameraParametersControl = 40963
	getCameraParametersControl = 40964

	cameraListCountSize       = 4
	cameraListEntrySize       = 60
	cameraListMax             = 16
	cameraListIndexOffset     = 0
	cameraListNameOffset      = 1
	cameraListNameSize        = 32
	cameraListModelOffset     = 33
	cameraListModelSize       = 16
	cameraListFWOffset        = 49
	cameraListMCUOffset       = 53
	cameraListUpdateOffset    = 57
	cameraListVersionSize     = 4
	cameraListUpdateableValue = 1
	deviceInfoRequestSize     = 4
	uint32Size                = 4
	cameraParametersSize      = 256
	hubInfoMinSize            = 56
	hubInfoModelOffset        = 0
	hubInfoModelSize          = 16
	hubInfoVendorOffset       = 16
	hubInfoVendorSize         = 16
	hubInfoVersionOffset      = 32
	hubInfoVersionSize        = 4
	hubInfoAPIVersionOff      = 36
	hubInfoTotalOffset        = 40
	hubInfoFreeOffset         = 44
	hubInfoStatusOffset       = 48
	hubInfoMCUOffset          = 56
	hubInfoMCUSize            = 16
	sdStatusNormal            = 0
	sdStatusBusy              = 32
	sdStatusNoCard            = 48
	sdStatusSlow              = 64
	sdStatusError             = 240
	megabyte                  = 1_000_000
	motionChannelOffset       = 0
	motionModeOffset          = 4
	motionSensitivityOff      = 5
	motionEnabledValue        = 1
	motionDisabledValue       = 0
	MaxMotionChannel          = 15
)
