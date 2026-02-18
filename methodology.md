# fluentfp Methodology

This document details how empirical claims in [analysis](analysis.md) were derived, enabling readers to verify or replicate the analysis.

**Contents:**
- [A. Loop Sampling Methodology](#a-loop-sampling-methodology)
- [B. Line Classification Rules](#b-line-classification-rules)
- [C. Density Calculation](#c-density-calculation)
- [D. Replication Guide](#d-replication-guide)
- [E. Limitations](#e-limitations)
- [F. Code Metrics Tool (scc)](#f-code-metrics-tool-scc)
- [G. Chain Formatting Rules](#g-chain-formatting-rules)
- [H. Real-World Loop Bugs](#h-real-world-loop-bugs)
- [I. Performance Analysis](#i-performance-analysis)

## A. Loop Sampling Methodology

How 11 representative loops were selected from a production codebase (608 total):

**What counts as "a loop":**
- Each `for` statement = 1 loop (nested loops count separately)
- `for range`, `for i := 0; ...`, and `for { ... }` all count
- Excluded: test files (table-driven tests skew toward simple patterns)

**Selection approach:**
- Systematic sample: every ~55th loop (608 ÷ 11 ≈ 55)
- Starting point chosen randomly
- No cherry-picking or exclusions after selection

**Source:** Analysis performed on an internal production Go project (~15k LOC excluding tests).

## B. Line Classification Rules

Explicit rules for semantic vs syntactic classification:

**Semantic (intent-carrying):**
- Condition expressions: `if x.IsActive()`, `switch`, `case`
- Accumulation statements: `count++`, `result = append(...)`, `total += x`
- Function calls that do work: `process(item)`, `db.Save(record)`
- Return statements with values: `return result`

**Syntactic (mechanics-only):**
- Variable declarations: `var x T`, `x := 0`, `x := make(...)`
- Loop headers: `for _, x := range xs {`, `for i := 0; i < n; i++ {`
- Closing braces: `}` (standalone line)
- Blank lines within loop body
- Comments (don't count toward either)

**Edge cases (judgment calls—reasonable people may differ):**
- `if x.IsActive() {` — semantic (condition is the point; brace is incidental)
- `x := slice.From(xs).` — syntactic (setup/scaffolding)
- `return nil` — we classify as syntactic (no semantic payload), but could argue either way
- `return result` — semantic (delivers computed value)
- `err != nil` checks — syntactic (error handling boilerplate), though essential

**Guiding principle:** If the line would disappear in a pseudocode version, it's syntactic. If it carries domain meaning, it's semantic.

## C. Density Calculation

**Formula:**
```
Semantic Density = Semantic Lines / Total Lines × 100%
```

**Worked example (filter + count):**

Loop version:
```
Total lines: 6
Semantic: 2 (condition, accumulation)
Syntactic: 4 (setup, header, 2 braces)
Density: 2/6 = 33%
```

fluentfp version:
```
Total lines: 3
Semantic: 2 (KeepIf, Len)
Syntactic: 1 (setup line)
Density: 2/3 = 67%
```

## D. Replication Guide

How readers can verify on their own codebase:

1. **Count loops** (excluding tests):
   ```bash
   grep -rn "^\s*for\s" --include="*.go" --exclude="*_test.go" . | wc -l
   ```
   This catches all forms: `for range`, `for i := ...`, `for condition {`, and `for {`.

2. **Sample systematically**: For N loops, take every (N ÷ 10)th loop. Random start point.

3. **For each sampled loop**:
   - Count total lines (from `for` to closing `}`)
   - Count *visual* lines as displayed, not logical statements
   - Multi-line statements: each line counts separately
   - Mark each line semantic or syntactic per Section B rules
   - Calculate: semantic ÷ total × 100

**Multi-line example:**
```go
count := slice.From(users).     // line 1: syntactic (setup)
    KeepIf(User.IsActive).      // line 2: SEMANTIC (filter)
    Len()                       // line 3: SEMANTIC (count)
```
This is 3 visual lines: 1 syntactic + 2 semantic = 67% density.

4. **Aggregate**: Average across all sampled loops

**Expected results for typical Go codebases:**
- Simple transforms: 30-40% semantic density
- Complex control flow: 40-60% semantic density
- Overall average: 35-45% semantic density

Results outside these ranges aren't wrong—they may indicate different coding styles or domain characteristics.

## E. Limitations

**What this metric measures:**
- Vertical space efficiency (lines consumed per unit of intent)
- Proportion of "meaningful" vs "mechanical" code

**What this metric does NOT measure:**
- Readability (dense code isn't always clearer)
- Correctness (fewer lines doesn't mean fewer bugs—though see [Error Prevention](analysis.md#error-prevention) for how fluentfp eliminates certain bug classes)
- Performance (no runtime implications)
- Maintainability (though reduced boilerplate can help)

**Caveats:**
- Classification involves judgment calls; different analysts may vary by ±5%
- Sample size (11 loops) provides directional insight, not statistical significance
- Results are specific to Go; other languages may differ

This is one lens among many. Use alongside other quality metrics, not as a sole criterion.

## F. Code Metrics Tool (scc)

The [Measuring the Correlation](analysis.md#measuring-the-correlation) section uses [scc](https://github.com/boyter/scc) (Sloc, Cloc and Code) for line counting and complexity measurement.

**Why scc:**
- Separates code lines from blanks and comments
- Provides complexity estimates at near-zero CPU cost
- Fast enough for large codebases

**Code vs Lines:**
scc distinguishes:
- **Lines**: Total lines including blanks and comments
- **Code**: Executable statements only
- **Blanks**: Empty lines
- **Comments**: Documentation lines

We report **Code** lines for accuracy. Total lines overcount by including whitespace and documentation.

**Complexity metric:**
scc's complexity is an approximation of [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity). It counts branch and loop tokens in the code:

- `for`, `if`, `switch`, `while`, `else`
- `||`, `&&`, `!=`, `==`

Each occurrence increments the file's complexity counter. This is cheaper than building an AST but provides a reasonable approximation for comparing files in the same language.

**Why complexity matters:**
Higher complexity = more execution paths = more levers available to pull incorrectly. This is why the 95% complexity reduction matters—it's correctness by construction. See [The Principle](analysis.md#the-principle) for why eliminating control structures eliminates the bugs they enable.

**Usage:**
```bash
# Compare two files
scc file1.go file2.go

# Find most complex files in a project
scc --by-file -s complexity .
```

**Limitation:** Complexity is comparable only within the same language. Don't compare Go complexity to Python complexity directly.

## G. Chain Formatting Rules

How fluentfp chains are formatted affects line counts. These rules ensure consistent measurement.

**Single operation = single line (ToX methods chain onto previous):**
```go
active := slice.From(users).KeepIf(User.IsActive)
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

**Two+ operations = multiline, each operation on its own line:**
```go
count := slice.From(users).
    KeepIf(User.IsActive).
    Len()
```

**What counts toward multiline:**
- `KeepIf`, `RemoveIf`, `Convert`, `Len`, `Each` — these count
- `slice.From()`, `slice.MapTo[R]()` — setup, doesn't count
- `ToString`, `ToInt`, `ToFloat64` — data extraction, chains onto previous line

**Why this matters:**
By convention, data transformation operations don't add lines—they chain onto existing structure. A single filter is one line (vs 5-7 for a loop). Adding a second operation (filter-then-count) adds one line, not another 5-7. This is the composability advantage: conventional patterns cost 4×N; FP operations cost entry + N.

## H. Real-World Loop Bugs

Loop mechanics create opportunities for error regardless of developer experience. These categories were found in production code:

### Index Arithmetic
```go
// Bug: i+i instead of i+1 (typo doubles the index)
p.Attributes = append(p.Attributes[:i], p.Attributes[i+i:]...)
```
FluentFP: No manual index math—`RemoveIf` handles element removal.

### Accumulator Assignment
```java
// Bug: passed 0 instead of accumulator, never incremented
page = getAllEntities(0, pageSize, cond);  // should be: start
// missing: start += page.getItems().size();
```
FluentFP: `Fold` manages the accumulator automatically.

### Iterator Bounds
```java
// Bug: assumes iterator has 3+ elements without checking
Iterator<String> parts = splitter.split(input).iterator();
String first = parts.next();   // assumes element exists
String second = parts.next();  // assumes element exists
String third = parts.next();   // assumes element exists
```
FluentFP: No manual iteration—element access is bounds-checked.

### Off-by-One
```c
// Bug: <= iterates one past array end (0-indexed)
for (i = 0; i <= num_channels; i++) {
    channels[i] = init_channel();  // accesses channels[num_channels] - OOB!
}
```
FluentFP: No manual bounds—iteration is over the collection itself.

### Loop Termination
```java
// Bug: no progress detection causes infinite loop
while (inflater.getRemaining() > 0) {
    inflater.inflate(buffer);  // what if inflate returns 0?
}
```
FluentFP: No while loops—operations are bounded by collection size.

### Defer in Loop (Go)
```go
// Bug: defer accumulates N times, all execute at function end
for _, item := range items {
    ctx, cancel := context.WithTimeout(parentCtx, timeout)
    defer cancel()  // leaks until function returns
}
```
FluentFP: No loop body reduces (but doesn't eliminate) misplacement risk.

**What FluentFP eliminates:**
- No accumulators to forget (`Fold` handles it)
- No manual indexing (`KeepIf`/`RemoveIf`)
- No index arithmetic (predicates operate on values)
- No manual iteration (no `.next()` calls)
- No off-by-one in bounds (iterate collection, not indices)

**What FluentFP reduces but doesn't eliminate:**
- Defer misplacement (no loop body, but still possible elsewhere)

**What FluentFP does NOT prevent:**
- Predicate logic errors—the user writes that logic either way

**Why this matters:**
These aren't junior developer mistakes. Off-by-one bugs made it into the Linux kernel—likely some of the most reviewed patches anywhere—and they inevitably recur. The same error pattern appears across kernel releases years apart. If the construct allows an error, it will eventually happen; loop mechanics errors are inherent to the construct itself.

## I. Performance Analysis

The [Performance Characteristics](analysis.md#performance-characteristics) section is based on static analysis of the fluentfp source code, not runtime benchmarks.

### Source Code Evidence

**Mapper type definition** (`slice/types.go`):
```go
type Mapper[T any] []T
```

`Mapper[T]` is a defined type with underlying type `[]T`—not a wrapper struct, not a lazy evaluator. Operations execute immediately.

**KeepIf allocation** (`slice/mapper.go:30-39`):
```go
func (ts Mapper[T]) KeepIf(fn func(T) bool) Mapper[T] {
    results := make([]T, 0, len(ts))  // Allocation happens here
    for _, t := range ts {
        if fn(t) {
            results = append(results, t)
        }
    }
    return results
}
```

Every `KeepIf` call allocates a new slice. Chaining `KeepIf(...).ToString(...)` produces two allocations.

**Exception: TakeFirst** (`slice/mapper.go:60-66`):
```go
func (ts Mapper[T]) TakeFirst(n int) Mapper[T] {
    if n > len(ts) {
        n = len(ts)
    }
    return ts[:n]  // Slice bounds—no allocation
}
```

`TakeFirst` returns a slice view, not a copy. This is the only non-allocating filter operation.

### Implications

| Operation | Allocations |
|-----------|-------------|
| `slice.From(xs)` | 0 (type conversion) |
| `.KeepIf(...)` | 1 |
| `.ToString(...)` | 1 |
| `.TakeFirst(n)` | 0 |
| `Fold(...)` | 0 (accumulator only) |

A 3-operation chain like `From(xs).KeepIf(p).Convert(f).ToString(g)` allocates 3 slices. A fused manual loop allocates 1.

### Benchmark Results

Runtime benchmarks comparing fluentfp chains vs properly-written loops (both pre-allocate; 1000 elements, Intel i5-1135G7):

```
Filter only (KeepIf vs pre-allocated loop):
BenchmarkFilter_Loop_1000     5625 ns/op   32768 B/op    1 allocs/op
BenchmarkFilter_Chain_1000    5524 ns/op   32768 B/op    1 allocs/op

Filter + Map (chain vs fused pre-allocated loop):
BenchmarkFilterMap_Loop_1000   3102 ns/op   16384 B/op    1 allocs/op
BenchmarkFilterMap_Chain_1000  7629 ns/op   40960 B/op    2 allocs/op

Count after filter+map (when result slice not needed):
BenchmarkFilterMapCount_Loop_1000    259 ns/op       0 B/op    0 allocs/op
BenchmarkFilterMapCount_Chain_1000  7599 ns/op   40960 B/op    2 allocs/op
```

**Key findings:**

1. **Single-operation chains are equal:** When loops properly pre-allocate (as they should when input size is known), single operations are equivalent. The ~2% difference is within measurement noise.

2. **Multi-operation chains have overhead:** Filter+Map chain is ~2.5× slower and uses ~2.5× memory vs a fused loop. Each chain operation allocates a new slice; a fused loop allocates once.

3. **Fused loops win dramatically when you don't need intermediate results:** Counting without building a result slice shows 29× speedup for loops (259 ns vs 7599 ns). The chain allocates slices it then discards.

**Guidance:**

- For single operations: fluentfp equals properly-written loops
- For multi-operation pipelines: expect 2-3× overhead in hot paths
- For count/reduce operations: use `Fold` (single pass, no intermediate allocation) or a fused loop
- Profile actual hot paths before optimizing—the overhead is measurable but rarely dominant

Source: `slice/benchmark_test.go`. Run with `go test -bench=. -benchmem ./slice/`
