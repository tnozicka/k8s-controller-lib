package kubecrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	clcrypto "github.com/tnozicka/k8s-controller-lib/pkg/crypto"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

const (
	CABundleKey           = "ca-bundle.crt"
	certsKeyType          = "certificates.internal.k8s.controller-lib.tnozicka.github.com/key-type"
	certsNotBeforeKey     = "certificates.internal.k8s.controller-lib.tnozicka.github.com/not-before"
	certsNotAfterKey      = "certificates.internal.k8s.controller-lib.tnozicka.github.com/not-after"
	certsCountKey         = "certificates.internal.k8s.controller-lib.tnozicka.github.com/count"
	certsIsCAKey          = "certificates.internal.k8s.controller-lib.tnozicka.github.com/is-ca"
	certsIssuerKey        = "certificates.internal.k8s.controller-lib.tnozicka.github.com/issuer"
	certsRefreshReasonKey = "certificates.internal.k8s.controller-lib.tnozicka.github.com/refresh-reason"
)

func getKeyInfo(key crypto.Signer) string {
	switch key := key.(type) {
	case *rsa.PrivateKey:
		return fmt.Sprintf("rsa/%d", key.Size()*8)
	default:
		return "<unknown>"
	}
}

func ApplyCertAnnotations(annotations map[string]string, certs []*x509.Certificate, refreshReason string) {
	annotations[certsCountKey] = strconv.Itoa(len(certs))
	annotations[certsNotBeforeKey] = certs[0].NotBefore.Format(time.RFC3339)
	annotations[certsNotAfterKey] = certs[0].NotAfter.Format(time.RFC3339)
	annotations[certsIsCAKey] = strconv.FormatBool(certs[0].IsCA)
	annotations[certsIssuerKey] = certs[0].Issuer.String()
	// TODO: double check if this can blip.
	annotations[certsRefreshReasonKey] = refreshReason
}

func ApplyCertKeyAnnotations(annotations map[string]string, certs []*x509.Certificate, key crypto.Signer, refreshReason string) {
	ApplyCertAnnotations(annotations, certs, refreshReason)
	annotations[certsKeyType] = getKeyInfo(key)
}

func GetCABundleDataFromConfigMap(cm *corev1.ConfigMap) ([]byte, error) {
	if cm.Data == nil {
		return nil, fmt.Errorf("configMap %q doesn't contain any data", naming.ObjNN(cm))
	}

	caBundle, found := cm.Data[CABundleKey]
	if !found {
		return nil, fmt.Errorf("configMap %q is missing data for key %q", naming.ObjNN(cm), CABundleKey)
	}

	if len(caBundle) == 0 {
		return nil, fmt.Errorf("configMap %q is missing ca-bundle content", naming.ObjNN(cm))
	}

	return []byte(caBundle), nil
}

func GetCABundleFromConfigMap(cm *corev1.ConfigMap) ([]*x509.Certificate, error) {
	caBundleData, err := GetCABundleDataFromConfigMap(cm)
	if err != nil {
		return nil, err
	}

	return clcrypto.DecodeCertificates(caBundleData)
}

func GetCertsDataFromSecret(secret *corev1.Secret) ([]byte, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret %q doesn't contain any data", naming.ObjNN(secret))
	}

	certBytes := secret.Data[corev1.TLSCertKey]
	if len(certBytes) == 0 {
		return nil, fmt.Errorf("secret %q is missing certificate data", naming.ObjNN(secret))
	}

	return certBytes, nil
}

func GetCertsFromSecret(secret *corev1.Secret) ([]*x509.Certificate, error) {
	certBytes, err := GetCertsDataFromSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("can't get certificate bytes from secret %q: %w", naming.ObjNN(secret), err)
	}

	certificates, err := clcrypto.DecodeCertificates(certBytes)
	if err != nil {
		return nil, fmt.Errorf("can't decode TLS certificates from secret %q: %w", naming.ObjNN(secret), err)
	}

	return certificates, nil
}

func SetCertDataOnSecret(secret *corev1.Secret, certBytes []byte) {
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	secret.Data[corev1.TLSCertKey] = certBytes
}

func SetCertsOnSecret(secret *corev1.Secret, certs []*x509.Certificate) error {
	certBytes, err := clcrypto.EncodeCertificates(certs...)
	if err != nil {
		return fmt.Errorf("can't encode TLS certificates for secret %q: %w", naming.ObjNN(secret), err)
	}

	SetCertDataOnSecret(secret, certBytes)

	return nil
}

func GetCertFromSecret(secret *corev1.Secret) (*x509.Certificate, error) {
	certs, err := GetCertsFromSecret(secret)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("secret %q is missing a certificate", naming.ObjNN(secret))
	}

	return certs[0], nil
}

func GetKeyDataFromSecret(secret *corev1.Secret) ([]byte, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret %q doesn't contain any data", naming.ObjNN(secret))
	}

	keyBytes := secret.Data[corev1.TLSPrivateKeyKey]
	if len(keyBytes) == 0 {
		return nil, fmt.Errorf("secret %q is missing certificate key", naming.ObjNN(secret))
	}

	return keyBytes, nil
}

