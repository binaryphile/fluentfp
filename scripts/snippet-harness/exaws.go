//go:build ignore

// Package snippet is the verification harness for the exaws showcase
// entry in docs/showcase.md (ExAws S3 / Hex multipart upload patterns).
// The showcase has THREE fluentfp fences demonstrating the FanOut
// idiom spectrum, each with a different return-shape contract:
//
//   slot=upload_all → UploadAll returns ([]ChunkUpload, error)
//   slot=fetch_both → FetchBoth returns ([]FetchedDep, []error)
//   slot=fetch_ok   → FetchOk   returns []FetchedDep
//
// Each slot is wrapped in its own function shell so the per-fence
// return types can be expressed precisely.
//
// The harness is excluded from default `go build ./...` via
// `//go:build ignore`; scripts/check-snippets.py strips the
// constraint when assembling into the tmpdir.
package snippet

import (
	"context"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
)

// Chunk + ChunkUpload stub the per-item upload types.
type Chunk struct{}
type ChunkUpload struct{}

// Dep + FetchedDep stub the per-item dependency-fetch types.
type Dep struct{}
type FetchedDep struct{}

func uploadChunk(ctx context.Context, c Chunk) (ChunkUpload, error) {
	return ChunkUpload{}, nil
}

func fetchDep(ctx context.Context, d Dep) (FetchedDep, error) {
	return FetchedDep{}, nil
}

func UploadAll(ctx context.Context, chunks []Chunk) ([]ChunkUpload, error) {
	// __SNIPPET_upload_all__
}

func FetchBoth(ctx context.Context, deps []Dep) ([]FetchedDep, []error) {
	// __SNIPPET_fetch_both__
}

func FetchOk(ctx context.Context, deps []Dep) []FetchedDep {
	// __SNIPPET_fetch_ok__
}

// Keep rslt referenced so the un-substituted harness parses with imports
// used (otherwise gopls/tools flag unused — fine for build ignore, but
// noisy). The fetch_both and fetch_ok snippets exercise rslt at build time.
var _ = rslt.CollectOk[FetchedDep]
var _ = slice.FanOutAll[Chunk, ChunkUpload]
