// Package web provides JSON HTTP adapter composition for net/http handlers.
//
// Handlers return rslt.Result[Response] instead of writing to ResponseWriter,
// making them testable expressions that return values. Adapt bridges handlers
// to http.HandlerFunc at the registration boundary.
//
// This package helps build HTTP adapters — the transport boundary layer.
// Domain logic should live in separate functions that handlers call.
// JSON request/response only — not a general web framework.
package web
