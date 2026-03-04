package option

import "encoding/json"

// MarshalJSON serializes Option: Ok(v) → v, NotOk → null
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if v, ok := o.Get(); ok {
		return json.Marshal(v)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes Option: null → NotOk, value → Ok(value)
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
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
