package controllerhelpers

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHashObjects(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		args         []any
		expectedHash string
		expectedErr  error
	}{
		{
			name:         "deterministic",
			args:         []any{"hello", 42},
			expectedHash: "NOcQZlZRcRNxrzSq2HeIaczKZEeYK0Ja9tCy2IHoOSpFpodBYHXZfspdm80OtUu15qSBUHhIAG79kwxkNll27w==",
			expectedErr:  nil,
		},
		{
			name:         "nil input",
			args:         nil,
			expectedHash: "z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg==",
			expectedErr:  nil,
		},
		{
			name:         "unjsonifiable input returns error",
			args:         []any{func() {}},
			expectedHash: "",
			expectedErr:  &json.UnsupportedTypeError{Type: reflect.TypeFor[func()]()},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lastGot string
			for i := range 2 {
				got, err := HashObjects(tc.args...)
				if !reflect.DeepEqual(err, tc.expectedErr) {
					t.Errorf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
				}

				if !reflect.DeepEqual(got, tc.expectedHash) {
					t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedHash, got))
				}

				if i != 0 && !reflect.DeepEqual(got, lastGot) {
					t.Errorf("hash value has changed for the same input:\n%s", cmp.Diff(lastGot, got))
				}
				lastGot = got
			}
		})
	}
}
