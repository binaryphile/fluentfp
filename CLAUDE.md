@/home/ted/projects/share/tandem-protocol/tandem-protocol.md

# Software Development Simulation

## Development Environment

- **Language**: Go
- **Package Management**: Nix flakes for ongoing development
- **Ephemeral Tools**: nix-shell for one-off tool needs

## Code Style: FluentFP

Use `github.com/binaryphile/fluentfp` for fluent, functional patterns where they afford concise but clear code.

### slice Package - Complete API

```go
import "github.com/binaryphile/fluentfp/slice"

// Factory functions
slice.From(ts []T) Mapper[T]           // For mapping to built-in types
slice.MapTo[R](ts []T) MapperTo[R,T]   // For mapping to arbitrary type R

// Mapper[T] methods (also on MapperTo)
.KeepIf(fn func(T) bool) Mapper[T]     // Filter: keep matching
.RemoveIf(fn func(T) bool) Mapper[T]   // Filter: remove matching
.Convert(fn func(T) T) Mapper[T]       // Map to same type
.TakeFirst(n int) Mapper[T]            // First n elements
.Each(fn func(T))                      // Side-effect iteration
.Len() int                             // Count elements

// Mapping methods (return Mapper of target type)
.ToAny(fn func(T) any) Mapper[any]
.ToBool(fn func(T) bool) Mapper[bool]
.ToByte(fn func(T) byte) Mapper[byte]
.ToError(fn func(T) error) Mapper[error]
.ToFloat32(fn func(T) float32) Mapper[float32]
.ToFloat64(fn func(T) float64) Mapper[float64]
.ToInt(fn func(T) int) Mapper[int]
.ToRune(fn func(T) rune) Mapper[rune]
.ToString(fn func(T) string) Mapper[string]

// MapperTo[R,T] additional method
.To(fn func(T) R) Mapper[R]            // Map to type R
```

### slice Patterns

```go
// Count matching elements
count := slice.From(tickets).KeepIf(Ticket.IsActive).Len()

// Extract field to strings
ids := slice.From(tickets).ToString(Ticket.GetID)

// Method expressions for clean chains
actives := slice.From(users).
    Convert(User.Normalize).
    KeepIf(User.IsValid)
```

### option Package

```go
import "github.com/binaryphile/fluentfp/option"

// Creating options
option.Of(t T) Basic[T]                // Always ok
option.New(t T, ok bool) Basic[T]      // Conditional ok
option.IfProvided(t T) Basic[T]        // Ok if non-zero value
option.FromOpt(ptr *T) Basic[T]        // From pointer (nil = not-ok)

// Using options
.Get() (T, bool)                       // Comma-ok unwrap
.Or(t T) T                             // Value or default
.OrZero() T                            // Value or zero
.OrEmpty() T                           // Alias for strings
.OrFalse() bool                        // For option.Bool
.Call(fn func(T))                      // Side-effect if ok

// Pre-defined types
option.String, option.Int, option.Bool, option.Error
```

### option Patterns

```go
// Nullable database field
func (r Record) GetHost() option.String {
    return option.IfProvided(r.NullableHost.String)
}

// Tri-state boolean (true/false/unknown)
type Result struct {
    IsConnected option.Bool  // OrFalse() gives default
}
connected := result.IsConnected.OrFalse()
```

### must Package

```go
import "github.com/binaryphile/fluentfp/must"

must.Get(t T, err error) T             // Return or panic
must.BeNil(err error)                  // Panic if error
must.Getenv(key string) string         // Env var or panic
must.Of(fn func(T) (R, error)) func(T) R  // Wrap fallible func
```

### must Patterns

```go
// Initialization sequences
db := must.Get(sql.Open("postgres", dsn))
must.BeNil(db.Ping())

// Validation-only (discard result, just validate)
_ = must.Get(strconv.Atoi(configID))

// Inline in expressions
devices = append(devices, must.Get(store.GetDevices(chunk))...)

// Time parsing
timestamp := must.Get(time.Parse("2006-01-02 15:04:05", s.ScannedAt))

// With slice operations
atoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(atoi)
```

