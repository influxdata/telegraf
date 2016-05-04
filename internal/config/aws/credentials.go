package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/awslabs/aws-sdk-go/aws/credentials"
	"github.com/kelseyhightower/confd/vendor/github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/kelseyhightower/confd/vendor/github.com/aws/aws-sdk-go/aws/session"
)

type AwsCredentials struct {
	Region    string `toml:"region"`                 // AWS Region
	AccessKey string `toml:"access_key"`             // Explicit AWS Access Key ID
	SecretKey string `toml:"secret_key"`             // Explicit AWS Secret Access Key
	RoleArn   string `toml:"role_arn"`               // Role ARN to assume
	Profile   string `toml:"profile"`                // the shared profile to use
	SharedCredentialFile  string `toml:"shared_credential_file"` // location of shared credential file
	Token     string `toml:"token"`                  // STS session token
}

func (c *AwsCredentials) Credentials() client.ConfigProvider {
	if c.RoleArn != "" {
		return c.assumeCredentials()
	} else {
		return c.rootCredentials()
	}
}

func (c *AwsCredentials) rootCredentials() client.ConfigProvider {
	config := &aws.Config{
		Region: aws.String(c.Region),
	}
	if c.AccessKey != "" || c.SecretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.Token)
	} else if c.Profile != "" || c.SharedCredentialFile != "" {
		config.Credentials = credentials.NewSharedCredentials(c.SharedCredentialFile, c.Profile)
	}

	return session.New(config)
}

func (c *AwsCredentials) assumeCredentials() client.ConfigProvider {
	rootCredentials := c.rootCredentials()
	config := &aws.Config{
		Region: aws.String(c.Region),
	}
	config.Credentials = stscreds.NewCredentials(rootCredentials, c.RoleArn)
	return session.New(config)
}
