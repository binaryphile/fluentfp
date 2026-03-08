# Roadmap

Competitor gap analysis (March 2026). Compared fluentfp against samber/lo (20.9k stars), samber/mo (2.8k stars), ahmetb/go-linq (3.6k stars), repeale/fp-go (325 stars), IBM/fp-go (1.9k stars), BooleanCat/go-functional (150 stars).

For usage-survey-based prioritization, see [feature-gaps.md](feature-gaps.md).

## High Priority

Features every serious competitor provides and fluentfp lacks.

### ~~Option.FlatMap / Result.FlatMap — monadic bind~~ (done v0.47.0)

### Set Operations — Intersect, Difference, Union

lo's most popular operations after Filter/Map. Commonly needed for reconciliation, diffing, and membership logic.

### Flatten — [][]T to []T

See [feature-gaps.md](feature-gaps.md) for details. Basic operation, lo and repeale have it.

## Medium Priority

Useful features with clear demand but existing workarounds.

### Iterator-Native Operations (Go 1.23+)

lo/it mirrors the entire lo API on `iter.Seq[T]`. fluentfp has `stream.Seq()` as a bridge but no iterator-native filter/map/take/etc. Go 1.23 iterators are mainstream. This is lo's biggest recent investment.

### JSON/SQL Serialization for Option

samber/mo's Option implements `json.Marshaler`/`json.Unmarshaler` and `sql.Scanner`/`sql.Valuer`. Practical for API and database work where nullable fields map to Option.

### Slice Operations

| Operation | Description | Who has it |
|-----------|-------------|------------|
| DropRight/DropRightWhile | Drop from end | lo |
| FindLast | Find from end | lo, IBM |
| IndexOf/LastIndexOf | By value, not predicate | lo |
| IsSorted/IsSortedBy | Check sort order | lo |
| Intersperse | Insert separator between elements | IBM |
| Shuffle | Randomize order | lo |
| Sample/Samples | Random element(s) | lo |
| CountBy | Count per group | lo |
| Repeat/RepeatBy | Generate slice by repeating | lo |

### Map Operations

| Operation | Description | Who has it |
|-----------|-------------|------------|
| Invert | Swap keys and values | lo |
| Merge/Assign | Combine maps | lo |
| PickByKeys/OmitByKeys | Filter by key set | lo |
| Entries/FromEntries | Map to/from slice of pairs | lo |

### Option/Result Extras

| Operation | Description | Who has it |
|-----------|-------------|------------|
| MapNone | Transform the absent case | mo |
| Match/Fold | Pattern-match dispatch | mo, repeale |

### Function Composition

| Operation | Description | Who has it |
|-----------|-------------|------------|
| Curry | Currying functions | repeale (Curry2-16) |

### Concurrency

| Operation | Description | Who has it |
|-----------|-------------|------------|
| Retry/Attempt | Retry with backoff | lo |
| Debounce/Throttle | Rate limiting | lo |
| Channel ops | SliceToChannel, ChannelToSlice, FanIn | lo |

## Skip — Academic/Niche

Not idiomatic Go. Excluded by design.

| Feature | Library | Why skip |
|---------|---------|----------|
| IO/Reader/State monads | IBM | Haskell-ism, not idiomatic Go |
| Optics (Lens/Prism/Traversal) | IBM | Very niche, no evidence of Go adoption |
| Do-notation (Do/Bind/Let/ApS) | IBM | Haskell-ism |
| Either3-5 sum types | mo | Rarely needed |
| String case conversion | lo | stdlib territory |
| Mutable in-place ops | lo | Contradicts immutable FP philosophy |
