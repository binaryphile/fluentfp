# fluentfp Use Cases

## System Scope

**System:** fluentfp
**In scope:** Collection transformation, lazy sequence processing, optional value handling, typed alternatives, invariant enforcement, conditional value selection, tuple operations, builtin function adapters for higher-order use, function composition, concurrency control, memoization, persistent data structures, combinatorics, iterator-native processing
**Out of scope:** General concurrency, I/O, serialization, error handling strategies, logging. Note: bounded concurrent traversal (`FanOut`) is in scope as a collection operation — it transforms a slice concurrently, not a general concurrency primitive.

## System Invariants

- **Immutability by default** — Operations produce new collections; inputs are never modified
- **Order preservation** — Transformations preserve element order unless explicitly sorting
- **Nil safety** — Collection and optional operations on nil/empty inputs produce valid empty results, never panic
- **Type safety** — All type mismatches are caught before the program runs; no runtime type errors
- **Interoperability** — Results work seamlessly with standard language constructs and existing code

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
| Process elements from a sequence on demand | Blue | med |
| Cache expensive function results | Blue | low |
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
- Developer has a collection

**Main Scenario:**
1. Developer selects the collection source.
2. Developer specifies transformations: filtering by criteria, converting elements, changing element types, reordering, expanding, deduplicating, or limiting count.
3. System applies each transformation in sequence.
4. System returns the final collection.

**Extensions:**
- 1a. Collection is nil or empty: System produces a valid empty collection.
- 1b. Collection source is a map: System extracts the map's values as a collection for further transformation.
- 1c. Collection source is a set: System extracts the set's members as a collection for further transformation.
- 1d. Developer needs to filter or transform map entries while preserving map structure: System applies predicates or value transforms to entries, returning a map for further map-level operations or value extraction.
- 1e. Collection source is a combinatorial construction: System generates all permutations, combinations, subsets, or pairwise products from input elements.
- 1f. Developer needs map entries as a flat collection of key-value pairs, or needs to construct a map from pairs: System converts between map and pair-slice representations. When constructing from pairs with duplicate keys, the last pair wins.
- 2a. Developer needs to expand each element into multiple: System applies expansion and concatenates in order. When the expansion produces a different type, the standalone variant infers both types.
- 2b. Developer needs duplicates removed: System removes duplicates preserving first occurrence.
- 2c. Developer needs a sorted copy: System produces sorted collection; original unchanged.
- 2d. Developer needs elements grouped by a derived key: System groups elements into a chainable collection of groups, each containing a key and the elements sharing that key.
- 2e. Developer needs to combine corresponding elements from two collections: System combines elements pairwise, either into pairs or through a provided function. If collections differ in length, system signals an error. For lazy sequences, system truncates to the shorter sequence.
- 2f. Developer needs transformations applied concurrently: System applies transformations concurrently with bounded parallelism, preserving element order in the result. For I/O-bound workloads, system reports success or failure per element, recovers panics as errors, and respects context cancellation.
- 2g. Developer needs an independent copy of the collection: System produces a copy not affected by changes to the original.
- 2h. Developer needs zero-value elements removed from a collection: System removes all elements equal to their type's zero value and returns the remaining elements. For string collections, the developer may use a string-specific variant that reads as "non-empty" for clarity.
- 2i. Developer needs to split a collection into fixed-size batches: System divides the collection into sub-collections of the specified size; the last batch may be smaller.
- 2j. Developer needs elements in random order: System produces a randomly shuffled copy of the collection.
- 2k. Developer needs a random subset of elements: System selects count random elements without replacement; if count exceeds length, returns all elements in random order.
- 4a. Developer needs to apply a side effect to each element rather than produce a new collection: System calls the function for every element in order.

**Sub-Variations:**
- Filtering: inclusion-based, exclusion-based, zero-value removal, or empty-string removal
- Type conversion: to built-in types or to arbitrary types
- Sorting: ascending or descending by extracted key
- Deduplication: by identity or by extracted key
- Batching: by fixed size
- Concurrent bounding: by item count (uniform cost) or by total cost (weighted)
- Randomization: full shuffle, random subset without replacement
- Combinatorial: permutations, combinations, power sets, Cartesian products

