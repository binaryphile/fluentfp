# fluentfp Use Cases

## System Scope

**System:** fluentfp
**In scope:** Collection transformation, lazy sequence processing (memoized and iterator-native), optional value handling (including JSON/SQL marshaling), typed alternatives, invariant enforcement, conditional value selection, builtin function adapters for higher-order use, function composition, function-level concurrency and control-flow wrappers, memoization, persistent min-heap, combinatorics, constrained stage processing (bounded worker with backpressure and stats), pipeline composition (multi-stage with error passthrough, fan-out/fan-in, and per-stage observability)
**Out of scope:** General concurrency primitives (channels, mutexes, goroutine lifecycle), I/O, error handling strategies, logging. Note: bounded concurrent traversal (`FanOut`) is in scope as a collection operation — it transforms a slice concurrently, not a general concurrency primitive. Constrained stage processing (`toc`) is in scope as a pipeline building block — it runs items through a known bottleneck with observability, not a general job queue.

## System Invariants

- **Immutability by default** — Operations produce new collections; inputs are never modified
- **Order preservation** — Operations preserve encounter order when the source has a defined order, unless explicitly sorting or randomizing. Map and set sources yield iteration order (unspecified by Go).
- **Nil safety** — Nil or empty collection inputs are handled without library-originated panics. Panics from user-supplied callbacks are outside this guarantee unless explicitly recovered (e.g., FanOut).
- **Type safety** — APIs use Go's static typing and generic constraints to prevent library-level runtime type mismatches.
- **Interoperability** — Collection types are convertible to/from standard Go types (`[]T`, `map[K]V`) without allocation. Results work with `range`, `len`, `append`, and standard library functions.

## System-in-Use Story

> Alex, maintaining a fleet management service, needs to find active devices, extract their signal strengths, and compute the average. Instead of writing a loop with index tracking and an accumulator variable, Alex describes the pipeline as business logic: keep the active devices, extract their signal strengths, sum and count them. When a nullable database field causes a nil pointer panic in staging, Alex makes the absence explicit with a sensible default — the crash disappears and the code documents the possibility. A colleague reviewing the PR reads the intent without tracing control flow through if-else branches.

## Actor-Goal List

### Go Developer

**Characterization:** Professional Go developer writing business logic, familiar with Go idioms, varying experience with functional programming concepts

| Goal | Level | Priority |
|------|-------|----------|
| Transform a collection into a new collection | Blue | high |
| Derive a single result from a collection | Blue | high |
| Handle values that might be absent | Blue | high |
| Model a value that is one of two typed outcomes | Blue | med |
| Enforce invariants during initialization | Blue | med |
| Select a value conditionally with a fallback | Blue | low |
| Construct reusable functions from existing ones | Blue | low |
| Process a memoized persistent sequence on demand | Blue | med |
| Process an iterator-native sequence on demand | Blue | med |
| Generate combinatorial selections from a collection | Blue | low |
| Cache expensive function results | Blue | low |
| Maintain a persistent priority queue | Blue | low |
| Process items through a known bottleneck with bounded concurrency and observability | Blue | med |
| Compose multi-stage processing pipelines with per-stage observability and error passthrough | Blue | med |
| Replace manual loop patterns with composable operations across the codebase | White | — |

## Use Cases

### UC-1: Transform a Collection

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct output with expected elements, types, ordering
- Code reviewer: transformation reads as intent, not iteration mechanics

**Postconditions:**
- A new collection exists with desired elements in expected order
- Original collection is unmodified

**Minimal Guarantee:** Original collection is never modified, regardless of transformation outcome.

**Preconditions:**
- Developer has a slice or map

**Main Scenario:**
1. Developer selects the collection source.
2. Developer specifies transformations: filtering by criteria, converting elements, changing element types, reordering, expanding, deduplicating, or limiting count.
3. System applies each transformation in sequence, returning a chainable collection.
4. System returns the final collection.

