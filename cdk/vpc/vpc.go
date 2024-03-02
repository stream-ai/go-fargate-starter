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
