// Package pagination compile-checks the showcase entry for Amazonka-style S3 pagination.
package pagination

import (
	"github.com/binaryphile/fluentfp/option"
	"github.com/binaryphile/fluentfp/stream"
)

// --- stubs for the S3 listing types ---

type ObjectPage struct {
	NextTokenOption option.String
}

func listObjects(bucket, token string) ObjectPage { return ObjectPage{} }

func pageContainsKey(p ObjectPage) bool { return false }

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func ListPages(bucket string) {
	// pageStep fetches one page and returns the optional next cursor.
	pageStep := func(token string) (ObjectPage, option.String) {
		page := listObjects(bucket, token)
		return page, page.NextTokenOption
	}

	var pages stream.Stream[ObjectPage] = stream.Paginate("", pageStep)

	pages.Collect()             // fetch everything
	pages.Take(3).Collect()     // first 3 pages only
	pages.Find(pageContainsKey) // stop at first match
}
