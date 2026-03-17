package kubecrypto

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	clcrypto "github.com/tnozicka/k8s-controller-lib/pkg/crypto"
)

func newTestKey(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("can't generate RSA key: %v", err)
	}

	keyBytes, err := clcrypto.EncodePrivateKey(key)
	if err != nil {
		t.Fatalf("can't encode key: %v", err)
	}

	return key, keyBytes
}

func newTestED25519Key(t *testing.T) (ed25519.PrivateKey, []byte) {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("can't generate ed25519 key: %v", err)
	}

	keyBytes, err := clcrypto.EncodePrivateKey(key)
	if err != nil {
		t.Fatalf("can't encode key: %v", err)
	}

	return key, keyBytes
}

func newSelfSignedCertificate(t *testing.T, template *x509.Certificate, signer crypto.Signer) (*x509.Certificate, []byte) {
	t.Helper()

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, signer.Public(), signer)
	if err != nil {
		t.Fatalf("can't create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("can't parse certificate: %v", err)
	}

	certBytes, err := clcrypto.EncodeCertificates(cert)
	if err != nil {
		t.Fatalf("can't encode certificate: %v", err)
	}

	return cert, certBytes
}

func TestGetCABundleDataFromConfigMap(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		cm           *corev1.ConfigMap
		expectedData []byte
		expectedErr  error
	}{
		{
			name: "valid configmap",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: map[string]string{
					CABundleKey: "certdata",
				},
			},
			expectedData: []byte("certdata"),
			expectedErr:  nil,
		},
		{
			name: "nil data",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: nil,
			},
			expectedData: nil,
			expectedErr: fmt.Errorf(
				"configMap %q doesn't contain any data",
				types.NamespacedName{Namespace: "ns", Name: "test"},
			),
		},
		{
			name: "missing key",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: map[string]string{
					"other-key": "value",
				},
			},
			expectedData: nil,
			expectedErr: fmt.Errorf(
				"configMap %q is missing data for key %q",
				types.NamespacedName{Namespace: "ns", Name: "test"},
				CABundleKey,
			),
		},
		{
			name: "empty value",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: map[string]string{
					CABundleKey: "",
				},
			},
			expectedData: nil,
			expectedErr: fmt.Errorf(
				"configMap %q is missing ca-bundle content",
				types.NamespacedName{Namespace: "ns", Name: "test"},
			),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetCABundleDataFromConfigMap(tc.cm)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if !reflect.DeepEqual(got, tc.expectedData) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedData, got))
			}
		})
	}
}

func TestGetCertsDataFromSecret(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		secret       *corev1.Secret
		expectedData []byte
		expectedErr  error
	}{
		{
			name: "valid secret",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: map[string][]byte{
					corev1.TLSCertKey: []byte("certdata"),
				},
			},
			expectedData: []byte("certdata"),
			expectedErr:  nil,
		},
		{
			name: "nil data",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: nil,
			},
			expectedData: nil,
			expectedErr: fmt.Errorf(
				"secret %q doesn't contain any data",
				types.NamespacedName{Namespace: "ns", Name: "test"},
			),
		},
		{
			name: "empty cert data",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: map[string][]byte{
					corev1.TLSCertKey: {},
				},
			},
			expectedData: nil,
			expectedErr: fmt.Errorf(
				"secret %q is missing certificate data",
				types.NamespacedName{Namespace: "ns", Name: "test"},
			),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetCertsDataFromSecret(tc.secret)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if !reflect.DeepEqual(got, tc.expectedData) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedData, got))
			}
		})
	}
}

func TestSetCertDataOnSecret(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		secret         *corev1.Secret
		expectedSecret *corev1.Secret
	}{
		{
			name: "existing data",
			secret: &corev1.Secret{
				Data: map[string][]byte{},
			},
			expectedSecret: &corev1.Secret{
				Data: map[string][]byte{
					corev1.TLSCertKey: []byte("certdata"),
				},
			},
		},
		{
			name: "nil data",
			secret: &corev1.Secret{
				Data: nil,
			},
			expectedSecret: &corev1.Secret{
				Data: map[string][]byte{
					corev1.TLSCertKey: []byte("certdata"),
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.secret.DeepCopy()
			SetCertDataOnSecret(got, []byte("certdata"))

			if !apiequality.Semantic.DeepEqual(got, tc.expectedSecret) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedSecret, got))
			}
		})
	}
}

