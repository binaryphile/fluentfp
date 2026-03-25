// Package call provides decorators for context-aware effectful functions.
//
// Every decorator in this package wraps func(context.Context, T) (R, error)
// and returns the same signature. This uniform shape is the organizing
// principle: decorators compose by stacking because the types match at
// every layer.
//
// For higher-order functions over plain signatures — func(A) B composition,
// partial application, debouncing — see the [hof] package. The seam between
// call and hof is the function signature: call operates on the context-aware
// error-returning call shape; hof operates on everything else.
package call
