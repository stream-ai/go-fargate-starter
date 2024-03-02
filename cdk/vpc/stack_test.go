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
