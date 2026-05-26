// Package consul_retry compile-checks the showcase entry for hashicorp/consul session TTL retry.
package consul_retry

import (
	"context"
	"time"

	"github.com/binaryphile/fluentfp/wrap"
)

// --- stubs mirroring the consul types referenced by the showcase ---

type EnterpriseMeta struct{}

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

// Logger stubs the structured logger consul uses.
type Logger struct{}

func (Logger) Debug(msg string, kv ...any) {}
func (Logger) Error(msg string, kv ...any) {}

// Metrics stubs the metrics package consul uses.
type Metrics struct{}

func (Metrics) MeasureSince(key []string, t time.Time) {}

var metrics = Metrics{}

// SessionTimers stubs the per-session timer registry.
type SessionTimers struct{}

func (SessionTimers) Del(id string) {}

// ServerConfig stubs the per-server configuration that supplies the datacenter.
type ServerConfig struct{ Datacenter string }

// Server is the showcase entry's receiver type.
type Server struct {
	config             ServerConfig
	sessionTimers      SessionTimers
	logger             Logger
	resilientRaftApply func(context.Context, SessionRequest) (any, error)
}

func (s *Server) leaderRaftApply(ctx context.Context, args SessionRequest) (any, error) {
	return nil, nil
}

// NewServer wires the retry policy at construction (showcase: "defined once,
// applied at every call site").
func NewServer() *Server {
	s := &Server{}
	s.resilientRaftApply = wrap.Func(s.leaderRaftApply).
		Retry(maxInvalidateAttempts, wrap.ExpBackoff(invalidateRetryBase), nil)
	return s
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func (s *Server) invalidateSession(ctx context.Context, id string, entMeta *EnterpriseMeta) {
	defer metrics.MeasureSince([]string{"session_ttl", "invalidate"}, time.Now())

	s.sessionTimers.Del(id)

	args := SessionRequest{
		Datacenter: s.config.Datacenter,
		Op:         SessionDestroy,
		Session:    Session{ID: id},
	}
	if entMeta != nil {
		args.Session.EnterpriseMeta = *entMeta
	}

	_, err := s.resilientRaftApply(ctx, args)
	if err != nil {
		s.logger.Error("maximum revoke attempts reached for session", "error", id)
		return
	}
	s.logger.Debug("Session TTL expired", "session", id)
}
