# lof: lower-order function wrappers

Wrap Go builtins for use with higher-order functions. When `len` or `fmt.Println` don't match the signature a HOF expects, lof provides compatible wrappers.

```go
names.Each(lof.Println)  // print each name
```

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

## When NOT to Use lof

- **Method expressions exist** — `User.Name` beats wrapping a getter
- **Direct calls work** — If you're not in a HOF chain, just call `len()` or `fmt.Println()`

## See Also

For the HOF methods that consume these wrappers, see [slice](../slice/).
