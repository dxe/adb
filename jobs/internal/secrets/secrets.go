// Package secrets resolves SecureString values from AWS SSM Parameter Store at
// runtime. Resolving secrets in-process (rather than passing them through
// `sam deploy --parameter-overrides`) keeps plaintext secrets out of the deploy
// command line and the CloudFormation template.
package secrets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Client fetches and decrypts SSM parameters. The region is taken from the
// Lambda-provided AWS_REGION environment variable via the default AWS config.
type Client struct {
	ssm *ssm.Client
}

// New builds a Client using the ambient AWS configuration.
func New(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return &Client{ssm: ssm.NewFromConfig(cfg)}, nil
}

// Get returns the decrypted value of the named SecureString parameter.
func (c *Client) Get(ctx context.Context, name string) (string, error) {
	out, err := c.ssm.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("getting SSM parameter %q: %w", name, err)
	}
	if out.Parameter == nil || out.Parameter.Value == nil {
		return "", fmt.Errorf("SSM parameter %q has no value", name)
	}
	return *out.Parameter.Value, nil
}