**Extensions:**
- 1a. Collection is nil or empty: System produces a valid empty collection.
- 1b. Collection source is a map: System extracts the map's values as a collection for further transformation.
- 1c. Collection source is a set (`map[T]struct{}`): System extracts the set's members as a collection for further transformation.
- 1d. Developer needs to filter or transform map entries while preserving map structure: System applies predicates or value transforms to entries, returning a map for further map-level operations or value extraction.
- 1e. Collection source is a combinatorial construction: See UC-11. Combinatorial results are chainable `Mapper` collections.
- 1f. Developer needs map entries as a flat collection of key-value pairs, or needs to construct a map from pairs: System converts between map and pair-slice representations. When constructing from pairs with duplicate keys, the last pair wins.
- 2a. Developer needs to expand each element into multiple: System applies expansion and concatenates in order. When the expansion produces a different type, the standalone variant infers both types.
- 2b. Developer needs duplicates removed by extracted key: System removes duplicates preserving first occurrence, using a caller-provided key function.
- 2c. Developer needs a sorted copy: System produces sorted collection; original unchanged.
- 2d. Developer needs elements grouped by a derived key: System groups elements into a chainable collection of groups, each containing a key and the elements sharing that key. Groups are ordered by first occurrence of each key; elements within each group preserve encounter order.
- 2e. Developer needs to combine corresponding elements from two slices: System combines elements pairwise, either into pairs or through a provided function. Panics if slices differ in length — length correspondence is a caller precondition.
- 2f. Developer needs transformations applied concurrently: System applies transformations concurrently with bounded parallelism, preserving element order in the result. Reports success or failure per element via `Result`, recovers panics as `PanicError`, and respects context cancellation. Bounding is by item count (uniform cost) or by total cost (weighted).
- 2g. Developer needs a shallow copy with independent backing array: System produces a copy whose backing array is not shared with the original. Element values are not deep-copied.
- 2h. Developer needs zero-value elements removed from a collection: System removes all elements equal to their type's zero value and returns the remaining elements. For string collections, the developer may use a string-specific variant that reads as "non-empty" for clarity.
- 2i. Developer needs to split a collection into fixed-size batches: System divides the collection into sub-collections of the specified size; the last batch may be smaller.
- 2j. Developer needs elements in random order: System produces a randomly shuffled copy of the collection using `math/rand/v2`.
- 2k. Developer needs a random subset of elements: System selects count random elements without replacement; if count exceeds length, returns all elements in random order.
- 2l. Developer needs to filter and transform in one step: System applies a function that returns both a transformed value and a keep/discard signal (`func(T) (R, bool)`). Elements where the function returns false are excluded; kept elements appear transformed in the result.
- 2m. Developer needs each element paired with its positional index: System pairs each element with its zero-based index, producing a collection of index-element `pair.Pair` values suitable for further transformation.
- 2n. Developer needs duplicate comparable elements removed while preserving first occurrence: System removes duplicates by comparable equality. No key function is required.

**Sub-Variations:**
- Filtering: inclusion-based, exclusion-based, conditional (`KeepIfWhen`/`RemoveIfWhen` — no-op when condition is false), zero-value removal, or empty-string removal
- Type conversion: to built-in types or to arbitrary types
- Sorting: ascending or descending by extracted key
- Deduplication: by comparable equality or by extracted key
- Batching: by fixed size
- Concurrent bounding: by item count (uniform cost) or by total cost (weighted)
- Randomization: full shuffle, random subset without replacement
- Combinatorial: permutations, combinations, power sets, Cartesian products (with optional mapping during generation)
- Filter+transform: combined in single pass with keep/discard signal
- Index pairing: each element paired with its zero-based position

---

### UC-2: Derive a Result from a Collection

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct scalar, optional, or aggregate result
- Code reviewer: derivation reads as intent, not accumulation mechanics

**Postconditions:**
- Result correctly summarizes, extracts from, or operates on the collection
- Original collection is unmodified

**Minimal Guarantee:** Original collection is never modified.

**Preconditions:**
- Developer has a collection

**Main Scenario:**
1. Developer selects the collection to derive from.
2. Developer specifies the derivation: combining elements progressively, finding a specific element, checking a condition, counting, or summing.
3. System processes the collection and returns the result.

**Extensions:**
- 1a. Collection is empty: System returns the appropriate empty result — zero for sums/counts, absence for lookups, initial value for accumulations, false for any-match checks, true for all-match checks, true for no-match checks, false for membership checks.
- 2a. Developer searches for first matching element: System returns the match as an option, or not-ok if absent.
- 2b. Developer searches for first matching element from the end: System returns the last match as an option, or not-ok if absent.
- 2c. Developer searches for the position of the first matching element: System returns the index as an option, or not-ok if no element matches.
- 2d. Developer searches for the position of the last matching element: System returns the index as an option, or not-ok if no element matches.
- 2e. Developer expects exactly one element: System returns it via `Either` — right with the value if exactly one, left with the actual count otherwise.
- 2f. Developer needs multiple fields extracted simultaneously: System returns one collection per field.
- 2g. Developer needs to accumulate state while also producing per-element output: System processes elements in order and returns both the final accumulated value and the per-element outputs.
- 2h. Developer needs to convert the collection to a set for membership checks: System returns a `map[K]struct{}` of extracted keys (or elements themselves when the element is the key).
- 2i. Developer checks whether all elements satisfy a criterion: System tests every element and returns true only if all match (short-circuits on first failure).
- 2j. Developer checks whether a specific comparable value exists in the collection: System tests membership by equality and returns true if found.
- 2k. Developer checks that no elements satisfy a criterion: System tests every element and returns true only if none match (short-circuits on first match).
- 2l. Developer needs elements indexed by a derived key for O(1) lookup: System produces a `map[K]T` from extracted keys to elements.
- 2m. Developer needs a random element from a collection: System returns a random element as an option, or not-ok if the collection is empty. Uses `math/rand/v2`.
- 2n. Developer needs the element with the minimum or maximum value of an extracted key: System extracts a comparable key from each element using `cmp.Compare` and returns the element with the smallest or largest key as an option, or not-ok if the collection is empty.
- 2o. Developer needs to combine elements without providing an initial value: System uses the first element as the seed and applies the combining function across remaining elements from left to right. Returns an option — not-ok if the collection is empty.
- 2p. Developer needs to build a map with both key and value derived from each element: System applies a function that returns a key-value pair for each element, producing a `map[K]V`. If multiple elements produce the same key, the last one wins.
- 2q. Developer needs to apply a side effect to each element: System calls the function for every element in order. No result is returned.

