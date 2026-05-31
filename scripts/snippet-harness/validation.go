//go:build ignore

// Package snippet is the verification harness for the validation
// showcase entry in docs/showcase.md (hashicorp/terraform
// replace_triggered_by validation accumulator rewrite). The snippet
// declares a package-level closure (validateTriggerExpr) and a top-
// level function (DecodeReplaceTriggeredBy); both go at package
// scope.
//
// Local stub names stand in for hcl.Expression / hcl.Diagnostic /
// hcl.Diagnostics / hcl.ExprList — same trade-off as annotation,
// sagas, temporal: lose the qualified-package texture in exchange
// for compile-verifiability without pulling in hashicorp/hcl. The
// addrs and langrefs package-level vars match the names the snippet
// uses; the harness defines them as singleton vars rather than
// actual packages.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"github.com/binaryphile/fluentfp/slice"
)

// Expression stubs hcl.Expression.
type Expression struct{}

func (Expression) Range() RangePtr { return RangePtr{} }

// RangePtr stubs hcl.Range pointer-returning wrapper.
type RangePtr struct{}

func (RangePtr) Ptr() *RangePtr { return nil }

// Diagnostic stubs hcl.Diagnostic.
type Diagnostic struct {
	Severity int
	Summary  string
	Detail   string
	Subject  *RangePtr
}

// Diagnostics stubs hcl.Diagnostics (a defined slice type — the
// showcase prose explains why validateTriggerExpr returns
// []*Diagnostic rather than this).
type Diagnostics []*Diagnostic

func (d Diagnostics) Extend(other Diagnostics) Diagnostics { return append(d, other...) }

const DiagError = 1

// ExprList stubs hcl.ExprList.
func ExprList(expr Expression) ([]Expression, Diagnostics) { return nil, nil }

// unwrapJSONRefExpr stubs the JSON-ref unwrap helper.
func unwrapJSONRefExpr(expr Expression) (Expression, Diagnostics) { return expr, nil }

// Reference stubs addrs.Reference.
type Reference struct{}

// AddrsClient stubs the addrs package's ParseRef entry point.
type AddrsClient struct{}

func (AddrsClient) ParseRef(_ Expression) (Reference, Diagnostics) { return Reference{}, nil }

var addrs = AddrsClient{}

// LangrefsClient stubs the langrefs package.
type LangrefsClient struct{}

func (LangrefsClient) ReferencesInExpr(parser func(Expression) (Reference, Diagnostics), expr Expression) ([]Reference, Diagnostics) {
	return nil, nil
}

var langrefs = LangrefsClient{}

// Force-reference slice so the unsubstituted harness parses with its
// imports used. The assembled snippet exercises slice.FlatMap.
var _ = slice.FlatMap[Expression, *Diagnostic]

// __SNIPPET__
