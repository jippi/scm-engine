package types_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewValue(t *testing.T) {
	t.Parallel()

	valid := types.NewValue(42, true)
	require.True(t, valid.Valid)
	require.Equal(t, 42, valid.V)

	invalid := types.NewValue(42, false)
	require.False(t, invalid.Valid)
	require.Equal(t, 42, invalid.V, "the value is kept even when marked invalid")
}

// ValueFrom treats the zero value as "not set", which is what lets an omitted
// YAML key and an explicit zero be told apart downstream.
func TestValueFrom(t *testing.T) {
	t.Parallel()

	require.True(t, types.ValueFrom(1).Valid)
	require.False(t, types.ValueFrom(0).Valid, "the zero value is not considered set")
	require.True(t, types.ValueFrom("text").Valid)
	require.False(t, types.ValueFrom("").Valid)
}

func TestValueFromPtr(t *testing.T) {
	t.Parallel()

	fromNil := types.ValueFromPtr[int](nil)
	require.False(t, fromNil.Valid)
	require.Equal(t, 0, fromNil.V)

	fromValue := types.ValueFromPtr(scm.Ptr(7))
	require.True(t, fromValue.Valid)
	require.Equal(t, 7, fromValue.V)

	// A pointer to the zero value is still treated as unset
	fromZero := types.ValueFromPtr(scm.Ptr(0))
	require.False(t, fromZero.Valid)
}

func TestValue_MarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value types.Value[int]
		want  string
	}{
		{name: "a set value", value: types.NewValue(5, true), want: "5\n"},
		{name: "an invalid value is null", value: types.NewValue(5, false), want: "null"},
		{name: "a valid zero is null", value: types.NewValue(0, true), want: "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := tt.value.MarshalYAML()
			require.NoError(t, err)
			require.Equal(t, tt.want, string(raw))
		})
	}
}

func TestValue_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		yaml      string
		wantValid bool
		wantValue int
	}{
		{name: "a number", yaml: "field: 5", wantValid: true, wantValue: 5},
		{name: "zero", yaml: "field: 0", wantValid: true, wantValue: 0},
		{name: "negative", yaml: "field: -3", wantValid: true, wantValue: -3},
		{name: "explicit null", yaml: "field: null", wantValid: false},
		{name: "tilde null", yaml: "field: ~", wantValid: false},
		{name: "empty value is null", yaml: "field:", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var target struct {
				Field types.Value[int] `yaml:"field"`
			}

			require.NoError(t, yaml.Unmarshal([]byte(tt.yaml), &target))
			require.Equal(t, tt.wantValid, target.Field.Valid)

			if tt.wantValid {
				require.Equal(t, tt.wantValue, target.Field.V)
			}
		})
	}
}

// An absent key must stay invalid, which is how an unset priority is told apart
// from one explicitly set to zero.
func TestValue_UnmarshalYAML_missingKeyStaysUnset(t *testing.T) {
	t.Parallel()

	var target struct {
		Field types.Value[int] `yaml:"field"`
	}

	require.NoError(t, yaml.Unmarshal([]byte("other: 1"), &target))
	require.False(t, target.Field.Valid)
}

// A value that cannot be parsed must be reported. The null check used to match
// any value beginning with "n", so `priority: nope` was silently swallowed as a
// null instead of failing.
func TestValue_UnmarshalYAML_rejectsWrongType(t *testing.T) {
	t.Parallel()

	for _, input := range []string{`field: "not-a-number"`, "field: nope", "field: nil", "field: no"} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			var target struct {
				Field types.Value[int] `yaml:"field"`
			}

			require.Error(t, yaml.Unmarshal([]byte(input), &target), "%q should not parse as an int", input)
		})
	}
}

func TestValue_JSONSchema(t *testing.T) {
	t.Parallel()

	require.Equal(t, "number", types.Value[int]{}.JSONSchema().Type)
}

// Only int is mapped today; anything else must fail loudly while generating the
// schema rather than emitting a schema with an empty type.
func TestValue_JSONSchema_panicsOnUnsupportedType(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		types.Value[string]{}.JSONSchema()
	})
}
