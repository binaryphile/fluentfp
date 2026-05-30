@/home/ted/projects/tandem-protocol/README.md

# fluentfp - Functional Programming Library for Go

## Development Environment

- **Language**: Go
- **Package Management**: Go modules

### evtctl — project task management

```
evtctl task <description>            # publish a task event
evtctl task --to <project> <desc>    # task for another project
evtctl inbox <app> <message>         # send inbox message
evtctl done <id>[,<id>...] [evidence] # publish a task-done event
evtctl open                          # list open tasks
evtctl audit                         # full task reconciliation
evtctl claim <id> <name>             # claim a task
evtctl claims                        # list active claims
```

Stream name automatically derived from project directory: `tasks.fluentfp`. To send tasks to other projects: `evtctl task --to <project> <description>`. To send inbox messages: `evtctl inbox <app> <message>`.

## Code Style: fluentfp

Use `mcp__era__code_search` for API signatures and package details.

### Container Variable Naming

Variables holding `Option[T]` use a `*Option` suffix; variables holding `Result[T]` use a `*Result` suffix. This matters because fluentfp container types share method names (`FlatMap`, `Transform`, `Get`, etc.) — without the suffix, a reader seeing `order.FlatMap(validate)` can't tell if `order` is a `Result`, an `Option`, or a `Mapper` (slice). The suffix makes the container type visible at the call site:

```go
order, err := web.DecodeJSON[Order](req)    // (Order, error)
rawMinTotalOption := option.NonEmpty(q.Get("min_total"))  // Option[string]
mtOption, err := option.MapResult(rawMinTotalOption, parseMinTotal).Unpack()  // (Option[int], error)
```

Exceptions: when the type is obvious from context (e.g., a one-line function return), the suffix can be omitted.

### Named vs Inline Functions

**Preference hierarchy** (best to worst):
1. **Method expressions** - `User.IsActive`, `Device.GetMAC` (cleanest, no function body)
2. **Named functions** - `isActive := func(u User) bool {...}` (readable, debuggable)

No inline lambdas — if the logic is simple enough to inline, it's simple enough to name and document. Exception: standard idioms (t.Run, http.HandlerFunc).

**Naming exception:** Standalone cross-type transforms may use ecosystem-standard names when the house-style alternative is materially worse. Current exception: `FilterMap` (not `KeepMap`) — universally recognized (Rust, lo, Elixir).

**Uniform commas rule — commas at one nesting level only.** When a call contains another call, only one level may have multiple arguments (commas). This keeps every comma at the same nesting depth, so the reader never has to mentally track which arguments belong to which call.

```go
// BAD: commas at both levels — outer has 2 args, inner has 2 args
slice.SortByDesc(kv.Map(m, toResult), sortKey)

// GOOD: extract inner call — commas only at outer level
results := kv.Map(m, toResult)
slice.SortByDesc(results, sortKey)

// OK: commas only at inner level — outer has 1 arg
slice.From(slice.NonZero(items))

// OK: commas only at outer level — each inner call has 1 arg
pair.ZipWith(slice.From(as), slice.From(bs), combine)
```

**Why name functions:**

Anonymous functions and higher-order functions require mental effort to parse. Named functions **reduce this cognitive load** by making code read like English:

```go
// Inline: reader must parse lambda syntax and infer meaning
slice.From(tickets).KeepIf(func(t Ticket) bool { return t.CompletedTick >= cutoff }).Len()

// Named: reads as intent - "keep if completed after cutoff"
slice.From(tickets).KeepIf(completedAfterCutoff).Len()
```

Named functions aren't ceremony—they're **documentation at the right boundary**. If logic is simple enough to consider inlining, it's simple enough to name and document. The godoc comment is there when you need to dig deeper—consistent with Go practices everywhere else.

**Locality:** Define named functions close to first usage, not at package level.

#### Method Expressions (preferred)

When a type has a method matching the required signature, use it directly:
```go
// Best: method expression
actives := users.KeepIf(User.IsActive)
names := users.ToString(User.Name)
```

#### Named Functions (when method expressions don't apply)

