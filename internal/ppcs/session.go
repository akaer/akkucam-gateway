package ppcs

import (
	"context"
	"time"
)

type ProbeResult struct {
	IP        string   `json:"ip"`
	Port      int      `json:"port"`
	DID       string   `json:"did"`
	Ready     bool     `json:"ready"`
	RoundTrip string   `json:"roundTrip"`
	Messages  []string `json:"messages"`
}

// Probe opens and immediately closes a direct LAN PPPP session. It performs
// no authentication and sends no camera commands.
func Probe(ctx context.Context, wantedIP, wantedDID string) (ProbeResult, error) {
	started := time.Now()
	session, messages, err := openLocal(ctx, wantedIP, wantedDID)
	if err != nil {
		return ProbeResult{}, err
	}
	peer, did := session.peer, session.did
	session.close()
	messages = append(messages, "CLOSE")

	return ProbeResult{
		IP:        peer.IP.String(),
		Port:      peer.Port,
		DID:       did,
		Ready:     true,
		RoundTrip: time.Since(started).Round(time.Millisecond).String(),
		Messages:  messages,
	}, nil
}
