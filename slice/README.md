# slice

Replace loop scaffolding with type-safe collection chains.

- **Interchangeable** (assignable) — `Mapper[T]` has underlying type `[]T`. No conversion needed in either direction.
- **Generics** — 100% type-safe. No `any`, no reflection, no type assertions.
- **Method expressions** — pass `User.IsActive` directly. No wrapper closures.
- **Comma-ok** — `Find`, `IndexWhere` return `option` with `.Get()` → `(value, ok)`.

```go
// Before: 5 lines of mechanics around 1 line of intent
var names []string
for _, u := range users {
    if u.IsActive() {
        names = append(names, u.Name)
    }
}

// After: intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.GetName)
```

## What It Looks Like

```go
// Ranking
top5 := slice.SortByDesc(players, Player.Score).TakeFirst(5)
```

```go
// Tag filtering (allowlist semantics — empty filter matches all)
if !slice.String(m.Tags).Matches(filter.Tags) {
    continue
}
```

```go
// Type mapping
users := slice.MapTo[User](ids).Map(FetchUser)
```

```go
// Multi-field extraction in one pass
leadTimes, deployFreqs, mttrs, cfrs := slice.Unzip4(history,
    Record.LeadTime, Record.DeployFreq, Record.MTTR, Record.ChangeFailRate,
)
```

```go
// Reduce to map
byMAC := slice.Fold(devices, make(map[string]Device), addDevice)
```

## It's Just a Slice

`Mapper[T]` is a defined type with underlying type `[]T`. Go's assignability rules make them interchangeable — no conversion ceremony to get values in or out.

```go
func activeNames(users []User) []string {
    names := slice.From(users).
        KeepIf(User.IsActive).
        ToString(User.Name)     // Mapper[string]
    names.Each(lof.Println)
    return names                // assigns to []string — no conversion
}
```

- `From()` is a type-cast, not a copy
- `append()`, `len()`, index, `range` all work directly on Mappers

Other Go FP libraries can't do this:
- **go-linq**, **fuego**: 6+ lines of `[]any` → `[]string` with type assertions to get results out
- **gofp**: conversion loops on both ends — `[]User` → `[]any` in, `[]any` → `[]string` out

See [comparison](../comparison.md) for the full library comparison.

## Operations

`From` creates `Mapper[T]`. `MapTo[R]` creates `MapperTo[R,T]` — all Mapper methods plus `Map` for arbitrary type mapping. `String` (`[]string`) and `Float64` (`[]float64`) are separate defined types with additional methods.

- **Filter**: `KeepIf`, `RemoveIf`, `TakeFirst`
- **Search**: `Find`, `IndexWhere`, `FindAs`, `Any`, `First`, `Single`, `Contains`, `ContainsAny`, `Matches` (String)
- **Transform**: `Convert`, `Map` (MapperTo), `ToString`, `ToInt`, other `To*`, `Clone`, `Unique` (String), `SortBy`, `SortByDesc`
- **Aggregate**: `Fold`, `Len`, `Sum` (Float64), `ToSet` (String), `Each`, `Unzip2`/`3`/`4`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/slice) for complete API documentation and the [main README](../README.md) for installation and performance characteristics.
