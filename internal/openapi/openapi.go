// Package openapi embeds the machine-readable OpenAPI contract and serves it
// from the gateway.
package openapi

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var spec []byte

// Spec returns the raw OpenAPI document.
func Spec() []byte {
	return spec
}

// Handler serves the OpenAPI document as JSON.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	}
}
