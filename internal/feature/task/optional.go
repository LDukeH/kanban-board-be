package task

import "encoding/json"

type OptionalState int

const (
	// OptionalUnset: the JSON key was absent. Leave the stored value unchanged.
	OptionalUnset OptionalState = iota
	// OptionalNull: the JSON key was an explicit null. Clear the stored value.
	OptionalNull
	// OptionalValue: the JSON key carried a value. Write that value.
	OptionalValue
)

type Optional[T any] struct {
	state OptionalState
	value T
}

// Some returns an Optional holding v.
func Some[T any](v T) Optional[T] {
	return Optional[T]{state: OptionalValue, value: v}
}

func Null[T any]() Optional[T] {
	return Optional[T]{state: OptionalNull}
}

func Unset[T any]() Optional[T] {
	return Optional[T]{state: OptionalUnset}
}

// State reports which of the three states the field is in.
func (o Optional[T]) State() OptionalState {
	return o.state
}

// IsSet reports whether the JSON key was present at all (null counts as present).
func (o Optional[T]) IsSet() bool {
	return o.state != OptionalUnset
}

// IsNull reports whether the JSON key was an explicit null.
func (o Optional[T]) IsNull() bool {
	return o.state == OptionalNull
}

// Get returns the value and whether one is present.
func (o Optional[T]) Get() (T, bool) {
	return o.value, o.state == OptionalValue
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		var zero T
		*o = Optional[T]{state: OptionalNull, value: zero}
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*o = Optional[T]{state: OptionalValue, value: value}
	return nil
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if o.state != OptionalValue {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}
