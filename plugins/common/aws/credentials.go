package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type CredentialConfig struct {
	Region               string `toml:"region"`
	AccessKey            string `toml:"access_key"`
	SecretKey            string `toml:"secret_key"`
	RoleARN              string `toml:"role_arn"`
	Profile              string `toml:"profile"`
	Filename             string `toml:"shared_credential_file"`
	Token                string `toml:"token"`
	EndpointURL          string `toml:"endpoint_url"`
	RoleSessionName      string `toml:"role_session_name"`
	WebIdentityTokenFile string `toml:"web_identity_token_file"`
}

func (c *CredentialConfig) Credentials() (aws.Config, error) {
	if c.RoleARN != "" {
		return c.configWithAssumeCredentials()
	}
	return c.configWithRootCredentials()
}

func (c *CredentialConfig) configWithRootCredentials() (aws.Config, error) {
	options := []func(*config.LoadOptions) error{
		config.WithRegion(c.Region),
	}

	if c.Profile != "" {
		options = append(options, config.WithSharedConfigProfile(c.Profile))
	}
	if c.Filename != "" {
		options = append(options, config.WithSharedCredentialsFiles([]string{c.Filename}))
	}

	if c.AccessKey != "" || c.SecretKey != "" {
		provider := credentials.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, c.Token)
		options = append(options, config.WithCredentialsProvider(provider))
	}

	return config.LoadDefaultConfig(context.Background(), options...)
}

func (c *CredentialConfig) configWithAssumeCredentials() (aws.Config, error) {
	// To generate credentials using assumeRole, we need to create AWS STS client with the default AWS endpoint,
	defaultConfig, err := c.configWithRootCredentials()
	if err != nil {
		return aws.Config{}, err
	}

	var provider aws.CredentialsProvider
	stsService := sts.NewFromConfig(defaultConfig)
	if c.WebIdentityTokenFile != "" {
		provider = stscreds.NewWebIdentityRoleProvider(
			stsService,
			c.RoleARN,
			stscreds.IdentityTokenFile(c.WebIdentityTokenFile),
			func(opts *stscreds.WebIdentityRoleOptions) {
				if c.RoleSessionName != "" {
					opts.RoleSessionName = c.RoleSessionName
				}
			},
		)
	} else {
		provider = stscreds.NewAssumeRoleProvider(stsService, c.RoleARN, func(opts *stscreds.AssumeRoleOptions) {
			if c.RoleSessionName != "" {
				opts.RoleSessionName = c.RoleSessionName
			}
		})
	}

	defaultConfig.Credentials = aws.NewCredentialsCache(provider)
	return defaultConfig, nil
}
