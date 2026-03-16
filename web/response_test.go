package web_test

import (
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

func TestOK(t *testing.T) {
	resp := web.OK(42)

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
