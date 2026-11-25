package ppcs

import "testing"

func FuzzParseHeaderNeverPanics(f *testing.F) {
	f.Add([]byte{0xf1, 0x30, 0x00, 0x00})
	f.Add([]byte{0xf1, 0xd0, 0xff, 0xff})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseHeader(data)
	})
}

func FuzzParsePunchNeverPanics(f *testing.F) {
	f.Add(specPunch())
	f.Add([]byte{0xf1, 0x41, 0x00, 0x14})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParsePunch(data)
	})
}

func FuzzParseMediaFragmentNeverPanics(f *testing.F) {
	f.Add([]byte{0xf1, 0xd0, 0x00, 0x09, 0xd1, 0x01, 0, 0, 0, 0, 0, 1, 0x67})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseMediaFragment(data)
	})
}

func FuzzAppAssemblerNeverPanics(f *testing.F) {
	f.Add([]byte{0x10, 0, 0, 1, 0, 0, 0, 0})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		// Keep fuzz memory bounded while still exercising framing and recovery.
		if len(data) > 64*1024 {
			data = data[:64*1024]
		}
		assembler := &appAssembler{}
		_ = assembler.feed(data)
	})
}
