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