func GetKeyFromSecret(secret *corev1.Secret) (crypto.Signer, error) {
	keyBytes, err := GetKeyDataFromSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("can't get key bytes from secret %q: %w", naming.ObjNN(secret), err)
	}

	privateKey, err := clcrypto.DecodePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("can't decode TLS private key from secret %q: %w", naming.ObjNN(secret), err)
	}

	return privateKey, nil
}

func SetKeyDataOnSecret(secret *corev1.Secret, keyBytes []byte) {
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	secret.Data[corev1.TLSPrivateKeyKey] = keyBytes
}

func SetKeyOnSecret(secret *corev1.Secret, key crypto.Signer) error {
	keyBytes, err := clcrypto.EncodePrivateKey(key)
	if err != nil {
		return fmt.Errorf("can't encode TLS key for secret %q: %w", naming.ObjNN(secret), err)
	}

	SetKeyDataOnSecret(secret, keyBytes)

	return nil
}

// TLSSecret is a wrapper around a corev1.Secret that provides convenient access to the certificate and key.
type TLSSecret struct {
	secret *corev1.Secret
	certs  []*x509.Certificate
	key    crypto.Signer
}

func NewEmptyTLSSecret(secretMeta metav1.ObjectMeta) *TLSSecret {
	return &TLSSecret{
		secret: &corev1.Secret{
			ObjectMeta: *secretMeta.DeepCopy(),
			Type:       corev1.SecretTypeTLS,
			Data:       map[string][]byte{},
		},
	}
}

func NewTLSSecret(secret *corev1.Secret) *TLSSecret {
	return &TLSSecret{
		secret: secret.DeepCopy(),
		// Accessors will lazily initialize the certs and key.
		certs: nil,
		key:   nil,
	}
}

func (s *TLSSecret) GetCerts() ([]*x509.Certificate, error) {
	if s.certs != nil {
		return s.certs, nil
	}

	var err error
	s.certs, err = GetCertsFromSecret(s.secret)
	if err != nil {
		return nil, fmt.Errorf("can't get certs from secret %q: %w", naming.ObjNN(s.secret), err)
	}

	return s.certs, nil
}

func (s *TLSSecret) GetCert() (*x509.Certificate, error) {
	certs, err := s.GetCerts()
	if err != nil {
		return nil, fmt.Errorf("can't get certs from secret %q: %w", naming.ObjNN(s.secret), err)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("secret %q doesn't contain any certificate", naming.ObjNN(s.secret))
	}

	return s.certs[0], nil
}

func (s *TLSSecret) GetKey() (crypto.Signer, error) {
	if s.key != nil {
		return s.key, nil
	}

	var err error
	s.key, err = GetKeyFromSecret(s.secret)
	if err != nil {
		return nil, fmt.Errorf("can't get key from secret %q: %w", naming.ObjNN(s.secret), err)
	}

	return s.key, nil
}

func (s *TLSSecret) GetCertsKey() ([]*x509.Certificate, crypto.Signer, error) {
	certs, err := s.GetCerts()
	if err != nil {
		return nil, nil, err
	}

	key, err := s.GetKey()
	if err != nil {
		return nil, nil, err
	}

	return certs, key, err
}

func (s *TLSSecret) GetCertKey() (*x509.Certificate, crypto.Signer, error) {
	cert, err := s.GetCert()
	if err != nil {
		return nil, nil, err
	}

	key, err := s.GetKey()
	if err != nil {
		return nil, nil, err
	}

	return cert, key, err
}

func (s *TLSSecret) SetCerts(certs []*x509.Certificate) error {
	s.certs = certs
	return SetCertsOnSecret(s.secret, certs)
}

func (s *TLSSecret) SetKey(key crypto.Signer) error {
	s.key = key
	return SetKeyOnSecret(s.secret, key)
}

func (s *TLSSecret) SetCertsKey(certs []*x509.Certificate, key crypto.Signer) error {
	err := s.SetCerts(certs)
	if err != nil {
		return err
	}

	err = s.SetKey(key)
	if err != nil {
		return err
	}

	return nil
}
func (s *TLSSecret) SetCertKey(certs *x509.Certificate, key crypto.Signer) error {
	return s.SetCertsKey([]*x509.Certificate{certs}, key)
}

func (s *TLSSecret) UpdateProjectedAnnotations(refreshReason string) error {
	a := s.GetSecret().GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}

	certs, err := s.GetCerts()
	if err != nil {
		return err
	}

	key, err := s.GetKey()
	if err != nil {
		return err
	}

	ApplyCertKeyAnnotations(a, certs, key, refreshReason)

	s.GetSecret().SetAnnotations(a)

	return nil
}

func (s *TLSSecret) GetSecret() *corev1.Secret {
	return s.secret
}

