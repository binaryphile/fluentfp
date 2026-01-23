# lof: lower-order function wrappers

Utility functions for functional programming. Wraps Go builtins for HOF use, plus helpers for common patterns.

A **lower-order function** is the flip side of a higher-order function—the function being passed, not the one receiving it.

Go builtins like `len` are operators, not functions—you can't pass `len` to `ToInt`. lof wraps them as passable functions.

```go
names.Each(lof.Println)  // print each name
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/lof) for complete API documentation.

## Quick Start

```go
import (
    "github.com/binaryphile/fluentfp/lof"
    "github.com/binaryphile/fluentfp/slice"
)

names := slice.From(users).ToString(User.Name)
names.Each(lof.Println)  // print each name

// When type has no Len() method, lof.Len bridges the gap
type Report struct { Pages []Page }

// pageCount returns the number of pages in a report.
pageCount := func(r Report) int { return lof.Len(r.Pages) }
pageCounts := slice.From(reports).ToInt(pageCount)
```

## API Reference

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Len` | `Len[T]([]T) int` | Wrap `len` for slices | `lengths = items.ToInt(lof.Len)` |
| `StringLen` | `StringLen(string) int` | Wrap `len` for strings | `lens = names.ToInt(lof.StringLen)` |
| `Println` | `Println(string)` | Wrap `fmt.Println` | `names.Each(lof.Println)` |
| `IfNotEmpty` | `IfNotEmpty(string) (string, bool)` | Comma-ok for strings | `diff, ok := lof.IfNotEmpty(result)` |

## IfNotEmpty: Comma-ok for Empty Strings

Some functions use empty string as "absent" (e.g., `cmp.Diff` returns `""` when equal). `IfNotEmpty` converts this to Go's comma-ok idiom.

```go
result := cmp.Diff(want, got)
if diff, ok := lof.IfNotEmpty(result); ok {
    t.Errorf("mismatch:\n%s", diff)
}
```

## When NOT to Use lof

- **Method expressions exist** — `User.Name` beats wrapping a getter
- **Direct calls work** — If you're not in a HOF chain, just call `len()` or `fmt.Println()`

## See Also

For the HOF methods that consume these wrappers, see [slice](../slice/).