func TestSetKeyDataOnSecret(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		secret         *corev1.Secret
		expectedSecret *corev1.Secret
	}{
		{
			name: "existing data",
			secret: &corev1.Secret{
				Data: map[string][]byte{},
			},
			expectedSecret: &corev1.Secret{
				Data: map[string][]byte{
					corev1.TLSPrivateKeyKey: []byte("keydata"),
				},
			},
		},
		{
			name: "nil data",
			secret: &corev1.Secret{
				Data: nil,
			},
			expectedSecret: &corev1.Secret{
				Data: map[string][]byte{
					corev1.TLSPrivateKeyKey: []byte("keydata"),
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.secret.DeepCopy()
			SetKeyDataOnSecret(got, []byte("keydata"))

			if !apiequality.Semantic.DeepEqual(got, tc.expectedSecret) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedSecret, got))
			}
		})
	}
}

func TestNewEmptyTLSSecret(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		meta           metav1.ObjectMeta
		expectedSecret *TLSSecret
	}{
		{
			name: "correct type and empty data",
			meta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "ns",
			},
			expectedSecret: &TLSSecret{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "ns",
					},
					Type: corev1.SecretTypeTLS,
					Data: map[string][]byte{},
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewEmptyTLSSecret(tc.meta)
			if !reflect.DeepEqual(got, tc.expectedSecret) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedSecret, got))
			}
		})
	}
}

func TestNewTLSSecret(t *testing.T) {
	t.Parallel()

	key, keyBytes := newTestKey(t)

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign,
	}
	_, certBytes := newSelfSignedCertificate(t, template, key)

	tt := []struct {
		name           string
		secret         *corev1.Secret
		expectedSecret *TLSSecret
	}{
		{
			name: "deep copies the secret",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "original",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					"tls.key": keyBytes,
					"tls.crt": certBytes,
				},
			},
			expectedSecret: &TLSSecret{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "original",
						Namespace: "ns",
					},
					Type: corev1.SecretTypeTLS,
					Data: map[string][]byte{
						"tls.key": keyBytes,
						"tls.crt": certBytes,
					},
				},
				key:   nil,
				certs: nil,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewTLSSecret(tc.secret)

			uncachedExpected := tc.expectedSecret.DeepCopy()
			uncachedExpected.key = nil
			uncachedExpected.certs = nil
			if !reflect.DeepEqual(got, uncachedExpected) {
				t.Errorf("initial expected and got differ:\n%s", cmp.Diff(uncachedExpected, got, cmp.AllowUnexported(TLSSecret{})))
			}

			if !reflect.DeepEqual(got, tc.expectedSecret) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedSecret, got, cmp.AllowUnexported(TLSSecret{})))
			}
		})
	}
}

func TestTLSSecretGetSecret(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name           string
		tlsSecret      *TLSSecret
		expectedSecret *corev1.Secret
	}{
		{
			name: "returns inner secret",
			tlsSecret: &TLSSecret{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inner",
						Namespace: "ns",
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inner",
					Namespace: "ns",
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.tlsSecret.GetSecret()

			if !apiequality.Semantic.DeepEqual(got, tc.expectedSecret) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expectedSecret, got))
			}
		})
	}
}