### ternary Package

```go
import "github.com/binaryphile/fluentfp/ternary"

ternary.If[R](cond bool).Then(t R).Else(e R) R
ternary.If[R](cond bool).ThenCall(fn).ElseCall(fn) R  // Lazy
```

### ternary Patterns

```go
// Factory alias for repeated use
If := ternary.If[string]
status := If(done).Then("complete").Else("pending")
```

### lof Package (Lower-Order Functions)

```go
import "github.com/binaryphile/fluentfp/lof"

lof.Println(s string)      // Wraps fmt.Println for Each
lof.Len(ts []T) int        // Wraps len
```

### pair Package (Tuples)

```go
import "github.com/binaryphile/fluentfp/tuple/pair"

// Pair type
pair.X[V1, V2]             // Struct with V1, V2 fields

// Creating pairs
pair.Of(v1, v2) X[V1,V2]   // Construct a pair

// Zipping slices
pair.Zip(as, bs) []X[A,B]           // Combine into pairs (panics if unequal length)
pair.ZipWith(as, bs, fn) []R        // Combine and transform (panics if unequal length)
```

### pair Patterns

```go
// Parallel slice iteration
pairs := pair.Zip(names, ages)
for _, p := range pairs {
    fmt.Printf("%s is %d\n", p.V1, p.V2)
}

// Direct transformation without intermediate pairs
users := pair.ZipWith(names, ages, NewUserFromNameAge)

// Chain with slice.From for filtering
adults := slice.From(pair.Zip(names, ages)).KeepIf(NameAgePairIsAdult)
```

### Predicate Naming Style

**Always name predicates** - never use inline anonymous functions for `KeepIf`/`RemoveIf`. Named predicates make code read like sentences.

**Naming patterns:**

| Pattern | When to use | Example |
|---------|-------------|---------|
| `Is[Condition]` | Simple check, subject obvious | `IsValidMAC` |
| `[Subject]Is[Condition]` | State check on specific type | `SliceOfScansIsEmpty` |
| `[Subject]Has[Condition](params)` | Parameterized predicate factory | `DeviceHasHWVersion("EX12")` |
| `Type.Is[Condition]` | Method expression | `Device.IsActive` |
| `Type.Has[Condition]` | Method expression | `Device.HasValidHost` |

**Simple predicate:**
```go
func IsValidMAC(mac string) bool {
    _, err := hex.DecodeString(mac)
    return len(mac) == 12 && err == nil
}

// Reads like English: "keep if is valid MAC"
macs := slice.From(inputs).KeepIf(IsValidMAC)
```

**Predicate factory** (when you need parameters):
```go
func DeviceHasHWVersion(versions ...string) func(Device) bool {
    return func(d Device) bool {
        return slices.Contains(versions, d.HWVersion)
    }
}

// Reads like: "remove if device has HW version EX12 or EX15"
devices = devices.RemoveIf(DeviceHasHWVersion("EX12", "EX15"))
```

**Method expression:**
```go
// Method on type
func (d Device) HasValidHost() bool { return d.Host != "" }

// Reads like: "keep if device has valid host"
devices = devices.KeepIf(Device.HasValidHost)
```

### Why Always Prefer FluentFP Over Loops

Loops are 3+ lines; FluentFP is 1 conceptual operation per line.
- Loops have multiple forms → mental load
- Loops force wasted syntax (discarded `_` values)
- Loops nest; FluentFP chains
- Loops describe *how*; FluentFP describes *what*

### When Loops Are Still Necessary

1. **Channel consumption** - `for r := range chan` has no FP equivalent
2. **Complex control flow** - break/continue/early return within loop

