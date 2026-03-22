// Package main demonstrates why identifying the correct drum (constraint)
// matters in a pipeline. Three scenarios process the same work through
// parse → transform → store, where transform is 10× slower.
//
// The demo tracks resource consumption over the pipeline's lifetime:
// total items in flight (aggregate WIP), estimated memory footprint,
// and per-stage queue depth.
//
// Run:
//
//	go run ./examples/drum-demo/
package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

const (
	totalItems = 200
	parseTime  = 2 * time.Millisecond  // fast
	xformTime  = 20 * time.Millisecond // bottleneck — 10× slower
	storeTime  = 2 * time.Millisecond  // fast
	tickRate   = 200 * time.Millisecond
	colWidth   = 84

	// Simulated per-item memory cost (e.g. an embedding vector or image chunk).
	itemWeightKB = 64
)

// ── ANSI ────────────────────────────────────────────────────────────────

const (
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

// snapshot captures one tick's resource state.
type snapshot struct {
	elapsed  time.Duration
	wip      int64
	memoryKB int64
	tput     float64
	done     int64
	xformQ   int64
}

// scenarioResult holds the full timeline + summary for one run.
type scenarioResult struct {
	name       string
	timeline   []snapshot
	elapsed    time.Duration
	throughput float64
	peakWIP    int64
	peakMemKB  int64
	peakXformQ int64
	wipSeconds float64 // area under the WIP curve
}

type stageOpts struct {
	parse toc.Options[int]
	xform toc.Options[int]
	store toc.Options[int]
}

func main() {
	fmt.Println(divider("DRUM-BUFFER-ROPE: RESOURCE COST DEMO"))
	fmt.Println()
	fmt.Println("Pipeline:  parse (2ms, 4 workers)")
	fmt.Println("         → transform (20ms, 1 worker)  ← constraint")
	fmt.Println("         → store (2ms, 2 workers)")
	fmt.Println()
	fmt.Printf("Each item weighs ~%dKB in memory. Sending %d items.\n", itemWeightKB, totalItems)
	fmt.Println("Watch what happens to throughput, memory, and queue depth")
	fmt.Println("depending on " + bold + "where" + reset + " the WIP limit goes.")

	type scenario struct {
		name string
		desc string
		opts stageOpts
	}

	defaultStore := toc.Options[int]{Capacity: 10, Workers: 2}

	scenarios := []scenario{
		{
			name: "No drum",
			desc: "No WIP limits. Parse floods freely. All 200 items rush into the pipeline.",
			opts: stageOpts{
				parse: toc.Options[int]{Capacity: 200, Workers: 4},
				xform: toc.Options[int]{Capacity: 200, Workers: 1},
				store: defaultStore,
			},
		},
		{
			name: "Limit on wrong stage",
			desc: "WIP limit=8 on parse. Slows entry — helps some, but transform still floods.",
			opts: stageOpts{
				parse: toc.Options[int]{Capacity: 10, Workers: 4, MaxWIP: 8},
				xform: toc.Options[int]{Capacity: 100, Workers: 1},
				store: defaultStore,
			},
		},
		{
			name: "Correct drum (transform)",
			desc: "WIP limit=3 on transform. Pipeline holds only what the drum can eat.",
			opts: stageOpts{
				parse: toc.Options[int]{Capacity: 10, Workers: 4},
				xform: toc.Options[int]{Capacity: 4, Workers: 1, MaxWIP: 3},
				store: defaultStore,
			},
		},
	}

	results := make([]scenarioResult, len(scenarios))

	for i, sc := range scenarios {
		fmt.Println()
		fmt.Println(divider(fmt.Sprintf("SCENARIO %d: %s", i+1, sc.name)))
		fmt.Println(dim + sc.desc + reset)
		fmt.Println()
		results[i] = runScenario(sc.name, sc.opts)
	}

	fmt.Println()
	printComparison(results)
}

func runScenario(name string, opts stageOpts) scenarioResult {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parseFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(parseTime)
		return item, nil
	}
	xformFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(xformTime)
		return item * 10, nil
	}
	storeFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(storeTime)
		return item, nil
	}

	parse := toc.Start[int, int](ctx, parseFn, opts.parse)
	xform := toc.Pipe[int, int](ctx, parse.Out(), xformFn, opts.xform)
	store := toc.Pipe[int, int](ctx, xform.Out(), storeFn, opts.store)

	var submitted atomic.Int64
	go func() {
		for i := range totalItems {
			if err := parse.Submit(ctx, i); err != nil {
				break
			}
			submitted.Add(1)
		}
		parse.CloseInput()
	}()

	var completed atomic.Int64
	var drainWg sync.WaitGroup
	drainWg.Add(1)
	go func() {
		defer drainWg.Done()
		for range store.Out() {
			completed.Add(1)
		}
	}()

	// Table header.
	fmt.Printf("  %-5s │ %4s %4s │ %-20s │ %5s │ %5s │ %-16s\n",
		"time", "fed", "done", "pipeline WIP", "mem", "t/s", "xform queue")
	fmt.Println("  " + strings.Repeat("─", colWidth-2))

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	start := time.Now()
	prevCompleted := int64(0)
	var timeline []snapshot
	var peakWIP, peakXformQ int64
	var wipSeconds float64
	prevWIP := int64(0)

	for {
		<-ticker.C

		xs := xform.Stats()
		elapsed := time.Since(start)

		fed := submitted.Load()
		done := completed.Load()
		pipelineWIP := fed - done
		if pipelineWIP < 0 {
			pipelineWIP = 0
		}

		memKB := pipelineWIP * itemWeightKB
		tput := float64(done-prevCompleted) / tickRate.Seconds()
		prevCompleted = done

		wipSeconds += float64(prevWIP+pipelineWIP) / 2.0 * tickRate.Seconds()
		prevWIP = pipelineWIP

		if pipelineWIP > peakWIP {
			peakWIP = pipelineWIP
		}
		if xs.BufferedDepth > peakXformQ {
			peakXformQ = xs.BufferedDepth
		}

		snap := snapshot{
			elapsed:  elapsed,
			wip:      pipelineWIP,
			memoryKB: memKB,
			tput:     tput,
			done:     done,
			xformQ:   xs.BufferedDepth,
		}
		timeline = append(timeline, snap)

		wipBar := resourceBar(pipelineWIP, int64(totalItems), 12)
		xfBar := fmtQ(xs.BufferedDepth, int64(xs.QueueCapacity))

		fmt.Printf("  %-5s │ %4d %4d │ %s %3d │ %4dM │ %5.0f │ %s\n",
			fmtDur(elapsed),
			fed, done,
			wipBar, pipelineWIP,
			memKB/1024,
			tput,
			xfBar,
		)

		if done >= totalItems {
			break
		}
	}

	drainWg.Wait()
	store.Wait()
	elapsed := time.Since(start)
	throughput := float64(totalItems) / elapsed.Seconds()
	peakMemKB := peakWIP * itemWeightKB

	fmt.Println("  " + strings.Repeat("─", colWidth-2))
	printScenarioSummary(elapsed, throughput, peakWIP, peakMemKB, peakXformQ, wipSeconds, timeline)

	return scenarioResult{
		name:       name,
		timeline:   timeline,
		elapsed:    elapsed,
		throughput: throughput,
		peakWIP:    peakWIP,
		peakMemKB:  peakMemKB,
		peakXformQ: peakXformQ,
		wipSeconds: wipSeconds,
	}
}

