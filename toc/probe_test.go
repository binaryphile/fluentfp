package toc

import (
	"runtime/metrics"
	"testing"
)

func TestAllocMetricsProbe(t *testing.T) {
	tests := []struct {
		name string
		desc []metrics.Description
		want bool
	}{
		{
			name: "both present with KindUint64",
			desc: []metrics.Description{
				{Name: metricAllocBytes, Kind: metrics.KindUint64},
				{Name: metricAllocObjects, Kind: metrics.KindUint64},
			},
			want: true,
		},
		{
			name: "bytes missing",
			desc: []metrics.Description{
				{Name: metricAllocObjects, Kind: metrics.KindUint64},
			},
			want: false,
		},
		{
			name: "objects missing",
			desc: []metrics.Description{
				{Name: metricAllocBytes, Kind: metrics.KindUint64},
			},
			want: false,
		},
		{
			name: "both missing",
			desc: nil,
			want: false,
		},
		{
			name: "bytes wrong kind",
			desc: []metrics.Description{
				{Name: metricAllocBytes, Kind: metrics.KindFloat64},
				{Name: metricAllocObjects, Kind: metrics.KindUint64},
			},
			want: false,
		},
		{
			name: "objects wrong kind",
			desc: []metrics.Description{
				{Name: metricAllocBytes, Kind: metrics.KindUint64},
				{Name: metricAllocObjects, Kind: metrics.KindFloat64},
			},
			want: false,
		},
		{
			name: "duplicate bytes replaces missing objects",
			desc: []metrics.Description{
				{Name: metricAllocBytes, Kind: metrics.KindUint64},
				{Name: metricAllocBytes, Kind: metrics.KindUint64},
			},
			want: false,
		},
		{
			name: "duplicate objects replaces missing bytes",
			desc: []metrics.Description{
				{Name: metricAllocObjects, Kind: metrics.KindUint64},
				{Name: metricAllocObjects, Kind: metrics.KindUint64},
			},
			want: false,
		},
		{
			name: "unrelated metrics only",
			desc: []metrics.Description{
				{Name: "/gc/heap/frees:bytes", Kind: metrics.KindUint64},
				{Name: "/sched/goroutines:goroutines", Kind: metrics.KindUint64},
			},
			want: false,
		},
		{
			name: "both present among unrelated",
			desc: []metrics.Description{
				{Name: "/gc/heap/frees:bytes", Kind: metrics.KindUint64},
				{Name: metricAllocBytes, Kind: metrics.KindUint64},
				{Name: "/sched/goroutines:goroutines", Kind: metrics.KindUint64},
				{Name: metricAllocObjects, Kind: metrics.KindUint64},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := allocMetricsProbe(tt.desc)
			if got != tt.want {
				t.Errorf("allocMetricsProbe() = %v, want %v", got, tt.want)
			}
		})
	}
}
