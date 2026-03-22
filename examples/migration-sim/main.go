// Package main simulates constraint migration using real toc stages
// to test whether tearing down the old rope causes the old constraint
// to dominate again.
//
// Three scenarios with real backpressure:
//  1. Baseline: embed is drum, rope active, no migration
//  2. Elevate embed, migrate rope to walk — keep embed workers
//  3. Same but steal embed worker at migration (the dangerous case)
//
// Run:
//
//	go run ./examples/migration-sim/
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

const (
	totalItems = 300
	tickRate   = 200 * time.Millisecond
	colWidth   = 80

	headTime  = 2 * time.Millisecond
	walkTime  = 5 * time.Millisecond
	embedTime = 20 * time.Millisecond
	storeTime = 2 * time.Millisecond
)

func main() {
	fmt.Println(divider("CONSTRAINT MIGRATION SIMULATION"))
	fmt.Println()
	fmt.Println("Pipeline: head(2ms,4w) → walk(5ms) → embed(20ms) → store(2ms,2w)")
	fmt.Println("Embed is 4× slower than walk — it's the initial constraint.")
	fmt.Println()
	fmt.Println("Question: when we elevate embed and migrate the rope to walk,")
	fmt.Println("does embed become the constraint again?")

	fmt.Println()
	fmt.Println(divider("SCENARIO 1: Baseline — no elevation, no migration"))
	fmt.Println("embed=1w (50/s), walk=1w (200/s). Rope on embed.")
	fmt.Println()
	runScenario(scenario{
		name:         "baseline",
		walkWorkers:  1,
		embedWorkers: 1,
		embedMaxWIP:  3,
	})

	fmt.Println()
	fmt.Println(divider("SCENARIO 2: Elevate embed, migrate rope, KEEP workers"))
	fmt.Println("Tick 0-9: embed=1w. Tick 10: embed=3w (150/s > walk 200/s).")
	fmt.Println("Tick 15: migrate rope to walk. Embed keeps 3 workers.")
	fmt.Println()
	runScenario(scenario{
		name:              "elevate+migrate (keep workers)",
		walkWorkers:       1,
		embedWorkers:      1,
		embedMaxWIP:       3,
		elevateEmbedAt:    10,
		elevateEmbedTo:    3,
		migrateToWalkAt:   15,
		walkMaxWIPAfter:   3,
	})

	fmt.Println()
	fmt.Println(divider("SCENARIO 3: Elevate embed, migrate rope, STEAL worker"))
	fmt.Println("Same as 2, but at tick 15 also move 1 worker embed→walk.")
	fmt.Println("embed drops to 2w (100/s), walk rises to 2w (400/s).")
	fmt.Println("This is the dangerous case.")
	fmt.Println()
	runScenario(scenario{
		name:              "elevate+migrate (steal worker)",
		walkWorkers:       1,
		embedWorkers:      1,
		embedMaxWIP:       3,
		elevateEmbedAt:    10,
		elevateEmbedTo:    3,
		migrateToWalkAt:   15,
		walkMaxWIPAfter:   3,
		stealWorkerAtMigr: true,
	})
}

type scenario struct {
	name              string
	walkWorkers       int
	embedWorkers      int
	embedMaxWIP       int
	elevateEmbedAt    int // tick to add embed workers (-1 = never)
	elevateEmbedTo    int
	migrateToWalkAt   int // tick to move MaxWIP from embed to walk (-1 = never)
	walkMaxWIPAfter   int
	stealWorkerAtMigr bool
}

type tickSnapshot struct {
	elapsed    time.Duration
	fed        int64
	done       int64
	walkQ      int64
	embedQ     int64
	walkW      int
	embedW     int
	embedMaxW  int
	throughput float64
	drum       string
}

