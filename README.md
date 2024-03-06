# Huddle Backend

## Useful Tools

### `printgo | xclip`

`printgo` prints all source files at the current directory and subdirectories to stdout.
`xclip` can be used to send stdin from a Devcontainer to the Windows clipboard by using `socat` running on the WSL2 host.

To set it up:

1. On the WSL host, install `socat` and set up aliases:
```bash
sudo apt install socat
alias runsocat='socat tcp-listen:8121,fork,bind=0.0.0.0 EXEC:"clip.exe"'
```
2. In the Devcontainer, install `socat` and set up aliases.
> NOTE: this is already being done in this project in `$/.devcontainer/post-create.sh`
```zsh
sudo apt install socat
alias xclip='socat - tcp:host.docker.internal:8121'
alias printgo='$(git rev-parse --show-toplevel)/.scripts/print_go_files.sh'
```
3. In the devcontainer, copy the code at the current subdirectory to the windows clipboard:
```zsh
cd service/cmd
printgo | xclip
```

After the above command, the Windows clipboard now contains:
```
==> Contents of cdk/appstacks/main.go <==
```(golang)
package appstacks

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"gitlab.con/stream-ai/huddle/backend/cdk/backend"
	"gitlab.con/stream-ai/huddle/backend/cdk/shared"
	"gitlab.con/stream-ai/huddle/backend/cdk/vpc"
)

type BuildReturn struct {
	VpcStack     vpc.Stack
	BackendStack backend.Stack
}

func Build(scope constructs.Construct,
	appEnvName string,
	// vpc props
	vpcMaxAzs float64,
	// backend props
	backendCpu float64,
	backendMemoryLimit float64,
	backendDomainName string,
	backendCertArn string,
) BuildReturn {
	// stackTags() returns a map of tags for a stack, including the "cdk-stack" tag
	stackTags := func(stackName string) map[string]*string {
		tags := make(map[string]*string)
		tags["cdk-stack"] = jsii.String(stackName)
		tags["app-environment"] = jsii.String(appEnvName)
		return tags
	}

	stackId := func(stackName string) shared.StackId {
		return shared.StackId(appEnvName + shared.Sep + "huddle" + shared.Sep + stackName)
	}

	vpcStack := vpc.NewStack(
		shared.NewDefaultEnvProvider(),
		scope,
		stackId("vpc"),
		stackTags("vpc"),
		vpcMaxAzs)

	vpc := vpcStack.Vpc()

	backendStack := backend.NewStack(
		shared.NewDefaultEnvProvider(),
		scope,
		stackId("backend"),
		stackTags("backend"),
		backendCpu,
		backendMemoryLimit,
		vpc,
		shared.NewZoneLookupProvider(backendDomainName),
		shared.NewCertificateLookupProvider(backendCertArn),
	)

	return BuildReturn{vpcStack, backendStack}
}
```

==> Contents of cdk/appstacks/main_test.go <==
```(golang)
package appstacks_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
	"gitlab.con/stream-ai/huddle/backend/cdk/appstacks"
)