func printScenarioSummary(elapsed time.Duration, throughput float64, peakWIP, peakMemKB, peakXformQ int64, wipSeconds float64, timeline []snapshot) {
	fmt.Printf("  Throughput:  %.0f/s   Time: %s\n", throughput, elapsed.Round(time.Millisecond))
	fmt.Printf("  Peak WIP:   %d items (%dMB)  ", peakWIP, peakMemKB/1024)

	switch {
	case peakWIP > 100:
		fmt.Printf("%s● HEAVY — most items parked in memory as dead weight%s\n", red, reset)
	case peakWIP > 30:
		fmt.Printf("%s● ELEVATED%s\n", yellow, reset)
	default:
		fmt.Printf("%s● LIGHT — minimal memory footprint%s\n", green, reset)
	}

	fmt.Printf("  WIP·time:   %.0f item-seconds", wipSeconds)
	fmt.Printf("  %s(total resource cost held over pipeline lifetime)%s\n", dim, reset)

	// Characterize the trajectory.
	if len(timeline) >= 3 {
		firstWIP := timeline[0].wip
		midIdx := len(timeline) / 2
		midWIP := timeline[midIdx].wip

		fmt.Print("  Shape:      ")
		if firstWIP > int64(totalItems/2) {
			fmt.Printf("%sspike-then-drain%s — front-loaded: %d items at %.1fs, drains to %d\n",
				red, reset, firstWIP, timeline[0].elapsed.Seconds(), timeline[len(timeline)-1].wip)
		} else if midWIP > 2*firstWIP && midWIP > 2*timeline[len(timeline)-1].wip {
			fmt.Printf("%shump%s — grows to %d at midpoint, then drains\n",
				yellow, reset, midWIP)
		} else {
			maxWIP := int64(0)
			minWIP := int64(math.MaxInt64)
			for _, s := range timeline[:len(timeline)-1] {
				if s.wip > maxWIP {
					maxWIP = s.wip
				}
				if s.wip < minWIP {
					minWIP = s.wip
				}
			}
			spread := maxWIP - minWIP
			if spread <= 15 {
				fmt.Printf("%ssteady-state%s — flat at ~%d items throughout (±%d)\n",
					green, reset, (maxWIP+minWIP)/2, spread)
			} else {
				fmt.Printf("variable — ranges %d to %d\n", minWIP, maxWIP)
			}
		}
	}
}