func TestTLSSecretUpdateProjectedAnnotations(t *testing.T) {
	t.Parallel()

	key, keyBytes := newTestKey(t)

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	cert, certBytes := newSelfSignedCertificate(t, template, key)

	tt := []struct {
		name           string
		secret         *TLSSecret
		expectedReason string
		expectedSecret *TLSSecret
	}{
		{
			name: "all annotations",
			secret: NewTLSSecret(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Namespace:   "ns",
					Annotations: nil,
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       certBytes,
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			}),
			expectedReason: "test-reason",
			expectedSecret: &TLSSecret{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "ns",
						Annotations: map[string]string{
							certsKeyType:          fmt.Sprintf("rsa/%d", key.Size()*8),
							certsNotBeforeKey:     cert.NotBefore.Format(time.RFC3339),
							certsNotAfterKey:      cert.NotAfter.Format(time.RFC3339),
							certsCountKey:         "1",
							certsIsCAKey:          "true",
							certsIssuerKey:        cert.Issuer.String(),
							certsRefreshReasonKey: "test-reason",
						},
					},
					Type: corev1.SecretTypeTLS,
					Data: map[string][]byte{
						corev1.TLSCertKey:       certBytes,
						corev1.TLSPrivateKeyKey: keyBytes,
					},
				},
				certs: []*x509.Certificate{cert},
				key:   key,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.expectedSecret.DeepCopy()
			err := got.UpdateProjectedAnnotations(tc.expectedReason)
			if err != nil {
				t.Fatalf("can't update projected annotations: %v", err)
			}

			if !reflect.DeepEqual(got, tc.expectedSecret) {
				t.Errorf("expected and got TLSSecrets differ:\n%s", cmp.Diff(tc.expectedSecret, got, cmp.AllowUnexported(TLSSecret{})))
			}
		})
	}
}

func TestTLSSecretCalculateRefreshReason(t *testing.T) {
	t.Parallel()

	now := time.Now()

	key, keyBytes := newTestKey(t)
	certTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(1 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	cert, certBytes := newSelfSignedCertificate(t, certTemplate, key)

	_, expiredKeyBytes := newTestKey(t)
	expiredTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "test-expired"},
		NotBefore:             now.Add(-2 * time.Hour),
		NotAfter:              now.Add(-1 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	_, expiredCertBytes := newSelfSignedCertificate(t, expiredTemplate, key)

	_, at80PercentKeyBytes := newTestKey(t)
	at80PercentTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "test-at80Percent"},
		NotBefore:             now.Add(-2 * time.Hour),
		NotAfter:              now.Add(3 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	at80PercentCert, at80PercentCertBytes := newSelfSignedCertificate(t, at80PercentTemplate, key)

	tt := []struct {
		name            string
		secret          *corev1.Secret
		now             time.Time
		refresh         time.Duration
		desiredCert     *x509.Certificate
		expectedContain string
	}{
		{
			name: "invalid secret type",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					corev1.TLSCertKey:       certBytes,
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			},
			now:             now,
			refresh:         24 * time.Hour,
			desiredCert:     cert,
			expectedContain: "invalid secret type",
		},
		{
			name: "no certs from non-PEM data",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       []byte("non-pem-data"),
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			},
			now:             now,
			refresh:         24 * time.Hour,
			desiredCert:     nil,
			expectedContain: "no certificates are present",
		},
		{
			name: "no certs in secret",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{},
			},
			now:             now,
			refresh:         24 * time.Hour,
			desiredCert:     cert,
			expectedContain: "can't get certs and key",
		},
		{
			name: "past 80% validity",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       certBytes,
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			},
			now:             cert.NotAfter.Add(-time.Minute),
			refresh:         24 * time.Hour,
			desiredCert:     cert,
			expectedContain: "past its latest possible refresh time",
		},
		{
			name: "expired caught at 80%",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       certBytes,
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			},
			now:             cert.NotAfter.Add(time.Hour),
			refresh:         24 * time.Hour,
			desiredCert:     cert,
			expectedContain: "already expired",
		},
		{
			name: "already expired",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       expiredCertBytes,
					corev1.TLSPrivateKeyKey: expiredKeyBytes,
				},
			},
			now:             now,
			refresh:         24 * time.Hour,
			desiredCert:     expiredTemplate,
			expectedContain: "already expired",
		},
		{
			name: "past refresh time",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       at80PercentCertBytes,
					corev1.TLSPrivateKeyKey: at80PercentKeyBytes,
				},
			},
			now:             now,
			refresh:         time.Hour,
			desiredCert:     at80PercentCert,
			expectedContain: "past its refresh time",
		},
		{
			name: "template diff",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       certBytes,
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			},
			now:     now,
			refresh: 24 * time.Hour,
			desiredCert: func() *x509.Certificate {
				tmpl := *certTemplate
				tmpl.Subject.CommonName += "t.Subject.CommonName"
				return &tmpl
			}(),
			expectedContain: "certificate needs an update",
		},
		{
			name: "up to date",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       certBytes,
					corev1.TLSPrivateKeyKey: keyBytes,
				},
			},
			now:             now,
			refresh:         24 * time.Hour,
			desiredCert:     certTemplate,
			expectedContain: "",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tlsSecret := NewTLSSecret(tc.secret)
			got := tlsSecret.CalculateRefreshReason(
				tc.now,
				tc.refresh,
				tc.desiredCert,
				nil,
				types.NamespacedName{Namespace: "ns", Name: "test"},
			)

			if tc.expectedContain == "" {
				if got != "" {
					t.Errorf("expected empty string, got %q", got)
				}
				return
			}

			if !strings.Contains(got, tc.expectedContain) {
				t.Errorf("expected result to contain %q, got %q", tc.expectedContain, got)
			}
		})
	}
}

