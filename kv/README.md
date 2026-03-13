# kv

Fluent operations on Go maps — filter, transform, extract.

```go
// Before: five lines to extract active config values from a map
var active []Config
for _, cfg := range configs {
    if cfg.Enabled {
        active = append(active, cfg)
    }
}

// After: intent only
// isEnabled returns true if the config is enabled.
isEnabled := func(_ string, cfg Config) bool { return cfg.Enabled }
active := kv.From(configs).KeepIf(isEnabled).Values()
```

Five lines become two.

## What It Looks Like

```go
// Transform map values (type change)
labels := kv.MapValues(counts, strconv.Itoa)
```

```go
// Filter entries, then chain
valid := kv.MapValues(raw, parseConfig).KeepIf(configIsValid)
```

```go
// Map entries to structs
items := kv.Map(s.Processes, toResult)
```

```go
// Extract and filter values
actives := kv.Values(userMap).KeepIf(User.IsActive)
```

```go
// Map entries to a built-in type
// formatEntry returns "key=value" for each map entry.
formatEntry := func(k string, v int) string { return fmt.Sprintf("%s=%d", k, v) }
labels := kv.From(m).ToString(formatEntry)
```

## It's Just a Map

`Entries[K,V]` is `map[K]V` — a defined type, not a wrapper. The zero value is a nil map — safe for reads (`len`, `range`) but panics on write. Indexing, `range`, and `len` all work as with a plain map. `From` does not copy; the `Entries` and the original map share the same backing data.

```go
entries := kv.From(m)
v := entries["key"]          // indexing works
for k, v := range entries {} // range works
n := len(entries)            // len works
```

## Operations

`Entries[K,V]` is a map with fluent methods. Filter and mapping methods take `func(K, V) T` predicates/transforms — filtering considers both key and value. Iteration order is not guaranteed (standard Go map behavior).

- **Wrap**: `From`
- **Filter**: `KeepIf`, `RemoveIf` — return `Entries[K,V]` for chaining
- **Extract**: `Values`, `Keys` — return `Mapper[V]`/`Mapper[K]` for slice chaining
- **Transform (standalone)**: `Map`, `MapValues`, `MapTo` — standalone because Go methods can't introduce new type parameters
- **Mapping**: `ToString`, `ToInt`, `ToFloat64`, `ToBool`, `ToAny`, `ToByte`, `ToError`, `ToFloat32`, `ToInt32`, `ToInt64`, `ToRune`
- **Utilities**: `Invert` (swap keys/values), `Merge` (combine maps, last-wins), `PickByKeys`/`OmitByKeys` (filter by key set)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/kv) for complete API documentation, the [main README](../README.md) for installation, and the [showcase](../docs/showcase.md) for real-world rewrites.
