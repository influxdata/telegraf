package aliyun

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
)

// CredentialConfig holds Aliyun authentication configuration
type CredentialConfig struct {
	AccessKeyID       string
	AccessKeySecret   string
	AccessKeyStsToken string
	RoleArn           string
	RoleSessionName   string
	PrivateKey        string
	PublicKeyID       string
	RoleName          string
}

// GetCredentials retrieves Aliyun credentials using the credential chain
// Credentials are loaded in the following order:
// 1) Ram RoleArn credential
// 2) AccessKey STS token credential
// 3) AccessKey credential
// 4) Ecs Ram Role credential
// 5) RSA keypair credential
// 6) Environment variables credential
// 7) Instance metadata credential
func GetCredentials(config CredentialConfig) (auth.Credential, error) {
	var (
		roleSessionExpiration = 3600
		sessionExpiration     = 3600
	)

	configuration := &providers.Configuration{
		AccessKeyID:           config.AccessKeyID,
		AccessKeySecret:       config.AccessKeySecret,
		AccessKeyStsToken:     config.AccessKeyStsToken,
		RoleArn:               config.RoleArn,
		RoleSessionName:       config.RoleSessionName,
		RoleSessionExpiration: &roleSessionExpiration,
		PrivateKey:            config.PrivateKey,
		PublicKeyID:           config.PublicKeyID,
		SessionExpiration:     &sessionExpiration,
		RoleName:              config.RoleName,
	}

	credentialProviders := []providers.Provider{
		providers.NewConfigurationCredentialProvider(configuration),
		providers.NewEnvCredentialProvider(),
		providers.NewInstanceMetadataProvider(),
	}

	credential, err := providers.NewChainProvider(credentialProviders).Retrieve()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	return credential, nil
}
