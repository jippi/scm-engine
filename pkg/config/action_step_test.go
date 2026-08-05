package config_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionStep_RequiredStringSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		step    config.ActionStep
		key     string
		want    []string
		wantErr string
	}{
		{
			name: "returns slice when key exists with valid value",
			step: config.ActionStep{
				"user_ids": []string{"1", "2", "3"},
			},
			key:     "user_ids",
			want:    []string{"1", "2", "3"},
			wantErr: "",
		},
		{
			name: "returns empty slice when key exists with empty slice",
			step: config.ActionStep{
				"user_ids": []string{},
			},
			key:     "user_ids",
			want:    []string{},
			wantErr: "",
		},
		{
			name:    "returns error when key is missing",
			step:    config.ActionStep{},
			key:     "user_ids",
			want:    nil,
			wantErr: "Required 'step' key 'user_ids' is missing",
		},
		{
			name: "returns error when value is wrong type (string)",
			step: config.ActionStep{
				"user_ids": "not a slice",
			},
			key:     "user_ids",
			want:    nil,
			wantErr: "Required 'step' key 'user_ids' must be of type []string, got string",
		},
		{
			name: "returns error when value is wrong type (int)",
			step: config.ActionStep{
				"user_ids": 123,
			},
			key:     "user_ids",
			want:    nil,
			wantErr: "Required 'step' key 'user_ids' must be of type []string, got int",
		},
		{
			name: "returns slice when YAML provides []interface{} with string values",
			step: config.ActionStep{
				"user_ids": []interface{}{"2223", "4456"},
			},
			key:     "user_ids",
			want:    []string{"2223", "4456"},
			wantErr: "",
		},
		{
			name: "returns error when []interface{} contains non-string values",
			step: config.ActionStep{
				"user_ids": []interface{}{"2223", 123},
			},
			key:     "user_ids",
			want:    nil,
			wantErr: "Required 'step' key 'user_ids' must be of type []string, but element at index 1 is int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.step.RequiredStringSlice(tt.key)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestActionStep_RequiredString(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"present": "value", "wrong-type": 1234}

	got, err := step.RequiredString("present")
	require.NoError(t, err)
	require.Equal(t, "value", got)

	_, err = step.RequiredString("missing")
	require.ErrorContains(t, err, "Required 'step' key 'missing' is missing")

	_, err = step.RequiredString("wrong-type")
	require.ErrorContains(t, err, "must be of type string, got int")
}

func TestActionStep_RequiredInt(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"present": 42, "wrong-type": "not-a-number"}

	got, err := step.RequiredInt("present")
	require.NoError(t, err)
	require.Equal(t, 42, got)

	_, err = step.RequiredInt("missing")
	require.ErrorContains(t, err, "Required 'step' key 'missing' is missing")

	_, err = step.RequiredInt("wrong-type")
	require.ErrorContains(t, err, "must be of type int, got string")
}

func TestActionStep_RequiredStringEnum(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"source": "codeowners", "unknown": "nope", "wrong-type": 1}

	got, err := step.RequiredStringEnum("source", "codeowners", "backstage")
	require.NoError(t, err)
	require.Equal(t, "codeowners", got)

	_, err = step.RequiredStringEnum("missing", "codeowners")
	require.ErrorContains(t, err, "is missing")

	_, err = step.RequiredStringEnum("unknown", "codeowners", "backstage")
	require.ErrorContains(t, err, "must be one of [codeowners backstage], got nope")

	_, err = step.RequiredStringEnum("wrong-type", "codeowners")
	require.ErrorContains(t, err, "must be of type string, got int")
}

// The Optional accessors return the fallback for an absent key, but a key that
// is present with the wrong type is still an error the user has to fix.
func TestActionStep_OptionalString(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"present": "value", "wrong-type": 1234}

	got, err := step.OptionalString("present", "fallback")
	require.NoError(t, err)
	require.Equal(t, "value", got)

	got, err = step.OptionalString("missing", "fallback")
	require.NoError(t, err)
	require.Equal(t, "fallback", got)

	got, err = step.OptionalString("wrong-type", "fallback")
	require.ErrorContains(t, err, "must be of type string, got int")
	require.Equal(t, "fallback", got, "the fallback is returned alongside the error")
}

func TestActionStep_OptionalInt(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"present": 42, "wrong-type": "nope"}

	got, err := step.OptionalInt("present", 7)
	require.NoError(t, err)
	require.Equal(t, 42, got)

	got, err = step.OptionalInt("missing", 7)
	require.NoError(t, err)
	require.Equal(t, 7, got)

	got, err = step.OptionalInt("wrong-type", 7)
	require.ErrorContains(t, err, "must be of type int, got string")
	require.Equal(t, 7, got)
}

func TestActionStep_OptionalStringEnum(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"mode": "random", "unknown": "nope"}

	got, err := step.OptionalStringEnum("mode", "random", "random")
	require.NoError(t, err)
	require.Equal(t, "random", got)

	// This is the path that keeps `source` optional on assign_reviewers
	got, err = step.OptionalStringEnum("missing", "codeowners", "codeowners", "backstage")
	require.NoError(t, err)
	require.Equal(t, "codeowners", got)

	got, err = step.OptionalStringEnum("unknown", "fallback", "a", "b")
	require.ErrorContains(t, err, "must be one of [a b], got nope")
	require.Equal(t, "fallback", got)
}

func TestActionStep_Get(t *testing.T) {
	t.Parallel()

	step := config.ActionStep{"nested": config.ActionStep{"key": "value"}}

	got, err := step.Get("nested")
	require.NoError(t, err)
	require.Equal(t, config.ActionStep{"key": "value"}, got)

	_, err = step.Get("missing")
	require.ErrorContains(t, err, "'step' key 'missing' is missing")
}
