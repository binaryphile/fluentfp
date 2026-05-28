# Code Shape: Conventional vs fluentfp

Two pairs of files for *seeing* how fluentfp changes the shape of code. Each pair implements the same task two ways — open them side-by-side and the difference is immediate. The visualizations below render each pair at small zoom, so total bulk and indent pattern read at a glance without the syntax demanding attention.

Code shape correlates with complexity because both come from the same source: nested control structures. Counting indent levels at a glance estimates branch points without running a tool. See [analysis.md § The Principle](../../analysis.md#the-principle) for the full argument.

---

## Pair 1 — Mixed code (typical case)

Mirrors a typical production ratio: ~36% of operations are filter/map/fold-convertible; the rest stay as conventional loops with break/continue/error returns.

![Mixed code shape comparison](../../images/code-shape-comparison.png)

| File | Code | Complexity |
|---|---:|---:|
| [conventional.go](conventional.go) | 91 | 23 |
| [fluentfp.go](fluentfp.go) | 80 | 17 |
| **Reduction** | **−12%** | **−26%** |

12% fewer lines, 26% fewer branch points — the convertible functions shed their `for`/`if`; the rest unchanged. Complexity is scc's count of branch and loop tokens; see [methodology.md § F](../../methodology.md#f-code-metrics-tool-scc).

### Error surfaces reduced — and preserved (Pair 1)

The first three converted functions (`getActiveUsers`, `getEmails`, `countAdmins`), and most of the fourth (`averageAge`), replace **accumulator bookkeeping**. Each conventional version declares a `result` slice or `count` int and feeds it inside a loop body. The most-quoted Go-shaped failures of that pattern:

- Forgotten `result = ` on `append` — works until the backing array reallocates, then silently drops elements.
- `count++` placed outside the `if` — counts everything instead of the predicate match.

`KeepIf`, `ToString`, `Len`, and `Fold` create the accumulator internally; in these four call sites, those mistakes are no longer expressible. (`averageAge` carries the additional risks of divide-by-zero and integer overflow, which fluentfp does not address — both forms have those risks identically.)

The seven functions kept as loops (5–11) retain other risk surfaces. The lines below point into `conventional.go`:

| Function | Risk preserved | Why fluentfp doesn't replace it |
|---|---|---|
| `findByEmail` (line 64) | explicit `users[i]` indexing — index-arithmetic class | the function requires a pointer to the original element; chains return values |
| `processWithRetry` (line 83) | nested-loop `break` — premature-exit class | multi-level imperative control flow with success-dependent early exit |
| `validateSequentialIDs` (line 94) | `i+1` arithmetic — index-arithmetic class | possible via `seq.Enumerate` + `Every`, but reads less clearly than the loop |
| `deactivateAll` (line 111) | input slice mutation | fluentfp deliberately returns new slices; a non-mutating equivalent would change the function's contract |

The examples therefore preserve both reduced and unreduced error surfaces. See [analysis.md § Error Prevention](../../analysis.md#error-prevention) for the broader category table (index typo, defer-in-loop, error shadowing, input mutation) the kept loops illustrate.

---

## Pair 2 — Pure data pipeline (best case)

A report generator with no I/O, only data transformations. The ceiling for fluentfp's impact.

![Best-case code shape comparison](../../images/best-case-code-shape-comparison.png)

| File | Code | Complexity |
|---|---:|---:|
| [best-case-conventional.go](best-case-conventional.go) | 281 | 57 |
| [best-case-fluentfp.go](best-case-fluentfp.go) | 148 | 3 |
| **Reduction** | **−47%** | **−95%** |

When every operation fits filter/map/fold, complexity drops from 57 to 3 — every `for` and `if` is gone; chains have no branch points to count. The remaining complexity-3 is three operators inside predicates (one `&&`, two `==`), not iteration or branching syntax. Because the conversion is total, Pair 2 has no preserved-risk table to enumerate — the bug-category catalog reduces to zero applicable rows.

---

## Indent tracks complexity

Indentation correlates with complexity because both trace to the same source. Verified on these same files using `scc` for complexity and `awk` for tab-sum (per [analysis.md § Measuring the Correlation](../../analysis.md#measuring-the-correlation)):

| Pair | Indent change | Complexity change |
|---|---:|---:|
| Mixed | −26% | −26% |
| Best-case | −80% | −95% |

In the mixed pair the two metrics drop identically — here, indent is an effective 1:1 estimator. In the pure-pipeline pair complexity drops *faster* than indent — a multi-line chain can have visual indentation with zero branch points. Indent is the conservative eyeball estimate.

---

## What to look for when reading the pairs

**Chain vs nested block.** A filter+map+extract in `fluentfp.go` is one expression on three lines, all at the same indent. The same task in `conventional.go` is a `var` declaration and a `for` header at the function indent, then a nested `if` one level deeper, then an `append` one level deeper still — three indents in four lines. The stairstep is the shape difference the visualizations expose.

**Where the loop stays.** `sendNotifications` in `fluentfp.go` keeps the imperative `for` body because it has an early return on error. fluentfp doesn't try to replace every loop, only the mechanical ones. The library is honest about its scope; see the "lower-yield code patterns" table in [analysis.md § Trade-offs](../../analysis.md#trade-offs).

**Method expressions.** `User.IsActive` and `User.GetEmail` plug directly into `KeepIf` and `ToString` — the receiver-as-first-arg shape matches the higher-order signature. No wrapper closures, no `func(u User) bool { return u.IsActive() }` adapters.

**Density.** The fluentfp files spend more lines per unit of intent on the chains (which carry meaning) and fewer on mechanics (variable declarations, loop headers, closing braces). In the mixed pair, fluentfp removes 11 lines while preserving the same behavior, primarily by eliminating those mechanical lines. [Methodology § B-C](../../methodology.md#b-line-classification-rules) defines the semantic-vs-syntactic line rules used to compute this.

---

## Reproducing the numbers

The files build with `//go:build ignore` — they don't compile as a package. Measure directly:

```bash
# Code + complexity per file (glob excludes this README)
scc --by-file examples/code-shape/*.go

# Indent sum (tabs)
awk 'BEGIN{s=0} {n=0; while(substr($0,n+1,1)=="\t") n++; s+=n} END{print s}' \
  examples/code-shape/conventional.go
```

scc reports the same code/complexity numbers shown in the tables above. The awk script gives per-file tab-sums; compute the percentage from any pair (e.g., (72−97)/97 = −26% for the mixed pair).

---

## See also

- [analysis.md](../../analysis.md) — full argument for code shape as a complexity proxy, plus categories of loop-mechanics bugs that fluentfp eliminates.
- [methodology.md](../../methodology.md) — measurement rules: line classification, scc usage, chain formatting conventions.
- [docs/showcase.md](../../docs/showcase.md) — 24 before/after rewrites of real-world code from public Go projects.
