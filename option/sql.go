package option

import (
	"database/sql"
	"database/sql/driver"
	"errors"
)

// Value implements driver.Valuer: Ok(v) → v, NotOk → nil.
// Note: Ok with a nil-typed value and NotOk both produce SQL NULL —
// the database cannot distinguish them.
func (o Option[T]) Value() (driver.Value, error) {
	n := sql.Null[T]{V: o.t, Valid: o.ok}
	return n.Value()
}

// Scan implements sql.Scanner: nil → NotOk, value → Ok with type conversion.
func (o *Option[T]) Scan(src any) error {
	if o == nil {
		return errors.New("option: Scan on nil receiver")
	}
	var n sql.Null[T]
	if err := n.Scan(src); err != nil {
		return err
	}
	if n.Valid {
		*o = Of(n.V)
	} else {
		*o = Option[T]{}
	}
	return nil
}
