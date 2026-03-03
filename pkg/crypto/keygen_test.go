package crypto

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/util/flowcontrol"
)

func TestRsaKeyGeneratorGenerateKey(t *testing.T) {
	t.Parallel()

	_, rsaErr := rsa.GenerateKey(rand.Reader, 0)

	tt := []struct {
		name         string
		bits         int
		expectedBits int
		expectedErr  error
	}{
		{
			name:         "generates a key with 2048 bits",
			bits:         2048,
			expectedBits: 2048,
			expectedErr:  nil,
		},

		{
			name:         "zero bits returns error",
			bits:         0,
			expectedBits: 0,
			expectedErr:  fmt.Errorf("can't generate rsa key (bits=%d): %w", 0, rsaErr),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rateLimiter := flowcontrol.NewFakeAlwaysRateLimiter()
			g := NewRSAKeyGenerator(rateLimiter, tc.bits)
			got, err := g.GenerateKey(context.Background())
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if got == nil {
				return
			}

			rsaKey, ok := got.(*rsa.PrivateKey)
			if !ok {
				t.Fatalf("expected *rsa.PrivateKey, got %T", got)
			}

			if rsaKey.N.BitLen() != tc.expectedBits {
				t.Errorf("expected and got key bit length differ:\n%s",
					cmp.Diff(tc.expectedBits, rsaKey.N.BitLen()))
			}
		})
	}
}
