package types

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/guregu/null/v6"
	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"
)

type Value[T comparable] struct {
	null.Value[T]
}

func (v Value[T]) JSONSchema() *jsonschema.Schema {
	var schemaType string

	switch reflect.TypeFor[T]().Kind() {
	case reflect.Int:
		schemaType = "number"

	default:
		panic(fmt.Errorf("don't know how to convert type [%s] to JSON Schema type", reflect.TypeFor[T]().Kind().String()))
	}

	return &jsonschema.Schema{
		Type: schemaType,
	}
}

// NewValue creates a new Value.
func NewValue[T comparable](t T, valid bool) Value[T] {
	return Value[T]{
		Value: null.Value[T]{
			Null: sql.Null[T]{
				V:     t,
				Valid: valid,
			},
		},
	}
}

// ValueFrom creates a new Value that will always be valid.
func ValueFrom[T comparable](t T) Value[T] {
	var zero T

	return NewValue(t, t != zero)
}

// ValueFromPtr creates a new Value that will be null if t is nil.
func ValueFromPtr[T comparable](input *T) Value[T] {
	var zero T

	if input == nil {
		return NewValue(zero, false)
	}

	return NewValue(*input, *input != zero)
}

// MarshalJSON implements yaml.Marshaler.
// It will encode null if this value is null or zero.
func (t Value[T]) MarshalYAML() ([]byte, error) {
	var zero T

	if !t.Valid || t.V == zero {
		return []byte("null"), nil
	}

	return yaml.Marshal(t.V)
}

// UnmarshalJSON implements yaml.Unmarshaler.
// It supports string and null input.
func (t *Value[T]) UnmarshalYAML(value *yaml.Node) error {
	// Ask the parser whether this is a null rather than sniffing the raw text.
	// Matching on a leading "n" swallowed every value starting with that letter,
	// so a typo such as `priority: nope` silently became null instead of being
	// reported, and the other YAML spellings of null (`~`, an empty value) were
	// not recognised at all.
	if value.Tag == "!!null" {
		t.Valid = false

		return nil
	}

	data := []byte(value.Value)

	if err := yaml.Unmarshal(data, &t.V); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	t.Valid = true

	return nil
}
