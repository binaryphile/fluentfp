# fluentfp Use Cases

## System Scope

**System:** fluentfp
**In scope:** Collection transformation, optional value handling, typed alternatives, invariant enforcement, conditional value selection, tuple operations, builtin function adapters for higher-order use
**Out of scope:** General concurrency, I/O, serialization, error handling strategies, logging

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

**Main Scenario:**
1. Developer has a collection and needs a derived collection with different elements, types, or ordering.
2. Developer specifies transformations: filtering by criteria, converting elements, changing element types, reordering, expanding, deduplicating, or limiting count.
3. System applies each transformation in sequence.
4. System returns the final collection.

**Extensions:**
- 1a. Collection is nil or empty: System produces a valid empty collection.
- 2a. Developer needs to expand each element into multiple: System applies expansion and concatenates in order.
- 2b. Developer needs duplicates removed: System removes duplicates preserving first occurrence.
- 2c. Developer needs a sorted copy: System produces sorted collection; original unchanged.
- 2d. Developer needs to combine corresponding elements from two collections: System combines elements pairwise, either into pairs or through a provided function. If collections differ in length, system signals an error.
- 2e. Developer needs transformations applied concurrently: System applies transformations concurrently, preserving element order in the result.
- 2f. Developer needs an independent copy of the collection: System produces a copy not affected by changes to the original.
- 4a. Developer needs to apply a side effect to each element rather than produce a new collection: System calls the function for every element in order.

**Sub-Variations:**
- Filtering: inclusion-based or exclusion-based criteria
- Type conversion: to built-in types or to arbitrary types
- Sorting: ascending or descending by extracted key
- Deduplication: by identity or by extracted key

---

### UC-2: Derive a Result from a Collection

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: correct scalar, optional, or aggregate result

**Postconditions:**
- Result correctly summarizes or extracts from the collection
- Original collection is unmodified

**Minimal Guarantee:** Original collection is never modified.

**Main Scenario:**
1. Developer has a collection and needs a single value derived from its elements.
2. Developer specifies the derivation: combining elements progressively, finding a specific element, checking a condition, counting, summing, or extracting multiple fields simultaneously.
3. System processes the collection and returns the result.

**Extensions:**
- 1a. Collection is empty: System returns the appropriate empty result — zero for sums/counts, absence for lookups, initial value for accumulations, false for any-match checks, true for all-match checks, false for membership checks.
- 2a. Developer searches for first matching element: System returns the match or indicates absence.
- 2b. Developer searches for the first element matching a specific type: System returns the first type-compatible match or indicates absence.
- 2c. Developer expects exactly one element: System returns it or indicates the actual count.
- 2d. Developer needs multiple fields extracted simultaneously: System returns one collection per field.
- 2e. Developer needs to accumulate state while also producing per-element output: System processes elements in order and returns both the final accumulated value and the per-element outputs.
- 2f. Developer needs to convert the collection to a set for membership checks: System returns a set of the elements or extracted keys.
- 2g. Developer checks whether all elements satisfy a criterion: System tests every element and returns true only if all match.
- 2h. Developer checks whether a specific value exists in the collection: System tests membership and returns true if found.

**Sub-Variations:**
- Numeric aggregation: sum, min, max on integer or floating-point collections
- Element search: first element, first matching, first type-compatible, index of first matching
- Condition checks: any match, all match, membership
- Multi-field extraction: 2, 3, or 4 fields simultaneously

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

**Main Scenario:**
1. Developer encounters a value that might be absent.
2. Developer wraps the value as optional, specifying what determines presence.
3. Developer transforms or extracts: providing a default, applying logic only when present, converting type, or filtering by additional criteria.
4. Developer uses the resolved value.

**Extensions:**
- 2a. Value comes from a pointer (nil means absent): System extracts the pointed-to value when non-nil.
- 2b. Value is a zero value that should mean absent: System treats zero as absent.
- 3a. Developer needs a side effect only when present: System calls the function only when present; does nothing when absent.
- 3b. Developer needs a side effect only when absent: System calls the function only when absent; does nothing when present.
- 3c. Developer needs to filter an already-present value: System applies filter, converting to absent if not met.
- 3d. Fallback is expensive to compute: System evaluates fallback only when absent.

**Sub-Variations:**
- Specialized variants for common value types (string, int, bool, error)
- Construction from: direct value, value-and-presence pair, pointer, zero-value check

---

### UC-4: Model Typed Alternatives

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: two possible outcomes handled exhaustively with correct types

**Postconditions:**
- Developer has processed the value through the appropriate branch
- Type system prevents accessing the wrong branch's value unsafely

**Minimal Guarantee:** Accessing the wrong branch returns a zero value and false, never corrupts state.

**Main Scenario:**
1. Developer has an operation producing one of two typed outcomes.
2. Developer constructs the value indicating which branch it represents.
3. Developer processes: extracting with a default, applying branch-specific logic, or handling both branches to produce a unified result.

**Extensions:**
- 3a. Developer needs both branches handled with different logic, producing unified result: System applies appropriate branch function.
- 3b. Developer needs to transform only the success branch: System transforms, passing failure through.

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

**Main Scenario:**
1. Developer has initialization steps that each might fail.
2. Developer wraps each step to enforce success.
3. System executes; if any step fails, program terminates immediately with the error.
4. Developer uses resulting values without further error handling.

**Extensions:**
- 2a. Developer needs to wrap a function for repeated use: System returns a new function enforcing success on every call.
- 2b. Developer needs to assert an error is nil without extracting a result: System checks, terminates if non-nil.
- 2c. Developer needs a required environment variable: System reads it, terminates if missing.

---

### UC-6: Select Values Conditionally

**Scope:** fluentfp | **Level:** Blue | **Actor:** Go Developer

**Stakeholders:**
- Developer: conditional logic expressed as value selection rather than control flow

**Postconditions:**
- Developer has the correct value: preferred when condition holds, fallback otherwise

**Minimal Guarantee:** The unused branch's computation is never evaluated when deferred.

**Main Scenario:**
1. Developer needs to choose between a preferred value and a fallback based on a condition.
2. Developer specifies the preferred value and condition.
3. Developer specifies the fallback value.
4. System evaluates: if condition holds, returns preferred value; otherwise returns fallback.

**Extensions:**
- 2a. Preferred value is expensive to compute: System defers computation until condition confirmed true. If false, expensive computation never runs.

**Sub-Variations:**
- Eager: value computed before condition check (cheap values)
- Lazy: value computed only when condition true (expensive computations)
