//go:build !linux

package discovery

// MACForIP is unavailable on non-Linux platforms and always returns an empty
// string.
func MACForIP(string) string { return "" }
