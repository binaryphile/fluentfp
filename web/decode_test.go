package web_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/binaryphile/fluentfp/web"
)

type testPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func makeRequest(body, contentType string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req
}

func TestDecodeJSON(t *testing.T) {
	t.Run("decodes valid JSON", func(t *testing.T) {
		req := makeRequest(`{"name":"alice","age":30}`, "application/json")
		result := web.DecodeJSON[testPayload](req)

		val, ok := result.Get()
		if !ok {
			_, err := result.Unpack()
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Name != "alice" || val.Age != 30 {
			t.Fatalf("got %+v, want {alice 30}", val)
		}
	})

	t.Run("accepts missing content-type", func(t *testing.T) {
		req := makeRequest(`{"name":"bob"}`, "")
		result := web.DecodeJSON[testPayload](req)

		if result.IsErr() {
			_, err := result.Unpack()
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("accepts charset variant", func(t *testing.T) {
		req := makeRequest(`{"name":"charlie"}`, "application/json; charset=utf-8")
		result := web.DecodeJSON[testPayload](req)

		if result.IsErr() {
			_, err := result.Unpack()
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("accepts +json suffix", func(t *testing.T) {
		req := makeRequest(`{"name":"diana"}`, "application/vnd.api+json")
		result := web.DecodeJSON[testPayload](req)

		if result.IsErr() {
			_, err := result.Unpack()
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects wrong content-type", func(t *testing.T) {
		req := makeRequest(`{"name":"eve"}`, "text/plain")
		result := web.DecodeJSON[testPayload](req)

		if result.IsOk() {
			t.Fatal("expected error for wrong content type")
		}

		_, err := result.Unpack()

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusUnsupportedMediaType {
			t.Fatalf("Status = %d, want 415", webErr.Status)
		}
	})

	t.Run("rejects empty body", func(t *testing.T) {
		req := makeRequest("", "application/json")
		result := web.DecodeJSON[testPayload](req)

		if result.IsOk() {
			t.Fatal("expected error for empty body")
		}

		_, err := result.Unpack()

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", webErr.Status)
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		req := makeRequest(`{invalid`, "application/json")
		result := web.DecodeJSON[testPayload](req)

		if result.IsOk() {
			t.Fatal("expected error for malformed JSON")
		}

		_, err := result.Unpack()

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", webErr.Status)
		}
	})

	t.Run("rejects unknown fields by default", func(t *testing.T) {
		req := makeRequest(`{"name":"frank","unknown":"field"}`, "application/json")
		result := web.DecodeJSON[testPayload](req)

		if result.IsOk() {
			t.Fatal("expected error for unknown fields")
		}

		_, err := result.Unpack()

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", webErr.Status)
		}
	})

	t.Run("rejects trailing garbage", func(t *testing.T) {
		req := makeRequest(`{"name":"grace"}{"extra":true}`, "application/json")
		result := web.DecodeJSON[testPayload](req)

		if result.IsOk() {
			t.Fatal("expected error for trailing garbage")
		}

		_, err := result.Unpack()

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", webErr.Status)
		}
	})
}

func TestDecodeJSONWith(t *testing.T) {
	t.Run("allows unknown fields when configured", func(t *testing.T) {
		req := makeRequest(`{"name":"helen","unknown":"field"}`, "application/json")
		result := web.DecodeJSONWith[testPayload](req, web.DecodeOpts{AllowUnknown: true})

		val, ok := result.Get()
		if !ok {
			_, err := result.Unpack()
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Name != "helen" {
			t.Fatalf("Name = %q, want %q", val.Name, "helen")
		}
	})

	t.Run("respects custom max bytes", func(t *testing.T) {
		req := makeRequest(`{"name":"iris"}`, "application/json")
		result := web.DecodeJSONWith[testPayload](req, web.DecodeOpts{MaxBytes: 5})

		if result.IsOk() {
			t.Fatal("expected error for body exceeding max bytes")
		}

		_, err := result.Unpack()

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusRequestEntityTooLarge {
			t.Fatalf("Status = %d, want 413", webErr.Status)
		}
	})
}
