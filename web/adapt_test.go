package web_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/web"
)

func TestAdapt(t *testing.T) {
	t.Run("ok response with body", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Ok(web.OK(map[string]string{"msg": "hello"}))
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("Status = %d, want 200", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}

		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["msg"] != "hello" {
			t.Fatalf("body = %v, want {msg: hello}", body)
		}
	})

	t.Run("no content response", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Ok(web.NoContent())
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodDelete, "/", nil))

		if w.Code != http.StatusNoContent {
			t.Fatalf("Status = %d, want 204", w.Code)
		}
		if w.Body.Len() != 0 {
			t.Fatalf("body should be empty, got %d bytes", w.Body.Len())
		}
	})

	t.Run("response with custom headers", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Ok(web.Response{
				Status:  http.StatusOK,
				Headers: http.Header{"X-Custom": {"value"}},
				Body:    "ok",
			})
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Header().Get("X-Custom") != "value" {
			t.Fatalf("X-Custom = %q, want %q", w.Header().Get("X-Custom"), "value")
		}
	})

	t.Run("web.Error renders as ClientError", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Err[web.Response](web.NotFound("user not found"))
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusNotFound {
			t.Fatalf("Status = %d, want 404", w.Code)
		}

		var ce web.ClientError
		if err := json.Unmarshal(w.Body.Bytes(), &ce); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if ce.Error != "user not found" {
			t.Fatalf("error = %q, want %q", ce.Error, "user not found")
		}
		if ce.Code != "NOT_FOUND" {
			t.Fatalf("code = %q, want %q", ce.Code, "NOT_FOUND")
		}
	})

	t.Run("wrapped web.Error found via errors.As", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			inner := web.BadRequest("bad input")

			return rslt.Err[web.Response](fmt.Errorf("handler: %w", inner))
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", w.Code)
		}
	})

	t.Run("non-web error becomes 500", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Err[web.Response](fmt.Errorf("database connection failed"))
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want 500", w.Code)
		}

		var ce web.ClientError
		if err := json.Unmarshal(w.Body.Bytes(), &ce); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if ce.Error != "internal error" {
			t.Fatalf("error = %q, want %q", ce.Error, "internal error")
		}
		// Verify internal error message is NOT exposed.
		body := w.Body.String()
		if strings.Contains(body, "database") {
			t.Fatal("internal error message leaked to client")
		}
	})

	t.Run("error with headers writes headers", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Err[web.Response](&web.Error{
				Status:  http.StatusUnauthorized,
				Message: "authentication required",
				Code:    "UNAUTHORIZED",
				Headers: http.Header{"Www-Authenticate": {"Bearer"}},
			})
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Status = %d, want 401", w.Code)
		}
		if w.Header().Get("Www-Authenticate") != "Bearer" {
			t.Fatalf("WWW-Authenticate = %q, want %q", w.Header().Get("Www-Authenticate"), "Bearer")
		}
	})

	t.Run("error with details includes details in response", func(t *testing.T) {
		type fieldErr struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		}

		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Err[web.Response](&web.Error{
				Status:  http.StatusBadRequest,
				Message: "validation failed",
				Code:    "VALIDATION_ERROR",
				Details: []fieldErr{{Field: "email", Message: "invalid"}},
			})
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if _, ok := raw["details"]; !ok {
			t.Fatal("response missing details field")
		}
	})

	t.Run("error with zero status defaults to 500", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Err[web.Response](&web.Error{Message: "oops"}) // Status: 0
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want 500 (zero-status guard)", w.Code)
		}
	})

	t.Run("no content omits Content-Type", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			return rslt.Ok(web.NoContent())
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodDelete, "/", nil))

		if ct := w.Header().Get("Content-Type"); ct != "" {
			t.Fatalf("Content-Type = %q, want empty for NoContent", ct)
		}
	})

	t.Run("marshal failure returns 500", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			// Channels cannot be marshaled to JSON.
			return rslt.Ok(web.OK(make(chan int)))
		})

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want 500", w.Code)
		}
	})
}

