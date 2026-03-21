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
		val, err := web.DecodeJSON[testPayload](req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Name != "alice" || val.Age != 30 {
			t.Fatalf("got %+v, want {alice 30}", val)
		}
	})

	t.Run("accepts missing content-type", func(t *testing.T) {
		req := makeRequest(`{"name":"bob"}`, "")
		_, err := web.DecodeJSON[testPayload](req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("accepts charset variant", func(t *testing.T) {
		req := makeRequest(`{"name":"charlie"}`, "application/json; charset=utf-8")
		_, err := web.DecodeJSON[testPayload](req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("accepts +json suffix", func(t *testing.T) {
		req := makeRequest(`{"name":"diana"}`, "application/vnd.api+json")
		_, err := web.DecodeJSON[testPayload](req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects wrong content-type", func(t *testing.T) {
		req := makeRequest(`{"name":"eve"}`, "text/plain")
		_, err := web.DecodeJSON[testPayload](req)

		if err == nil {
			t.Fatal("expected error for wrong content type")
		}

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
		_, err := web.DecodeJSON[testPayload](req)

		if err == nil {
			t.Fatal("expected error for empty body")
		}

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
		_, err := web.DecodeJSON[testPayload](req)

		if err == nil {
			t.Fatal("expected error for malformed JSON")
		}

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
		_, err := web.DecodeJSON[testPayload](req)

		if err == nil {
			t.Fatal("expected error for unknown fields")
		}

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
		_, err := web.DecodeJSON[testPayload](req)

		if err == nil {
			t.Fatal("expected error for trailing garbage")
		}

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
		val, err := web.DecodeJSONWith[testPayload](req, web.DecodeOpts{AllowUnknown: true})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Name != "helen" {
			t.Fatalf("Name = %q, want %q", val.Name, "helen")
		}
	})

	t.Run("respects custom max bytes", func(t *testing.T) {
		req := makeRequest(`{"name":"iris"}`, "application/json")
		_, err := web.DecodeJSONWith[testPayload](req, web.DecodeOpts{MaxBytes: 5})

		if err == nil {
			t.Fatal("expected error for body exceeding max bytes")
		}

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("expected *web.Error")
		}
		if webErr.Status != http.StatusRequestEntityTooLarge {
			t.Fatalf("Status = %d, want 413", webErr.Status)
		}
	})
}
