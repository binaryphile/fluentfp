# Roadmap

Competitor gap analysis (March 2026). Compared fluentfp against samber/lo (20.9k stars), samber/mo (2.8k stars), ahmetb/go-linq (3.6k stars), repeale/fp-go (325 stars), IBM/fp-go (1.9k stars), BooleanCat/go-functional (150 stars).

For usage-survey-based prioritization, see [feature-gaps.md](feature-gaps.md).

## Status: Complete

All identified competitor gaps have been addressed or explicitly decided against. The only deferred item is FanIn (requires concurrent merge, not sequential Concat).

## Delivered

| Category | Feature | Version/Package |
|----------|---------|-----------------|
| High | Option.FlatMap / Result.FlatMap | v0.47.0 |
| High | Set Operations (Intersect, Difference, Union) | v0.48.0 |
| High | Flatten | v0.49.0 |
| Medium | Iterator-native operations | seq package |
| Medium | JSON/SQL serialization for Option | JSON v0.14.0, SQL v0.51.0 |
| Slice | DropLast, DropLastWhile, FindLast, IndexOf, LastIndexOf | slice |
| Slice | IsSorted, IsSortedBy, Intersperse, Shuffle, Sample, Samples | slice |
| Slice | Tally (supersedes CountBy), RepeatN | slice |
| Map | Invert, Merge, PickByKeys, OmitByKeys, ToPairs, FromPairs | kv |
| Option | OrWrap (MapNone), rslt.Fold, either.Fold | option, rslt, either |
| Concurrency | Retry with ConstantBackoff/ExponentialBackoff | hof |
| Concurrency | Throttle, ThrottleWeighted | hof v0.56.0 |
| Concurrency | Debouncer with MaxWait, Cancel, Flush, Close | hof |
| Concurrency | FromChannel, ToChannel | seq |
| Combinatorics | CartesianProduct, Permutations, Combinations, PowerSet | combo |

## Decided Against

| Feature | Reason |
|---------|--------|
| Curry | Only repeale (325 stars) provides it, zero real-world adoption. `Bind`/`BindR` cover practical partial application. Go type inference breaks with curried returns. |
| IO/Reader/State monads | Haskell-ism, not idiomatic Go |
| Optics (Lens/Prism/Traversal) | Very niche, no evidence of Go adoption |
| Do-notation (Do/Bind/Let/ApS) | Haskell-ism |
| Either3-5 sum types | Rarely needed |
| String case conversion | stdlib territory |
| Mutable in-place ops | Contradicts immutable FP philosophy |

## Deferred

| Feature | Reason |
|---------|--------|
| FanIn | Requires concurrent merge, not sequential Concat |
