package web_test

import . "github.com/binaryphile/fluentfp/web"
import "github.com/binaryphile/fluentfp/rslt"

// Compile-time export verification. Every fluentfp package uses this pattern
// to ensure exported symbols remain available across refactors.
func _() {
	// Response constructors
	_ = JSON[int]
	_ = OK[int]
	_ = Created[int]
	_ = NoContent
	_ = Response{}

	// Error type and constructors
	_ = ClientError{}
	_ = Error{}
	_ = BadRequest
	_ = NotFound
	_ = Conflict
	_ = Forbidden
	_ = TooManyRequests
	_ = StatusError

	// Params
	_ = PathParam

	// Decode
	_ = DecodeJSON[int]
	_ = DecodeJSONWith[int]
	_ = DecodeOpts{}

	// Steps
	_ = Steps[int]

	// Adapt
	_ = Handler(nil)
	_ = Adapt
	_ = WithErrorMapper
	_ = AdaptOption(nil)

	// rslt dependency (used by Handler return type)
	_ = rslt.Ok[int]
}