When you need custom logic or the type lacks an appropriate method. **Include godoc-style comments**:
```go
// isRecentlyActive returns true if user is active and was seen after cutoff.
isRecentlyActive := func(u User) bool {
    return u.IsActive() && u.LastSeen.After(cutoff)
}
actives := users.KeepIf(isRecentlyActive)
```

#### Predicate Naming Patterns

| Pattern | When to use | Example |
|---------|-------------|---------|
| `Is[Condition]` | Simple check, subject obvious | `IsValidMAC` |
| `[Subject]Is[Condition]` | State check on specific type | `SliceOfScansIsEmpty` |
| `[Subject]Has[Condition](params)` | Parameterized predicate factory | `DeviceHasHWVersion("EX12")` |
| `Type.Is[Condition]` | Method expression | `Device.IsActive` |

#### Reducer Naming

```go
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
total := slice.Fold(amounts, 0.0, sumFloat64)
```

**See also:** [naming-in-hof.md](naming-in-hof.md) for complete naming patterns.

### Why Always Prefer fluentfp Over Loops

**Concrete example - field extraction:**

```go
// fluentfp: one expression stating intent
return slice.From(f.History).ToFloat64(FeverSnapshot.GetPercentUsed)

// Loop: four concepts interleaved
// Extract percent used values from history
var result []float64                           // 1. variable declaration
for _, s := range f.History {                  // 2. iteration mechanics (discarded _)
    result = append(result, s.PercentUsed)     // 3. append mechanics
}
return result                                  // 4. return
```

The loop forces you to think about *how* (declare, iterate, append, return). fluentfp expresses *what* (extract PercentUsed as float64s).

**General principles:**
- Loops have multiple forms → mental load
- Loops force wasted syntax (discarded `_` values)
- Loops nest; fluentfp chains
- Loops describe *how*; fluentfp describes *what*

### When Loops Are Still Necessary

1. **Channel consumption** - `for r := range chan` has no FP equivalent
2. **Complex control flow** - break/continue/early return within loop

## Testing: Khorikov Principles

### Khorikov's Four Quadrants

| Quadrant | Complexity | Collaborators | Test Strategy |
|----------|------------|---------------|---------------|
| **Domain/Algorithms** | High | Few | Unit test heavily (edge cases) |
| **Controllers** | Low | Many | ONE integration test per happy path |
| **Trivial** | Low | Few | **Don't test at all** |
| **Overcomplicated** | High | Many | Refactor first, then test |

### Domain/Algorithms: Unit Test Heavily

**fluentfp-specific domain code:**
- `slice.KeepIf`, `slice.RemoveIf` - conditional inclusion logic
- `slice.Take`, `slice.TakeLast`, `slice.Drop`, `slice.DropLast` - boundary handling (`if n > len`, negative n)
- `slice.TakeWhile`, `slice.DropWhile`, `slice.DropLastWhile` - predicate-based prefix/suffix logic
- `slice.Fold`, `slice.Scan`, `slice.Unzip2/3/4` - accumulation and multi-output logic
- `slice.Zip`, `slice.ZipWith` - length-mismatch truncation
- `slice.Intersperse` - separator insertion edge cases (empty, single)
- `slice.Range`, `slice.RangeFrom`, `slice.RangeStep` - half-open integer generation with direction/step edge cases
- `slice.Window` - sliding window with backing array aliasing
- `stream.RemoveIf` - complement of KeepIf (delegation correctness)
- `stream.Every`, `stream.None` - universal/negative quantification with short-circuit
- `seq.Unfold` - stateful lazy generation with termination
- `option.New`, `option.NonZero`, `option.NonNil` - conditional construction
- `option.Or`, `option.OrCall`, `option.MustGet` - conditional extraction
- `option.KeepIf`, `option.RemoveIf` - double conditional (filter)
- `option.OrWrap` - absent-case recovery staying in Option (lazy evaluation)
- `option.ZipWith` - combine two Options (both-present gate)
- `option.WhenCall` - conditional function call with eager nil check
- `slice.FilterMap` - combined filter+transform with comma-ok callback
- `slice.MinBy`, `slice.MaxBy` - extremum by key with cmp.Compare (NaN ordering)
- `slice.Reduce` - fold without initial, single-element returns without calling fn
- `slice.Associate` - key+value extraction to map (last wins for duplicates)
- `slice.Enumerate` - index pairing
- `slice.Unique` - comparable dedup (NaN never deduplicates)
- `seq.Enumerate` - lazy index pairing with per-iteration reset
- `seq.FilterMap` - lazy filter+transform with comma-ok callback
- `seq.Reduce` - terminal fold without initial, unconditional nil-fn panic
- `seq.Unique`, `seq.UniqueBy` - lazy dedup with per-iteration seen-set reset
- `seq.Contains` - terminal short-circuit membership check
- `seq.Intersperse` - lazy separator insertion (O(1) state)
- `seq.Chunk` - lazy batching with stable independent snapshots

