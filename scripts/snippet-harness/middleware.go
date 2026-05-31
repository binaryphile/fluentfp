//go:build ignore

// Package snippet is the verification harness for the middleware showcase
// entry in docs/showcase.md (kubernetes/apiserver middleware-as-data
// rewrite). The snippet is function-body code (declares a function-local
// type Middleware, an applyMiddleware var, a middlewares slice literal,
// and a handler variable; ends with `return handler`).
//
// The harness helpers return the unnamed type `func(http.Handler)
// http.Handler` rather than the named `Middleware` type. Go's
// assignability rules allow values of the underlying unnamed type to
// populate a `[]Middleware` slice literal, so the snippet's local
// `type Middleware` declaration coexists with these helpers without
// duplicate-declaration or type-mismatch errors.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"net/http"

	"github.com/binaryphile/fluentfp/slice"
)

// Config stubs the k8s apiserver config carried through middleware
// constructors. Only the type identity is used by the snippet.
type Config struct{}

// passthrough is the no-op handler middleware used as a placeholder
// return value for every stubbed middleware constructor.
func passthrough(h http.Handler) http.Handler { return h }

// Middleware constructors used by the snippet. They each return the
// unnamed type `func(http.Handler) http.Handler`, which is assignable
// to the snippet's local `type Middleware` per Go's assignability rules.
func withAuth(c *Config) func(http.Handler) http.Handler {
	return passthrough
}
func withLatencyTracking(c *Config, name string) func(http.Handler) http.Handler {
	return passthrough
}
func withAuthentication(c *Config) func(http.Handler) http.Handler { return passthrough }
func withCORS(c *Config) func(http.Handler) http.Handler           { return passthrough }
func withTimeout(c *Config) func(http.Handler) http.Handler        { return passthrough }
func withPanicRecovery(c *Config) func(http.Handler) http.Handler  { return passthrough }

// Package-style stubs mirror the snippet's `pkg.Symbol` references.
type filterlatencyStub struct{}

func (filterlatencyStub) TrackCompleted(h http.Handler) http.Handler { return h }

var filterlatency = filterlatencyStub{}

type genericapifiltersStub struct{}

func (genericapifiltersStub) WithWarningRecorder(h http.Handler) http.Handler { return h }
func (genericapifiltersStub) WithAuditInit(h http.Handler) http.Handler       { return h }

var genericapifilters = genericapifiltersStub{}

// Keep slice as imported even before substitution.
var _ = slice.Fold[int, int]

func DefaultBuildHandlerChain(apiHandler http.Handler, c *Config) http.Handler {
	// __SNIPPET__
}
