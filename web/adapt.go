package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/binaryphile/fluentfp/rslt"
)

// Handler is a function from request to result. Not pure in the FP sense
// (*http.Request has mutable state), but more functional than ResponseWriter
// mutation — handlers are testable expressions that return values.
type Handler = func(*http.Request) rslt.Result[Response]

// AdaptOption configures Adapt behavior. Scope is strictly response/error
// rendering. Cross-cutting concerns (logging, metrics, tracing, auth)
// belong in standard func(http.Handler) http.Handler middleware.
type AdaptOption func(*adaptConfig)

type adaptConfig struct {
	errorMapper func(error) (*Error, bool)
}

// WithErrorMapper maps domain errors to *Error at the adapter boundary.
// Only called for errors that are NOT already *Error (transport errors from
// DecodeJSON, validation errors that are already *Error bypass this).
// Return (*Error, true) to handle, or (nil, false) to fall through to 500.
// At most one mapper per Adapt call; last wins if called multiple times.
// Panics if fn is nil.
func WithErrorMapper(fn func(error) (*Error, bool)) AdaptOption {
	if fn == nil {
		panic("web.WithErrorMapper: fn must not be nil")
	}

	return func(cfg *adaptConfig) {
		cfg.errorMapper = fn
	}
}

// Adapt converts a Handler into an http.HandlerFunc.
//
// Rendering: marshals Body to buffer via json.Marshal before writing any
// response bytes. If marshaling fails, returns 500 (no partial writes).
//
// Error rendering flow (uses errors.As, not type assertion):
//  1. errors.As(err, &webErr) → render *Error directly (with Headers, Details)
//  2. Else if ErrorMapper set → call mapper
//  3. If mapper returns (*Error, true) → render that
//  4. Else → 500 with generic "internal error"
//
// Panics if h is nil.
func Adapt(h Handler, opts ...AdaptOption) http.HandlerFunc {
	if h == nil {
		panic("web.Adapt: handler must not be nil")
	}

	var cfg adaptConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		result := h(r)

		resp, err := result.Unpack()
		if err != nil {
			writeError(w, err, &cfg)

			return
		}

		writeResponse(w, resp)
	}
}

// writeResponse writes a successful Response.
func writeResponse(w http.ResponseWriter, resp Response) {
	if resp.Body == nil {
		copyHeaders(w, resp.Headers)
		w.WriteHeader(resp.Status)

		return
	}

	data, err := json.Marshal(resp.Body)
	if err != nil {
		writeInternalError(w)

		return
	}

	copyHeaders(w, resp.Headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Status)
	w.Write(data)
}

// writeError renders an error response using the error rendering flow.
func writeError(w http.ResponseWriter, err error, cfg *adaptConfig) {
	// Step 1: Check for *Error via errors.As (finds wrapped errors too).
	var webErr *Error
	if errors.As(err, &webErr) {
		writeWebError(w, webErr)

		return
	}

	// Step 2: Try error mapper for domain errors.
	if cfg.errorMapper != nil {
		if mapped, ok := cfg.errorMapper(err); ok && mapped != nil {
			writeWebError(w, mapped)

			return
		}
	}

	// Step 3: Unmapped error → 500.
	writeInternalError(w)
}

// writeWebError renders a *Error as a ClientError JSON response.
func writeWebError(w http.ResponseWriter, e *Error) {
	ce := ClientError{
		Error:   e.Message,
		Code:    e.Code,
		Details: e.Details,
	}

	data, err := json.Marshal(ce)
	if err != nil {
		writeInternalError(w)

		return
	}

	status := e.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}

	copyHeaders(w, e.Headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// writeInternalError writes a generic 500 response.
func writeInternalError(w http.ResponseWriter) {
	ce := ClientError{
		Error: "internal error",
		Code:  "INTERNAL_ERROR",
	}

	data, _ := json.Marshal(ce) // ClientError with string fields cannot fail

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(data)
}

// copyHeaders copies src headers into the ResponseWriter, avoiding mutation
// of the original.
func copyHeaders(w http.ResponseWriter, src http.Header) {
	for key, values := range src {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
}
