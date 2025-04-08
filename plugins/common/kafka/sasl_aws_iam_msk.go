package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"
)

const saslTypeOAuthAWSMSKIAM = "AWS-MSK-IAM"

type SASLOAuthAWSMSKIAMConfig struct {
	SASLAWSRegion  string `toml:"sasl_aws_msk_iam_region"`
	SASLAWSProfile string `toml:"sasl_aws_msk_iam_profile"`
	SASLAWSRole    string `toml:"sasl_aws_msk_iam_role"`
	SASLAWSSession string `toml:"sasl_aws_msk_iam_session"`
}

func (c *SASLOAuthAWSMSKIAMConfig) tokenProvider(extensions map[string]string) (sarama.AccessTokenProvider, error) {
	if c.SASLAWSRegion == "" {
		return nil, errors.New("region cannot be empty")
	}

	if c.SASLAWSProfile != "" && (c.SASLAWSRole != "" || c.SASLAWSSession != "") {
		return nil, errors.New("cannot mix profile based and role based authentication")
	}

	if c.SASLAWSProfile == "" && (c.SASLAWSRole == "" || c.SASLAWSSession == "") {
		return nil, errors.New("both role and session must be set for role based authentication")
	}

	if c.SASLAWSProfile != "" {
		return &oauthAWSMSKIAM{
			generator: func(ctx context.Context) (string, error) {
				t, _, err := signer.GenerateAuthTokenFromProfile(ctx, c.SASLAWSRegion, c.SASLAWSProfile)
				return t, err
			},
			extensions: extensions,
		}, nil
	}

	// Generate using role/session
	if c.SASLAWSRole != "" && c.SASLAWSSession != "" {
		return &oauthAWSMSKIAM{
			generator: func(ctx context.Context) (string, error) {
				t, _, err := signer.GenerateAuthTokenFromRole(ctx, c.SASLAWSRegion, c.SASLAWSRole, c.SASLAWSSession)
				return t, err
			},
			extensions: extensions,
		}, nil
	}

	return &oauthAWSMSKIAM{
		generator: func(ctx context.Context) (string, error) {
			t, _, err := signer.GenerateAuthToken(ctx, c.SASLAWSRegion)
			return t, err
		},
		extensions: extensions,
	}, nil
}

type oauthAWSMSKIAM struct {
	generator  func(context.Context) (string, error)
	extensions map[string]string
}

// Token generates a token using the provided AWS MSK IAM generator function.
func (a *oauthAWSMSKIAM) Token() (*sarama.AccessToken, error) {
	token, err := a.generator(context.Background())
	if err != nil {
		return nil, fmt.Errorf("getting AWS MSK IAM token failed: %w", err)
	}
	return &sarama.AccessToken{
		Token:      token,
		Extensions: a.extensions,
	}, nil
}