**Sub-Variations:**
- Numeric aggregation: sum, min, max on integer or floating-point collections
- Element search: first element, last element, first matching, last matching, index of first matching, index of last matching, random element
- Condition checks: any match, all match, no match, membership
- Multi-field extraction: 2, 3, or 4 fields simultaneously
- Indexing: by extracted key for O(1) lookup
- Extremum by key: element with smallest or largest extracted comparable key
- Reduction: combine elements using first element as seed
- Map construction: key-value pairs derived from elements

---

### UC-3: Handle Absent Values

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: absent values handled consistently without scattered nil/zero checks
- Code reviewer: absence handling explicit at point of use

**Postconditions:**
- Developer has either the value (if present) or an appropriate fallback
- Absence is handled explicitly, not silently ignored

**Minimal Guarantee:** No silent zero-value substitution — absence is always distinct from a present zero value.

**Preconditions:**
- Developer has a value that might be absent

**Main Scenario:**
1. Developer wraps the value as optional, specifying what determines presence.
2. Developer transforms or extracts: providing a default, applying logic only when present, converting type, or filtering by additional criteria.
3. Developer uses the resolved value.

**Extensions:**
- 1a. Value comes from a pointer (nil means absent): System extracts the pointed-to value when non-nil.
- 1b. Value is a zero value that should mean absent: System treats zero as absent.
- 1c. Value is present but needs transformation to a different type: System checks presence and transforms in one step, returning absence if the original was absent.
- 1d. Value comes from a map lookup: System wraps the comma-ok result as an option.
- 1e. Value comes from an environment variable: System treats unset or empty as absent.
- 2a. Developer needs a side effect only when present: System calls the function only when present; does nothing when absent.
- 2b. Developer needs a side effect only when absent: System calls the function only when absent; does nothing when present.
- 2c. Developer needs to filter an already-present value: System applies filter, converting to absent if not met.
- 2d. Fallback is expensive to compute: System evaluates fallback only when absent.
- 2e. Developer needs to combine two optional values when both are present: System applies a combiner function to both values when both are present, or returns absent if either is absent.
- 2f. Developer needs a lazily computed default while remaining in the optional context for further chaining: System calls a function only when the value is absent, producing a present result. When already present, the value passes through unchanged.
- 2g. Developer needs to chain operations that each may produce absence: System applies each operation in sequence, short-circuiting to absent if any step produces absence. No manual unwrapping between steps.
- 2h. Developer needs a multi-level fallback chain staying in the optional context: System tries each fallback in order, short-circuiting on the first that produces a present value.
- 3a. Developer stores optional value in a database column: System implements `driver.Valuer` and `sql.Scanner` — present maps to the column value, absent maps to SQL NULL.
- 3b. Developer serializes optional value to JSON: System implements `json.Marshaler` and `json.Unmarshaler` — present maps to the JSON value, absent maps to `null`. Note: round-tripping collapses `Ok(nil)` and `NotOk` into the same representation.

**Sub-Variations:**
- Specialized type aliases for common types: `String`, `Int`, `Bool`, `Error`, etc.
- Pre-declared not-ok values: `NotOkString`, `NotOkInt`, etc.
- Construction from: direct value (`Of`), value-and-presence pair (`New`), pointer (`NonNil`), zero-value check (`NonZero`), empty-string check (`NonEmpty`), error check (`NonErr`), map lookup (`Lookup`), environment variable (`Env`), condition (`When`, `WhenCall`)
- Create + transform: check presence and map to a new type in one call (`NonZeroCall`, `NonEmptyCall`, `NonNilCall`)
- Combining: two optional values via combiner function (`ZipWith`)
- Recovery: lazily computed default staying in optional context (`OrWrap`, `OrElse`)

---

### UC-4: Model Typed Alternatives

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: two possible outcomes handled exhaustively with correct types

**Postconditions:**
- Developer has processed the value through the appropriate branch
- Type system prevents accessing the wrong branch's value unsafely

**Minimal Guarantee:** Accessing the wrong branch returns a zero value and false, never corrupts state.

**Preconditions:**
- Developer has an operation producing one of two typed outcomes

**Main Scenario:**
1. Developer constructs the value indicating which branch it represents.
2. Developer processes: extracting with a default, applying branch-specific logic, or handling both branches to produce a unified result.

**Extensions:**
- 1a. Developer has a fallible function returning `(R, error)` and needs it to return `Result` instead: System wraps the function via `rslt.Lift`, producing a new function with the same input that returns a `Result`.
- 2a. Developer needs both branches handled with different logic, producing unified result: System applies the appropriate branch function via `Fold`.
- 2b. Developer needs to transform only the success branch: System transforms via `Convert` (same type) or `Map` (cross-type), passing failure through.
- 2c. Developer needs to chain operations that each may fail: System applies each operation in sequence via `FlatMap`, short-circuiting to failure if any step fails. No manual error checking between steps.

**Sub-Variations:**
- Either: general two-branch sum type, left = failure convention, right = success convention
- Result: specialized `Either[error, T]` with `Ok`/`Err` constructors, `PanicError` for recovered panics
- Collectors: `CollectAll`, `CollectOk`, `CollectErr`, `CollectOkAndErr` — batch result processing returning plain slices
- Extraction: value-and-presence pair (`Get`), default value (`Or`), lazy default (`OrCall`)

