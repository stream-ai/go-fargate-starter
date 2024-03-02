package shared

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ZoneProvider is an interface for providing a HostedZone
type ZoneProvider interface {
	HostedZone(scope constructs.Construct, id ResourceId) awsroute53.IHostedZone
}

// NewZoneLookupProvider returns a ZoneProvider that looks up the HostedZone by domain name
func NewZoneLookupProvider(domainName string) ZoneProvider {
	return &zoneLookupProvider{domainName: domainName}
}

type zoneLookupProvider struct {
	domainName string
	hostedZone awsroute53.IHostedZone
}

func (z *zoneLookupProvider) HostedZone(scope constructs.Construct, id ResourceId) awsroute53.IHostedZone {
	zone := awsroute53.HostedZone_FromLookup(scope, id.String(), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String(z.domainName),
	})
	return zone
}

// NewZoneProvider returns a ZoneProvider that creates a new HostedZone from mocked domain name and environment
func NewMockZoneProvider(domainName string, env awscdk.Environment) ZoneProvider {
	return &mockZoneProvider{domainName, env}
}

type mockZoneProvider struct {
	domainName string
	env        awscdk.Environment
}

func (z *mockZoneProvider) HostedZone(scope constructs.Construct, id ResourceId) awsroute53.IHostedZone {
	return awsroute53.NewHostedZone(scope, id.String(), &awsroute53.HostedZoneProps{
		ZoneName: jsii.String(z.domainName),
	})
}
