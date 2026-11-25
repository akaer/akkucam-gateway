package apierr

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()
	var body envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	return body
}

func TestWriteIncludesStableCode(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	apiErr := New(CodeValidation, http.StatusBadRequest, "bad input")

	// Act
	Write(rec, apiErr)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", rec.Code)
	}
	body := decode(t, rec)
	if body.Code != CodeValidation {
		t.Fatalf("code=%q, want %q", body.Code, CodeValidation)
	}
	if body.Error != "bad input" {
		t.Fatalf("message=%q", body.Error)
	}
}

func TestWriteUnknownErrorIsInternal(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()

	// Act
	Write(rec, errors.New("boom"))

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500", rec.Code)
	}
	body := decode(t, rec)
	if body.Code != CodeInternal {
		t.Fatalf("code=%q, want %q", body.Code, CodeInternal)
	}
	if body.Error != "internal server error" {
		t.Fatalf("message=%q", body.Error)
	}
}

func TestWriteUnwrapsWrappedError(t *testing.T) {
	// Arrange
	rec := httptest.NewRecorder()
	wrapped := errors.Join(errors.New("context"), New(CodeConflict, http.StatusConflict, "busy"))

	// Act
	Write(rec, wrapped)

	// Assert
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d, want 409", rec.Code)
	}
	body := decode(t, rec)
	if body.Code != CodeConflict {
		t.Fatalf("code=%q, want %q", body.Code, CodeConflict)
	}
}
