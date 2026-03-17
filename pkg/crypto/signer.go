package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"math/big"
	"reflect"
	"time"
)

var (
	SerialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
)

type NowFunc func() time.Time

type Signer interface {
	GetPublicKey() (crypto.PublicKey, error)
	GetCertificate() (*x509.Certificate, error)
	SignCertificate(template *x509.Certificate, requestKey crypto.PublicKey) (*x509.Certificate, error)
}

type SelfSignedSigner struct {
	privateKey crypto.Signer
	nowFunc    NowFunc
}

var _ Signer = &SelfSignedSigner{}

func NewSelfSignedSigner(nowFunc NowFunc, privateKey crypto.Signer) Signer {
	return &SelfSignedSigner{
		nowFunc:    nowFunc,
		privateKey: privateKey,
	}
}

func (s *SelfSignedSigner) Now() time.Time {
	return s.nowFunc()
}

func (s *SelfSignedSigner) GetPublicKey() (crypto.PublicKey, error) {
	return s.privateKey.Public(), nil
}

func (s *SelfSignedSigner) GetCertificate() (*x509.Certificate, error) {
	return nil, nil
}

func (s *SelfSignedSigner) SignCertificate(template *x509.Certificate, requestKey crypto.PublicKey) (*x509.Certificate, error) {
	// Sanity check.
	if !reflect.DeepEqual(requestKey, s.privateKey.Public()) {
		return nil, fmt.Errorf("self-signed signer: public key mismatch")
	}

	var err error
	template.SerialNumber, err = rand.Int(rand.Reader, SerialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("can't generate certificate serial number: %w", err)
	}

	cert, err := SignCertificate(template, requestKey, template, s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("can't self-sign certificate for %q: %w", template.Subject, err)
	}

	return cert, nil
}
