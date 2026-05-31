//go:build ignore

// Package snippet is the verification harness for the pagination
// showcase entry in docs/showcase.md (Amazonka-style S3 cursor
// pagination as a lazy stream). The showcase has TWO fluentfp
// fences sharing variables:
//
//   slot=define  → declares pageStep + pages
//   slot=consume → invokes three different consumers on `pages`
//
// Both slots land inside the same function shell so `pages` flows
// from define into consume naturally.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/stream"
)

// ObjectPage stubs an S3 listing page. NextTokenOption is what
// pageStep emits as the lazy-cursor signal for stream.Paginate.
type ObjectPage struct {
	NextTokenOption option.String
}

// listObjects stubs the AWS S3 list call.
func listObjects(bucket, token string) ObjectPage { return ObjectPage{} }

// pageContainsKey stubs the per-page predicate used by the consume
// slot's Find call.
func pageContainsKey(p ObjectPage) bool { return false }

func ListPages(bucket string) {
	// __SNIPPET_define__

	// __SNIPPET_consume__
}
