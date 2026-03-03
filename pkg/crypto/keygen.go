package crypto

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"k8s.io/client-go/util/flowcontrol"
)

type KeyGenerator interface {
	GenerateKey(ctx context.Context) (crypto.Signer, error)
}

type rsaKeyGenerator struct {
	rateLimiter flowcontrol.RateLimiter
	bits        int
}

var _ KeyGenerator = &rsaKeyGenerator{}

func NewRSAKeyGenerator(rateLimiter flowcontrol.RateLimiter, bits int) KeyGenerator {
	return &rsaKeyGenerator{
		rateLimiter: rateLimiter,
		bits:        bits,
	}
}

func (g *rsaKeyGenerator) GenerateKey(_ context.Context) (crypto.Signer, error) {
	g.rateLimiter.Accept()

	privateKey, err := rsa.GenerateKey(rand.Reader, g.bits)
	if err != nil {
		return nil, fmt.Errorf("can't generate rsa key (bits=%d): %w", g.bits, err)
	}

	return privateKey, nil
}
