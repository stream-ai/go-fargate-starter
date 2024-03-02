package backend

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"gitlab.con/stream-ai/huddle/backend/cdk/shared"
)

type FargateConstruct interface {
	constructs.Construct
	FargateService() awsecspatterns.ApplicationLoadBalancedFargateService
	HealthCheck() *awselasticloadbalancingv2.HealthCheck
}

type fargateConstruct struct {
	constructs.Construct
	fargateService awsecspatterns.ApplicationLoadBalancedFargateService
}

func (f *fargateConstruct) FargateService() awsecspatterns.ApplicationLoadBalancedFargateService {
	return f.fargateService
}

func (f *fargateConstruct) HealthCheck() *awselasticloadbalancingv2.HealthCheck {
	return f.fargateService.TargetGroup().HealthCheck()
}

func NewFargateConstruct(
	// Common construct props
	scope constructs.Construct,
	id shared.ConstructId,
	// Fargate construct props
	memoryLimitMiB float64,
	cpu float64,
	vpc awsec2.IVpc,
	zoneProvider shared.ZoneProvider,
	certificateProvider shared.CertificateProvider,
) FargateConstruct {
	// Load Balanced Fargate Service
	dockerBuildArgs := map[string]*string{
		"TRAFFIC_PORT": jsii.String("80"),
	}
	assetImage := awsecs.ContainerImage_FromAsset(jsii.String("/workspaces/backend/"), &awsecs.AssetImageProps{
		File:      jsii.String("./service/Dockerfile"),
		BuildArgs: &dockerBuildArgs,
	})

	domainZone := zoneProvider.HostedZone(scope, id.Resource("zone"))

	logDriver := awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
		LogGroup: awslogs.NewLogGroup(scope, id.Resource("logGroup").String(), &awslogs.LogGroupProps{
			LogGroupName: id.Path(),
			Retention:    awslogs.RetentionDays_ONE_MONTH,
		}),
		StreamPrefix: jsii.String("service"),
	})

	loadBalancedFargateService := awsecspatterns.NewApplicationLoadBalancedFargateService(scope, id.Resource("service").String(), &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Vpc:            vpc,
		Certificate:    certificateProvider.Certificate(scope, id.Resource("certificate")),
		SslPolicy:      awselasticloadbalancingv2.SslPolicy_RECOMMENDED,
		MemoryLimitMiB: jsii.Number(memoryLimitMiB),
		Cpu:            jsii.Number(cpu),
		TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
			Image:         assetImage,
			LogDriver:     logDriver,
			ContainerName: jsii.String("http"),
		},
		PublicLoadBalancer: jsii.Bool(true),
		DomainName:         domainZone.ZoneName(),
		DomainZone:         domainZone,
		RedirectHTTP:       jsii.Bool(true),
		Protocol:           awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
	})
	loadBalancedFargateService.TargetGroup().ConfigureHealthCheck(&awselasticloadbalancingv2.HealthCheck{
		Port:                    jsii.String("80"),
		Path:                    jsii.String("/healthz"),
		Interval:                awscdk.Duration_Seconds(jsii.Number(5)),
		Timeout:                 awscdk.Duration_Seconds(jsii.Number(4)),
		HealthyThresholdCount:   jsii.Number(5),
		UnhealthyThresholdCount: jsii.Number(2),
	})
	return &fargateConstruct{scope, loadBalancedFargateService}
}
