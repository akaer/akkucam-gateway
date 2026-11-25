package openapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSpecIsValidJSON(t *testing.T) {
	// Arrange
	raw := Spec()

	// Act
	var document map[string]any
	err := json.Unmarshal(raw, &document)

	// Assert
	if err != nil {
		t.Fatalf("OpenAPI document is not valid JSON: %v", err)
	}
	if document["openapi"] == nil {
		t.Fatal("OpenAPI document has no openapi version")
	}
	if document["paths"] == nil {
		t.Fatal("OpenAPI document has no paths")
	}
}

// TestSpecUsesRelativeServerURL keeps "Try it out" working from any host. An
// absolute URL would make the browser call its own localhost.
func TestSpecUsesRelativeServerURL(t *testing.T) {
	// Arrange
	var document struct {
		Servers []struct {
			URL string `json:"url"`
		} `json:"servers"`
	}

	// Act
	err := json.Unmarshal(Spec(), &document)

	// Assert
	if err != nil {
		t.Fatalf("invalid OpenAPI JSON: %v", err)
	}
	if len(document.Servers) == 0 {
		t.Fatal("OpenAPI document has no servers")
	}
	for _, server := range document.Servers {
		if strings.Contains(server.URL, "://") || strings.HasPrefix(server.URL, "//") {
			t.Fatalf("server URL must be relative, got %q", server.URL)
		}
	}
}
