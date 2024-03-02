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
