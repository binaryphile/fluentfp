//go:build ignore

// Tee/Join write-ahead log + primary store pattern.
//
// Run:
//
//	go run examples/tee_join_wal.go
//
// Classic dual-write problem solved with toc primitives:
//   - Tee broadcasts each write to WAL and primary store
//   - Join recombines the results for a consistency check
//   - If either branch fails, the combined result is Err
//
// Join is one-shot by design: it consumes the first result from each
// branch and combines them. This example submits items one at a time,
// each getting its own Tee→Join pipeline via the Start worker function.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

// WriteResult captures the outcome of a write operation.
type WriteResult struct {
	WAL     string
	Primary string
}

// dualWrite runs a one-shot Tee→Join pipeline for a single item:
// broadcast to WAL and primary store, recombine results.
func dualWrite(ctx context.Context, item string) (WriteResult, error) {
	ch := make(chan rslt.Result[string], 1)
	ch <- rslt.Ok(item)
	close(ch)

	tee := toc.NewTee(ctx, ch, 2)

	// WAL branch — write-ahead log.
	walWrite := func(_ context.Context, s string) (string, error) {
		return fmt.Sprintf("WAL: %s", s), nil
	}
	walPipe := toc.Pipe(ctx, tee.Branch(0), walWrite, toc.Options[string]{})

	// Primary branch — main data store.
	primaryWrite := func(_ context.Context, s string) (string, error) {
		return fmt.Sprintf("Primary: %s", s), nil
	}
	primaryPipe := toc.Pipe(ctx, tee.Branch(1), primaryWrite, toc.Options[string]{})

	// Join recombines: both must succeed for the write to be consistent.
	combineWrites := func(wal, primary string) WriteResult {
		return WriteResult{WAL: wal, Primary: primary}
	}
	join := toc.NewJoin(ctx, walPipe.Out(), primaryPipe.Out(), combineWrites)

	r := <-join.Out()
	for range join.Out() {
	}
	join.Wait()
	return r.Unpack()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Process items through dual-write pipeline.
	stage := toc.Start(ctx, dualWrite, toc.Options[string]{
		Capacity:        5,
		Workers:         2,
		ContinueOnError: true,
	})

	items := []string{"order-1", "order-2", "order-3", "order-4", "order-5"}
	for _, item := range items {
		if err := stage.Submit(ctx, item); err != nil {
			log.Printf("submit failed: %v", err)
		}
	}
	stage.CloseInput()

	// Drain results.
	for r := range stage.Out() {
		if err := r.Err(); err != nil {
			fmt.Printf("  error: %v\n", err)
			continue
		}
		wr, _ := r.Get()
		fmt.Printf("  %s | %s\n", wr.WAL, wr.Primary)
	}

	s := stage.Stats()
	fmt.Printf("\nStats: submitted=%d completed=%d failed=%d\n", s.Submitted, s.Completed, s.Failed)
}