See [fluentfp/slice/README.md](https://github.com/binaryphile/fluentfp/blob/develop/slice/README.md#when-loops-are-still-necessary) for detailed examples.

## Testing: Red/Green TDD + Khorikov Principles

Reference: Khorikov, Vladimir. "Unit Testing: Principles, Practices, and Patterns." Manning, 2020.
Summary available at: `/home/ted/projects/urma-obsidian/sources/tier2-silver/practitioner-blogs/khorikov-unit-testing-olano-summary.md`

### TDD Cycle (MANDATORY)

**CRITICAL**: Never implement before writing a failing test. If caught implementing first, STOP and revert.

1. **Red**: Write a failing test first - NO EXCEPTIONS
2. **Green**: Write minimal code to make it pass
3. **Refactor**: Clean up while keeping tests green
4. **Prune**: After refactoring, evaluate if test should be kept (see quadrants below)

### Khorikov's Four Quadrants

Categorize code BEFORE deciding what to test:

| Quadrant | Complexity | Collaborators | Test Strategy |
|----------|------------|---------------|---------------|
| **Domain/Algorithms** | High | Few | Unit test heavily (edge cases) |
| **Controllers** | Low | Many | ONE integration test per happy path |
| **Trivial** | Low | Few | **Don't test at all** |
| **Overcomplicated** | High | Many | Refactor first, then test |

### Domain/Algorithms: Unit Test Heavily
- Pure functions with business logic
- Functions with conditional logic that determines behavior
- Test all edge cases with table-driven tests

**FluentFP-specific domain code:**
- `slice.KeepIf`, `slice.RemoveIf` - conditional inclusion logic
- `slice.TakeFirst` - boundary handling (`if n > len`)
- `slice.Fold`, `slice.Unzip2/3/4` - accumulation and multi-output logic
- `option.New`, `option.IfProvided`, `option.FromOpt` - conditional construction
- `option.Or`, `option.OrCall`, `option.MustGet` - conditional extraction
- `option.KeepOkIf`, `option.ToNotOkIf` - double conditional (filter)
- `ternary.Else`, `ternary.ElseCall` - 3 code paths each

### Representative Pattern Testing
When multiple methods share **identical logic**, test ONE representative:
- `option.ToInt` covers all `option.ToX` methods (same if-ok-then-transform pattern)
- `option.Or` covers `OrZero`/`OrEmpty`/`OrFalse` (same if-ok-return-value pattern)
- `slice.KeepIf` on `Mapper` covers `MapperTo.KeepIf` (identical implementation)

### Controllers: ONE Integration Test
- **One happy path** per major workflow - not multiple!
- Plus edge cases that **cannot** be covered by unit tests
- Example: `Export()` gets ONE test that verifies files created, not separate tests for each file

**Anti-pattern**: Writing TestWriteMetrics, TestWriteTickets, TestWriteSprints separately - these are all ONE controller operation.

### Trivial Code: Don't Test
Examples of trivial code to **skip**:
- Simple getters/setters
- Constructors with no logic
- Code that just delegates
- Loop + apply with no branching (e.g., `ToInt`, `ToString` - just iterate and call fn)
- Wrappers around stdlib (e.g., `lof.Len` wraps `len()`)
- Type aliases with identical logic (e.g., `OrZero`, `OrEmpty`, `OrFalse` are same impl)

**FluentFP-specific trivial code:**
- `slice.From`, `slice.MapTo` - just return input
- `slice.Len` - wraps `len()`
- `slice.ToX` methods - loop + apply, no branching
- `option.Of`, `option.NotOk` - just construct struct
- `option.Get`, `option.IsOk` - just return fields
- `ternary.If`, `ternary.Then`, `ternary.ThenCall` - just store values

### What Makes a Test Bad (Delete It)
- Tests implementation details (e.g., checking for specific string ",5," in CSV)
- Breaks on every refactor but functionality still works
- Redundant with other tests (multiple controller tests for same operation)
- Tests trivial code that can't meaningfully fail

### Test Behavior, Not Implementation
- Test **observable outcomes**: "file exists", "has 5 lines", "returns error"
- Don't test **how**: checking specific string formats, column order, internal state
- Black-box by default: verify outputs, not steps
- Mocks only for **external boundaries** (filesystem, HTTP), never internal collaborators

### Coverage: Diagnostic Only

**How we track:**
1. Run `go test -cover ./...` at end of each phase
2. Update baseline in this file
3. Investigate any package that drops significantly
4. Note: Low coverage is fine IF the code is trivial (see quadrants)

```bash
go test -cover ./...
```

**Current baseline (2026-01-03):**
| Package | Coverage | Notes |
|---------|----------|-------|
| must | 100% | All domain (conditional + panic) |
| ternary | 100% | All domain code paths tested |
| option | 51.9% | Domain tested, trivial aliases untested |
| slice | 28.8% | Domain tested, ToX methods trivial |
| lof | 0.0% | All trivial wrappers - acceptable |

Per Khorikov: Coverage is a "good negative indicator, bad positive one."
- **High coverage**: Means nothing about quality
- **Low coverage**: Fine if code is trivial (ToX methods, aliases, wrappers)
- **Don't target a number**: Creates perverse incentive for useless tests
- **Drops matter more than absolutes**: If a package drops 20%, investigate

### Test Performance Budget

**How we track:**
```bash
go test -v ./... 2>&1 | grep -E "^(ok|---)"  # Time per package
go test -bench=. -benchmem ./...              # Memory allocation
```

**Current baseline (2026-01-03):**
| Package | Time | Notes |
|---------|------|-------|
| must | <0.01s | Target: <0.1s |
| ternary | <0.01s | Target: <0.1s |
| option | <0.01s | Target: <0.1s |
| slice | <0.01s | Target: <0.1s |

**Thresholds:**
- **Per-package**: <0.1s for unit tests, <1s for integration
- **Total suite**: <5s (enables fast TDD cycles)
- **If exceeded**: Profile with `go test -cpuprofile` and optimize or split

### Go Test Style
- Prefer **table-driven tests** for domain algorithms
- Use descriptive test names that explain the behavior being tested
- Group related assertions in subtests with `t.Run()`

```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"zero input", 0, 0},
        {"positive input", 5, 10},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Foo(tt.input)
            if got != tt.expected {
                t.Errorf("Foo(%d) = %d, want %d", tt.input, got, tt.expected)
            }
        })
    }
}
```

Run tests: `go test ./...`
Run with coverage: `go test -cover ./...`

## Development Process

Our development workflow follows this sequence:

1. **Use Cases** (`/use-case-skill`): Define use cases as the origin of development
2. **Design Documents**: Create design documents from use cases
3. **Implementation Plan**: Plan the implementation with Tandem Protocol contract

Each phase produces artifacts that inform the next, ensuring traceability from user needs to code.

## Branching Strategy: Trunk-Based Development

We use trunk-based development (TBD) to minimize merge complexity and enable continuous integration.

### Core Principles

- **Single trunk**: `main` is the only long-lived branch
- **Short-lived feature branches**: If used, live < 1-2 days max
- **Small, frequent commits**: Commit directly to main when safe
- **Feature flags over branches**: Hide incomplete work behind flags, not branches
- **Always releasable**: Main should always be in a deployable state

### When to Use a Branch

| Situation | Approach |
|-----------|----------|
| Small change (< 1 day) | Commit directly to main |
| Larger feature (1-2 days) | Short-lived branch, merge quickly |
| Experimental/risky | Feature flag on main |
| Multi-day work | Break into smaller pieces that can merge daily |

### Practices

- **No long-lived feature branches**: They create merge hell
- **No release branches**: Tag releases on main instead
- **Continuous integration**: All commits trigger CI on main
- **Code review**: Use small PRs or pair programming
- **Revert over rollback**: If main breaks, revert the commit

### Feature Flags

For incomplete features that span multiple commits:
```go
if config.EnableDataExport {
    // new feature code
}
```

This lets us merge to main continuously without exposing unfinished work.

## Data Output Requirements

The simulation must produce sufficient data output to:
- Compare actual runs against theoretical predictions
- Validate variance models and incident rates
- Enable statistical analysis across multiple seeds
- Support hypothesis testing (DORA vs TameFlow)

## Project Overview

A software development simulation with two evolutionary paths:
1. Full video game
2. LLM-based laboratory for automated software development experimentation
