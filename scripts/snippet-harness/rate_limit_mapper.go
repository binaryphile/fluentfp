//go:build ignore

// Package snippet is the verification harness for the rate-limit
// mapper closure block in web/README.md (~line 171). The block
// declares `mapRateLimit := func(err error) (*web.Error, bool) {...}`
// inside an implicit function shell — the harness wraps it in Demo
// and consumes the closure via web.WithErrorMapper so Go's
// declared-and-not-used check doesn't trip.
package snippet

import (
	"errors"
	"net/http"

	"github.com/binaryphile/fluentfp/web"
)

// errRateLimited stubs the domain error the snippet sentinel-compares
// via errors.Is.
var errRateLimited = errors.New("rate limited")

func Demo() web.AdaptOption {
	// __SNIPPET__
	return web.WithErrorMapper(mapRateLimit)
}

// Force-reference each import for pre-substitution parse parity.
var _ = http.Header{}
