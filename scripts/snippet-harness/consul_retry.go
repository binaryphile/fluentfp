//go:build ignore

// Package snippet is the verification harness for the consul_retry
// showcase entry in docs/showcase.md (hashicorp/consul session_ttl
// invalidate rewrite). The showcase has TWO fluentfp fences — the
// construction-time wiring (`s.resilientRaftApply = wrap.Func(...)`)
// and the call-site method (`func (s *Server) invalidateSession`).
// Multi-slot semantics let both substitute into one assembled file:
//
//   slot=construct → // __SNIPPET_construct__  (inside NewServer body)
//   slot=call      → // __SNIPPET_call__       (at package level)
//
// Local stub names stand in for structs.Session* / acl.EnterpriseMeta
// (same pattern as other migrated entries).
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

// EnterpriseMeta stubs acl.EnterpriseMeta.
type EnterpriseMeta struct{}

// Session + SessionRequest stub structs.* — the showcase Session has
// an EnterpriseMeta field that the call-site assigns into.
type Session struct {
	ID             string
	EnterpriseMeta EnterpriseMeta
}

type SessionRequest struct {
	Datacenter string
	Op         int
	Session    Session
}

const (
	SessionDestroy        = 1
	maxInvalidateAttempts = 3
	invalidateRetryBase   = time.Second
)

// Logger stubs the structured logger.
type Logger struct{}

func (Logger) Debug(msg string, kv ...any) {}
func (Logger) Error(msg string, kv ...any) {}

// Metrics stubs the metrics package; MeasureSince is exercised by the
// call-site snippet via defer.
type Metrics struct{}

func (Metrics) MeasureSince(key []string, t time.Time) {}

var metrics = Metrics{}

// SessionTimers stubs the per-session timer registry.
type SessionTimers struct{}

func (SessionTimers) Del(id string) {}

// ServerConfig stubs the per-server config carrying the datacenter.
type ServerConfig struct{ Datacenter string }

// Server is the receiver type for invalidateSession. The
// resilientRaftApply field is populated by the construct snippet and
// read by the call snippet — wiring across slots is the point of the
// multi-fence presentation.
type Server struct {
	config             ServerConfig
	sessionTimers      SessionTimers
	logger             Logger
	resilientRaftApply func(context.Context, SessionRequest) (any, error)
}

func (s *Server) leaderRaftApply(ctx context.Context, args SessionRequest) (any, error) {
	return nil, nil
}

// NewServer is the construction shell whose body holds the construct
// slot. The showcase prose introduces the construct fence with "At
// server construction"; this scaffold makes that concrete.
func NewServer() *Server {
	s := &Server{}
	// __SNIPPET_construct__
	return s
}

// __SNIPPET_call__
