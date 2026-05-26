# 10 Go FP Libraries Side-by-Side

A single runnable comparison: 10 Go functional libraries each implement the same task — filter active users, extract names, print them, return as `[]string` — then exercise operations beyond filter+map (Find, Reduce, chaining, Unzip).

## Why a separate Go module

This directory has its own [`go.mod`](go.mod) because the comparison imports 10 third-party FP libraries. Keeping them out of the main module avoids transitive dependencies for users who just want fluentfp.

## Run

```bash
go run examples/comparison/main.go
```

Or, from inside this directory:

```bash
go run main.go
```

## What it compares

Each library is wrapped in its own `{ ... }` block in [main.go](main.go) so they can all define identically-named functions (`printActiveNames`) without name collision. The package-level doc comment at the top of `main.go` has a quick-reference table showing line counts and four binary properties per library: type-safe, concise, supports method expressions, fluent.

For the synthesized summary, design rationale, and benchmarks, see [`comparison.md`](../../comparison.md) at the project root.
