package kubecrypto

import (
	"context"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	clcrypto "github.com/tnozicka/k8s-controller-lib/pkg/crypto"
	"github.com/tnozicka/k8s-controller-lib/pkg/helpers/maps"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type MetaConfig struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
}

type CertificateConfig struct {
	MetaConfig
	Validity    time.Duration
	Refresh     time.Duration
	CertCreator clcrypto.CertCreator
}

type CAConfig struct {
	MetaConfig
	Validity              time.Duration
	Refresh               time.Duration
	BundleConfigOverrides *MetaConfig
}

func (cac CAConfig) GetBundleConfigMeta() MetaConfig {
	if cac.BundleConfigOverrides == nil {
		return cac.MetaConfig
	}

	return *cac.BundleConfigOverrides
}

type CertChainConfig struct {
	// Intermediate CA is always present and signs the leaf certificate.
	// Can be a top level CA, when self-signed.
	CAConfig CAConfig

	// CertConfigs are leaf certificates that will be signed by the intermediate CA.
	// At least one certificate config must be present.
	CertConfigs []CertificateConfig
}

type CertificateManager interface {
	MakeCertificateChain(
		ctx context.Context,
		certChainConfig *CertChainConfig,
		existingSecrets map[string]*corev1.Secret,
		existingConfigMaps map[string]*corev1.ConfigMap,
	) ([]*corev1.Secret, []*corev1.ConfigMap, bool, error)
}

type certificateManager struct {
	namespace    string
	keyGenerator clcrypto.KeyGenerator
	now          func() time.Time
}

var _ CertificateManager = &certificateManager{}

func NewCertificateManager(
	namespace string,
	keyGenerator clcrypto.KeyGenerator,
	nowFunc func() time.Time,
) CertificateManager {
	return &certificateManager{
		namespace:    namespace,
		keyGenerator: keyGenerator,
		now:          nowFunc,
	}
}

// TODO: unify the CA and cert methods
func (cm *certificateManager) makeCASecret(
	ctx context.Context,
	caConfig CAConfig,
	keyGenerator clcrypto.KeyGenerator,
	existingSecrets map[string]*corev1.Secret,
) (*TLSSecret, bool, error) {
	var err error
	var caTLSSecret *TLSSecret
	existingCASecret, found := existingSecrets[caConfig.Name]
	if found {
		caTLSSecret = NewTLSSecret(existingCASecret)
	} else {
		caTLSSecret = NewEmptyTLSSecret(metav1.ObjectMeta{
			Namespace: cm.namespace,
			Name:      caConfig.Name,
		})
	}

	maps.Merge(caTLSSecret.GetSecret().Annotations, caConfig.Annotations)
	maps.Merge(caTLSSecret.GetSecret().Labels, caConfig.Labels)

	caCertCreator := (&clcrypto.CACertCreatorConfig{
		Subject: pkix.Name{
			CommonName: caConfig.Name,
		},
	}).ToCreator()
	caCertTemplate := caCertCreator.MakeCertificateTemplate(cm.now(), caConfig.Validity)

	// Self-signed signer doesn't have a public key.
	// (This can be expanded to support external signers in the future.)
	var issuerCert *x509.Certificate
	caRefreshReason := caTLSSecret.CalculateRefreshReason(
		cm.now(),
		caConfig.Refresh,
		caCertTemplate,
		issuerCert,
		types.NamespacedName{Namespace: cm.namespace, Name: caConfig.Name},
	)

	if len(caRefreshReason) != 0 {
		var caPrivateKey crypto.Signer
		caPrivateKey, err = keyGenerator.GenerateKey(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("can't generate key: %w", err)
		}

		caSigner := clcrypto.NewSelfSignedSigner(cm.now, caPrivateKey)
		var caCert *x509.Certificate
		caCert, err = caSigner.SignCertificate(caCertTemplate, caPrivateKey.Public())
		if err != nil {
			return nil, false, fmt.Errorf("can't sign CA certificate %q: %w", naming.ObjNN(caTLSSecret.GetSecret()), err)
		}

		err = caTLSSecret.SetCertKey(caCert, caPrivateKey)
		if err != nil {
			return nil, false, fmt.Errorf("can't set CA certificate %q: %w", naming.ObjNN(caTLSSecret.GetSecret()), err)
		}

		err = caTLSSecret.UpdateProjectedAnnotations(caRefreshReason)
		if err != nil {
			return nil, false, fmt.Errorf("can't update projected annotations on secret %q: %w", naming.ObjNN(caTLSSecret.GetSecret()), err)
		}

		return caTLSSecret, false, nil
	}

	return caTLSSecret, true, nil
}

