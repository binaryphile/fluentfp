# fluentfp -- Functional programming in Go with an emphasis on fluent interfaces

practicalfp is a collection of small packages for Go, inspired by functional programming
principles.

Why another Go fp package?  

[valor]: https://github.com/phelmkamp/valor
[fp-go]: https://github.com/repeale/fp-go

Valor has 

It includes the following packages:

- **mappable** -- mappable implements versions of the *map* and *filter* collection methods
  for slices in a fluent style, along with conventional methods like *contains*
    - **anymappable** -- a version of mappable for types that aren't comparable (no
      *contains* since it requires comparison)
- **iterator** -- an iterator function, where calling the function returns the next value
  and whether it is valid (if not, the iterator is done)
- **must** -- functions that turn errors into panics, making some fluent approaches possible
- **option** -- a basic option package to serve as the foundation for bespoke option types
- **ternary** -- a ternary implemented in a fluent style, i.e. If(cond).Then(alt1).Else(alt2)

## Mappable

`mappable.SliceOf` is the basic type of the mappable package.  It has the following
structure:

| Method | 
SliceOf []T
  Contains(t T) bool
  Index(t T) int
  KeepIf(fn func(T) bool) SliceOf[T] 
  Map
  MapToBool
  MapToInt
  MapToStr
  MapToStrOption
  MapToStrSlice
  RemoveIf

to run *filter* and *map* as methods on a slice.  If you are unfamiliar with map and slice,

[fluent interface]: https://en.wikipedia.org/wiki/Fluent_interface