func TestAdaptWithErrorMapper(t *testing.T) {
	errDuplicate := fmt.Errorf("duplicate entity")

	// domainMapper maps domain errors to *web.Error.
	domainMapper := func(err error) (*web.Error, bool) {
		if errors.Is(err, errDuplicate) {
			return &web.Error{
				Status:  http.StatusConflict,
				Message: "already exists",
				Code:    "CONFLICT",
			}, true
		}

		return nil, false
	}

	t.Run("maps domain error", func(t *testing.T) {
		handler := web.Adapt(
			func(r *http.Request) rslt.Result[web.Response] {
				return rslt.Err[web.Response](errDuplicate)
			},
			web.WithErrorMapper(domainMapper),
		)

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodPost, "/", nil))

		if w.Code != http.StatusConflict {
			t.Fatalf("Status = %d, want 409", w.Code)
		}

		var ce web.ClientError
		if err := json.Unmarshal(w.Body.Bytes(), &ce); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if ce.Code != "CONFLICT" {
			t.Fatalf("code = %q, want %q", ce.Code, "CONFLICT")
		}
	})

	t.Run("unmapped domain error falls through to 500", func(t *testing.T) {
		handler := web.Adapt(
			func(r *http.Request) rslt.Result[web.Response] {
				return rslt.Err[web.Response](fmt.Errorf("unknown error"))
			},
			web.WithErrorMapper(domainMapper),
		)

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want 500", w.Code)
		}
	})

	t.Run("*web.Error bypasses mapper", func(t *testing.T) {
		mapperCalled := false
		// spyMapper tracks whether the mapper is called.
		spyMapper := func(err error) (*web.Error, bool) {
			mapperCalled = true

			return nil, false
		}

		handler := web.Adapt(
			func(r *http.Request) rslt.Result[web.Response] {
				return rslt.Err[web.Response](web.BadRequest("from decode"))
			},
			web.WithErrorMapper(spyMapper),
		)

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", w.Code)
		}
		if mapperCalled {
			t.Fatal("mapper should not be called for *web.Error")
		}
	})

	t.Run("wrapped *web.Error bypasses mapper", func(t *testing.T) {
		mapperCalled := false
		// spyMapper tracks whether the mapper is called.
		spyMapper := func(err error) (*web.Error, bool) {
			mapperCalled = true

			return nil, false
		}

		handler := web.Adapt(
			func(r *http.Request) rslt.Result[web.Response] {
				return rslt.Err[web.Response](fmt.Errorf("wrap: %w", web.NotFound("gone")))
			},
			web.WithErrorMapper(spyMapper),
		)

		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusNotFound {
			t.Fatalf("Status = %d, want 404", w.Code)
		}
		if mapperCalled {
			t.Fatal("mapper should not be called for wrapped *web.Error")
		}
	})
}

func TestAdaptDecodeIntegration(t *testing.T) {
	t.Run("decode error renders as ClientError through Adapt", func(t *testing.T) {
		handler := web.Adapt(func(r *http.Request) rslt.Result[web.Response] {
			type Payload struct {
				Name string `json:"name"`
			}

			p, err := web.DecodeJSON[Payload](r)
			if err != nil {
				return rslt.Err[web.Response](err)
			}

			return rslt.Ok(web.OK(p))
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}

		var ce web.ClientError
		if err := json.Unmarshal(w.Body.Bytes(), &ce); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if ce.Code != "BAD_REQUEST" {
			t.Fatalf("code = %q, want %q", ce.Code, "BAD_REQUEST")
		}
	})

	t.Run("decode error bypasses error mapper", func(t *testing.T) {
		mapperCalled := false

		handler := web.Adapt(
			func(r *http.Request) rslt.Result[web.Response] {
				type Payload struct {
					Name string `json:"name"`
				}

				p, err := web.DecodeJSON[Payload](r)
				if err != nil {
					return rslt.Err[web.Response](err)
				}

				return rslt.Ok(web.OK(p))
			},
			web.WithErrorMapper(func(err error) (*web.Error, bool) {
				mapperCalled = true

				return nil, false
			}),
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Status = %d, want 400", w.Code)
		}
		if mapperCalled {
			t.Fatal("mapper should not be called for decode errors (already *web.Error)")
		}
	})
}

func TestAdaptPanics(t *testing.T) {
	t.Run("nil handler", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		web.Adapt(nil)
	})

	t.Run("nil error mapper fn", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		web.WithErrorMapper(nil)
	})
}

