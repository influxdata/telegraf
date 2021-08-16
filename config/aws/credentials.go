package aws

import (
	"context"
	awsV2 "github.com/aws/aws-sdk-go-v2/aws"
	configV2 "github.com/aws/aws-sdk-go-v2/config"
	credentialsV2 "github.com/aws/aws-sdk-go-v2/credentials"
	stscredsV2 "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
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

func (c *CredentialConfig) Credentials() (client.ConfigProvider, error) {
	if c.RoleARN != "" {
		return c.assumeCredentials()
	}

	return c.rootCredentials()
}

func (c *CredentialConfig) rootCredentials() (client.ConfigProvider, error) {
	config := &aws.Config{
		Region: aws.String(c.Region),
	}
	if c.EndpointURL != "" {
		config.Endpoint = &c.EndpointURL
	}
	if c.AccessKey != "" || c.SecretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.Token)
	} else if c.Profile != "" || c.Filename != "" {
		config.Credentials = credentials.NewSharedCredentials(c.Filename, c.Profile)
	}

	return session.NewSession(config)
}

func (c *CredentialConfig) assumeCredentials() (client.ConfigProvider, error) {
	rootCredentials, err := c.rootCredentials()
	if err != nil {
		return nil, err
	}
	config := &aws.Config{
		Region:   aws.String(c.Region),
		Endpoint: &c.EndpointURL,
	}

	if c.WebIdentityTokenFile != "" {
		config.Credentials = stscreds.NewWebIdentityCredentials(rootCredentials, c.RoleARN, c.RoleSessionName, c.WebIdentityTokenFile)
	} else {
		config.Credentials = stscreds.NewCredentials(rootCredentials, c.RoleARN)
	}

	return session.NewSession(config)
}

func (c *CredentialConfig) CredentialsV2() (awsV2.Config, error) {
	if c.RoleARN != "" {
		return c.assumeCredentialsV2()
	}
	return c.rootCredentialsV2()
}

func (c *CredentialConfig) rootCredentialsV2() (awsV2.Config, error) {
	options := []func(*configV2.LoadOptions) error{
		configV2.WithRegion(c.Region),
	}

	if c.EndpointURL != "" {
		resolver := awsV2.EndpointResolverFunc(func(service, region string) (awsV2.Endpoint, error) {
			return awsV2.Endpoint{
				URL:               c.EndpointURL,
				HostnameImmutable: true,
				Source:            awsV2.EndpointSourceCustom,
			}, nil
		})
		options = append(options, configV2.WithEndpointResolver(resolver))
	}

	if c.Profile != "" {
		options = append(options, configV2.WithSharedConfigProfile(c.Profile))
	}
	if c.Filename != "" {
		options = append(options, configV2.WithSharedCredentialsFiles([]string{c.Filename}))
	}

	if c.AccessKey != "" || c.SecretKey != "" {
		provider := credentialsV2.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, c.Token)
		options = append(options, configV2.WithCredentialsProvider(provider))
	}

	return configV2.LoadDefaultConfig(context.Background(), options...)
}

func (c *CredentialConfig) assumeCredentialsV2() (awsV2.Config, error) {
	rootCredentials, err := c.rootCredentialsV2()
	if err != nil {
		return awsV2.Config{}, err
	}

	var provider awsV2.CredentialsProvider
	stsService := sts.NewFromConfig(rootCredentials)
	if c.WebIdentityTokenFile != "" {
		provider = stscredsV2.NewWebIdentityRoleProvider(stsService, c.RoleARN, stscredsV2.IdentityTokenFile(c.WebIdentityTokenFile), func(opts *stscredsV2.WebIdentityRoleOptions) {
			if c.RoleSessionName != "" {
				opts.RoleSessionName = c.RoleSessionName
			}
		})
	} else {
		provider = stscredsV2.NewAssumeRoleProvider(stsService, c.RoleARN, func(opts *stscredsV2.AssumeRoleOptions) {
			if c.RoleSessionName != "" {
				opts.RoleSessionName = c.RoleSessionName
			}
		})
	}

	rootCredentials.Credentials = awsV2.NewCredentialsCache(provider)
	return rootCredentials, nil
}
