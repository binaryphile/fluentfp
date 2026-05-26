// Package middleware compile-checks the showcase entry for kubernetes/apiserver middleware fold.
package middleware

import (
	"net/http"

	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the k8s apiserver types and middleware packages ---

type Config struct{}

// Helper functions (per the showcase note: "withAuth(c) partially applies
// config to produce a Middleware from a multi-arg wrapper").
func withAuth(c *Config) Middleware                          { return passthrough }
func withLatencyTracking(c *Config, name string) Middleware  { return passthrough }
func withAuthentication(c *Config) Middleware                { return passthrough }
func withCORS(c *Config) Middleware                          { return passthrough }
func withTimeout(c *Config) Middleware                       { return passthrough }
func withPanicRecovery(c *Config) Middleware                 { return passthrough }

func passthrough(h http.Handler) http.Handler { return h }

// Package-style stubs to match the showcase's `pkg.Symbol` references.
type filterlatencyStub struct{}

func (filterlatencyStub) TrackCompleted(h http.Handler) http.Handler { return h }

var filterlatency = filterlatencyStub{}

type genericapifiltersStub struct{}

func (genericapifiltersStub) WithWarningRecorder(h http.Handler) http.Handler { return h }
func (genericapifiltersStub) WithAuditInit(h http.Handler) http.Handler       { return h }

var genericapifilters = genericapifiltersStub{}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

type Middleware func(http.Handler) http.Handler

// applyMiddleware wraps the handler with the next middleware layer.
var applyMiddleware = func(h http.Handler, mw Middleware) http.Handler {
	return mw(h)
}

func DefaultBuildHandlerChain(apiHandler http.Handler, c *Config) http.Handler {
	// Build middleware stack as data — inspectable, testable, reorderable
	middlewares := []Middleware{
		filterlatency.TrackCompleted,
		withAuth(c),
		withLatencyTracking(c, "authorization"),
		withAuthentication(c),
		withCORS(c),
		genericapifilters.WithWarningRecorder,
		withTimeout(c),
		// ... (additional layers omitted in the showcase)
		withPanicRecovery(c),
		genericapifilters.WithAuditInit,
	}

	handler := slice.Fold(middlewares, apiHandler, applyMiddleware)
	return handler
}