func (s *TLSSecret) CalculateRefreshReason(
	now time.Time,
	refresh time.Duration,
	desiredCert *x509.Certificate,
	issuerCert *x509.Certificate,
	secretRef types.NamespacedName,
) string {
	if s.secret.Type != corev1.SecretTypeTLS {
		return fmt.Sprintf("invalid secret type %q, expected %q", s.secret.Type, corev1.SecretTypeTLS)
	}

	certs, _, err := s.GetCertsKey()
	if err != nil {
		return fmt.Sprintf("can't get certs and key from secret: %v", err)
	}

	if len(certs) == 0 {
		return fmt.Sprintf("no certificates are present")
	}

	for _, existingCert := range certs {
		if now.After(existingCert.NotAfter) {
			return "already expired"
		}

		validity := existingCert.NotAfter.Sub(existingCert.NotBefore)

		at80Percent := existingCert.NotAfter.Add(-validity / 5)
		if now.After(at80Percent) {
			return fmt.Sprintf("past its latest possible refresh time %v", at80Percent)
		}

		refreshDate := existingCert.NotBefore.Add(refresh)
		if now.After(refreshDate) {
			return fmt.Sprintf("past its refresh time %v", refreshDate)
		}

		existingCertTemplate := clcrypto.ExtractDesiredFieldsFromTemplate(existingCert)
		desiredCertTemplate := clcrypto.ExtractDesiredFieldsFromTemplate(desiredCert)
		if !reflect.DeepEqual(desiredCertTemplate, existingCertTemplate) {
			klog.V(2).InfoS(
				"Existing certificate template differs from the desired one",
				"Secret", secretRef,
				"Diff", cmp.Diff(existingCertTemplate, desiredCertTemplate),
			)
			return "certificate needs an update"
		}

		if issuerCert != nil && !reflect.DeepEqual(existingCert.AuthorityKeyId, issuerCert.SubjectKeyId) {
			klog.V(2).InfoS(
				"Issuers key hashes differ",
				"Secret", secretRef,
				"Existing64", base64.StdEncoding.EncodeToString(existingCert.AuthorityKeyId),
				"Issuer64", base64.StdEncoding.EncodeToString(issuerCert.SubjectKeyId),
			)
			return "issuer changed, new cert needs to be signed"
		}
	}

	return ""
}

func (s *TLSSecret) MakeCABundle(
	name string,
	existingCM *corev1.ConfigMap,
	now time.Time,
) (*corev1.ConfigMap, error) {
	var err error
	var existingCertificates []*x509.Certificate

	if existingCM != nil {
		existingCertificates, err = GetCABundleFromConfigMap(existingCM)
		if err != nil {
			// If the ConfigMap is empty or the data couldn't be decoded we'll act like there are
			// no valid certificates and use only the current certificate as the desired data.
			// This will make sure the controller can always repair the state and valid certs make it though.
			klog.V(2).InfoS(
				"Couldn't extract existing certificates from CABundle. Using only the current certificate",
				"CABundle", klog.KObj(existingCM),
				"Error", err,
			)
		}
	}

	currentCertificate, err := s.GetCert()
	if err != nil {
		return nil, fmt.Errorf("can't get current certificate: %w", err)
	}

	caBundle := clcrypto.MakeCABundle(currentCertificate, existingCertificates, now)

	caBundleBytes, err := clcrypto.EncodeCertificates(caBundle...)
	if err != nil {
		return nil, fmt.Errorf("can't encode ca bundle bytes: %w", err)
	}

	res := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.secret.Namespace,
			Name:      name,
		},
		Data: map[string]string{
			CABundleKey: string(caBundleBytes),
		},
	}

	return res, nil
}

func (s *TLSSecret) DeepCopy() *TLSSecret {
	if s == nil {
		return nil
	}

	out := *s

	out.secret = s.secret.DeepCopy()

	return &out
}

type SigningTLSSecret struct {
	TLSSecret
	nowFunc func() time.Time
}

var _ clcrypto.Signer = &SigningTLSSecret{}

func NewSigningTLSSecret(tlsSecret *TLSSecret, nowFunc func() time.Time) *SigningTLSSecret {
	return &SigningTLSSecret{
		TLSSecret: *tlsSecret,
		nowFunc:   nowFunc,
	}
}

func (s *SigningTLSSecret) GetPublicKey() (crypto.PublicKey, error) {
	key, err := s.GetKey()
	if err != nil {
		return nil, err
	}

	return key.Public(), nil
}

func (s *SigningTLSSecret) GetCertificate() (*x509.Certificate, error) {
	return s.GetCert()
}

func (s *SigningTLSSecret) SignCertificate(template *x509.Certificate, requestKey crypto.PublicKey) (*x509.Certificate, error) {
	signerKey, err := s.GetKey()
	if err != nil {
		return nil, err
	}

	signerCert, err := s.GetCert()
	if err != nil {
		return nil, err
	}

	template.SerialNumber, err = rand.Int(rand.Reader, clcrypto.SerialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("can't generate certificate serial number: %w", err)
	}

	cert, err := clcrypto.SignCertificate(template, requestKey, signerCert, signerKey)
	if err != nil {
		return nil, fmt.Errorf("can't sign certificate for %q: %w", template.Subject, err)
	}

	return cert, nil
}
