//go:build ignore

// Package snippet is the verification harness for the sieve showcase
// entry in docs/showcase.md (the golang/go test/chan/sieve1.go
// equivalent — channel-based prime sieve rewritten as a lazy stream).
// The snippet is function-body code (isPrime declared via :=, primes
// produced by a stream chain). The harness wraps it in FirstPrimes.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/stream"
)

func FirstPrimes() []int {
	// __SNIPPET__
}
