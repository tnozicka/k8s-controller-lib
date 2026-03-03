package crypto

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractDesiredFieldsFromTemplate(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name             string
		template         *x509.Certificate
		expectedTemplate *DesiredCertTemplate
	}{
		{
			name: "extracts CommonName and fields from template correctly",
			template: &x509.Certificate{
				Subject: pkix.Name{
					CommonName:   "my-cert",
					Organization: []string{"my-org"},
				},
				IsCA:                  true,
				BasicConstraintsValid: true,
				KeyUsage:              x509.KeyUsageCertSign,
				DNSNames:              []string{"example.com"},
			},
			expectedTemplate: &DesiredCertTemplate{
				Subject: pkixName{
					Organization: []string{"my-org"},
					CommonName:   "my-cert",
				},
				KeyUsage:              x509.KeyUsageCertSign,
				BasicConstraintsValid: true,
				IsCA:                  true,
				DNSNames:              []string{"example.com"},
				IPAddresses:           []net.IP{},
			},
		},
		{
			name: "extracts empty fields correctly",
			template: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "",
				},
			},
			expectedTemplate: &DesiredCertTemplate{
				IPAddresses: []net.IP{},
			},
		},
		{
			name: "extracts IPAddresses and ExtKeyUsage",
			template: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "serving-cert",
				},
				DNSNames:    []string{"svc.ns.svc.cluster.local"},
				IPAddresses: []net.IP{net.ParseIP("10.0.0.1")},
				KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			expectedTemplate: &DesiredCertTemplate{
				Subject: pkixName{
					CommonName: "serving-cert",
				},
				KeyUsage: x509.KeyUsageDigitalSignature |
					x509.KeyUsageKeyEncipherment,
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
				DNSNames:    []string{"svc.ns.svc.cluster.local"},
				IPAddresses: []net.IP{net.ParseIP("10.0.0.1").To16()},
			},
		},
		{
			name: "extracts client cert fields",
			template: &x509.Certificate{
				Subject: pkix.Name{
					CommonName:   "client",
					Organization: []string{"system:masters"},
				},
				KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
			expectedTemplate: &DesiredCertTemplate{
				Subject: pkixName{
					Organization: []string{"system:masters"},
					CommonName:   "client",
				},
				KeyUsage: x509.KeyUsageDigitalSignature |
					x509.KeyUsageKeyEncipherment,
				ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
				IPAddresses: []net.IP{},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ExtractDesiredFieldsFromTemplate(tc.template)
			if !reflect.DeepEqual(got, tc.expectedTemplate) {
				t.Errorf("expected and got template differ:\n%s", cmp.Diff(tc.expectedTemplate, got))
			}
		})
	}
}

func TestDesiredCertTemplateToJSON(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		template     *DesiredCertTemplate
		expectedJSON []byte
		expectedErr  error
	}{
		{
			name:         "empty template returns valid JSON",
			template:     &DesiredCertTemplate{},
			expectedJSON: []byte(`{"Subject":{"Country":null,"Organization":null,"OrganizationalUnit":null,"Locality":null,"Province":null,"StreetAddress":null,"PostalCode":null,"CommonName":""},"KeyUsage":0,"ExtKeyUsage":null,"BasicConstraintsValid":false,"IsCA":false,"MaxPathLen":0,"MaxPathLenZero":false,"OCSPServer":null,"IssuingCertificateURL":null,"DNSNames":null,"EmailAddresses":null,"IPAddresses":null,"URIs":null,"PermittedDNSDomainsCritical":false,"PermittedDNSDomains":null,"ExcludedDNSDomains":null,"PermittedIPRanges":null,"ExcludedIPRanges":null,"PermittedEmailAddresses":null,"ExcludedEmailAddresses":null,"PermittedURIDomains":null,"ExcludedURIDomains":null,"CRLDistributionPoints":null,"PolicyIdentifiers":null}`),
			expectedErr:  nil,
		},
		{
			name: "returns valid JSON with expected fields",
			template: &DesiredCertTemplate{
				Subject: pkixName{
					CommonName:   "json-test",
					Organization: []string{"test-org"},
				},
				IsCA:        true,
				DNSNames:    []string{"example.com"},
				IPAddresses: []net.IP{net.ParseIP("10.0.0.1")},
			},
			expectedJSON: []byte(`{"Subject":{"Country":null,"Organization":["test-org"],"OrganizationalUnit":null,"Locality":null,"Province":null,"StreetAddress":null,"PostalCode":null,"CommonName":"json-test"},"KeyUsage":0,"ExtKeyUsage":null,"BasicConstraintsValid":false,"IsCA":true,"MaxPathLen":0,"MaxPathLenZero":false,"OCSPServer":null,"IssuingCertificateURL":null,"DNSNames":["example.com"],"EmailAddresses":null,"IPAddresses":["10.0.0.1"],"URIs":null,"PermittedDNSDomainsCritical":false,"PermittedDNSDomains":null,"ExcludedDNSDomains":null,"PermittedIPRanges":null,"ExcludedIPRanges":null,"PermittedEmailAddresses":null,"ExcludedEmailAddresses":null,"PermittedURIDomains":null,"ExcludedURIDomains":null,"CRLDistributionPoints":null,"PolicyIdentifiers":null}`),
			expectedErr:  nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.template.ToJSON()
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}
			if !reflect.DeepEqual(got, tc.expectedJSON) {
				t.Errorf("expected and got JSON differ:\n%s", cmp.Diff(string(tc.expectedJSON), string(got)))
			}
		})
	}
}
