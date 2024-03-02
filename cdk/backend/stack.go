package backend

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"gitlab.con/stream-ai/huddle/backend/cdk/shared"
)

type stack struct {
	fargateConstruct FargateConstruct
}

func (b *stack) LoadBalancerDNS() *string {
	return b.fargateConstruct.FargateService().LoadBalancer().LoadBalancerDnsName()
}

func (b *stack) FargateConstruct() constructs.Construct {
	return b.fargateConstruct
}

type Stack interface {
	FargateConstruct() constructs.Construct
	LoadBalancerDNS() *string
}

func NewStack(
	// Common Stack Properties
	envProvider shared.EnvProvider,
	scope constructs.Construct,
	id shared.StackId,
	tags map[string]*string,
	// Backend Stack Properties
	ecsCpu float64,
	ecsMemoryLimit float64,
	vpc awsec2.IVpc,
	zoneProvider shared.ZoneProvider,
	certificateProvider shared.CertificateProvider,
) Stack {
	cdkStack := awscdk.NewStack(scope, id.String(), &awscdk.StackProps{
		Tags: &tags,
		Env:  envProvider.Env(),
	})

	fargateConstruct := NewFargateConstruct(
		cdkStack,
		id.Construct("fargate"),
		ecsMemoryLimit,
		ecsCpu,
		vpc,
		zoneProvider,
		certificateProvider,
	)

	awscdk.NewCfnOutput(cdkStack, id.CfnOutput("LoadBalancerDNS"), &awscdk.CfnOutputProps{Value: fargateConstruct.FargateService().LoadBalancer().LoadBalancerDnsName()})

	return &stack{fargateConstruct}
}
