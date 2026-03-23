package memctl_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/memctl"
)

func TestDefaultBudget(t *testing.T) {
	tests := []struct {
		name    string
		budget  memctl.DefaultBudget
		memory  uint64
		cores   int
		want    int
	}{
		{
			name: "basic",
			budget: memctl.DefaultBudget{
				PerWorkerCost:  500 * 1024 * 1024, // 500MB per worker
				BudgetFraction: 0.5,
			},
			memory: 4 * 1024 * 1024 * 1024, // 4GB available
			cores:  8,
			want:   4, // 4GB × 0.5 / 500MB = 4
		},
		{
			name: "with_overhead",
			budget: memctl.DefaultBudget{
				BaseOverhead:   1024 * 1024 * 1024, // 1GB overhead
				PerWorkerCost:  500 * 1024 * 1024,
				BudgetFraction: 0.5,
			},
			memory: 4 * 1024 * 1024 * 1024,
			cores:  8,
			want:   2, // (4GB × 0.5 - 1GB) / 500MB = 1GB / 500MB = 2
		},
		{
			name: "cpu_cap",
			budget: memctl.DefaultBudget{
				PerWorkerCost:  100 * 1024 * 1024, // 100MB per worker
				BudgetFraction: 1.0,
				MaxCoresRatio:  1.0,
			},
			memory: 10 * 1024 * 1024 * 1024, // 10GB
			cores:  4,
			want:   4, // memory says 100, CPU cap says 4
		},
		{
			name: "max_workers_cap",
			budget: memctl.DefaultBudget{
				PerWorkerCost:  100 * 1024 * 1024,
				BudgetFraction: 1.0,
				MaxWorkers:     6,
			},
			memory: 10 * 1024 * 1024 * 1024,
			cores:  100,
			want:   6,
		},
		{
			name: "zero_memory",
			budget: memctl.DefaultBudget{
				PerWorkerCost:  500 * 1024 * 1024,
				BudgetFraction: 0.5,
			},
			memory: 0,
			cores:  8,
			want:   1, // floor
		},
		{
			name: "zero_per_worker_cost",
			budget: memctl.DefaultBudget{
				PerWorkerCost:  0,
				BudgetFraction: 0.5,
			},
			memory: 4 * 1024 * 1024 * 1024,
			cores:  8,
			want:   1, // safety floor
		},
		{
			name: "overhead_exceeds_budget",
			budget: memctl.DefaultBudget{
				BaseOverhead:   10 * 1024 * 1024 * 1024,
				PerWorkerCost:  500 * 1024 * 1024,
				BudgetFraction: 0.5,
			},
			memory: 4 * 1024 * 1024 * 1024,
			cores:  8,
			want:   1, // floor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.budget.ComputeWorkers(tt.memory, tt.cores)
			if got != tt.want {
				t.Errorf("ComputeWorkers(%d, %d) = %d, want %d",
					tt.memory, tt.cores, got, tt.want)
			}
		})
	}
}

func TestBudgetStrategyInterface(t *testing.T) {
	// Verify DefaultBudget implements BudgetStrategy.
	var _ memctl.BudgetStrategy = memctl.DefaultBudget{}
}