func Test_Integrationm(t *testing.T) {
	defer jsii.Close()

	type test struct {
		// vpc props
		vpcMaxAzs float64
		// backend props
		backendCpu         float64
		backendMemoryLimit float64
		backendDomainName  string
		backendTrafficPort int
		backendCertArn     string
	}

	tests := []test{
		{
			// vpc props
			vpcMaxAzs: 2,
			// backend props
			backendCpu:         256,
			backendMemoryLimit: 512,
			backendDomainName:  "test.example.com",
			backendTrafficPort: 8080,
			backendCertArn:     "arn:aws:acm:us-west-2:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		},
	}

	for _, tc := range tests {
		app := awscdk.NewApp(nil)

		// get an environment from the current AWS profile

		ret := appstacks.Build(app,
			"jdibling",
			tc.vpcMaxAzs,
			tc.backendCpu,
			tc.backendMemoryLimit,
			tc.backendDomainName,
			tc.backendCertArn,
		)

		// check the vpc stack
		azs := ret.VpcStack.Vpc().AvailabilityZones()
		assert.NotEmpty(t, azs)
		assert.Len(t, *azs, int(tc.vpcMaxAzs))

		// template := assertions.Template_FromStack(ret.VpcStack.Vpc().Stack(), nil)
		// b, e := json.MarshalIndent(template.ToJSON(), "", " ")
		// require.NoError(t, e)
		// log.Printf("%s\n", b)

		// check the backend stack
		assert.NotNil(t, ret.BackendStack.FargateConstruct())
		assert.NotNil(t, ret.BackendStack.LoadBalancerDNS())
	}
}
```

==> Contents of cdk/backend/fargate.go <==
```(golang)
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
```

==> Contents of cdk/backend/fargate_test.go <==
```(golang)
package backend_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
	"gitlab.con/stream-ai/huddle/backend/cdk/backend"
	"gitlab.con/stream-ai/huddle/backend/cdk/shared"
)

type mockVpcConstruct struct {
	awsec2.IVpc
}

var mockEnv = awscdk.Environment{
	Account: jsii.String("123456789012"),
	Region:  jsii.String("us-west-2"),
}

func TestNewFargateConstruct(t *testing.T) {
	defer jsii.Close()

	type test struct {
		memory       float64
		cpu          float64
		domain       string
		zoneProvider shared.ZoneProvider
		certProvider shared.CertificateProvider
	}

	tests := []test{
		{
			memory:       1024,
			cpu:          256,
			domain:       "test",
			zoneProvider: shared.NewMockZoneProvider("test", mockEnv),
			certProvider: shared.NewMockCertificateProvider("example.com", jsii.Strings("api.example.com")),
		},
	}

	for _, tc := range tests {
		app := awscdk.NewApp(nil)
		stackId := shared.StackId("testStack")
		stack := awscdk.NewStack(app, stackId.String(), &awscdk.StackProps{
			Env: &mockEnv,
		})

		constructId := stackId.Construct("network")
		vpc := awsec2.NewVpc(stack, constructId.Resource("vpc").String(), &awsec2.VpcProps{
			MaxAzs: jsii.Number(2),
		})

		// Create the FargateConstruct
		fargate := backend.NewFargateConstruct(stack, "fargate",
			tc.memory,
			tc.cpu,
			vpc,
			tc.zoneProvider,
			tc.certProvider,
		)

		// Assert that the FargateConstruct is created correctly
		assert.NotNil(t, fargate)
		// Assert that the health check is on port 80 at /healthz
		healthCheck := fargate.HealthCheck()
		assert.NotNil(t, healthCheck)
		assert.NotNil(t, healthCheck.Port)
		assert.NotNil(t, healthCheck.Path)
		if *healthCheck.Port != "80" || *healthCheck.Path != "/healthz" {
			t.Errorf("Expected health check on port 80 at /healthz, got port %s and path %s", *healthCheck.Port, *healthCheck.Path)
		}
	}
}
```

==> Contents of cdk/backend/stack.go <==
```(golang)
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
```

==> Contents of cdk/main.go <==
```(golang)
package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"gitlab.con/stream-ai/huddle/backend/cdk/appstacks"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Build app stacks
	appstacks.Build(app,
		"jdibling",
		2,                  // vpc.maxAzs
		256,                // backend.cpu
		512,                // backend.memoryLimit
		"jdibling.hudl.ai", // backend.domainName
		"arn:aws:acm:us-east-1:590184032693:certificate/68789120-b333-423e-bd3a-2573a95b534d", // backend.certArn
	)

	app.Synth(nil)
}
```

==> Contents of cdk/shared/certificate-provider.go <==
```(golang)
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
```

==> Contents of cdk/shared/env-provider.go <==
```(golang)
// Package shared provides environment provider implementations for AWS CDK.
package shared

