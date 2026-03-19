package web

import (
	"net/http"

	"github.com/binaryphile/fluentfp/option"
)

// PathParam returns the named path parameter from the request as an Option.
// Returns not-ok if the parameter is missing or empty.
// Wraps [http.Request.PathValue] with [option.NonEmpty].
func PathParam(req *http.Request, name string) option.String {
	return option.NonEmpty(req.PathValue(name))
}
