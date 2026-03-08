package option

import (
	"database/sql"
	"database/sql/driver"
)

// Value implements driver.Valuer: Ok(v) → v, NotOk → nil.
func (o Option[T]) Value() (driver.Value, error) {
	n := sql.Null[T]{V: o.t, Valid: o.ok}
	return n.Value()
}

// Scan implements sql.Scanner: nil → NotOk, value → Ok with type conversion.
func (o *Option[T]) Scan(src any) error {
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
