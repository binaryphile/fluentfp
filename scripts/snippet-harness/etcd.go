//go:build ignore

// Package snippet is the verification harness for the etcd showcase
// entry in docs/showcase.md (etcd-io/etcd retry interceptor rewrite —
// labeled "fluentfp (conceptual)" in the showcase).
//
// The snippet is function-body code (three := declarations + return),
// so the harness wraps it in SetupResilientInvoke. The harness stubs
// the etcd client + call vocabulary (Client with shouldRefreshToken
// and refreshToken methods; CallOptions with lowercase `max` matching
// etcd's private field as shown in the snippet; InvokeRequest /
// InvokeResponse + invoker function value; isContextError /
// isSafeRetryError + retryBase).
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"context"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

// CallOptions stubs etcd's clientv3 per-call options. The lowercase
// `max` field matches what the showcase snippet references (and
// etcd's actual unexported field name).
type CallOptions struct {
	max int
}

// Client stubs etcd's *clientv3.Client; methods used by the snippet.
type Client struct{}

func (c Client) shouldRefreshToken(err error, opts CallOptions) bool { return false }
func (c Client) refreshToken()                                       {}

func isContextError(err error) bool                               { return false }
func isSafeRetryError(c Client, err error, opts CallOptions) bool { return true }

// InvokeRequest / InvokeResponse stub the gRPC request / response types.
type InvokeRequest struct{}
type InvokeResponse struct{}

const retryBase = time.Second

// invoker matches wrap.Func's expected signature: func(ctx, T) (R, error).
func invoker(ctx context.Context, req InvokeRequest) (InvokeResponse, error) {
	return InvokeResponse{}, nil
}

func SetupResilientInvoke(c Client, callOpts CallOptions) func(context.Context, InvokeRequest) (InvokeResponse, error) {
	// __SNIPPET__
}
