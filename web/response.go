package web

import (
	"net/http"

	"github.com/binaryphile/fluentfp/rslt"
)

// Response is the return value of a handler — status + headers + body.
// Body is any because a generic Response[T] would infect the entire Handler
// signature. This is a boundary escape hatch, not full type safety.
// Constructors are generic to preserve call-site type intent.
type Response struct {
	Status  int
	Headers http.Header // nil means no extra headers; copied in Adapt
	Body    any         // JSON-serialized by Adapt; nil means no body
}

// JSON returns a Response with the given status and body.
func JSON[T any](status int, body T) Response {
	return Response{Status: status, Body: body}
}

// Ok returns a 200 Response with the given body.
func Ok[T any](body T) Response {
	return Response{Status: http.StatusOK, Body: body}
}

// Created returns a 201 Response with the given body.
func Created[T any](body T) Response {
	return Response{Status: http.StatusCreated, Body: body}
}

// NoContent returns a 204 Response with no body.
func NoContent() Response {
	return Response{Status: http.StatusNoContent}
}

// Result-shaped siblings of the plain constructors above, for the common
// case where a Handler returns immediately with no post-construction
// mutation (setting a cookie header, etc.) -- direct-return shortcuts for
// the `rslt.Ok(web.Xxx(...))` shape. Additive alongside the plain
// constructors and rslt.Ok/rslt.Err, not a replacement: changing the
// plain constructors' own return type to rslt.Result[Response] would
// break every existing call site across every Handler already written
// against the current signature, AND would remove the ability to mutate
// a Response (e.g. resp.Headers = setCookieHeader(...)) before wrapping
// it -- a real, still-needed shape (see handleSignup/handleLogin in
// internal/platformapi for TNT's own examples).

// OkResult returns a 200 Response with the given body, already wrapped
// in rslt.Result.
func OkResult[T any](body T) rslt.Result[Response] {
	return rslt.Ok(Ok(body))
}

// CreatedResult returns a 201 Response with the given body, already
// wrapped in rslt.Result.
func CreatedResult[T any](body T) rslt.Result[Response] {
	return rslt.Ok(Created(body))
}

// NoContentResult returns a 204 Response with no body, already wrapped
// in rslt.Result.
func NoContentResult() rslt.Result[Response] {
	return rslt.Ok(NoContent())
}

// JSONResult returns a Response with the given status and body, already
// wrapped in rslt.Result.
func JSONResult[T any](status int, body T) rslt.Result[Response] {
	return rslt.Ok(JSON(status, body))
}

// ErrResult wraps err as an Err[Response] -- the Result-side counterpart
// to the constructors above, replacing the need to spell out
// `rslt.Err[web.Response](err)`'s explicit type parameter at every call
// site.
func ErrResult(err error) rslt.Result[Response] {
	return rslt.Err[Response](err)
}