### Representative Pattern Testing

When multiple methods share **identical logic**, test ONE representative:
- `option.ToInt` covers all `option.ToX` methods (same if-ok-then-transform pattern)
- `option.Or` covers `OrZero`/`OrEmpty` (same if-ok-return-value pattern)

### Trivial Code: Don't Test

- Loop + apply with no branching (e.g., `ToInt`, `ToString` - just iterate and call fn)
- Wrappers around stdlib (e.g., `lof.Len` wraps `len()`)
- Type aliases with identical logic (e.g., `OrZero`, `OrEmpty` are same impl)
- `slice.From` - just returns input
- `option.Of`, `option.NotOk` - just construct struct
- `option.Get`, `option.IsOk` - just return fields
- `option.When` - trivial delegation to option.New

### Coverage

The load-bearing rule is the Khorikov posture above (domain heavily,
controllers integration-once, trivial untested). Per-package numbers
rot too fast to keep in this file. Current snapshot:

```bash
nix develop -c go test -cover ./...
```

Low coverage on a wrapper package (e.g. `lof`) is acceptable when the
wrapped operations are stdlib; high coverage on a domain package is the
target. Don't refactor to chase a number — refactor when the test class
(domain / controller / trivial / overcomplicated) is wrong.

### Go Test Style

- Prefer **table-driven tests** for domain algorithms
- Use descriptive test names that explain the behavior being tested
- Group related assertions in subtests with `t.Run()`

### Build and test via ./mk

Use `./mk` for build/test/clean — it wraps the underlying `go` commands and
prints what it ran (via `mk.Cue`), so the operator sees the actual command.

| Command | What it runs |
|---|---|
| `./mk build` | `mkdir -p bin && go build -o bin/orders ./examples/orders/` |
| `./mk test`  | `go test ./...` |
| `./mk clean` | `rm -f bin/orders && rmdir bin 2>/dev/null \|\| true` |

`./mk -h` shows the full usage. Coverage still goes through `go` directly:
`go test -cover ./...`.

**Nix shell required.** The `flake.nix` provides `go`, `gopls`,
`golangci-lint`, `gh`, `nodejs`, and `sqlite`; these are NOT on the system
PATH outside the dev shell. From a plain bash session, prefix commands with
`nix develop -c bash -c '...'` (or use `nix develop` interactively). `./mk`
and direct `go` calls silently fail-as-not-found in unactivated shells —
the symptom is a "command not found" swallowed by piping or a hang waiting
for a missing toolchain.

### CI drift gates

Two scripts feed the `docs-check` workflow at `.github/workflows/docs-check.yml`;
both fail CI if regenerated output diverges from what's committed.

- **`scripts/check-docs.py`** — enforces showcase anchor/count consistency
  between `docs/showcase.md` (24 entries) and `README.md`. Run locally with
  `python3 scripts/check-docs.py`. Edit a showcase entry → re-run → commit
  any updated counts.
