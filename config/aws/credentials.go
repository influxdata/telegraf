package aws

import (
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

func (c *CredentialConfig) Credentials() client.ConfigProvider {
	if c.RoleARN != "" {
		return c.assumeCredentials()
	}

	return c.rootCredentials()
}

func (c *CredentialConfig) rootCredentials() client.ConfigProvider {
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

	return session.New(config)
}

func (c *CredentialConfig) assumeCredentials() client.ConfigProvider {
	rootCredentials := c.rootCredentials()
	config := &aws.Config{
		Region:   aws.String(c.Region),
		Endpoint: &c.EndpointURL,
	}

	if c.WebIdentityTokenFile != "" {
		config.Credentials = stscreds.NewWebIdentityCredentials(rootCredentials, c.RoleARN, c.RoleSessionName, c.WebIdentityTokenFile)
	} else {
		config.Credentials = stscreds.NewCredentials(rootCredentials, c.RoleARN)
	}

	return session.New(config)
}
