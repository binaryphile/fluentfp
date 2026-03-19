//go:build ignore

// DAG pipeline: CSV ingest → parse → validate → Tee to DB + audit log.
//
// Run:
//
//	go run examples/pipeline_fanout.go
//
// Demonstrates toc pipeline composition without a framework:
//   - Start: bounded worker stage with backpressure
//   - Pipe: chain stages, forwarding errors
//   - NewTee: broadcast each item to N branches
//   - Stats: built-in observability on every stage
//
// Each stage has bounded capacity, configurable workers, and automatic
// error propagation via rslt.Result channels.
package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

// Record represents a parsed CSV row.
type Record struct {
	Name   string
	Amount int
}

// parseLine splits a CSV line into a Record.
func parseLine(_ context.Context, line string) (Record, error) {
	parts := strings.SplitN(line, ",", 2)
	if len(parts) != 2 {
		return Record{}, fmt.Errorf("bad format: %q", line)
	}
	amount, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return Record{}, fmt.Errorf("bad amount in %q: %w", line, err)
	}
	return Record{Name: strings.TrimSpace(parts[0]), Amount: amount}, nil
}

// validateRecord rejects negative amounts.
func validateRecord(_ context.Context, r Record) (Record, error) {
	if r.Amount < 0 {
		return r, fmt.Errorf("negative amount for %s: %d", r.Name, r.Amount)
	}
	return r, nil
}

// writeDB simulates a database insert.
func writeDB(_ context.Context, r Record) (string, error) {
	return fmt.Sprintf("db: inserted %s ($%d)", r.Name, r.Amount), nil
}

// writeAudit simulates an audit log entry.
func writeAudit(_ context.Context, r Record) (string, error) {
	return fmt.Sprintf("audit: logged %s ($%d)", r.Name, r.Amount), nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulated CSV input.
	lines := []string{
		"Alice, 100",
		"Bob, 250",
		"bad-line",
		"Carol, -50",
		"Dave, 300",
	}

	// Pipeline: parse → validate → tee(db, audit)
	parseStage := toc.Start(ctx, parseLine, toc.Options[string]{
		Capacity:        5,
		Workers:         2,
		ContinueOnError: true,
	})

	validateStage := toc.Pipe(ctx, parseStage.Out(), validateRecord, toc.Options[Record]{
		ContinueOnError: true,
	})

	tee := toc.NewTee(ctx, validateStage.Out(), 2)
	dbPipe := toc.Pipe(ctx, tee.Branch(0), writeDB, toc.Options[Record]{})
	auditPipe := toc.Pipe(ctx, tee.Branch(1), writeAudit, toc.Options[Record]{})

	// Submit lines.
	for _, line := range lines {
		if err := parseStage.Submit(ctx, line); err != nil {
			log.Printf("submit failed: %v", err)
		}
	}
	parseStage.CloseInput()

	// Drain both branches.
	done := make(chan struct{})
	drain := func(name string, ch <-chan rslt.Result[string]) {
		for r := range ch {
			if err := r.Err(); err != nil {
				fmt.Printf("  %s error: %v\n", name, err)
			} else {
				v, _ := r.Get()
				fmt.Printf("  %s: %s\n", name, v)
			}
		}
		done <- struct{}{}
	}
	go drain("db", dbPipe.Out())
	go drain("audit", auditPipe.Out())
	<-done
	<-done

	// Print stats.
	ps := parseStage.Stats()
	fmt.Printf("\nParse:    submitted=%d completed=%d failed=%d\n", ps.Submitted, ps.Completed, ps.Failed)
	vs := validateStage.Stats()
	fmt.Printf("Validate: completed=%d failed=%d\n", vs.Completed, vs.Failed)
	ts := tee.Stats()
	fmt.Printf("Tee:      received=%d delivered=%d\n", ts.Received, ts.FullyDelivered)
}
