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
