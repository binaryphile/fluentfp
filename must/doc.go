// Package must provides panic-on-error helpers for enforcing invariants.
//
// Use must for startup-time configuration, code generation, tests, and
// programmer-error invariants — cases where failure means a bug or
// misconfiguration, not a recoverable runtime condition.
//
// Do not use must for user input, expected I/O failures, or exported
// library APIs unless panic semantics are explicitly part of the contract.
//
// [BeNil], [Get], [Get2], and [Of] panic with the original error value,
// preserving error chains for errors.Is/errors.As after recovery.
// [NonEmptyEnv] panics with a descriptive error wrapping [ErrEnvUnset]
// or [ErrEnvEmpty] for machine-checkable classification after recovery.
//
// [Of] panics immediately if given a nil function, wrapping [ErrNilFunction].
package must
