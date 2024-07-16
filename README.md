# funcTrunk - functional junk in a trunk

funcTrunk is a collection of small packages for Go, are inspired by functional
programming principles.  It is motivated by a desire to bring a more fluent
style to my own Go programs, and maybe to yours as well.

It includes the following packages:

- **mappable** -- fp collection methods for slices in a fluent style, such as *map* and *filter*, along with conventional methods like *contains*
  - **anymappable** -- a version of mappable for types that aren't comparable (no *contains* since it requires comparison)
- **iterator** -- an iterator function, where calling the function returns the next value and another, optional iterator (see option)
- **must** -- functions that turn errors into panics, making some fluent approaches possible
- **option** -- a basic option package to serve as the foundation for bespoke option types
- **ternary**
