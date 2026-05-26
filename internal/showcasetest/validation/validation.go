// Package validation compile-checks the showcase entry for hashicorp/terraform
// replace_triggered_by validation. The showcase elides "// classify refs,
// build diagnostics" inside the helper; compile-check uses minimal stubs so
// the slice.FlatMap pattern verifies.
package validation

import (
	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for hcl / langrefs / addrs types (heavily simplified) ---

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

// Diagnostics stubs hcl.Diagnostics (a defined slice type — see the showcase
// note about why validateTriggerExpr returns []*Diagnostic instead).
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

// --- the fluentfp rewrite from docs/showcase.md (verbatim structural pattern) ---

// validateTriggerExpr validates a single replace_triggered_by expression
// and returns all diagnostics found. Returns []*Diagnostic (not the
// Diagnostics defined type) so slice.FlatMap's generic inference
// resolves R = *Diagnostic.
func validateTriggerExpr(expr Expression) []*Diagnostic {
	expr, jsDiags := unwrapJSONRefExpr(expr)
	_, refDiags := langrefs.ReferencesInExpr(addrs.ParseRef, expr)
	// ... classify refs, build diagnostics ...
	return append(jsDiags, refDiags...)
}

func DecodeReplaceTriggeredBy(expr Expression) ([]Expression, Diagnostics) {
	exprs, diags := ExprList(expr)
	diags = append(diags, slice.FlatMap(exprs, validateTriggerExpr)...)
	return exprs, diags
}
