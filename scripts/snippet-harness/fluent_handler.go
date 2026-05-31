//go:build ignore

// Package snippet is the verification harness for the fluentfp
// handler closure in examples/orders/README.md (~line 64). The
// block declares `handleCreateOrder := func(...) rslt.Result[...]
// {...}` inside an implicit function shell. The harness wraps it
// in Demo and returns the closure so Go's declared-and-not-used
// check is satisfied.
package snippet

import (
	"context"
	"log"
	"net/http"

	"github.com/binaryphile/fluentfp/ctxval"
	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/web"
)

// RequestID stubs the per-request correlation type carried in ctx.
type RequestID string

// Order stubs the order domain type.
type Order struct{}

// priceFn stubs the async-pricing dependency the closure binds
// req.Context() into.
var priceFn func(context.Context, Order) (Order, error)

// store stubs the storage type with the put method the snippet
// rebinds as saveOrder.
type store struct{}

func (s *store) put(o Order) {}

// s stubs the package-level *store the snippet method-binds.
var s = &store{}

// validateOrder stubs the named business-rule validator.
func validateOrder(o Order) rslt.Result[Order] { return rslt.Ok(o) }

// withNewID stubs the infallible ID-assignment step.
func withNewID(o Order) Order { return o }

func Demo() func(*http.Request) rslt.Result[web.Response] {
	// __SNIPPET__
	return handleCreateOrder
}

// Force-reference for pre-substitution parse parity.
var _ = ctxval.Lookup[RequestID]
var _ = log.Printf