---

### UC-5: Enforce Initialization Invariants

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: program fails immediately when preconditions violated
- Operator: initialization failures surface at startup, not under load

**Postconditions:**
- All values available without error checking downstream
- If any precondition violated, program panicked with clear error before operational code

**Minimal Guarantee:** A violated precondition always panics. No silent continuation with invalid state.

**Preconditions:**
- Developer has initialization steps that return `(T, error)` or `error`

**Main Scenario:**
1. Developer wraps each step with `must.Get` (for `(T, error)`) or `must.BeNil` (for `error`).
2. System executes; if any step returns a non-nil error, system panics immediately with that error.
3. Developer uses resulting values without further error handling.

**Extensions:**
- 1a. Developer needs to wrap a fallible function for repeated use: System returns a new function via `must.Of` that panics on error on every call. Panics immediately if given a nil function.
- 1b. Developer needs a required environment variable: System reads it via `must.Env`, panics if missing or empty.

---

### UC-6: Select Values Conditionally

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: conditional logic expressed as value selection rather than control flow

**Postconditions:**
- Developer has the correct value: preferred when condition holds, fallback otherwise

**Minimal Guarantee:** The unused branch's computation is never evaluated when deferred.

**Preconditions:**
- Developer has a preferred value, a fallback value, and a selection condition

**Main Scenario:**
1. Developer specifies the preferred value and condition via `option.When` (eager) or `option.WhenCall` (lazy).
2. Developer specifies the fallback value via `.Or` (eager) or `.OrCall` (lazy).
3. System evaluates: if condition holds, returns preferred value; otherwise returns fallback.

**Extensions:**
- 1a. Preferred value is expensive to compute: System defers computation via `WhenCall` until condition confirmed true. If false, expensive computation never runs.

**Sub-Variations:**
- Eager: value computed before condition check — `When(cond, val).Or(fallback)`
- Lazy: value computed only when condition true — `WhenCall(cond, fn).Or(fallback)`
- Note: `cmp.Or` (Go 1.22+ stdlib) covers first-non-zero selection for comparable types; fluentfp does not duplicate this.

---

### UC-7: Construct Reusable Functions

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: new function behaves correctly, types checked at compile time
- Code reviewer: construction intent is clear from combinator name

**Postconditions:**
- A new function exists with the combined behavior
- Original functions are unmodified

**Minimal Guarantee:** Original functions are never modified. Constructed function is type-safe — mismatched signatures fail at compile time, not runtime.

**Preconditions:**
- Developer has existing functions to combine

**Main Scenario:**
1. Developer combines functions using composition, partial application, or standard building blocks.
2. System returns a new function with the combined behavior.

**Extensions:**
- 1a. Developer needs left-to-right composition of two unary transforms: System composes them via `Pipe` so the first feeds into the second.
- 1b. Developer needs to fix one argument of a two-argument function: System returns a one-argument function with the fixed argument captured. `Bind` fixes the first argument; `BindR` fixes the second.
- 1c. Developer needs to apply separate functions to separate arguments: System applies each function independently via `Cross` and returns the results as a pair.
- 1d. Developer needs an identity function: System provides `hof.Identity`, which returns its argument unchanged.
- 1e. Developer needs a predicate that checks equality to a known value: System returns a `func(T) bool` via `Eq` that tests its argument against the captured value.
- 1f. Developer needs to pass a Go builtin as a higher-order argument: `lof` provides first-class function wrappers for builtins (`Len`, `Println`, `HasPrefix`, `Inc`, etc.).
- 1g. Developer needs to bound concurrent access to a function: `Throttle` returns a function with the same signature (`func(A) (R, error)`) that blocks callers until a semaphore slot is available. `ThrottleWeighted` bounds by per-call cost rather than count.
- 1h. Developer needs a side-effect triggered when a function call returns an error: `OnErr` returns a function with the same signature (`func(A) (R, error)`) that calls the original, then invokes the handler if the error is non-nil.
- 1i. Developer needs to retry a function on failure with configurable delays: `Retry` returns a function with the same signature (`func(context.Context, A) (R, error)`) that retries on error according to a backoff strategy (`ConstantBackoff` or `ExponentialBackoff` with full jitter), respecting context cancellation during waits.
- 1j. Developer needs to coalesce rapid calls, executing once after activity stops: `NewDebouncer` creates a stateful debouncer. `Send` stores the latest value and resets a quiet-period timer. After the quiet period elapses, the callback executes with the stored value. `MaxWait` caps total deferral. The debouncer must be closed via `Close` (or deferred) to release its goroutine.

**Sub-Variations:**
- Composition: left-to-right (`Pipe`)
- Partial application: fix first arg (`Bind`), fix second arg (`BindR`)
- Building blocks: identity function (`Identity`), equality predicate (`Eq`)
- Builtin adapters (`lof`): `Len`, `Println`, `HasPrefix`, `HasSuffix`, `Contains`, `Inc`
- Concurrency control: by count (`Throttle`), by cost (`ThrottleWeighted`)
- Side-effect on error (`OnErr`)
- Retry with backoff: constant delay (`ConstantBackoff`), exponential with full jitter (`ExponentialBackoff`)
- Call coalescing: trailing-edge debounce with optional maximum wait (`NewDebouncer`)

