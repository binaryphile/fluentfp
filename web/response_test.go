package web_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/binaryphile/fluentfp/web"
)

func TestJSON(t *testing.T) {
	t.Run("sets status and body", func(t *testing.T) {
		resp := web.JSON(http.StatusAccepted, "hello")

		if resp.Status != http.StatusAccepted {
			t.Fatalf("Status = %d, want %d", resp.Status, http.StatusAccepted)
		}
		if resp.Body != "hello" {
			t.Fatalf("Body = %v, want %q", resp.Body, "hello")
		}
		if resp.Headers != nil {
			t.Fatalf("Headers = %v, want nil", resp.Headers)
		}
	})
}

func TestOk(t *testing.T) {
	resp := web.Ok(42)

	if resp.Status != http.StatusOK {
		t.Fatalf("Status = %d, want %d", resp.Status, http.StatusOK)
	}
	if resp.Body != 42 {
		t.Fatalf("Body = %v, want 42", resp.Body)
	}
}

func TestCreated(t *testing.T) {
	resp := web.Created("new-resource")

	if resp.Status != http.StatusCreated {
		t.Fatalf("Status = %d, want %d", resp.Status, http.StatusCreated)
	}
	if resp.Body != "new-resource" {
		t.Fatalf("Body = %v, want %q", resp.Body, "new-resource")
	}
}

func TestNoContent(t *testing.T) {
	resp := web.NoContent()

	if resp.Status != http.StatusNoContent {
		t.Fatalf("Status = %d, want %d", resp.Status, http.StatusNoContent)
	}
	if resp.Body != nil {
		t.Fatalf("Body = %v, want nil", resp.Body)
	}
}

// TestOkResult is the representative test for the Result-shaped
// constructors (OkResult/CreatedResult/NoContentResult/JSONResult share
// the identical delegate-then-wrap shape -- see the guide's
// "Representative Pattern Testing" section).
func TestOkResult(t *testing.T) {
	result := web.OkResult(42)

	resp, err := result.Unpack()
	if err != nil {
		t.Fatalf("Unpack() err = %v, want nil", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("Status = %d, want %d", resp.Status, http.StatusOK)
	}
	if resp.Body != 42 {
		t.Fatalf("Body = %v, want 42", resp.Body)
	}
}

func TestErrResult(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := web.ErrResult(wantErr).Unpack()

	if !errors.Is(err, wantErr) {
		t.Fatalf("Unpack() err = %v, want %v", err, wantErr)
	}
}
