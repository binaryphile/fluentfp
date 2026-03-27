// Package wrap provides chainable decorators for context-aware effectful functions.
//
// Start with [Func] to wrap a plain function, then chain With* methods:
//
//	safe := wrap.Func(fetchOrder).
//	    WithRetry(3, wrap.ExpBackoff(time.Second), nil).
//	    WithBreaker(breaker).
//	    WithThrottle(10)
//
// Each method returns [Fn], preserving the func(context.Context, T) (R, error)
// signature so decorators compose freely. For custom decorators, use [Fn.With]
// with [Decorator] values.
//
// For higher-order functions over plain signatures — func(A) B composition,
// partial application, debouncing — see the [hof] package.
package wrap
