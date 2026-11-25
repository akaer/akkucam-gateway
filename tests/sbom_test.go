// Package tests contains cross-package contract checks for repository-level
// artifacts. Unit tests stay next to the code they exercise.
package tests

import (
	"encoding/json"
	"os"
	"testing"
)

const (
	cycloneDXPath = "../sbom/akkucam-gateway.cyclonedx.json"
	spdxPath      = "../sbom/akkucam-gateway.spdx.json"
)

func readJSON(t *testing.T, path string, target any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (run ./scripts/sbom.sh)", path, err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func TestCycloneDXSBOMIsValid(t *testing.T) {
	// Arrange
	var document struct {
		BOMFormat   string `json:"bomFormat"`
		SpecVersion string `json:"specVersion"`
		Components  []struct {
			Name string `json:"name"`
		} `json:"components"`
	}

	// Act
	readJSON(t, cycloneDXPath, &document)

	// Assert
	if document.BOMFormat != "CycloneDX" {
		t.Fatalf("bomFormat=%q, want CycloneDX", document.BOMFormat)
	}
	if document.SpecVersion == "" {
		t.Fatal("specVersion is empty")
	}
	if len(document.Components) == 0 {
		t.Fatal("components are empty")
	}
}

func TestSPDXSBOMIsValid(t *testing.T) {
	// Arrange
	var document struct {
		SPDXVersion string `json:"spdxVersion"`
		Packages    []struct {
			Name string `json:"name"`
		} `json:"packages"`
	}

	// Act
	readJSON(t, spdxPath, &document)

	// Assert
	if document.SPDXVersion != "SPDX-2.3" {
		t.Fatalf("spdxVersion=%q, want SPDX-2.3", document.SPDXVersion)
	}
	if len(document.Packages) == 0 {
		t.Fatal("packages are empty")
	}
}

func TestSBOMsListRuntimeComponents(t *testing.T) {
	// Arrange
	var cycloneDX struct {
		Components []struct {
			Name string `json:"name"`
		} `json:"components"`
	}
	var spdx struct {
		Packages []struct {
			Name string `json:"name"`
		} `json:"packages"`
	}
	readJSON(t, cycloneDXPath, &cycloneDX)
	readJSON(t, spdxPath, &spdx)

	// Act
	cycloneDXHasFFmpeg := false
	for _, component := range cycloneDX.Components {
		if component.Name == "ffmpeg" {
			cycloneDXHasFFmpeg = true
			break
		}
	}
	spdxHasFFmpeg := false
	for _, pkg := range spdx.Packages {
		if pkg.Name == "ffmpeg" {
			spdxHasFFmpeg = true
			break
		}
	}

	// Assert
	if !cycloneDXHasFFmpeg {
		t.Fatal("CycloneDX SBOM does not list ffmpeg")
	}
	if !spdxHasFFmpeg {
		t.Fatal("SPDX SBOM does not list ffmpeg")
	}
}
