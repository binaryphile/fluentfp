package web

import (
	"fmt"
	"net/http"
)

// ClientError is the JSON shape written to the client for all errors.
// Extensible via the Details field for field-level validation, rate-limit
// metadata, or any structured payload.
type ClientError struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// Error carries an HTTP status code, client-safe message, and optional
// structured details. Message is what the client sees. Err (if set) is for
// logs and errors.Is/As only — never sent to clients.
type Error struct {
	Status  int
	Message string
	Code    string
	Details any
	Headers http.Header // optional response headers (WWW-Authenticate, Retry-After)
	Err     error       // internal cause, never sent to client
}

// Error returns the client-safe message. If an internal cause is set,
// it is included for logging but should not be exposed to clients.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

// Unwrap returns the internal cause for errors.Is/As.
func (e *Error) Unwrap() error {
	return e.Err
}

// BadRequest returns a 400 error.
func BadRequest(msg string) error {
	return &Error{Status: http.StatusBadRequest, Message: msg, Code: "BAD_REQUEST"}
}

// NotFound returns a 404 error.
func NotFound(msg string) error {
	return &Error{Status: http.StatusNotFound, Message: msg, Code: "NOT_FOUND"}
}

// Conflict returns a 409 error.
func Conflict(msg string) error {
	return &Error{Status: http.StatusConflict, Message: msg, Code: "CONFLICT"}
}

// Forbidden returns a 403 error.
func Forbidden(msg string) error {
	return &Error{Status: http.StatusForbidden, Message: msg, Code: "FORBIDDEN"}
}

// StatusError returns an error with the given status, code, and message.
func StatusError(status int, code, msg string) error {
	return &Error{Status: status, Message: msg, Code: code}
}
