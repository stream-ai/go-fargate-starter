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