---

### UC-8: Process a Memoized Stream

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct elements processed without materializing entire sequence
- Code reviewer: lazy pipeline reads as intent, evaluation boundaries are clear

**Postconditions:**
- Requested elements have been produced or processed
- Full sequence was not required to exist in memory simultaneously
- Source stream is unchanged and can be reused (persistence via structural sharing)

**Minimal Guarantee:** Partially consumed streams remain valid for further operations. Panicking thunks reset to pending for retry; no corrupted cell is cached.

**Preconditions:**
- Developer has a source that is large, infinite, or expensive to compute and benefits from memoized traversal

**Main Scenario:**
1. Developer constructs a stream from a source.
2. Developer specifies transformations: filtering, converting, limiting count, skipping elements, changing element types, expanding and flattening, or concatenating streams.
3. Developer terminates the pipeline: collecting to a slice, iterating for side effects, searching for a match, checking a condition across all elements, or accumulating a result.

**Extensions:**
- 1a. Stream is empty: System produces a valid empty result for any terminal operation.
- 2a. Source is a slice: System wraps it; elements are produced on demand from the underlying slice.
- 2b. Source is an infinite mathematical series: System generates elements from a seed and step function via `Generate`; the stream never terminates.
- 2c. Source is a constant value repeated indefinitely: System produces the same value on each access via `Repeat`.
- 2d. Source is a step function with termination: System unfolds from a seed via `Unfold`, producing elements until the step function returns not-ok.
- 2e. Source is a step function that always produces an element: System produces an element from each step via `Paginate`; an optional next-state controls whether to continue. Every step emits, including the last.
- 2f. Source is a recursive definition: System accepts a head value and a deferred tail computation via `Prepend`/`PrependLazy`, building the stream lazily.
- 2g. Developer needs to chain two streams end-to-end: System produces all elements from the first, then all from the second, via `Concat`.
- 3a. Developer needs cross-type transformation: System transforms elements to a different type via standalone `Map`.
- 3b. Developer needs to expand each element into a sub-stream and flatten: System applies expansion lazily via standalone `FlatMap`.
- 3c. Developer needs to combine corresponding elements from two streams: System pairs elements via `Zip`/`ZipWith`, truncating to the shorter stream.
- 3d. Developer needs running accumulator values as a stream: System produces the initial value followed by each intermediate accumulation via standalone `Scan`.
- 4a. Developer needs to bridge to a Go range loop: `Stream.All` provides an `iter.Seq[T]` for use with Go's range-over-func protocol.
- 4b. Developer needs to bridge to slice operations: `Collect` materializes to a `[]T` for use with eager collection operations.

**Sub-Variations:**
- Construction: `From` (slice), `Of` (variadic), `Generate` (infinite), `Repeat` (constant), `Unfold` (step function), `Paginate` (always-emit step function), `Prepend` (eager cons), `PrependLazy` (deferred cons)
- Filtering: `KeepIf`, `RemoveIf`
- Limiting: `Take` (by count), `TakeWhile` (by predicate)
- Skipping: `Drop` (by count), `DropWhile` (by predicate)
- Transformation: `Convert` (same-type method), `Map` (cross-type standalone)
- Expanding: `FlatMap` (expand and flatten sub-streams)
- Concatenating: `Concat`
- Pairing: `Zip`, `ZipWith`
- Accumulating: `Scan` (running intermediate values)
- Termination: `Collect`, `Each`, `Find`, `Any`, `Every`, `None`, `Fold`

---

### UC-9: Process an Iterator-Native Sequence

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct elements processed with minimal overhead
- Code reviewer: pipeline reads as intent, compatible with Go's range protocol

**Postconditions:**
- Requested elements have been produced or processed
- Pipeline is reusable — each iteration re-evaluates from source (except channel-backed sources, which are consumptive)

**Minimal Guarantee:** Pipelines do not cache intermediate results. Each iteration is independent. A broken iteration (early break) does not corrupt the pipeline.

**Preconditions:**
- Developer has a source suitable for lightweight, non-memoized lazy processing

**Main Scenario:**
1. Developer constructs a `Seq` from a source.
2. Developer specifies transformations: filtering, converting, filter+transform, limiting count, skipping elements, changing element types, expanding and flattening, concatenating sequences, deduplicating, interspersing separators, batching into chunks, pairing corresponding elements, or accumulating intermediate values.
3. Developer terminates the pipeline: collecting to a slice, iterating for side effects, searching for a match, checking membership, or reducing to a single value.