- **`scripts/render-shape-viz.py`** — regenerates the two heatmap SVGs at
  `images/code-shape-comparison.svg` + `images/best-case-code-shape-comparison.svg`
  from the source files in `examples/loop-to-chain/`. Run with
  `nix develop -c python3 scripts/render-shape-viz.py`. CI runs
  `git diff --exit-code images/` after regeneration. If you edit any
  `examples/loop-to-chain/*.go`, re-run the script and commit the resulting
  SVG changes in the same commit as the source edit. The script is
  byte-deterministic; non-zero diff means stale output.

## Documentation Updates

When adding or changing packages, always update these docs as part of the same cycle (docs-first per protocol §3a — UC → design → README+CHANGELOG → impl → tests):

1. **Use cases** (`docs/use-cases.md`) — Cockburn-style. Existing UC-1 through UC-13 are the canonical shape: Scope/Level/Actor line, Stakeholders, Postconditions, Minimal Guarantee, Preconditions, Main Scenario, Extensions, Sub-Variations. Update the scope line (top of file) and the Actor-Goal table when adding a new domain. UCs come first — they define WHAT the package does before design explains HOW.
2. **Design** (`docs/design.md`) — Package structure table, mermaid diagram, design decisions (D-numbered; last entry is D38). Edge cases and rationale live here, not in the UC.
3. **README** (`README.md`) — Packages table, package highlights if warranted.
4. **Package README** (`<pkg>/README.md`) — Per-package README following existing patterns (see `web/README.md`, `slice/README.md`, etc.).
5. **CHANGELOG.md** — User-visible changes go under a new version heading (semver, see existing entries v0.59 / v0.60 / v0.61). Behavior changes get explicit "Behavior change:" notes; new features get one bullet.

### Showcase compile-check suite

`docs/showcase.md` carries 24 before/after rewrites of real-world code. Each entry has a corresponding compile-check package under `internal/showcasetest/<entry>/<entry>.go` that verifies the fluentfp snippet compiles against the real API. When adding or editing a showcase entry:

1. Update or add the matching `internal/showcasetest/<entry>/<entry>.go` package so the snippet builds.
2. Run `nix develop -c go build ./internal/showcasetest/...` locally to verify.
3. Run `python3 scripts/check-docs.py` to re-verify the showcase ↔ README count/anchor consistency (also enforced in CI).

The compile-checks have caught real bugs that editorial review missed (e.g., wrong return types in chained calls, stale API names after renames). Treat them as a load-bearing part of the showcase.

### Cycle attestation: criterion names verbatim

When publishing `evtctl complete` for a contract, copy the criterion-name strings verbatim from the published contract event into the completion's `criteria[].name` fields. The `validate-attestation` superset-match join is string-equality on these names — paraphrasing (e.g., "writeResponse honors caller Content-Type" → "writeResponse honors caller's Content-Type") will cause the audit to leave the contract unmatched even though the work shipped. Get the names from `era query "tasks.fluentfp" 'id = <contract-event-id>' --json` before drafting the attestation.

## Branching Strategy: Trunk-Based Development

- **Single trunk**: `main` is the only long-lived branch
- **Small, frequent commits**: Commit directly to main
- **Tag releases**: Use semantic versioning (v0.6.0, etc.)

### Visual assets: SVG canonical, PNG generated on demand

Visual assets in `images/` are SVG. PNGs are never committed; if a PNG is
needed (preview, external embed, chat screenshot), generate it on demand
from the SVG via:

```bash
nix develop -c magick images/<name>.svg images/<name>.preview.png
```

The `.preview.png` filename signals it is disposable. Delete after use.
This keeps the repo SVG-only and avoids the binary-asset failure mode
described below.

### Pre-commit hook: no binaries in commits

The global `~/dotfiles/githooks/pre-commit` hook refuses any commit whose
staged changes include binary files — and it does NOT distinguish
additions from deletions. The SVG-canonical policy above eliminates the
trip at the source: never commit a PNG, JPG, etc., and the hook is happy.

If a one-off binary commit is genuinely intended (e.g., removing a legacy
binary asset that's already tracked), the hook's own help text documents
`git commit --no-verify` as the escape. Use it deliberately, and document
the reason in the commit message so the audit trail is intact.
