// Package etcd compile-checks the showcase entry for etcd-io retry interceptor.
// The showcase labels this entry "fluentfp (conceptual)" — compile-check
// verifies the Func → OnError → Retry chain shape is valid.
package etcd

import (
	"context"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

// --- stubs for the etcd client and call types ---

// CallOptions stubs the per-call options.
type CallOptions struct {
	Max int
}

// Client stubs etcd's *clientv3.Client.
type Client struct{}

func (c Client) shouldRefreshToken(err error, opts CallOptions) bool { return false }
func (c Client) refreshToken()                                       {}

func isContextError(err error) bool                                { return false }
func isSafeRetryError(c Client, err error, opts CallOptions) bool  { return true }

// InvokeRequest and InvokeResponse stub the gRPC request/response types.
type InvokeRequest struct{}
type InvokeResponse struct{}

const retryBase = time.Second

// invoker matches wrap.Func's expected signature: func(ctx, T) (R, error).
func invoker(ctx context.Context, req InvokeRequest) (InvokeResponse, error) {
	return InvokeResponse{}, nil
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func SetupResilientInvoke(c Client, callOpts CallOptions) func(context.Context, InvokeRequest) (InvokeResponse, error) {
	// isSafeRetry returns true for errors safe to retry.
	isSafeRetry := func(err error) bool {
		return !isContextError(err) && isSafeRetryError(c, err, callOpts)
	}

	// refreshOnAuthErr refreshes the token only for authentication errors.
	refreshOnAuthErr := func(err error) {
		if c.shouldRefreshToken(err, callOpts) {
			c.refreshToken()
		}
	}

	resilientInvoke := wrap.Func(invoker).
		OnError(refreshOnAuthErr).
		Retry(callOpts.Max, wrap.ExpBackoff(retryBase), isSafeRetry)
	return resilientInvoke
}
