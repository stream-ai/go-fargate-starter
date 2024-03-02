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
