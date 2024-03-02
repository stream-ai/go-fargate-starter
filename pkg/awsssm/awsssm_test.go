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
