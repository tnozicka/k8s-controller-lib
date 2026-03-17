package slices

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "int to string",
			input:    []int{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "empty input",
			input:    []int{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []int{42},
			expected: []string{"42"},
		},
		{
			name:     "nil input returns empty slice",
			input:    nil,
			expected: []string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Convert(strconv.Itoa, tc.input...)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}
