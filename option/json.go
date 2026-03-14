package option

import (
	"bytes"
	"encoding/json"
	"errors"
)

// MarshalJSON serializes Option: Ok(v) → json(v), NotOk → null.
// Note: Ok(nil) and NotOk both serialize to null — the receiver cannot
// distinguish them. If round-trip fidelity matters, wrap the value in a
// non-nil container before storing.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if v, ok := o.Get(); ok {
		return json.Marshal(v)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes Option: null → NotOk, value → Ok(value).
// Note: because JSON null becomes NotOk, a round-trip through Ok(nil) →
// JSON → unmarshal yields NotOk, not Ok(nil). This is intentional — null
// means absent.
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if o == nil {
		return errors.New("option: UnmarshalJSON on nil receiver")
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*o = Option[T]{} // NotOk
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*o = Of(v)
	return nil
}
