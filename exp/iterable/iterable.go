package iterable

import "errors"

var Done = errors.New("no more items in iterator") // definition borrowed from google.golang.org/api/iterator

type Iterable[T any] func() (T, error)