**Extensions:**
- 1a. Sequence is empty: System produces a valid empty result for any terminal operation.
- 2a. Source is a slice: System wraps it via `From`.
- 2b. Source is an `iter.Seq[T]`: System wraps it via `FromIter`.
- 2c. Source is a step function with termination: System unfolds from a seed via `Unfold`, producing elements until the step function returns not-ok.
- 2d. Source is a Go channel: System creates a `Seq` from a receive channel via `FromChannel`. Each iteration step blocks on receive. The sequence ends when the channel is closed or the provided context is canceled. Cancellation is best-effort — one additional value may be yielded if channel receive and cancellation are simultaneously ready.
- 2e. Developer needs to chain two sequences end-to-end: System produces all elements from the first, then all from the second, via `Concat`.
- 3a. Developer needs cross-type transformation: System transforms elements to a different type via standalone `Map`.
- 3b. Developer needs to expand each element into a sub-sequence and flatten: System applies expansion lazily via standalone `FlatMap`.
- 3c. Developer needs to combine corresponding elements from two sequences: System pairs elements via `Zip`/`ZipWith`, truncating to the shorter sequence.
- 3d. Developer needs running accumulator values as a sequence: System produces the initial value followed by each intermediate accumulation via standalone `Scan`.
- 3e. Developer needs each element paired with its positional index: System lazily pairs each element with its zero-based index via `Enumerate`.
- 3f. Developer needs to filter and transform in one step: System applies a `func(T) (R, bool)` via `FilterMap`. Elements returning false are excluded; kept elements appear transformed.
- 3g. Developer needs to combine elements without providing an initial value: System uses the first element as the seed via `Reduce`. Returns an option — not-ok if the sequence is empty.
- 3h. Developer needs duplicate elements removed while preserving first occurrence: System lazily removes duplicates via `Unique` (by comparable equality) or `UniqueBy` (by extracted key).
- 3i. Developer needs to check whether a sequence contains a specific value: System checks membership via `Contains`, short-circuiting on first match.
- 3j. Developer needs a separator inserted between elements: System lazily inserts a separator via `Intersperse`. Empty and single-element sequences pass through unchanged.
- 3k. Developer needs to process a sequence in fixed-size batches: System groups elements via `Chunk` into slices of the specified size. The last batch may be smaller. Each emitted slice is a stable snapshot safe to retain.
- 4a. Developer needs to bridge to a Go range loop: `Seq.All` provides an `iter.Seq[T]`.
- 4b. Developer needs to bridge to slice operations: `Collect` materializes to a `[]T`.
- 4c. Developer needs to bridge a sequence to a Go channel: `ToChannel` sends values into a new channel via a spawned goroutine. The channel closes when the sequence is exhausted or the context is canceled. Cancellation is cooperative — if the sequence blocks internally before yielding, cancellation cannot interrupt it.

**Sub-Variations:**
- Construction: `From` (slice), `FromIter` (iter.Seq), `Unfold` (step function), `FromChannel` (blocking receive with context)
- Filtering: `KeepIf`, `RemoveIf`, `FilterMap` (filter+transform)
- Limiting: `Take` (by count), `TakeWhile` (by predicate)
- Skipping: `Drop` (by count), `DropWhile` (by predicate)
- Transformation: `Convert` (same-type method), `Map` (cross-type standalone)
- Deduplication: `Unique` (by comparable equality), `UniqueBy` (by key)
- Expanding: `FlatMap`
- Concatenating: `Concat`
- Separating: `Intersperse`
- Batching: `Chunk`
- Pairing: `Zip`, `ZipWith`
- Accumulating: `Scan` (running intermediate values), `Reduce` (no initial value)
- Index pairing: `Enumerate`
- Membership: `Contains` (short-circuit equality check)
- Termination: `Collect`, `Each`, `Find`, `Any`, `Every`, `None`, `Fold`, `Reduce`, `Contains`, `ToChannel`

---

### UC-10: Maintain a Persistent Priority Queue

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: always-available minimum element without manual heap management
- Code reviewer: insertion and extraction read as value operations, not mutable state manipulation

**Postconditions:**
- A new heap exists with the desired elements
- Original heap is unmodified (persistence via structural sharing)

**Minimal Guarantee:** Original heap is never modified. Empty heap operations return absence, not panic.

**Preconditions:**
- Developer has elements that need priority ordering and a comparator function

**Main Scenario:**
1. Developer creates a heap, optionally from existing elements.
2. Developer inserts elements, producing new heaps.
3. Developer queries the minimum or removes it, producing a new heap.

**Extensions:**
- 1a. Heap is empty: `Min` returns not-ok. `DeleteMin` returns (zero, empty heap).
- 2a. Developer needs to merge two heaps: System merges in O(1) via `Merge`, producing a new heap.
- 3a. Developer needs the minimum element without removing it: System returns it as an option via `Min`.
- 3b. Developer needs all elements as a sorted slice: System extracts elements in order via `Collect`.

**Sub-Variations:**
- Construction: `New` (empty), `From` (slice)
- Operations: `Insert`, `Min`, `DeleteMin`, `Merge`, `Len`, `IsEmpty`
- Ordering: caller-provided `func(T, T) int` comparator (use `slice.Asc`/`slice.Desc` with a key function for common cases)

---

### UC-11: Generate Combinatorial Selections

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct combinatorial output without manual recursive or loop-based generation
- Code reviewer: generation reads as a single call, not nested loops

**Postconditions:**
- A chainable collection exists with all requested combinatorial elements
- Input collection is unmodified

**Minimal Guarantee:** Input collection is never modified. Results are eagerly materialized — the caller controls input size.

**Preconditions:**
- Developer has a small collection (combinatorial output grows rapidly)

**Main Scenario:**
1. Developer selects the combinatorial operation and input.
2. System generates all results as a chainable `Mapper` collection.
3. Developer chains further operations (`.KeepIf`, `.Convert`, etc.) on the result.