import (
	"context"
	"log"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// EnvProvider is an interface that defines methods for retrieving AWS CDK environment information.
type EnvProvider interface {
	Env() *awscdk.Environment
}

// defaultEnvProvider is a implementation of the EnvProvider interface which uses the default AWS configuration to retrieve the environment.
type defaultEnvProvider struct{}

// Env retrieves the AWS CDK environment using the default AWS configuration.
func (p *defaultEnvProvider) Env() *awscdk.Environment {
	// Load the default AWS configuration (credentials, region from environment or config file)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error loading AWS configuration:", err)
	}

	// Create an STS client
	stsClient := sts.NewFromConfig(cfg)

	// Get the account ID using GetCallerIdentity
	identity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal("Error getting caller identity:", err)
	}

	// TODO: We may need to add a way to assume a role here

	// Get the current region from the configuration
	return &awscdk.Environment{
		Account: identity.Account,
		Region:  aws.String(cfg.Region),
	}
}

// NewDefaultEnvProvider returns a new EnvProvider that uses the default AWS configuration.
func NewDefaultEnvProvider() EnvProvider {
	return &defaultEnvProvider{}
}

// mockEnvProvider is a mock implementation of the EnvProvider interface.
type mockEnvProvider struct {
	env awscdk.Environment
}

// Env retrieves the AWS CDK environment from the mock environment.
func (p *mockEnvProvider) Env() *awscdk.Environment {
	return &p.env
}

// NewMockEnvProvider returns a new EnvProvider that returns the given environment.
func NewMockEnvProvider(env awscdk.Environment) EnvProvider {
	return &mockEnvProvider{env}
}
```

==> Contents of cdk/shared/hostedZone-provider.go <==
```(golang)
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
```

==> Contents of cdk/shared/id.go <==
```(golang)
package shared

import (
	"strings"

	"github.com/aws/jsii-runtime-go"
)

type (
	StackId     string
	ConstructId string
	ResourceId  string
)

const Sep = "-"

type StringPointer interface {
	String() *string
}

type StringPather interface {
	Path() *string
}

func fmtPath(sp StringPointer) *string {
	return jsii.String("/" + strings.ReplaceAll(*sp.String(), Sep, "/"))
}

func (s StackId) String() *string {
	return jsii.String(string(s))
}

func (s StackId) Path() *string {
	return fmtPath(&s)
}

func (s StackId) Construct(id string) ConstructId {
	return ConstructId(*s.String() + Sep + id)
}

func (s StackId) CfnOutput(id string) *string {
	return jsii.String(*s.String() + id)
}

func (c ConstructId) String() *string {
	return jsii.String(string(c))
}

func (c ConstructId) Path() *string {
	return fmtPath(&c)
}

func (c ConstructId) Resource(id string) ResourceId {
	return ResourceId(*c.String() + Sep + id)
}

func (r ResourceId) String() *string {
	return jsii.String(string(r))
}

func (r ResourceId) Path() *string {
	return fmtPath(&r)
}
```

==> Contents of cdk/vpc/stack.go <==
```(golang)
package vpc

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"gitlab.con/stream-ai/huddle/backend/cdk/shared"
)

type stack struct {
	// awscdk.Stack
	vpc awsec2.IVpc
}

func (v *stack) Vpc() awsec2.IVpc {
	return v.vpc
}

type Stack interface {
	// awscdk.Stack
	Vpc() awsec2.IVpc
}

func NewStack(
	// Common Stack Properties
	envProvider shared.EnvProvider,
	scope constructs.Construct,
	id shared.StackId,
	tags map[string]*string,
	// VPC Stack Properties
	maxAzs float64,
) Stack {
	cdkStack := awscdk.NewStack(scope, id.String(), &awscdk.StackProps{
		Tags: &tags,
		Env:  envProvider.Env(),
	})

	vpcConstruct := NewVpcConstruct(cdkStack, id.Construct("backend"), maxAzs)

	return &stack{
		// awscdk.Stack(stack),
		vpc: vpcConstruct.Vpc(),
	}
}
```

==> Contents of cdk/vpc/stack_test.go <==
```(golang)
package vpc_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
	"gitlab.con/stream-ai/huddle/backend/cdk/vpc"
)

