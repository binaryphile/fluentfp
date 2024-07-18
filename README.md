# funcTrunk - functional junk in a trunk

funcTrunk is a collection of small packages for Go, are inspired by functional programming
principles.  It is motivated by a desire to bring a more fluent style to my own Go programs,
and maybe to yours as well.

It includes the following packages:

- **mappable** -- collection methods for slices in a fluent style, such as *map* and *filter*, along with conventional methods like *contains*
  - **anymappable** -- a version of mappable for types that aren't comparable (no *contains* since it requires comparison)
- **iterator** -- an iterator function, where calling the function returns the next value and whether it is valid (if not, the iterator is done)
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
