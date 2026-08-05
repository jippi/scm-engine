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
