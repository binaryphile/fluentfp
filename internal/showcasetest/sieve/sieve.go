// Package sieve compile-checks the showcase entry for golang/go sieve1.go.
package sieve

import (
	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/stream"
)

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

// isPrime returns true if n has no divisors other than 1 and itself.
func isPrime(n int) bool {
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func FirstPrimes() []int {
	primes := stream.Generate(2, lof.Inc).KeepIf(isPrime).Take(25).Collect()
	return primes
}
