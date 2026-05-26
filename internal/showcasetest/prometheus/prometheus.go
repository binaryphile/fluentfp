// Package prometheus compile-checks the showcase entry for prometheus/prometheus init.
package prometheus

import (
	"time"

	"github.com/binaryphile/fluentfp/must"
)

// --- stubs for the prometheus packages referenced in the showcase ---

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

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func init() {
	prometheus.MustRegister(versioncollector.NewCollector(appName))

	defaultRetentionDuration = must.Get(model.ParseDuration(defaultRetentionString))
}
