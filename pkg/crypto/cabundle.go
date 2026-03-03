package crypto

import (
	"crypto/x509"
	"maps"
	"slices"
	"time"
)

func MakeCABundle(currentCert *x509.Certificate, previousCerts []*x509.Certificate, now time.Time) []*x509.Certificate {
	certMap := map[string]*x509.Certificate{
		string(currentCert.Raw): currentCert,
	}

	for _, cert := range previousCerts {
		if now.After(cert.NotAfter) {
			continue
		}

		k := string(cert.Raw)

		_, isDuplicate := certMap[k]
		if isDuplicate {
			continue
		}

		certMap[k] = cert
	}

	return slices.Collect(maps.Values(certMap))
}
