package task

import "encoding/json"

// OptionalState records whether an Optional field was absent from a JSON body,
// an explicit null, or carried a value. The distinction matters on the update
// path: for a nullable column "absent" must leave the stored value alone while
// "null" must clear it, and a plain pointer cannot tell those two apart.
type OptionalState int

const (
	// OptionalUnset: the JSON key was absent. Leave the stored value unchanged.
	OptionalUnset OptionalState = iota
	// OptionalNull: the JSON key was an explicit null. Clear the stored value.
	OptionalNull
	// OptionalValue: the JSON key carried a value. Write that value.
	OptionalValue
)

// Optional is a three-state field: absent, null, or a value.
//
// It implements json.Unmarshaler, so encoding/json leaves the zero value
// (OptionalUnset) in place when a key is missing — UnmarshalJSON is only called
// for keys that are actually present. It also implements json.Marshaler, which
// renders both Unset and Null as JSON null so a response never leaks the
// tri-state and a stored NULL round-trips as null.
//
// It deliberately does NOT implement database/sql.Scanner. Repositories scan
// into a concrete *time.Time and convert with the date helper, which keeps this
// generic type free of pgx-specific concerns.
type Optional[T any] struct {
	state OptionalState
	value T
}

// Some returns an Optional holding v.
func Some[T any](v T) Optional[T] {
	return Optional[T]{state: OptionalValue, value: v}
}

// Null returns an Optional that marshals to JSON null and, in a SET clause,
// clears the column.
func Null[T any]() Optional[T] {
	return Optional[T]{state: OptionalNull}
}

// Unset returns an Optional that marshals to JSON null and, in a SET clause,
// leaves the column untouched. It is the zero state, provided for explicitness.
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