---

### UC-2: Derive a Result from a Collection

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct scalar, optional, or aggregate result
- Code reviewer: derivation reads as intent, not accumulation mechanics

**Postconditions:**
- Result correctly summarizes or extracts from the collection
- Original collection is unmodified

**Minimal Guarantee:** Original collection is never modified.

**Preconditions:**
- Developer has a collection

**Main Scenario:**
1. Developer selects the collection to derive from.
2. Developer specifies the derivation: combining elements progressively, finding a specific element, checking a condition, counting, summing, or extracting multiple fields simultaneously.
3. System processes the collection and returns the result.

**Extensions:**
- 1a. Collection is empty: System returns the appropriate empty result — zero for sums/counts, absence for lookups, initial value for accumulations, false for any-match checks, true for all-match checks, true for no-match checks, false for membership checks.
- 2a. Developer searches for first matching element: System returns the match or indicates absence.
- 2b. Developer searches for the first element matching a specific type: System returns the first type-compatible match or indicates absence.
- 2c. Developer expects exactly one element: System returns it or indicates the actual count.
- 2d. Developer needs multiple fields extracted simultaneously: System returns one collection per field.
- 2e. Developer needs to accumulate state while also producing per-element output: System processes elements in order and returns both the final accumulated value and the per-element outputs.
- 2f. Developer needs to convert the collection to a set for membership checks: System returns a set of the elements or extracted keys.
- 2g. Developer checks whether all elements satisfy a criterion: System tests every element and returns true only if all match.
- 2h. Developer checks whether a specific value exists in the collection: System tests membership and returns true if found.
- 2i. Developer checks that no elements satisfy a criterion: System tests every element and returns true only if none match.
- 2j. Developer needs elements indexed by a derived key for O(1) lookup: System produces a map from extracted keys to elements.
- 2k. Developer needs to efficiently track the minimum or maximum across a growing dataset: System maintains a persistent sorted structure where the extremum is always available in constant time and insertions produce a new structure without modifying the original.
- 2l. Developer needs a random element from a collection: System returns a random element or indicates absence if the collection is empty.

**Sub-Variations:**
- Numeric aggregation: sum, min, max on integer or floating-point collections
- Element search: first element, first matching, first type-compatible, index of first matching, random element
- Condition checks: any match, all match, no match, membership
- Multi-field extraction: 2, 3, or 4 fields simultaneously
- Indexing: by extracted key for O(1) lookup

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
- 2a. Developer needs a side effect only when present: System calls the function only when present; does nothing when absent.
- 2b. Developer needs a side effect only when absent: System calls the function only when absent; does nothing when present.
- 2c. Developer needs to filter an already-present value: System applies filter, converting to absent if not met.
- 2d. Fallback is expensive to compute: System evaluates fallback only when absent.
- 2e. Developer needs to chain operations that each may produce absence: System applies each operation in sequence, short-circuiting to absent if any step produces absence. No manual unwrapping between steps.
- 3a. Developer stores optional value in a database column: System maps present to the column value and absent to SQL NULL. Type conversion between Go types and SQL driver types is handled automatically.
- 3b. Developer serializes optional value to JSON: System maps present to the JSON value and absent to null.

**Sub-Variations:**
- Specialized variants for common value types (string, int, bool, error)
- Construction from: direct value, value-and-presence pair, pointer, zero-value check
- Create + transform: check presence and map to a new type in one call (zero-value, empty-string, nil-pointer variants)

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
- 1a. Developer has a fallible function returning (R, error) and needs it to return Result instead: System wraps the function, producing a new function with the same input that returns a Result.
- 2a. Developer needs both branches handled with different logic, producing unified result: System applies appropriate branch function.
- 2b. Developer needs to transform only the success branch: System transforms, passing failure through.
- 2c. Developer needs to chain operations that each may fail: System applies each operation in sequence, short-circuiting to failure if any step fails. No manual error checking between steps.

