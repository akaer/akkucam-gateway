package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"akkucam-gateway/internal/openapi"
)

func TestOpenAPIDocumentsEveryRoute(t *testing.T) {
	// Arrange
	spec := decodeSpec(t)
	routes := New(t.TempDir(), &fakeStreams{}).routeTable()

	// Act
	missing := make([]string, 0)
	for _, route := range routes {
		operations, ok := spec[route.path]
		if !ok {
			missing = append(missing, route.method+" "+route.path)
			continue
		}
		if _, ok := operations[strings.ToLower(route.method)]; !ok {
			missing = append(missing, route.method+" "+route.path)
		}
	}

	// Assert
	if len(missing) > 0 {
		t.Fatalf("routes missing from OpenAPI: %v", missing)
	}
}

func TestOpenAPIHasNoUnknownRoutes(t *testing.T) {
	// Arrange
	spec := decodeSpec(t)
	routes := New(t.TempDir(), &fakeStreams{}).routeTable()
	registered := make(map[string]bool, len(routes))
	for _, route := range routes {
		registered[route.method+" "+route.path] = true
	}

	// Act
	unknown := make([]string, 0)
	for path, operations := range spec {
		for method := range operations {
			if !registered[strings.ToUpper(method)+" "+path] {
				unknown = append(unknown, strings.ToUpper(method)+" "+path)
			}
		}
	}

	// Assert
	if len(unknown) > 0 {
		t.Fatalf("OpenAPI documents unknown routes: %v", unknown)
	}
}

func decodeSpec(t *testing.T) map[string]map[string]json.RawMessage {
	t.Helper()
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(openapi.Spec(), &document); err != nil {
		t.Fatalf("invalid OpenAPI document: %v", err)
	}
	return document.Paths
}
