package discovery

import "time"

// LAN discovery constants.
const (
	waiPort  = 7200
	ppcsPort = 32108

	ppcsMagic     = 0xf1
	ppcsLanSearch = 0x30
	ppcsPunch     = 0x41

	waiMagic       = "WAI1"
	waiMagicLength = 4
	waiMinLength   = 64
	waiDIDOffset   = 16
	waiDIDLength   = 17
	waiModelOffset = 48
	waiModelLength = 16

	ppcsMinLength    = 22
	ppcsHeaderSize   = 4
	ppcsPrefixOffset = 4
	ppcsPrefixLength = 8
	ppcsSerialOffset = 12
	ppcsSerialLength = 4
	ppcsCheckOffset  = 16
	ppcsCheckLength  = 6

	discoveryProbeInterval = 100 * time.Millisecond
	discoveryReadWindow    = 150 * time.Millisecond
	discoveryBufferSize    = 2048
)
