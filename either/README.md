# either

A two-way sum type. `Either[L, R]` stores exactly one of two values.

Left = failure/alternative, Right = success/primary. Mnemonic: "right is right." The zero value is `Left` with the zero value of `L`.

```go
// Construct
cfg := either.Right[error, Config](loadedCfg)

// Fold forces both branches — can't compile without them
name := either.Fold(cfg,
    func(err error) string { return "default" },
    func(c Config) string { return c.Name },
)
```

## What It Looks Like

```go
// Mode dispatch — either isn't just for errors
timeout := either.Fold(config,
    func(l LocalConfig) int { return l.Timeout },
    func(r RemoteConfig) int { return r.Timeout },
)
```

```go
// Comma-ok extraction
if cfg, ok := parsed.Get(); ok {
    fmt.Println("loaded:", cfg.Name())
}
```

```go
// Default value
cfg := parsed.Or(fallbackConfig)
```

```go
// Side effect — fires only if Right
parsed.IfRight(func(cfg Config) { fmt.Println("loaded:", cfg.Name()) })
```

## Exhaustive at the Call Site

`Fold` forces both branches — you can't compile without providing handlers for `L` and `R`:

- Both handler functions are required parameters — you can't forget a case
- Both must return the same type — the result is always well-typed
- No default/fallback branch — every state is explicitly handled

Other accessors like `Get`, `Or`, and `IsRight` are available when you only need one side.

## Either as Architecture

A single `Either[L, R]` type can flow through multiple dispatch sites. Each `Fold` handles both cases — every new dispatch site must account for both. Change either type's interface and the compiler catches every site that needs updating.

Consider an application with two modes — setup and running. The mode flows through the entire render path:

```go
type AppState = either.Either[Setup, Running]

// Multiple sites, all exhaustive
title   := either.Fold(state, Setup.Title, Running.Title)
canEdit := either.Fold(state, Setup.CanEdit, Running.CanEdit)
view    := either.Fold(state, Setup.Render, Running.Render)
```

Every site that touches `AppState` must handle both cases. The compiler enforces this everywhere `Fold` appears.

## Operations

`Either[L, R]` represents either a Left or a Right value.

- **Create**: `Left`, `Right`
- **Extract**: `Get`, `GetLeft`, `IsLeft`, `IsRight`, `MustGet`, `MustGetLeft`, `Or`, `LeftOr`, `OrCall`, `LeftOrCall`
- **Transform**: `.Convert` (same-type Right), `.FlatMap` (same-type bind), `.FlatMapLeft` (recovery from Left), `.Swap`, `FlatMap` (cross-type bind), `Map` (cross-type Right), `MapLeft` (cross-type Left), `Fold`
- **Side effects**: `IfRight`, `IfLeft`

Methods are used when the return type can be expressed with existing type parameters. Standalone functions are needed when new type parameters must be introduced (`Map`, `MapLeft`, cross-type `FlatMap`, `Fold`).

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/either) for complete API documentation, the [main README](../README.md) for installation, and [option](../option/) for absent values without failure context.
