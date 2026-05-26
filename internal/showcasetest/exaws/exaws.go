// Package exaws compile-checks the showcase entry for ExAws S3 / Hex multipart upload.
package exaws

import (
	"context"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the upload types ---

type Chunk struct{}
type ChunkUpload struct{}
type Dep struct{}
type FetchedDep struct{}

func uploadChunk(ctx context.Context, c Chunk) (ChunkUpload, error) {
	return ChunkUpload{}, nil
}

func fetchDep(ctx context.Context, d Dep) (FetchedDep, error) {
	return FetchedDep{}, nil
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func UploadAll(ctx context.Context, chunks []Chunk) ([]ChunkUpload, error) {
	uploads, err := slice.FanOutAll(ctx, 4, chunks, uploadChunk)
	return uploads, err
}

func FetchBothHalves(ctx context.Context, deps []Dep) ([]FetchedDep, []error) {
	downloaded, errs := rslt.CollectOkAndErr(slice.FanOut(ctx, 8, deps, fetchDep))
	return downloaded, errs
}

func FetchSuccessesOnly(ctx context.Context, deps []Dep) []FetchedDep {
	downloaded := rslt.CollectOk(slice.FanOut(ctx, 8, deps, fetchDep))
	return downloaded
}
