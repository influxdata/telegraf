package aliyun

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name        string
		config      CredentialConfig
		expectError bool
	}{
		{
			name: "with access key credentials",
			config: CredentialConfig{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
			},
			expectError: false,
		},
		{
			name: "with access key and STS token",
			config: CredentialConfig{
				AccessKeyID:       "test-key-id",
				AccessKeySecret:   "test-key-secret",
				AccessKeyStsToken: "test-sts-token",
			},
			expectError: false,
		},
		{
			name: "with role ARN",
			config: CredentialConfig{
				AccessKeyID:     "test-key-id",
				AccessKeySecret: "test-key-secret",
				RoleArn:         "acs:ram::123456:role/test-role",
				RoleSessionName: "test-session",
			},
			expectError: false,
		},
		{
			name: "with RSA keypair",
			config: CredentialConfig{
				PublicKeyID: "test-public-key-id",
				PrivateKey:  "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
			},
			expectError: true,
		},
		{
			name: "with ECS RAM role name",
			config: CredentialConfig{
				RoleName: "test-ecs-role",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := GetCredentials(tt.config)

			if tt.expectError {
				if err == nil {
					require.NotNil(t, cred)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, cred)
			}
		})
	}
}

func TestGetCredentialsEmpty(t *testing.T) {
	cred, err := GetCredentials(CredentialConfig{})

	if err != nil {
		require.Contains(t, err.Error(), "failed to retrieve credential")
	} else {
		require.NotNil(t, cred)
	}
}

func TestCredentialConfigFields(t *testing.T) {
	config := CredentialConfig{
		AccessKeyID:       "key-id",
		AccessKeySecret:   "key-secret",
		AccessKeyStsToken: "sts-token",
		RoleArn:           "role-arn",
		RoleSessionName:   "session-name",
		PrivateKey:        "private-key",
		PublicKeyID:       "public-key-id",
		RoleName:          "role-name",
	}

	require.Equal(t, "key-id", config.AccessKeyID)
	require.Equal(t, "key-secret", config.AccessKeySecret)
	require.Equal(t, "sts-token", config.AccessKeyStsToken)
	require.Equal(t, "role-arn", config.RoleArn)
	require.Equal(t, "session-name", config.RoleSessionName)
	require.Equal(t, "private-key", config.PrivateKey)
	require.Equal(t, "public-key-id", config.PublicKeyID)
	require.Equal(t, "role-name", config.RoleName)
}