func TestVpcConstruct(t *testing.T) {
	defer jsii.Close()
	// app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(nil, nil, nil)

	construct := vpc.NewVpcConstruct(stack, "TestStack",
		2,
	)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::EC2::VPC"), map[string]any{
		"Tags": assertions.Match_AnyValue(),
	})
	// verify the VPC is created with the correct number of AZs
	azs := construct.Vpc().AvailabilityZones()
	assert.NotNil(t, azs)
	assert.Equal(t, len(*azs), 2)
}
```

==> Contents of cdk/vpc/vpc.go <==
```(golang)
package vpc

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"gitlab.con/stream-ai/huddle/backend/cdk/shared"
)

type VpcConstruct interface {
	constructs.Construct
	Vpc() awsec2.IVpc
}

type vpcConstruct struct {
	constructs.Construct
	vpc *awsec2.Vpc
}

func (v *vpcConstruct) Vpc() awsec2.IVpc {
	return *v.vpc
}

func NewVpcConstruct(
	// Common construct props
	scope constructs.Construct,
	id shared.ConstructId,
	// VPC construct props
	maxAzs float64,
) VpcConstruct {
	this := constructs.NewConstruct(scope, id.String())

	// Create VPC and Cluster
	vpc := awsec2.NewVpc(this, id.Resource("vpc").String(), &awsec2.VpcProps{
		MaxAzs: jsii.Number(maxAzs),
	})

	return &vpcConstruct{this, &vpc}
}
```

==> Contents of pkg/awsssm/awsssm.go <==
```(golang)
package awsssm

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var ErrReadFailed = errors.New("read failed")

type GetParameterAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

func NewParameterReader(client GetParameterAPI, in *ssm.GetParameterInput) io.Reader {
	return &parameterReader{
		client: client,
		in:     in,
	}
}

type parameterReader struct {
	client GetParameterAPI
	in     *ssm.GetParameterInput
}

func (r *parameterReader) Read(p []byte) (n int, err error) {
	out, err := r.client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           r.in.Name,
		WithDecryption: r.in.WithDecryption,
	})
	if err != nil {
		return 0, errors.Join(ErrReadFailed, err)
	}

	return copy(p, []byte(*out.Parameter.Value)), io.EOF
}
```

==> Contents of pkg/awsssm/awsssm_test.go <==
```(golang)
package awsssm_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"gitlab.con/stream-ai/huddle/backend/pkg/awsssm"
)

type mockGetParameterAPI struct{}

func (m *mockGetParameterAPI) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	if params.Name == nil {
		return nil, errors.New("name is required")
	}
	return &ssm.GetParameterOutput{
		Parameter: &types.Parameter{
			Value: aws.String("mock-value"),
		},
	}, nil
}

func Test_ParameterReader(t *testing.T) {
	type test struct {
		param   *ssm.GetParameterInput
		want    string
		wantErr error
	}
	tests := []test{
		{
			param: &ssm.GetParameterInput{
				Name:           aws.String("/my/parameter"),
				WithDecryption: aws.Bool(false),
			},
			want: "mock-value",
		},
		{
			param: &ssm.GetParameterInput{
				Name:           aws.String("/my/parameter"),
				WithDecryption: aws.Bool(true),
			},
			want: "mock-value",
		},
		{
			param: &ssm.GetParameterInput{
				Name: nil,
			},
			wantErr: awsssm.ErrReadFailed,
		},
	}

	for _, tc := range tests {
		client := &mockGetParameterAPI{}
		reader := awsssm.NewParameterReader(client, tc.param)

		b, err := io.ReadAll(reader)
		if tc.wantErr != nil {
			assert.ErrorIs(t, err, tc.wantErr)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.want, string(b))
		}
	}
}
```

==> Contents of service/cmd/root.go <==
```(golang)
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.con/stream-ai/huddle/backend/service/server"
)

