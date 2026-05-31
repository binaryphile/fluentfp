//go:build ignore

// Package snippet is the verification harness for the conventional-Go
// raw handler in examples/orders/README.md (~line 15). The block is
// a top-level `func handleCreateOrder(w, req)` using pure stdlib
// (no fluentfp); the harness stubs the Order type, priceOrder, and
// the package-level store variable so the function body compiles.
package snippet

import (
	"context"
	"encoding/json"
	"net/http"
)

// Order stubs the order domain type with the Customer field the
// snippet's validation checks.
type Order struct {
	Customer string
}

// priceOrder stubs the synchronous pricing call.
func priceOrder(ctx context.Context, o Order) (Order, error) {
	return o, nil
}

// orderStore stubs the in-memory store with the put method the
// snippet's terminal write exercises.
type orderStore struct{}

func (orderStore) put(o Order) {}

// store stubs the package-level store variable the snippet writes to.
var store orderStore

// __SNIPPET__

// Force-reference each import for pre-substitution parse parity.
var _ = json.NewEncoder
var _ = http.MaxBytesReader
