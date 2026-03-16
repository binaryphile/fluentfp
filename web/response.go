package web

import "net/http"

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

// OK returns a 200 Response with the given body.
func OK[T any](body T) Response {
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
