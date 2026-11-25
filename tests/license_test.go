package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	// License entries in the dashboard carry a stable data-component attribute
	// so the drift check does not depend on the visible product names.
	licenseItemPattern    = regexp.MustCompile(`(?s)<li data-component="([^"]+)">(.*?)</li>`)
	licenseVersionPattern = regexp.MustCompile(`class="license-version">([^<]+)<`)
	// hls.js exposes its version through the Hls.version getter.
	hlsVersionPattern = regexp.MustCompile(`key:"version",get:function\(\)\{return"([0-9]+\.[0-9]+\.[0-9]+)"`)
	dockerBasePattern = regexp.MustCompile(`(?m)^FROM\s+([a-z0-9./-]+):([^\s]+)`)
)

type sbomComponent struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// TestLicenseVersionsMatchArtifacts fails when the license dialog drifts from
// the shipped bundle, the SBOM snapshot or the base images. The SBOM is the
// source of truth for runtime packages; the Dockerfile pins the base images.
//
// This lives in tests/ because it reads repository artifacts that are not part
// of the image build context.
func TestLicenseVersionsMatchArtifacts(t *testing.T) {
	// Arrange
	index, err := os.ReadFile(filepath.Join("..", "web", "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	versions := make(map[string]string)
	for _, item := range licenseItemPattern.FindAllStringSubmatch(string(index), -1) {
		version := ""
		if match := licenseVersionPattern.FindStringSubmatch(item[2]); match != nil {
			version = strings.TrimSpace(match[1])
		}
		versions[item[1]] = version
	}
	if len(versions) == 0 {
		t.Fatal("index.html has no license entries with data-component")
	}

	hls, err := os.ReadFile(filepath.Join("..", "web", "hls.min.js"))
	if err != nil {
		t.Fatalf("read hls.min.js: %v", err)
	}
	hlsMatch := hlsVersionPattern.FindStringSubmatch(string(hls))
	if hlsMatch == nil {
		t.Fatal("cannot read the hls.js version from web/hls.min.js")
	}

	var document struct {
		Components []sbomComponent `json:"components"`
	}
	readJSON(t, cycloneDXPath, &document)
	sbom := make(map[string]string)
	for _, component := range document.Components {
		sbom[component.Name] = component.Version
	}

	dockerfile, err := os.ReadFile(filepath.Join("..", "Dockerfile"))
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}
	images := make(map[string]string)
	for _, match := range dockerBasePattern.FindAllStringSubmatch(string(dockerfile), -1) {
		images[match[1]] = match[2]
	}

	// Act and Assert
	if versions["hls"] != hlsMatch[1] {
		t.Errorf("hls.js: dashboard shows %q, web/hls.min.js is %q", versions["hls"], hlsMatch[1])
	}
	assertShownVersion(t, "FFmpeg", versions["ffmpeg"], sbom["ffmpeg"], "SBOM")
	assertShownVersion(t, "Alpine", versions["alpine"], sbom["alpine"], "SBOM")
	if images["alpine"] != versions["alpine"] {
		t.Errorf("Alpine: dashboard shows %q, Dockerfile pins alpine:%q", versions["alpine"], images["alpine"])
	}
	assertShownVersion(t, "Go", versions["go"], strings.TrimPrefix(sbom["stdlib"], "go"), "SBOM")
	if goImage := strings.TrimSuffix(images["golang"], "-alpine"); goImage != versions["go"] {
		t.Errorf("Go: dashboard shows %q, Dockerfile pins golang:%s", versions["go"], images["golang"])
	}

	// A version without a machine-readable source is exactly the drift this
	// test exists to prevent.
	known := map[string]bool{"hls": true, "ffmpeg": true, "alpine": true, "go": true}
	for name, version := range versions {
		if version != "" && !known[name] {
			t.Errorf("license entry %q shows %q without a source this test can verify", name, version)
		}
	}
}

func assertShownVersion(t *testing.T, component, shown, artifact, source string) {
	t.Helper()
	if shown == "" {
		t.Errorf("%s: dashboard shows no version", component)
		return
	}
	if artifact == "" {
		t.Errorf("%s: %s has no version", component, source)
		return
	}
	if !strings.HasPrefix(artifact, shown) {
		t.Errorf("%s: dashboard shows %q, %s is %q", component, shown, source, artifact)
	}
}