func (cm *certificateManager) makeCertSecret(
	ctx context.Context,
	certConfig CertificateConfig,
	signer clcrypto.Signer,
	keyGenerator clcrypto.KeyGenerator,
	existingSecrets map[string]*corev1.Secret,
) (*TLSSecret, error) {
	var err error
	var certTLSSecret *TLSSecret
	existingCertSecret, found := existingSecrets[certConfig.Name]
	if found {
		certTLSSecret = NewTLSSecret(existingCertSecret)
	} else {
		certTLSSecret = NewEmptyTLSSecret(metav1.ObjectMeta{
			Namespace: cm.namespace,
			Name:      certConfig.Name,
		})
	}

	maps.Merge(certTLSSecret.GetSecret().Annotations, certConfig.Annotations)
	maps.Merge(certTLSSecret.GetSecret().Labels, certConfig.Labels)

	certTemplate := certConfig.CertCreator.MakeCertificateTemplate(cm.now(), certConfig.Validity)

	signerCert, err := signer.GetCertificate()
	if err != nil {
		return nil, fmt.Errorf("can't get signer certificate: %w", err)
	}

	certRefreshReason := certTLSSecret.CalculateRefreshReason(
		cm.now(),
		certConfig.Refresh,
		certTemplate,
		signerCert,
		types.NamespacedName{Namespace: cm.namespace, Name: certConfig.Name},
	)

	if len(certRefreshReason) != 0 {
		var certPrivateKey crypto.Signer
		certPrivateKey, err = keyGenerator.GenerateKey(ctx)
		if err != nil {
			return nil, fmt.Errorf("can't generate key: %w", err)
		}

		var cert *x509.Certificate
		cert, err = signer.SignCertificate(certTemplate, certPrivateKey.Public())
		if err != nil {
			return nil, fmt.Errorf("can't sign certificate %q: %w", naming.ObjNN(certTLSSecret.GetSecret()), err)
		}

		err = certTLSSecret.SetCertKey(cert, certPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("can't set certificate %q: %w", naming.ObjNN(certTLSSecret.GetSecret()), err)
		}

		err = certTLSSecret.UpdateProjectedAnnotations(certRefreshReason)
		if err != nil {
			return nil, fmt.Errorf("can't update projected annotations on secret %q: %w", naming.ObjNN(certTLSSecret.GetSecret()), err)
		}
	}

	return certTLSSecret, nil
}

// MakeCertificateChain creates a certificate chain from the given config.
// Note that this is an iterative process that makes sure the signers get applied (present in existing objects)
// before using them to sign new certificates. This makes sure that we don't ever apply the target certificate
// while failing to update the signer.
func (cm *certificateManager) MakeCertificateChain(
	ctx context.Context,
	certChainConfig *CertChainConfig,
	existingSecrets map[string]*corev1.Secret,
	existingConfigMaps map[string]*corev1.ConfigMap,
) ([]*corev1.Secret, []*corev1.ConfigMap, bool, error) {
	var secrets []*corev1.Secret
	var configMaps []*corev1.ConfigMap

	caTLSSecret, isCATLSSecretUpToDate, err := cm.makeCASecret(
		ctx,
		certChainConfig.CAConfig,
		cm.keyGenerator,
		existingSecrets,
	)
	if err != nil {
		return nil, nil, false, fmt.Errorf("can't make CA TLS secret: %w", err)
	}
	secrets = append(secrets, caTLSSecret.GetSecret())
	// Wait for the secret to be applied and observed.
	if !isCATLSSecretUpToDate {
		klog.V(4).InfoS("Waiting for CA TLS secret to be applied", "Secret", naming.ObjNN(caTLSSecret.GetSecret()))
		return secrets, configMaps, false, nil
	}

	// At this point the TLS secret lives in the API, so we can use it as a controller for CA bundle projection.
	var caCABundle *corev1.ConfigMap
	caCABundle, err = caTLSSecret.MakeCABundle(
		certChainConfig.CAConfig.GetBundleConfigMeta().Name,
		existingConfigMaps[certChainConfig.CAConfig.GetBundleConfigMeta().Name],
		cm.now(),
	)
	if err != nil {
		return nil, nil, false, fmt.Errorf("can't make CA bundle for secret %q: %w", naming.ObjNN(caTLSSecret.GetSecret()), err)
	}

	// Bundles don't have internal dependencies, so we don't have to return when updated.
	configMaps = append(configMaps, caCABundle)

	caSigningTLSSecret := NewSigningTLSSecret(caTLSSecret, cm.now)

	for _, cc := range certChainConfig.CertConfigs {
		var certTLSSecret *TLSSecret
		certTLSSecret, err = cm.makeCertSecret(
			ctx,
			cc,
			caSigningTLSSecret,
			cm.keyGenerator,
			existingSecrets,
		)
		if err != nil {
			return nil, nil, false, fmt.Errorf(
				"can't make certificate TLS secret %q: %w",
				naming.ObjNN(&metav1.ObjectMeta{Namespace: cm.namespace, Name: cc.Name}),
				err,
			)
		}

		secrets = append(secrets, certTLSSecret.GetSecret())
	}

	return secrets, configMaps, true, nil
}