**Extensions:**
- 1a. Developer needs all orderings: System generates n! permutations via `Permutations`.
- 1b. Developer needs k-element subsets: System generates C(n,k) combinations via `Combinations`.
- 1c. Developer needs all subsets: System generates 2^n subsets via `PowerSet`.
- 1d. Developer needs all pairs from two collections: System generates the Cartesian product via `CartesianProduct`, returning `Mapper[pair.Pair[A,B]]`.
- 1e. Developer needs all pairs mapped to a domain type: System generates the Cartesian product and maps via `CartesianProductWith`, avoiding intermediate pair allocation.
- 2a. Input is nil or empty: `Permutations` and `PowerSet` return `[[]]` (one empty result). `Combinations` returns `[[]]` for k=0, nil for k<0 or k>len. `CartesianProduct` returns nil if either input is empty.

**Sub-Variations:**
- `Permutations`: all orderings (n! results)
- `Combinations`: k-element subsets (C(n,k) results)
- `PowerSet`: all subsets (2^n results)
- `CartesianProduct`: all pairs from two collections
- `CartesianProductWith`: all pairs mapped during generation

---

### UC-12: Memoize Function Results

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: repeated calls return cached results without redundant computation
- Code reviewer: caching boundary explicit at point of wrapping

**Postconditions:**
- Repeated calls with the same input return the same result without re-executing the original function
- Original function is unmodified

**Minimal Guarantee:** If the wrapped function panics, no corrupted result is cached. Future calls retry. Concurrent callers for the same key may execute the function multiple times (no single-flight deduplication).

**Preconditions:**
- Developer has a function whose results are safe to cache (pure or idempotent)

**Main Scenario:**
1. Developer wraps the function with memoization.
2. First call executes the function and caches the result.
3. Subsequent calls with the same input return the cached result.

**Extensions:**
- 1a. Function takes no arguments (deferred initialization): System wraps a zero-arg function via `memo.Of`; first call evaluates and caches; subsequent calls return cached value. Thread-safe.
- 1b. Function is fallible (returns value and error): System caches only successes via `memo.FnErr` — errors trigger retry on subsequent calls.
- 1c. Developer needs bounded cache size: System provides an LRU cache via `memo.NewLRU` that evicts least recently used entries when capacity is exceeded.
- 1d. Developer needs a custom caching strategy: System accepts a caller-provided cache implementing `memo.Cache` (Load/Store). The caller is responsible for thread-safety of the provided implementation.
- 2a. Wrapped function panics: System resets to un-cached state; panic propagates; future calls retry the function.

**Sub-Variations:**
- Zero-arg memoization (`Of`): deferred initialization, `sync.Once` replacement with retry-on-panic
- Keyed memoization (`Fn`, `FnErr`): function caching by input
- Cache strategies: unbounded map (`NewMap`), bounded LRU (`NewLRU`), custom (`Cache` interface with `Load`/`Store`)

---

### UC-13: Process Items Through a Constrained Stage

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: items processed through a known bottleneck with bounded concurrency and backpressure
- Operator: constraint utilization visible via stats — can verify the bottleneck is real and detect downstream shifts

**Postconditions:**
- Every submitted item has a corresponding result
- Stage has shut down cleanly — no goroutine leaks, no abandoned resources
- Stats reflect actual constraint behavior (service time, idle time, output-blocked time, and optionally observed allocation bytes/objects)

**Minimal Guarantee:** Stage always shuts down if the consumer drains output. No goroutine leaks on normal completion, fail-fast, or parent cancellation. Panics in the processing function are recovered as error results, not process crashes.

**Preconditions:**
- Developer has a processing function and a known bottleneck stage in a pipeline
- Developer knows the constraint's concurrency limit (often 1 for serial resources)

**Main Scenario:**
1. Developer starts a stage with a processing function and options (capacity, workers).
2. Developer submits items from one or more goroutines; the stage buffers them up to capacity.
3. Workers dequeue items, process them through the function, and emit results.
4. Developer reads results from the output channel until it closes.
5. Developer calls Wait to confirm clean shutdown and retrieve any stage-level error.

**Extensions:**
- 2a. Buffer is full: Submit blocks the caller until capacity is available (backpressure / "rope").
- 2b. Stage is shut down: Submit returns an error without blocking.
- 2c. Caller's context is canceled while Submit is blocked: Submit returns the context error.
- 3a. Processing function returns an error (fail-fast mode): Stage cancels remaining work, drains buffered items as canceled results, and shuts down. Wait returns the triggering error.
- 3b. Processing function returns an error (continue-on-error mode): Stage continues processing remaining items. The error appears in the individual result.
- 3c. Processing function panics: Stage recovers the panic and emits an error result with the panic value and stack trace. Stage continues processing (or shuts down if fail-fast).
- 3d. Parent context is canceled: Stage stops accepting new items, drains buffered items as canceled results, and shuts down. Wait returns nil; Cause returns the parent's cancel cause.
- 4a. Developer does not need individual results: Developer calls DiscardAndWait or DiscardAndCause to drain and retrieve stage-level status in one step.
- 5a. Developer needs to distinguish success, fail-fast error, and parent cancellation: Developer calls Cause instead of Wait for the latched terminal cause.
- 5b. Developer enables TrackAllocations and checks AllocTrackingActive to confirm the runtime supports it, then reads ObservedAllocBytes/ObservedAllocObjects from Stats as a directional signal for allocation-heavy stages. These are process-global counters sampled around each fn invocation — not exclusive to the stage, not additive across stages.

