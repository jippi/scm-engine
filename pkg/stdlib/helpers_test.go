package stdlib_test

import (
	"testing"
	"time"

	"github.com/jippi/scm-engine/pkg/stdlib"
	"github.com/stretchr/testify/require"
)

func TestToDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want time.Duration
	}{
		{name: "passes a duration straight through", in: 5 * time.Minute, want: 5 * time.Minute},
		{name: "zero duration", in: time.Duration(0), want: 0},
		{name: "parses a plain duration string", in: "90s", want: 90 * time.Second},
		{name: "parses the extended day unit", in: "2d", want: 48 * time.Hour},
		{name: "parses the extended week unit", in: "1w", want: 7 * 24 * time.Hour},
		{name: "parses a combined string", in: "1d6h30m", want: 30*time.Hour + 30*time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, stdlib.ToDuration(tt.in))
		})
	}
}

// ToDuration is called while building label/action configuration, where a bad
// value is a configuration error the user must fix, so it panics rather than
// returning an error.
func TestToDuration_panics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
	}{
		{name: "unparsable string", in: "not-a-duration"},
		{name: "empty string", in: ""},
		{name: "integer", in: 42},
		{name: "nil", in: nil},
		{name: "bool", in: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Panics(t, func() {
				stdlib.ToDuration(tt.in)
			})
		})
	}
}