func printComparison(results []scenarioResult) {
	fmt.Println(divider("RESOURCE COMPARISON"))
	fmt.Println()

	fmt.Printf("  %-28s │ %5s │ %8s │ %8s │ %8s │ %12s │ %s\n",
		"Scenario", "t/s", "peak WIP", "peak mem", "xform q", "WIP·seconds", "shape")
	fmt.Println("  " + strings.Repeat("─", colWidth-2))

	for _, r := range results {
		shape := characterize(r.timeline)
		memStr := fmt.Sprintf("%dMB", r.peakMemKB/1024)
		wipSec := fmt.Sprintf("%.0f", r.wipSeconds)

		var wipColor string
		switch {
		case r.peakWIP > 100:
			wipColor = red
		case r.peakWIP > 30:
			wipColor = yellow
		default:
			wipColor = green
		}

		var xfColor string
		switch {
		case r.peakXformQ > 100:
			xfColor = red
		case r.peakXformQ > 10:
			xfColor = yellow
		default:
			xfColor = green
		}

		fmt.Printf("  %-28s │ %5.0f │ %s%8d%s │ %8s │ %s%8d%s │ %12s │ %s\n",
			r.name, r.throughput,
			wipColor, r.peakWIP, reset,
			memStr,
			xfColor, r.peakXformQ, reset,
			wipSec, shape)
	}

	fmt.Println()

	fmt.Println("  " + bold + "WIP over time" + reset + " (each char = one tick, height = items in pipeline):")
	fmt.Println()

	for _, r := range results {
		spark := wipSparkline(r.timeline, int64(totalItems))
		fmt.Printf("  %-14s %s\n", r.name+":", spark)
	}

	fmt.Println()
	fmt.Println(bold + "  The rope must reference the drum." + reset)
	fmt.Println("  Scenario 2 limits parse — it helps some (6MB vs 11MB) but the limit")
	fmt.Println("  doesn't know about the constraint. Items still flood transform's queue.")
	fmt.Println("  Scenario 3 limits at the constraint and backpressures the entry point.")
	fmt.Println("  Same throughput, 1MB peak, steady-state. That's what the rope does.")

	if len(results) == 3 && results[2].wipSeconds > 0 {
		memRatio := float64(results[0].peakMemKB) / float64(results[2].peakMemKB)
		costRatio := results[0].wipSeconds / results[2].wipSeconds
		fmt.Printf("\n  Scenario 1 vs 3: %.0f× peak memory, %.0f× total resource cost — same work done.\n",
			memRatio, costRatio)

		if results[1].throughput < results[2].throughput*0.95 {
			tputPct := (1.0 - results[1].throughput/results[2].throughput) * 100
			fmt.Printf("  Scenario 2 vs 3: %.0f%% throughput loss from throttling the wrong stage.\n", tputPct)
		}
	}
}

