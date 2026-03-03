package maps

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHasKey(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		m        map[string]int
		key      string
		expected bool
	}{
		{
			name: "present key",
			m: map[string]int{
				"a": 1,
				"b": 2,
			},
			key:      "a",
			expected: true,
		},
		{
			name: "absent key",
			m: map[string]int{
				"a": 1,
				"b": 2,
			},
			key:      "c",
			expected: false,
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "a",
			expected: false,
		},
		{
			name:     "empty map",
			m:        map[string]int{},
			key:      "a",
			expected: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := HasKey(tc.m, tc.key)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		maps     []map[string]int
		expected map[string]int
	}{
		{
			name:     "zero maps",
			maps:     nil,
			expected: map[string]int{},
		},
		{
			name: "one map",
			maps: []map[string]int{
				{
					"a": 1,
					"b": 2,
				},
			},
			expected: map[string]int{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "two non-overlapping",
			maps: []map[string]int{
				{
					"a": 1,
				},
				{
					"b": 2,
				},
			},
			expected: map[string]int{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "overlapping last wins",
			maps: []map[string]int{
				{
					"a": 1,
					"b": 2,
				},
				{
					"b": 3,
					"c": 4,
				},
			},
			expected: map[string]int{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
		{
			name: "nil map in list",
			maps: []map[string]int{
				{
					"a": 1,
				},
				nil,
				{
					"b": 2,
				},
			},
			expected: map[string]int{
				"a": 1,
				"b": 2,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Merge(tc.maps...)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}

func TestLookupValues(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		m        map[string]int
		keys     []string
		expected []int
	}{
		{
			name: "all present",
			m: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			keys:     []string{"a", "b", "c"},
			expected: []int{1, 2, 3},
		},
		{
			name: "some missing",
			m: map[string]int{
				"a": 1,
				"c": 3,
			},
			keys:     []string{"a", "b", "c"},
			expected: []int{1, 3},
		},
		{
			name: "none present",
			m: map[string]int{
				"a": 1,
			},
			keys:     []string{"x", "y"},
			expected: []int{},
		},
		{
			name:     "empty map",
			m:        map[string]int{},
			keys:     []string{"a"},
			expected: []int{},
		},
		{
			name: "zero keys",
			m: map[string]int{
				"a": 1,
				"b": 2,
			},
			keys:     []string{},
			expected: []int{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := LookupValues(tc.m, tc.keys...)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}
