//go:build ignore

// Package snippet is the verification harness for the query-string
// parsing block in examples/orders/README.md (~line 194). The block
// reads optional `status` and `min_total` query params, builds two
// closure predicates, and applies SortBy + KeepIf to s.list().
//
// The snippet has an early `return rslt.Err[web.Response](err)`
// inside an `if err != nil`, so the function shell must return
// rslt.Result[web.Response]. The terminal `orders` is wrapped into
// rslt.Ok(web.OK(orders)) so the variable is live.
package snippet

import (
	"net/url"

	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
	"github.com/binaryphile/fluentfp/web"
)

// Order stubs the order domain type with the fields the predicates
// inspect (Status, TotalCents).
type Order struct {
	Status     string
	TotalCents int
}

// orderNum stubs the sort-key function the snippet's SortBy uses.
func orderNum(o Order) int { return 0 }

// parseMinTotal stubs the parser the option.MapResult chain calls;
// signature matches main.go: `func(string) rslt.Result[int]`.
func parseMinTotal(raw string) rslt.Result[int] { return rslt.Ok(0) }

// orderStore stubs the storage type with the list() method.
type orderStore struct{}

func (orderStore) list() []Order { return nil }

// s stubs the package-level *store-equivalent.
var s = orderStore{}

func Demo(q url.Values) rslt.Result[web.Response] {
	// __SNIPPET__
	return rslt.Ok(web.OK(orders))
}

// Force-reference for pre-substitution parse parity.
var _ = option.NonEmpty
var _ = slice.SortBy[Order, int]
