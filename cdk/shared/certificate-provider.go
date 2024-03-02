package shared

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// CertificateProvider is an interface for providing a Certificate
type CertificateProvider interface {
	// Certificate returns the certificate
	Certificate(scope constructs.Construct, id ResourceId) awscertificatemanager.ICertificate
}

// NewCertificateProvider returns a CertificateProvider that looks up a certificate by ARN
func NewCertificateLookupProvider(arn string) CertificateProvider {
	return &certificateLookupProvider{arn}
}

type certificateLookupProvider struct {
	arn string
}

func (p *certificateLookupProvider) Certificate(scope constructs.Construct, id ResourceId) awscertificatemanager.ICertificate {
	return awscertificatemanager.Certificate_FromCertificateArn(scope, id.String(), jsii.String(p.arn))
}

// NewMockCertificateProvider returns a CertificateProvider that creates a new Certificate from mocked domain name and alternative names
func NewMockCertificateProvider(domain string, subjectAlternativeNames *[]*string) CertificateProvider {
	return &mockCertificateProvider{}
}

type mockCertificateProvider struct {
	domain                  string
	subjectAlternativeNames *[]*string
}

func (p *mockCertificateProvider) Certificate(scope constructs.Construct, id ResourceId) awscertificatemanager.ICertificate {
	return awscertificatemanager.NewCertificate(scope, id.String(), &awscertificatemanager.CertificateProps{
		DomainName:              jsii.String(p.domain),
		SubjectAlternativeNames: p.subjectAlternativeNames,
	})
}
