package gateway

import (
	"net/http"
	"testing"
)

func TestDocsRoutesAreRegistered(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	for _, path := range []string{"/api/openapi.json", "/api/docs/"} {
		t.Run(path, func(t *testing.T) {
			// Act
			response := request(t, handler, http.MethodGet, path, "")

			// Assert
			if response.Code != http.StatusOK {
				t.Fatalf("%s status=%d", path, response.Code)
			}
		})
	}
}

func TestDocsRedirectsToTrailingSlash(t *testing.T) {
	// Arrange
	_, handler, _ := testServer(t)

	// Act
	response := request(t, handler, http.MethodGet, "/api/docs", "")

	// Assert
	if response.Code != http.StatusMovedPermanently {
		t.Fatalf("status=%d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "/api/docs/" {
		t.Fatalf("location=%q", location)
	}
}
