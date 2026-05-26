// Package difference compile-checks the showcase entry for hashicorp/go-secure-stdlib.
package difference

import (
	"strings"

	"github.com/binaryphile/fluentfp/hof"
	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/slice"
)

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func Difference(a, b slice.Mapper[string], lowercase bool) []string {
	// trimAndLower trims whitespace and lowercases.
	trimAndLower := hof.Pipe(strings.TrimSpace, strings.ToLower)
	// normalize trims whitespace, adding lowercasing when requested.
	normalize := option.When(lowercase, trimAndLower).Or(strings.TrimSpace)

	normA := a.Transform(normalize).KeepIf(lof.IsNonEmpty)
	normB := b.Transform(normalize).KeepIf(lof.IsNonEmpty)

	return slice.Difference(normA, normB).Sort(lof.StringAsc)
}