func runScenario(sc scenario) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	headFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(headTime)
		return item, nil
	}
	walkFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(walkTime)
		return item, nil
	}
	embedFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(embedTime)
		return item * 10, nil
	}
	storeFn := func(_ context.Context, item int) (int, error) {
		time.Sleep(storeTime)
		return item, nil
	}

	head := toc.Start[int, int](ctx, headFn, toc.Options[int]{
		Capacity: 10, Workers: 4,
	})
	walk := toc.Pipe[int, int](ctx, head.Out(), walkFn, toc.Options[int]{
		Capacity: 10, Workers: sc.walkWorkers,
	})
	embed := toc.Pipe[int, int](ctx, walk.Out(), embedFn, toc.Options[int]{
		Capacity: 10, Workers: sc.embedWorkers, MaxWIP: sc.embedMaxWIP,
	})
	store := toc.Pipe[int, int](ctx, embed.Out(), storeFn, toc.Options[int]{
		Capacity: 10, Workers: 2,
	})

	// Feed items.
	var submitted atomic.Int64
	go func() {
		for i := range totalItems {
			if err := head.Submit(ctx, i); err != nil {
				break
			}
			submitted.Add(1)
		}
		head.CloseInput()
	}()

	// Drain.
	var completed atomic.Int64
	var drainWg sync.WaitGroup
	drainWg.Add(1)
	go func() {
		defer drainWg.Done()
		for range store.Out() {
			completed.Add(1)
		}
	}()

	// Print header.
	fmt.Printf("  %-5s │ %4s %4s │ %-10s │ %-10s │ %-10s │ %5s │ %s\n",
		"time", "fed", "done", "walk", "embed", "embed WIP", "t/s", "drum")
	fmt.Println("  " + strings.Repeat("─", colWidth-2))

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	start := time.Now()
	prevDone := int64(0)
	tick := 0
	ropeOn := "embed"

	var snapshots []tickSnapshot

	for {
		<-ticker.C

		// Events.
		if tick == sc.elevateEmbedAt && sc.elevateEmbedTo > 0 {
			embed.SetWorkers(sc.elevateEmbedTo)
		}
		if tick == sc.migrateToWalkAt && sc.walkMaxWIPAfter > 0 {
			// Remove MaxWIP on embed (open it up).
			embed.SetMaxWIP(sc.embedMaxWIP + 10) // loosen
			// Set MaxWIP on walk (new rope target).
			walk.SetMaxWIP(sc.walkMaxWIPAfter)
			ropeOn = "walk"

			if sc.stealWorkerAtMigr {
				embed.SetWorkers(sc.elevateEmbedTo - 1)
				walk.SetWorkers(sc.walkWorkers + 1)
			}
		}

		elapsed := time.Since(start)
		fed := submitted.Load()
		done := completed.Load()
		ws := walk.Stats()
		es := embed.Stats()
		tput := float64(done-prevDone) / tickRate.Seconds()
		prevDone = done

		// Identify drum by queue depth (crude but visible).
		drum := ropeOn
		if ws.BufferedDepth > es.BufferedDepth+5 {
			drum = "walk"
		} else if es.BufferedDepth > ws.BufferedDepth+5 {
			drum = "embed"
		}

		snap := tickSnapshot{
			elapsed:   elapsed,
			fed:       fed,
			done:      done,
			walkQ:     ws.BufferedDepth,
			embedQ:    es.BufferedDepth,
			walkW:     ws.ActiveWorkers,
			embedW:    es.ActiveWorkers,
			embedMaxW: es.MaxWIP,
			throughput: tput,
			drum:      drum,
		}
		snapshots = append(snapshots, snap)

		event := ""
		if tick == sc.elevateEmbedAt && sc.elevateEmbedTo > 0 {
			event = " ← ELEVATE"
		}
		if tick == sc.migrateToWalkAt && sc.walkMaxWIPAfter > 0 {
			event = " ← MIGRATE"
			if sc.stealWorkerAtMigr {
				event += "+STEAL"
			}
		}

		fmt.Printf("  %-5s │ %4d %4d │ q=%-3d w=%-2d │ q=%-3d w=%-2d │ wip=%-2d/%-2d  │ %5.0f │ %s%s\n",
			fmtDur(elapsed),
			fed, done,
			ws.BufferedDepth, ws.ActiveWorkers,
			es.BufferedDepth, es.ActiveWorkers,
			es.Admitted, es.MaxWIP,
			tput,
			drum, event)

		tick++
		if done >= int64(totalItems) {
			break
		}
		if tick > 80 {
			break // safety
		}
	}

	drainWg.Wait()
	store.Wait()
	elapsed := time.Since(start)

	fmt.Println("  " + strings.Repeat("─", colWidth-2))
	fmt.Printf("  Time: %s  Throughput: %.0f/s  walk q peak: %d  embed q peak: %d\n",
		elapsed.Round(time.Millisecond),
		float64(totalItems)/elapsed.Seconds(),
		maxQ(snapshots, func(s tickSnapshot) int64 { return s.walkQ }),
		maxQ(snapshots, func(s tickSnapshot) int64 { return s.embedQ }))
}

func maxQ(snaps []tickSnapshot, f func(tickSnapshot) int64) int64 {
	var m int64
	for _, s := range snaps {
		if v := f(s); v > m {
			m = v
		}
	}
	return m
}

func fmtDur(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
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