**Sub-Variations:**
- Convention: left = failure, right = success
- Extraction: value-and-presence pair, default value, lazy default

---

### UC-5: Enforce Initialization Invariants

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: program fails immediately when preconditions violated
- Operator: initialization failures surface at startup, not under load

**Postconditions:**
- All values available without error checking downstream
- If any precondition violated, program terminated with clear error before operational code

**Minimal Guarantee:** A violated precondition always terminates. No silent continuation with invalid state.

**Preconditions:**
- Developer has initialization steps that each might fail

**Main Scenario:**
1. Developer wraps each step to enforce success.
2. System executes; if any step fails, program terminates immediately with the error.
3. Developer uses resulting values without further error handling.

**Extensions:**
- 1a. Developer needs to wrap a function for repeated use: System returns a new function enforcing success on every call.
- 1b. Developer needs to assert an error is nil without extracting a result: System checks, terminates if non-nil.
- 1c. Developer needs a required environment variable: System reads it, terminates if missing.

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
1. Developer specifies the preferred value and condition.
2. Developer specifies the fallback value.
3. System evaluates: if condition holds, returns preferred value; otherwise returns fallback.

**Extensions:**
- 1a. Developer needs the first non-zero value from a sequence of candidates: System evaluates candidates in order and returns the first non-zero, or zero if all are zero.
- 1b. Preferred value is expensive to compute: System defers computation until condition confirmed true. If false, expensive computation never runs.

**Sub-Variations:**
- Eager: value computed before condition check (cheap values)
- Lazy: value computed only when condition true (expensive computations)
- FirstNonZero: first non-zero from candidates (zero = absent)

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
- 1a. Developer needs left-to-right composition of two transforms: System composes them so the first feeds into the second.
- 1b. Developer needs to fix one argument of a two-argument function: System returns a one-argument function with the fixed argument captured. Either the first or second argument can be fixed.
- 1c. Developer needs to apply separate functions to separate arguments: System applies each function independently and returns the results together.
- 1d. Developer needs a pass-through or identity key extractor: System provides a function that returns its argument unchanged.
- 1e. Developer needs a predicate that checks equality to a known value: System returns a function that tests its argument against the captured value.
- 1f. Developer needs to pass a Go builtin as a higher-order argument: System provides a first-class function wrapping the builtin, usable anywhere a function value is expected.
- 1g. Developer needs a function that enforces a concurrency budget when called from multiple goroutines: System returns a function with the same signature that blocks callers until budget is available, bounding by call count or per-call cost.
- 1h. Developer needs a function that triggers a side-effect when a call fails: System returns a function with the same signature that calls the original, then invokes the handler on error.
- 1i. Developer needs a function that retries on failure with configurable delays: System returns a function with the same signature that retries the original on error, waiting between attempts according to a backoff strategy, and respecting context cancellation during waits.

**Sub-Variations:**
- Composition: left-to-right (Pipe)
- Partial application: fix first arg (Bind), fix second arg (BindR)
- Building blocks: identity function, equality predicate
- Builtin adapters: length, printing, string predicates, successor
- Concurrency control: by count (Throttle), by cost (ThrottleWeighted)
- Side-effect on error (OnErr)
- Retry with backoff: constant delay, exponential with full jitter

---

### UC-8: Process a Lazy Sequence

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct elements processed without materializing entire sequence
- Code reviewer: lazy pipeline reads as intent, evaluation boundaries are clear

**Postconditions:**
- Requested elements have been produced or processed
- Full sequence was not required to exist in memory simultaneously
- Source sequence is unchanged and can be reused (persistence)

**Minimal Guarantee:** Partially consumed sequences remain valid for further operations. Evaluation failures do not corrupt the sequence.

**Preconditions:**
- Developer has a source that is large, infinite, or expensive to compute

