// Package option provides types and functions to work with optional values.
//
// Option[T] holds a value plus an ok flag. The zero value is not-ok,
// so it works without initialization. Construct via Of, New, When, WhenCall,
// NonZero, NonEmpty, NonNil, or Env.
package option
