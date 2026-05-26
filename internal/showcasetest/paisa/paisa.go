// Package paisa compile-checks the showcase entry for ananthakumaran/paisa.
package paisa

import (
	"regexp"
	"strings"

	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/slice"
)

// --- shared between original and rewrite (also in showcase) ---

func splitTokens(s string) slice.Mapper[string] {
	return regexp.MustCompile("[ .()/:]+").Split(s, -1)
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func Tokenize(s string) []string {
	tokens := splitTokens(s).Transform(strings.ToLower).KeepIf(lof.IsNonBlank)
	return tokens
}