**Sub-Variations:**
- Capacity: zero (unbuffered — Submit blocks until a worker dequeues), positive (buffered queue)
- Workers: 1 (serial constraint), N (limited concurrency)
- Error mode: fail-fast (default), continue-on-error
- Stats: submitted/completed/failed/panicked/canceled counts, service time, idle time, output-blocked time, buffered depth, in-flight weight
- Allocation tracking: disabled (default), enabled via TrackAllocations
- Shutdown: explicit (CloseInput), fail-fast (automatic), parent cancel (automatic via cancel watcher)

### UC-14: Compose Multi-Stage Processing Pipelines

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: items processed through a chain of constrained stages with per-stage observability, error passthrough, and backpressure
- Operator: per-stage stats reveal which stage is the bottleneck (the DBR "drum")

**Postconditions:**
- Every item from the source has been processed or accounted for (completed, forwarded, or dropped)
- All stages have shut down cleanly — no goroutine leaks, no upstream deadlocks
- Per-stage stats reflect actual constraint behavior independently

**Minimal Guarantee:** Pipeline always shuts down if the tail consumer drains output. Source ownership rule: every library-owned operator drains its upstream to completion, preventing deadlocks on unbuffered channels.

**Preconditions:**
- Developer has multiple processing stages with different concurrency or resource profiles
- Developer knows the constraint topology (which stages are CPU-bound, I/O-bound, serial)

**Main Scenario:**
1. Developer starts a head stage with Start, submitting items from a producer goroutine.
2. Developer creates intermediate stages with Pipe (or NewBatcher for accumulation, or NewWeightedBatcher for weight-based accumulation, or NewTee for broadcast, or NewMerge for fan-in, or NewJoin for branch recombination), each reading from the previous stage's Out().
3. Each Pipe stage's feeder reads upstream results: Ok values go to workers, Err values pass through directly to the output.
4. Developer drains the tail stage's Out() to receive final results (successes and forwarded errors).
5. Developer calls Wait on each stage in reverse order to confirm clean shutdown.

**Extensions:**
- 3a. Upstream error in a Pipe stage: Error bypasses workers and appears directly in the output (data-plane error). The stage continues processing subsequent items. Wait returns nil — forwarded errors do not trigger fail-fast.
- 3b. Pipe stage's fn returns an error (fail-fast mode): Stage cancels its own workers. Upstream stages continue until backpressure stalls them or parent context is canceled. Feeder continues draining source (source ownership rule).
- 3c. Parent context canceled mid-pipeline: All stages observe cancellation. Pipe feeders and Batchers switch to discard mode — continue draining source but discard items. Partial Batcher batches are discarded (not flushed). Stats reflect all drops.
- 2a. Batcher between stages: Batcher accumulates up to n Ok items. Errors act as batch boundaries — flush partial batch, forward error, start fresh accumulator.
- 2b. WeightedBatcher between stages: accumulates items by weight or count (whichever reaches threshold first), preventing unbounded accumulation of zero/low-weight items. Same error-as-batch-boundary semantics.
- 2c. Tee between stages: Developer fans out one stream to multiple downstream paths. All branches observe the same logical sequence. Backpressure is preserved across the fan-out.
- 2d. Merge between stages: Developer recombines multiple upstream paths into a single stream for downstream processing. Items from all sources appear in the merged output.
- 2e. Join between stages: Developer recombines results from two Tee branches. Join reads one result from each source, combines Ok/Ok pairs via a function, and forwards errors. Structural mismatches (extra or missing items from either source) are contract violations visible in stats.
- 5a. Wait called in forward order: Also valid — Wait may be called in any order after tail Out() is drained. Reverse order is recommended but not required.

**Sub-Variations:**
- Pipeline shape: Start → Pipe, Start → Batcher → Pipe, Start → Batcher → Pipe → Pipe (4-handle)
- Fan-out shape: Start → Tee → (Pipe, Pipe) — broadcast to N branches with independent downstream processing
- Per-stage workers: different worker counts per stage (e.g., N chunkers, E embedders, 1 writer)
- Error modes: per-stage fail-fast or continue-on-error (independent per stage)
- Pipe stats: Received = Submitted + Forwarded + Dropped
- Batcher stats: Received = Emitted + Forwarded + Dropped
- WeightedBatcher stats: Received = Emitted + Forwarded + Dropped, BufferedWeight tracks accumulated cost
- Tee stats: Received = FullyDelivered + PartiallyDelivered + Undelivered, per-branch BranchDelivered/BranchBlockedTime
- DAG shape: Start → Tee → (Pipe, Pipe) → Merge → Pipe — fan-out then fan-in with independent downstream processing
- Merge stats: per-source received, forwarded, dropped
- Join shape: Start → Tee → (Pipe, Pipe) → Join(fn) → Pipe — fan-out, independent processing, branch recombination
- Join stats: ReceivedA, ReceivedB, Combined, Errors, DiscardedA, DiscardedB, ExtraA, ExtraB
