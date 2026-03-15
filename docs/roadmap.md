# Roadmap

Competitor gap analysis (March 2026). Compared fluentfp against samber/lo (20.9k stars), samber/mo (2.8k stars), ahmetb/go-linq (3.6k stars), repeale/fp-go (325 stars), IBM/fp-go (1.9k stars), BooleanCat/go-functional (150 stars).

For usage-survey-based prioritization, see [feature-gaps.md](feature-gaps.md).

## High Priority

Features every serious competitor provides and fluentfp lacks.

### ~~Option.FlatMap / Result.FlatMap — monadic bind~~ (done v0.47.0)

### ~~Set Operations — Intersect, Difference, Union~~ (done v0.48.0)

### ~~Flatten — [][]T to []T~~ (done v0.49.0)

## Medium Priority

Useful features with clear demand but existing workarounds.

### ~~Iterator-Native Operations (Go 1.23+)~~ (done — seq package)

### ~~JSON/SQL Serialization for Option~~ (done — JSON v0.14.0, SQL v0.51.0)

### Slice Operations

| Operation | Description | Who has it |
|-----------|-------------|------------|
| ~~DropRight/DropRightWhile~~ | ~~Drop from end~~ | done — `DropLast`, `DropLastWhile` |
| ~~FindLast~~ | ~~Find from end~~ | done |
| ~~IndexOf/LastIndexOf~~ | ~~By value, not predicate~~ | done |
| ~~IsSorted/IsSortedBy~~ | ~~Check sort order~~ | done |
| ~~Intersperse~~ | ~~Insert separator between elements~~ | done |
| ~~Shuffle~~ | ~~Randomize order~~ | done — `Mapper.Shuffle` |
| ~~Sample/Samples~~ | ~~Random element(s)~~ | done — `Mapper.Sample`/`Mapper.Samples` |
| ~~CountBy~~ | ~~Count per group~~ | Superseded by `Tally` (v0.35.0) |
| ~~Repeat/RepeatBy~~ | ~~Generate slice by repeating~~ | done — `RepeatN` |

### Map Operations

| Operation | Description | Who has it |
|-----------|-------------|------------|
| ~~Invert~~ | ~~Swap keys and values~~ | done |
| ~~Merge/Assign~~ | ~~Combine maps~~ | done — `Merge` |
| ~~PickByKeys/OmitByKeys~~ | ~~Filter by key set~~ | done |
| ~~Entries/FromEntries~~ | ~~Map to/from slice of pairs~~ | done — `ToPairs`, `FromPairs` |

### Option/Result Extras

| Operation | Description | Who has it |
|-----------|-------------|------------|
| ~~MapNone~~ | ~~Transform the absent case~~ | done — `OrWrap` |
| ~~Match/Fold~~ | ~~Pattern-match dispatch~~ | done — `rslt.Fold`, `either.Fold` |

### Function Composition

| Operation | Description | Who has it |
|-----------|-------------|------------|
| Curry | Currying functions | repeale (Curry2-16) |

### Concurrency

| Operation | Description | Who has it |
|-----------|-------------|------------|
| ~~Retry/Attempt~~ | ~~Retry with backoff~~ | done — `hof.Retry` with `ConstantBackoff`/`ExponentialBackoff` |
| ~~Throttle~~ | ~~Rate limiting~~ | done v0.56.0 as `hof.Throttle`/`ThrottleWeighted` |
| ~~Debounce~~ | ~~Call coalescing~~ | done — `hof.NewDebouncer` with `MaxWait`, `Cancel`, `Flush`, `Close` |
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
