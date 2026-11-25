//go:build linux

package discovery

import (
	"net"
	"os"
	"strings"
)

// MACForIP looks up the MAC address for an IPv4 address in the Linux ARP
// table. It returns an empty string when the lookup fails or on non-Linux
// platforms.
func MACForIP(ip string) string {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] != ip {
			continue
		}
		mac, err := net.ParseMAC(fields[3])
		if err == nil {
			return strings.ToUpper(mac.String())
		}
	}
	return ""
}
