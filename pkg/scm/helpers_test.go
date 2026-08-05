package scm_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/stretchr/testify/require"
)

func TestFindModifiedFiles(t *testing.T) {
	t.Parallel()

	type args struct {
		files    []string
		patterns []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "patterns",
			args: args{
				files: []string{
					".dockerignore",
					"docker-compose.ci-override.yaml",
					"docker-compose.ci-override.yml",
					"docker-compose.test.yaml",
					"docker-compose.yaml",
					"docker-compose.yml",
					"docker-sync.yaml",
					"docker-sync.yml",
					"Dockerfile",
					"some-other-file.txt",
				},
				patterns: []string{
					".dockerignore",
					"*.Dockerfile",
					"docker-compose.*.y*ml",
					"docker-compose.y*ml",
					"docker-sync.y*ml",
					"Dockerfile",
				},
			},
			want: []string{
				".dockerignore",
				"docker-compose.ci-override.yaml",
				"docker-compose.ci-override.yml",
				"docker-compose.test.yaml",
				"docker-compose.yaml",
				"docker-compose.yml",
				"docker-sync.yaml",
				"docker-sync.yml",
				"Dockerfile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, scm.FindModifiedFiles(tt.args.files, tt.args.patterns...))
		})
	}
}

func TestPtr(t *testing.T) {
	t.Parallel()

	str := scm.Ptr("hello")
	require.NotNil(t, str)
	require.Equal(t, "hello", *str)

	num := scm.Ptr(0)
	require.NotNil(t, num, "a zero value must still yield a non-nil pointer")
	require.Equal(t, 0, *num)

	flag := scm.Ptr(false)
	require.NotNil(t, flag)
	require.False(t, *flag)
}

func TestMergeSlices(t *testing.T) {
	t.Parallel()

	identity := func(in string) string { return in }

	tests := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{name: "both empty", a: nil, b: nil, want: []string{}},
		{name: "only a", a: []string{"a", "b"}, b: nil, want: []string{"a", "b"}},
		{name: "only b", a: nil, b: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "disjoint", a: []string{"a"}, b: []string{"b"}, want: []string{"a", "b"}},
		{
			name: "b entries already present in a are dropped",
			a:    []string{"a", "b"},
			b:    []string{"b", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "a wins on conflict and order follows a then b",
			a:    []string{"z", "y"},
			b:    []string{"y", "x"},
			want: []string{"z", "y", "x"},
		},
		{
			name: "b fully contained in a is a no-op",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "a"},
			want: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, scm.MergeSlices(tt.a, tt.b, identity))
		})
	}
}

// The de-duplication only guards entries coming from the second slice, so
// duplicates already present in the first slice are preserved as-is.
func TestMergeSlices_keepsDuplicatesWithinTheFirstSlice(t *testing.T) {
	t.Parallel()

	identity := func(in string) string { return in }

	require.Equal(t,
		[]string{"a", "a", "b"},
		scm.MergeSlices([]string{"a", "a"}, []string{"a", "b"}, identity),
	)
}

// MergeSlices is used with a key function so structs can be merged on one field,
// which is how actions and labels are merged by name.
func TestMergeSlices_mergesStructsByKey(t *testing.T) {
	t.Parallel()

	type entry struct {
		Name  string
		Value int
	}

	got := scm.MergeSlices(
		[]entry{{Name: "one", Value: 1}},
		[]entry{{Name: "one", Value: 999}, {Name: "two", Value: 2}},
		func(e entry) string { return e.Name },
	)

	require.Equal(t, []entry{{Name: "one", Value: 1}, {Name: "two", Value: 2}}, got,
		"the entry from the first slice must win on a key collision")
}
