package queuehelpers

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/tools/cache"
)

func TestUnwrapTombstoneIfPresent(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name     string
		obj      any
		expected any
	}{
		{
			name:     "regular object returned as-is",
			obj:      "just-a-string",
			expected: "just-a-string",
		},
		{
			name: "tombstone returns inner object",
			obj: cache.DeletedFinalStateUnknown{
				Key: "some/key",
				Obj: "inner-object",
			},
			expected: "inner-object",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := UnwrapTombstoneIfPresent(tc.obj)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}
