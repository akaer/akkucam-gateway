package version

import (
	"regexp"
	"testing"
)

func TestVersionIsSemver(t *testing.T) {
	semver := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
		`(?:-(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*))*)?` +
		`(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

	if !semver.MatchString(Version) {
		t.Fatalf("version = %q, want a semantic version", Version)
	}
	if !regexp.MustCompile(`^0\.`).MatchString(Version) {
		t.Fatalf("version = %q, want major 0 during beta", Version)
	}
}

func TestDisplayIncludesShortCommit(t *testing.T) {
	original := commit
	t.Cleanup(func() { commit = original })

	commit = "0123456789abcdef"
	if got, want := Display(), Version+" (0123456)"; got != want {
		t.Fatalf("Display() = %q, want %q", got, want)
	}
}

func TestShortKeepsShortHashes(t *testing.T) {
	if got, want := short("abc1234"), "abc1234"; got != want {
		t.Fatalf("short() = %q, want %q", got, want)
	}
}
