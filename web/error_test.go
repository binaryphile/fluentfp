package web_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/binaryphile/fluentfp/web"
)

func TestErrorMessage(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := web.BadRequest("bad input")

		if err.Error() != "bad input" {
			t.Fatalf("Error() = %q, want %q", err.Error(), "bad input")
		}
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("parse failed")
		e := &web.Error{Status: 400, Message: "bad input", Err: cause}

		want := "bad input: parse failed"
		if e.Error() != want {
			t.Fatalf("Error() = %q, want %q", e.Error(), want)
		}
	})
}

func TestErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	e := &web.Error{Status: 500, Message: "failed", Err: cause}

	if !errors.Is(e, cause) {
		t.Fatal("errors.Is should find the wrapped cause")
	}
}

func TestErrorsAs(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		err := web.NotFound("user not found")

		var webErr *web.Error
		if !errors.As(err, &webErr) {
			t.Fatal("errors.As should find *web.Error")
		}
		if webErr.Status != http.StatusNotFound {
			t.Fatalf("Status = %d, want %d", webErr.Status, http.StatusNotFound)
		}
	})

	t.Run("wrapped", func(t *testing.T) {
		inner := web.BadRequest("bad")
		wrapped := fmt.Errorf("handler: %w", inner)

		var webErr *web.Error
		if !errors.As(wrapped, &webErr) {
			t.Fatal("errors.As should find *web.Error through wrapping")
		}
		if webErr.Status != http.StatusBadRequest {
			t.Fatalf("Status = %d, want %d", webErr.Status, http.StatusBadRequest)
		}
	})
}

func TestBadRequest(t *testing.T) {
	err := web.BadRequest("invalid")

	var e *web.Error
	if !errors.As(err, &e) {
		t.Fatal("expected *web.Error")
	}
	if e.Status != http.StatusBadRequest {
		t.Fatalf("Status = %d, want 400", e.Status)
	}
	if e.Code != "BAD_REQUEST" {
		t.Fatalf("Code = %q, want %q", e.Code, "BAD_REQUEST")
	}
}

func TestNotFound(t *testing.T) {
	err := web.NotFound("missing")

	var e *web.Error
	if !errors.As(err, &e) {
		t.Fatal("expected *web.Error")
	}
	if e.Status != http.StatusNotFound {
		t.Fatalf("Status = %d, want 404", e.Status)
	}
	if e.Code != "NOT_FOUND" {
		t.Fatalf("Code = %q, want %q", e.Code, "NOT_FOUND")
	}
}

func TestConflict(t *testing.T) {
	err := web.Conflict("duplicate")

	var e *web.Error
	if !errors.As(err, &e) {
		t.Fatal("expected *web.Error")
	}
	if e.Status != http.StatusConflict {
		t.Fatalf("Status = %d, want 409", e.Status)
	}
	if e.Code != "CONFLICT" {
		t.Fatalf("Code = %q, want %q", e.Code, "CONFLICT")
	}
}

func TestForbidden(t *testing.T) {
	err := web.Forbidden("denied")

	var e *web.Error
	if !errors.As(err, &e) {
		t.Fatal("expected *web.Error")
	}
	if e.Status != http.StatusForbidden {
		t.Fatalf("Status = %d, want 403", e.Status)
	}
	if e.Code != "FORBIDDEN" {
		t.Fatalf("Code = %q, want %q", e.Code, "FORBIDDEN")
	}
}

func TestStatusError(t *testing.T) {
	err := web.StatusError(http.StatusTooManyRequests, "RATE_LIMITED", "slow down")

	var e *web.Error
	if !errors.As(err, &e) {
		t.Fatal("expected *web.Error")
	}
	if e.Status != http.StatusTooManyRequests {
		t.Fatalf("Status = %d, want 429", e.Status)
	}
	if e.Code != "RATE_LIMITED" {
		t.Fatalf("Code = %q, want %q", e.Code, "RATE_LIMITED")
	}
	if e.Message != "slow down" {
		t.Fatalf("Message = %q, want %q", e.Message, "slow down")
	}
}

func TestErrorWithDetails(t *testing.T) {
	type FieldError struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}

	e := &web.Error{
		Status:  400,
		Message: "validation failed",
		Code:    "VALIDATION_ERROR",
		Details: []FieldError{{Field: "email", Message: "invalid format"}},
	}

	if e.Details == nil {
		t.Fatal("Details should not be nil")
	}

	details, ok := e.Details.([]FieldError)
	if !ok {
		t.Fatal("Details should be []FieldError")
	}
	if len(details) != 1 || details[0].Field != "email" {
		t.Fatalf("unexpected details: %v", details)
	}
}

func TestErrorWithHeaders(t *testing.T) {
	e := &web.Error{
		Status:  http.StatusUnauthorized,
		Message: "authentication required",
		Code:    "UNAUTHORIZED",
		Headers: http.Header{"Www-Authenticate": {"Bearer"}},
	}

	if e.Headers.Get("Www-Authenticate") != "Bearer" {
		t.Fatalf("Headers = %v, want WWW-Authenticate: Bearer", e.Headers)
	}
}
