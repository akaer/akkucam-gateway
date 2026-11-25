// Package apierr defines stable, machine-readable API error codes and the
// JSON error envelope used by the HTTP API.
package apierr

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Code is a stable identifier for an API error. Clients may rely on it.
type Code string

const (
	CodeInvalidJSON      Code = "invalid_json"
	CodeValidation       Code = "validation_failed"
	CodeMethodNotAllowed Code = "method_not_allowed"
	CodeNotFound         Code = "not_found"
	CodeConflict         Code = "conflict"
	CodeUnauthorized     Code = "unauthorized"
	CodeUpstreamTimeout  Code = "upstream_timeout"
	CodeUpstreamError    Code = "upstream_error"
	CodeInternal         Code = "internal_error"
)

// AllCodes returns every stable API error code. Tests use it to verify that
// clients can translate all codes.
func AllCodes() []Code {
	return []Code{
		CodeInvalidJSON,
		CodeValidation,
		CodeMethodNotAllowed,
		CodeNotFound,
		CodeConflict,
		CodeUnauthorized,
		CodeUpstreamTimeout,
		CodeUpstreamError,
		CodeInternal,
	}
}

// Error carries a stable code, an HTTP status and a human-readable message.
// Err is the optional underlying cause and is never serialized.
type Error struct {
	Code    Code
	Status  int
	Message string
	Err     error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

// Unwrap exposes the underlying cause to errors.Is and errors.As.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// New creates an API error without an underlying cause.
func New(code Code, status int, message string) *Error {
	return &Error{Code: code, Status: status, Message: message}
}

// Wrap creates an API error that keeps the underlying cause.
func Wrap(code Code, status int, message string, cause error) *Error {
	return &Error{Code: code, Status: status, Message: message, Err: cause}
}

type envelope struct {
	Error string `json:"error"`
	Code  Code   `json:"code"`
}

// Write sends err as a JSON error response. Unknown errors are mapped to a
// generic 500 response so internal details never leak to clients.
func Write(w http.ResponseWriter, err error) {
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr == nil {
		apiErr = &Error{
			Code:    CodeInternal,
			Status:  http.StatusInternalServerError,
			Message: "internal server error",
			Err:     err,
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(envelope{Error: apiErr.Error(), Code: apiErr.Code})
}