func characterize(timeline []snapshot) string {
	if len(timeline) < 3 {
		return "unknown"
	}
	firstWIP := timeline[0].wip

	if firstWIP > int64(totalItems/2) {
		return red + "spike-drain" + reset
	}

	maxWIP := int64(0)
	minWIP := int64(math.MaxInt64)
	for _, s := range timeline[:len(timeline)-1] {
		if s.wip > maxWIP {
			maxWIP = s.wip
		}
		if s.wip < minWIP {
			minWIP = s.wip
		}
	}
	if maxWIP-minWIP <= 15 {
		return green + "steady-state" + reset
	}
	return yellow + "variable" + reset
}

func wipSparkline(timeline []snapshot, maxPossible int64) string {
	if maxPossible <= 0 {
		maxPossible = 1
	}

	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var b strings.Builder

	for _, s := range timeline {
		pct := float64(s.wip) / float64(maxPossible)
		if pct < 0 {
			pct = 0
		}
		if pct > 1 {
			pct = 1
		}

		idx := int(pct * float64(len(bars)-1))
		if idx >= len(bars) {
			idx = len(bars) - 1
		}

		var color string
		switch {
		case pct >= 0.5:
			color = red
		case pct >= 0.15:
			color = yellow
		default:
			color = green
		}

		b.WriteString(color)
		b.WriteRune(bars[idx])
		b.WriteString(reset)
	}

	return b.String()
}

// ── Formatting helpers ──────────────────────────────────────────────────

func resourceBar(current, max int64, width int) string {
	if max <= 0 {
		max = 1
	}
	pct := float64(current) / float64(max)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(math.Round(pct * float64(width)))
	if filled > width {
		filled = width
	}
	empty := width - filled

	var color string
	switch {
	case pct >= 0.5:
		color = red
	case pct >= 0.15:
		color = yellow
	default:
		color = green
	}

	return color + strings.Repeat("█", filled) + strings.Repeat("░", empty) + reset
}

func fmtQ(depth, capacity int64) string {
	if capacity <= 0 {
		capacity = 1
	}
	pct := float64(depth) / float64(capacity)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	const barWidth = 6
	filled := int(math.Round(pct * barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	var color string
	switch {
	case pct >= 0.66:
		color = red
	case pct >= 0.33:
		color = yellow
	default:
		color = green
	}

	return color + strings.Repeat("█", filled) + strings.Repeat("░", empty) + reset + fmt.Sprintf(" %3d", depth)
}

func divider(text string) string {
	pad := colWidth - len(text) - 4
	if pad < 0 {
		pad = 0
	}
	left := pad / 2
	right := pad - left
	return "══" + strings.Repeat("═", left) + " " + text + " " + strings.Repeat("═", right) + "══"
}

func fmtDur(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}
