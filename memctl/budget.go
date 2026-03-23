package memctl

// BudgetStrategy computes the number of workers a stage should have
// based on available memory and CPU cores. Implementations encapsulate
// the sizing formula — consumers plug in a strategy instead of
// inline math in Watch callbacks.
type BudgetStrategy interface {
	// ComputeWorkers returns the recommended worker count for the
	// given available memory (bytes) and CPU core count.
	// Must return >= 1.
	ComputeWorkers(availableMemory uint64, cpuCores int) int
}

// DefaultBudget computes workers from memory headroom using a simple
// linear model:
//
//	usable = availableMemory × BudgetFraction - BaseOverhead
//	workers = min(usable / PerWorkerCost, cpuCores × MaxCoresRatio)
//	workers = clamp(workers, 1, MaxWorkers)
type DefaultBudget struct {
	// BaseOverhead is memory reserved for non-worker use (bytes).
	// Subtracted before dividing by PerWorkerCost.
	BaseOverhead uint64

	// PerWorkerCost is the memory cost per worker (bytes).
	// Typically: model size + batch buffer + working set per worker.
	PerWorkerCost uint64

	// BudgetFraction is the fraction of available memory to use.
	// Must be in (0, 1]. The rest is reserved as safety buffer.
	BudgetFraction float64

	// MaxCoresRatio caps workers relative to CPU cores.
	// 0 means no CPU-based cap. Example: 2.0 = at most 2× cores.
	MaxCoresRatio float64

	// MaxWorkers is the absolute maximum. 0 means no cap.
	MaxWorkers int
}

// ComputeWorkers implements [BudgetStrategy].
func (b DefaultBudget) ComputeWorkers(availableMemory uint64, cpuCores int) int {
	if b.PerWorkerCost == 0 {
		return 1
	}
	if b.BudgetFraction <= 0 {
		return 1
	}

	usable := uint64(float64(availableMemory) * b.BudgetFraction)
	if usable <= b.BaseOverhead {
		return 1
	}
	usable -= b.BaseOverhead

	workers := int(usable / b.PerWorkerCost)
	if workers < 1 {
		workers = 1
	}

	// CPU cap.
	if b.MaxCoresRatio > 0 && cpuCores > 0 {
		cpuCap := int(float64(cpuCores) * b.MaxCoresRatio)
		if cpuCap < 1 {
			cpuCap = 1
		}
		if workers > cpuCap {
			workers = cpuCap
		}
	}

	// Absolute cap.
	if b.MaxWorkers > 0 && workers > b.MaxWorkers {
		workers = b.MaxWorkers
	}

	return workers
}