**Main Scenario:**
1. Developer selects the sequence source.
2. Developer constructs a lazy sequence from the source.
3. Developer specifies transformations: filtering, converting, limiting count, skipping elements, changing element types, expanding and flattening, concatenating sequences, pairing corresponding elements, or accumulating intermediate values.
4. Developer terminates the pipeline: collecting to a slice, iterating for side effects, searching for a match, or reducing to a single value.

**Extensions:**
- 1a. Sequence is empty: System produces a valid empty result for any terminal operation.
- 2a. Source is a slice: System wraps it as a lazy sequence; elements are produced on demand from the underlying slice.
- 2b. Source is an infinite mathematical series: System generates elements from a seed and step function; the sequence never terminates.
- 2c. Source is a constant value repeated indefinitely: System produces the same value on each access.
- 2d. Source is a step function with termination: System unfolds from a seed, producing elements until the step function signals stop.
- 2e. Source is a step function that always produces an element: System produces an element from each step; an optional next-state controls whether to continue. Every step emits, including the last.
- 2f. Source is a recursive definition: System accepts a head value and a deferred tail computation, building the sequence lazily.
- 2g. Developer needs to chain two lazy sequences end-to-end: System produces all elements from the first, then all from the second.
- 3a. Developer needs cross-type transformation: System transforms elements to a different type.
- 3b. Developer needs to expand each element into a lazy sub-sequence and flatten: System applies expansion lazily, producing elements from inner sequences on demand.
- 3c. Developer needs to combine corresponding elements from two lazy sequences: System pairs elements, truncating to the shorter sequence.
- 3d. Developer needs running accumulator values as a lazy sequence: System produces the initial value followed by each intermediate accumulation.
- 4a. Developer needs to bridge to a Go range loop: System provides an iterator compatible with Go's range protocol.
- 4b. Developer needs to bridge to slice operations: System materializes to a plain slice for use with eager collection operations.
- 4c. Developer needs iterator-native processing without memoization: System provides lazy pipelines that re-evaluate on each iteration, compatible with Go's range protocol. Unlike memoized sequences, these pipelines do not cache intermediate results.

**Sub-Variations:**
- Construction: from slice, variadic, generate (infinite), repeat (constant), unfold (step function), paginate (always-emit step function), prepend (eager cons), prepend lazy (deferred cons)
- Filtering: by predicate
- Limiting: by count (take), by predicate (take-while)
- Skipping: by count (drop), by predicate (drop-while)
- Transformation: same-type (method), cross-type (standalone)
- Expanding: flatmap (expand and flatten sub-sequences)
- Concatenating: concat (chain sequences end-to-end)
- Pairing: zip (combine corresponding elements)
- Accumulating: scan (running intermediate values)
- Termination: collect, each, find, any, fold, seq

---

### UC-9: Memoize Function Results

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: repeated calls return cached results without redundant computation
- Code reviewer: caching boundary explicit at point of wrapping

**Postconditions:**
- Repeated calls with the same input return the same result without re-executing the original function
- Original function is unmodified

**Minimal Guarantee:** If the wrapped function panics, no corrupted result is cached. Future calls retry.

**Preconditions:**
- Developer has a function whose results are safe to cache (pure or idempotent)

**Main Scenario:**
1. Developer wraps the function with memoization.
2. First call executes the function and caches the result.
3. Subsequent calls with the same input return the cached result.

**Extensions:**
- 1a. Function takes no arguments (deferred initialization): System wraps a zero-arg function; first call evaluates and caches; subsequent calls return cached value.
- 1b. Function is fallible (returns value and error): System caches only successes — errors trigger retry on subsequent calls.
- 1c. Developer needs bounded cache size: System provides an LRU cache that evicts least recently used entries when capacity is exceeded.
- 1d. Developer needs a custom caching strategy: System accepts a caller-provided cache implementing Load/Store.
- 2a. Wrapped function panics: System resets to un-cached state; panic propagates; future calls retry the function.

**Sub-Variations:**
- Zero-arg memoization (Of): deferred initialization, sync.Once replacement with retry
- Keyed memoization (Fn, FnErr): function caching by input
- Cache strategies: unbounded map (NewMap), bounded LRU (NewLRU), custom (Cache interface)