func TestApplyCertAnnotations(t *testing.T) {
	t.Parallel()

	key, _ := newTestKey(t)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
	}
	cert, _ := newSelfSignedCertificate(t, template, key)

	tt := []struct {
		name                   string
		refreshReason          string
		expectedAnnotationKeys []string
	}{
		{
			name:          "cert annotation keys",
			refreshReason: "reason",
			expectedAnnotationKeys: []string{
				certsCountKey,
				certsNotBeforeKey,
				certsNotAfterKey,
				certsIsCAKey,
				certsIssuerKey,
				certsRefreshReasonKey,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			annotations := map[string]string{}
			ApplyCertAnnotations(annotations, []*x509.Certificate{cert}, tc.refreshReason)

			for _, expectedKey := range tc.expectedAnnotationKeys {
				_, found := annotations[expectedKey]
				if !found {
					t.Errorf("expected annotation key %q not found", expectedKey)
				}
			}
		})
	}
}

func TestApplyCertKeyAnnotations(t *testing.T) {
	t.Parallel()

	key, _ := newTestKey(t)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
	}
	cert, _ := newSelfSignedCertificate(t, template, key)

	tt := []struct {
		name                   string
		refreshReason          string
		expectedAnnotationKeys []string
	}{
		{
			name:          "cert and key annotation keys",
			refreshReason: "reason",
			expectedAnnotationKeys: []string{
				certsCountKey,
				certsNotBeforeKey,
				certsNotAfterKey,
				certsIsCAKey,
				certsIssuerKey,
				certsRefreshReasonKey,
				certsKeyType,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			annotations := map[string]string{}
			ApplyCertKeyAnnotations(annotations, []*x509.Certificate{cert}, key, tc.refreshReason)

			for _, expectedKey := range tc.expectedAnnotationKeys {
				_, found := annotations[expectedKey]
				if !found {
					t.Errorf("expected annotation key %q not found", expectedKey)
				}
			}
		})
	}
}

func TestGetCABundleFromConfigMap(t *testing.T) {
	t.Parallel()

	key, _ := newTestKey(t)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "ca"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
	}
	_, certBytes := newSelfSignedCertificate(t, template, key)

	tt := []struct {
		name              string
		cm                *corev1.ConfigMap
		expectedCertCount int
		expectedErr       error
	}{
		{
			name: "valid configmap",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
				Data: map[string]string{
					CABundleKey: string(certBytes),
				},
			},
			expectedCertCount: 1,
			expectedErr:       nil,
		},
		{
			name: "nil data",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ns",
				},
			},
			expectedCertCount: 0,
			expectedErr: fmt.Errorf(
				"configMap %q doesn't contain any data",
				types.NamespacedName{Namespace: "ns", Name: "test"},
			),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetCABundleFromConfigMap(tc.cm)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if len(got) != tc.expectedCertCount {
				t.Errorf("expected and got cert count differ:\n%s",
					cmp.Diff(tc.expectedCertCount, len(got)))
			}
		})
	}
}

func TestGetKeyInfo(t *testing.T) {
	t.Parallel()

	key, _ := newTestKey(t)

	ed25519Key, _ := newTestED25519Key(t)

	tt := []struct {
		name     string
		key      crypto.Signer
		expected string
	}{
		{
			name:     "RSA 2048 key",
			key:      key,
			expected: fmt.Sprintf("rsa/%d", key.Size()*8),
		},

		{
			name:     "non-RSA key returns unknown",
			key:      ed25519Key,
			expected: "<unknown>",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := getKeyInfo(tc.key)

			if got != tc.expected {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
			}
		})
	}
}
