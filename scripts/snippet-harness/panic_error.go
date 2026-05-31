//go:build ignore

// Package snippet is the verification harness for the PanicError
// type declaration in docs/parallelism-research.md (~line 532).
// The block declares a package-level type + Error() method, so the
// marker sits at package scope.
package snippet

import (
	"fmt"
)

// __SNIPPET__

// Force-reference for pre-substitution parse parity. After
// substitution the snippet uses fmt.Sprintf in the Error() method.
var _ = fmt.Sprintf
