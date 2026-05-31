//go:build ignore

// Package snippet is the verification harness for the rate-limit
// middleware showcase block in web/README.md (~line 144). The block
// declares two top-level functions: withRateLimit (the middleware
// constructor) and writeRateLimitError (the JSON-error writer).
// Both land at package scope, so the marker sits at package scope.
//
// External dep: golang.org/x/time/rate — declared in the companion
// rate_limit_mw.gomod file so the script's tmpdir go.mod gets the
// extra require.
package snippet

import (
	"encoding/json"
	"net/http"

	"github.com/binaryphile/fluentfp/web"
	"golang.org/x/time/rate"
)

// __SNIPPET__

// Force-reference each import so the un-substituted harness's static
// shape doesn't depend on which symbols the snippet exercises. After
// substitution the snippet itself uses every import; these lines
// remain harmless and absorb future snippet edits.
var _ = json.NewEncoder
var _ = http.StatusTooManyRequests
var _ = web.ClientError{}
var _ = rate.Limit(0)