var (
	DefaultServerPort = "80"

	ErrProcessingOptions = fmt.Errorf("error processing options")

	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "cobra-cli",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// ctx := context.Background()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			slog.SetDefault(logger)
			slog.Info("app started", "config-file", viper.ConfigFileUsed())

			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return errors.Join(ErrProcessingOptions, err)
			}
			if configFile != "" {
				logger.Info("config file used", "file", viper.ConfigFileUsed())
			}

			// addr := fmt.Sprintf(":%s", viper.GetString("port"))
			addr := fmt.Sprintf(":%s", viper.GetString("port"))
			return server.Run(cmd.Context(), logger, addr)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.huddle-backend.yaml)")
	rootCmd.PersistentFlags().StringP("port", "p", DefaultServerPort, "server port")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("/app")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".huddle-backend")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			cobra.CheckErr(err)
		}
	}
}
```

==> Contents of service/main.go <==
```(golang)
package main

import (
	"gitlab.con/stream-ai/huddle/backend/service/cmd"
)

func main() {
	cmd.Execute()
}
```

==> Contents of service/main_test.go <==
```(golang)
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.con/stream-ai/huddle/backend/service/server"
)

func TestIntegration(t *testing.T) {
	// Start the server in a separate goroutine
	go func() {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
		addr := fmt.Sprintf(":8080")
		err := server.Run(context.Background(), logger, addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running server: %s\n", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// path, expectedStatusCode, expectedBody
	// Make a request to the server
	resp, err := http.Get("http://localhost:8080/healthz")
	assert.NoError(t, err)
	defer resp.Body.Close()
	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
	}
	assert.Equal(t, "OK", string(body))
	// Assert the response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func RepoRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to determine repository root directory: %v", err)
	}
	return strings.TrimSpace(string(output))
}
```

==> Contents of service/middleware/loggermw/loggermw.go <==
```(golang)
package loggermw

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
)

type rw struct {
	http.ResponseWriter
	statusCode int
	resp       bytes.Buffer
}

func (r *rw) Write(b []byte) (int, error) {
	// copy b to r.resp using io.Copy
	_, err := io.Copy(&r.resp, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	return r.ResponseWriter.Write(b)
}

func (r *rw) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.statusCode = statusCode
}

func (r *rw) Header() http.Header {
	return r.ResponseWriter.Header()
}

func New(logger *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &rw{ResponseWriter: w}
		h.ServeHTTP(rw, r)
		if logger != nil {
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"response", string(rw.resp.String()),
			)
		}
	})
}
```

==> Contents of service/server/health.go <==
```(golang)
package server

import (
	"io"
	"log/slog"
	"net/http"
)

func handleHealthZ() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("health check")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}
}
```

==> Contents of service/server/routes.go <==
```(golang)
package server

import (
	"log/slog"
	"net/http"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.Handle("GET /healthz", handleHealthZ())
}
```

==> Contents of service/server/server.go <==
```(golang)
package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"gitlab.con/stream-ai/huddle/backend/service/middleware/loggermw"
)

func new(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger)

	var handler http.Handler = mux
	handler = loggermw.New(logger, handler)

	return handler
}

func Run(ctx context.Context, logger *slog.Logger, addr string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	srv := new(logger)
	httpServer := http.Server{
		Addr:    addr,
		Handler: srv,
	}
	go func() {
		log.Printf("server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
		log.Println("server stopped")
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("shutting down server")
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}
```




`printgo` is a ZSH alias which calls `$/.scripts/print_go_files.sh`. This script enumerates all source code files int eh current directory and all subdirectories, and prints the contents of each to stdout.

This is especially useful when used in combination with `xclip` which is another alias. `xclip` works similarly to the `xclip` from standard Linux packages, but it 
works within a Devcontainer. `xcip` receives input on stdin, and sends that input to a socket server using `socat`. The WSL host is running `socat
