package web

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/binaryphile/fluentfp/rslt"
)

const defaultMaxBytes int64 = 1 << 20 // 1MB

// DecodeOpts provides per-call overrides for decode policy.
type DecodeOpts struct {
	MaxBytes     int64 // 0 = use default (1MB)
	AllowUnknown bool  // false = reject unknown fields (default)
}

// DecodeJSON reads and JSON-decodes the request body into T.
//
// Policy:
//   - Content-Type: accepts application/json, application/json;charset=utf-8,
//     and application/*+json variants. Missing Content-Type is accepted
//     (lenient for ad hoc clients). Wrong Content-Type returns 415.
//   - Max body size: 1MB (returns 413 if exceeded)
//   - Disallows unknown fields (returns 400)
//   - Empty body returns 400
//   - Malformed JSON returns 400
//   - Trailing garbage after valid JSON returns 400
func DecodeJSON[T any](req *http.Request) rslt.Result[T] {
	return DecodeJSONWith[T](req, DecodeOpts{})
}

// DecodeJSONWith overrides default decode policy.
func DecodeJSONWith[T any](req *http.Request, opts DecodeOpts) rslt.Result[T] {
	if err := checkContentType(req); err != nil {
		return rslt.Err[T](err)
	}

	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}

	body := http.MaxBytesReader(nil, req.Body, maxBytes)
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		if isMaxBytesError(err) {
			return rslt.Err[T](&Error{
				Status:  http.StatusRequestEntityTooLarge,
				Message: "request body too large",
				Code:    "REQUEST_TOO_LARGE",
			})
		}

		return rslt.Err[T](BadRequest("failed to read request body"))
	}

	if len(data) == 0 {
		return rslt.Err[T](BadRequest("request body is empty"))
	}

	var result T

	dec := json.NewDecoder(bytes.NewReader(data))
	if !opts.AllowUnknown {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(&result); err != nil {
		return rslt.Err[T](BadRequest("malformed JSON: " + sanitizeJSONError(err)))
	}

	// Check for trailing garbage.
	if dec.More() {
		return rslt.Err[T](BadRequest("malformed JSON: trailing data after value"))
	}

	return rslt.Ok(result)
}

// checkContentType validates the Content-Type header. Missing is accepted;
// wrong is 415. Accepts application/json and application/*+json variants.
func checkContentType(req *http.Request) error {
	ct := req.Header.Get("Content-Type")
	if ct == "" {
		return nil // lenient for ad hoc clients
	}

	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return &Error{
			Status:  http.StatusUnsupportedMediaType,
			Message: "unsupported content type",
			Code:    "UNSUPPORTED_MEDIA_TYPE",
		}
	}

	if mediaType == "application/json" {
		return nil
	}

	// Accept application/*+json (e.g. application/vnd.api+json).
	if strings.HasPrefix(mediaType, "application/") && strings.HasSuffix(mediaType, "+json") {
		return nil
	}

	return &Error{
		Status:  http.StatusUnsupportedMediaType,
		Message: "unsupported content type: " + mediaType,
		Code:    "UNSUPPORTED_MEDIA_TYPE",
	}
}

// sanitizeJSONError returns a safe error message from a JSON decode error,
// avoiding leaking internal type information.
func sanitizeJSONError(err error) string {
	switch e := err.(type) {
	case *json.SyntaxError:
		return "syntax error"
	case *json.UnmarshalTypeError:
		if e.Field != "" {
			return "type mismatch for field " + e.Field
		}

		return "type mismatch"
	default:
		msg := err.Error()
		// json.Decoder unknown field errors start with "json: unknown field"
		if strings.HasPrefix(msg, "json: unknown field") {
			return msg
		}

		return "invalid JSON"
	}
}

// isMaxBytesError detects http.MaxBytesError.
func isMaxBytesError(err error) bool {
	// http.MaxBytesError was added in Go 1.19.
	_, ok := err.(*http.MaxBytesError)

	return ok
}
