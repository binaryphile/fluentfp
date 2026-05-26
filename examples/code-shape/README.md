# Code Shape: Conventional vs fluentfp Side-by-Side

Two pairs of files showing what fluentfp changes about the **shape** of code — and where it deliberately doesn't help.

| Pair | Files | Domain |
|---|---|---|
| Basic | [conventional.go](conventional.go) ↔ [fluentfp.go](fluentfp.go) | User filter / extract / count / fold / iterate (~7 → 3 lines per operation) |
| Best-case | [best-case-conventional.go](best-case-conventional.go) ↔ [best-case-fluentfp.go](best-case-fluentfp.go) | Employee report generator: chained transformations and aggregations (330 → 206 lines) |

Each file builds with `//go:build ignore`, so they don't compile as a package. Open the matched pair in two panes and diff visually.

## What to look for

- **Where the chain wins:** filter/map/fold compositions that read as one expression instead of three loops with intermediate `var` declarations.
- **Where the loop stays:** `sendNotifications` in `fluentfp.go` keeps the imperative form because the loop body has an early return on error. fluentfp doesn't try to replace every loop — only the mechanical ones.
- **Method expressions:** `User.IsActive`, `User.GetEmail`, etc. plug into `KeepIf` and `ToString` directly because each method's receiver-as-first-arg shape matches the higher-order function's expected signature. No wrapper closures.

For the full set of real-world before/after rewrites, see [`docs/showcase.md`](../../docs/showcase.md).
