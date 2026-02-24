# lof

Go builtins wrapped as passable functions for collection chains.

`len` and `fmt.Println` are operators, not functions — you can't pass them to higher-order methods. lof bridges the gap.

```go
names.Each(lof.Println)
```

## What It Looks Like

```go
// Length of each inner slice — reports is [][]Page
pageCounts := slice.From(reports).ToInt(lof.Len)
```

```go
// String lengths
charCounts := names.ToInt(lof.StringLen)
```

## Operations

- `Len[T]([]T) int` — wrap `len` for slices
- `StringLen(string) int` — wrap `len` for strings
- `Println(string)` — wrap `fmt.Println`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/lof) for complete API documentation, the [main README](../README.md) for installation, and [slice](../slice/) for the collection methods that consume these wrappers.
