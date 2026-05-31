//go:build ignore

// Package snippet is the verification harness for the prometheus
// showcase entry in docs/showcase.md (prometheus/prometheus cmd init
// rewrite). The snippet is a top-level `func init()` declaration; the
// harness provides the package-level vars and stubs the prometheus +
// versioncollector + model packages it references.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"time"

	"github.com/binaryphile/fluentfp/must"
)

// Collector stubs prometheus.Collector.
type Collector struct{}

// MetricsClient stubs the prometheus metrics registry.
type MetricsClient struct{}

func (MetricsClient) MustRegister(c Collector) {}

// VersionCollectorClient stubs the versioncollector subpackage.
type VersionCollectorClient struct{}

func (VersionCollectorClient) NewCollector(name string) Collector { return Collector{} }

// ModelClient stubs the model package's ParseDuration entry point.
type ModelClient struct{}

func (ModelClient) ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration("360h") // any non-zero stub
}

var (
	prometheus               = MetricsClient{}
	versioncollector         = VersionCollectorClient{}
	model                    = ModelClient{}
	defaultRetentionDuration time.Duration
	defaultRetentionString   = "15d"
	appName                  = "prometheus"
)

// __SNIPPET__
