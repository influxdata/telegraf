package opcua

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestGenerateCertEmptyPaths(t *testing.T) {
	// Case 1: Both paths empty - should generate in temp directory
	certPath, keyPath, err := generateCert("urn:test:client", 2048, "", "", 24*365*3600)
	require.NoError(t, err)
	require.NotEmpty(t, certPath)
	require.NotEmpty(t, keyPath)

	// Verify files were created
	require.FileExists(t, certPath)
	require.FileExists(t, keyPath)

	// Verify they're in a temp directory (check they start with /tmp or similar)
	require.True(t, filepath.IsAbs(certPath))
	require.True(t, filepath.IsAbs(keyPath))
}

func TestGenerateCertPersistentPaths(t *testing.T) {
	// Case 2: Both paths specified but files don't exist - should generate at specified paths
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Verify files don't exist yet
	require.NoFileExists(t, certPath)
	require.NoFileExists(t, keyPath)

	// Generate certificates
	returnedCertPath, returnedKeyPath, err := generateCert("urn:test:client", 2048, certPath, keyPath, 24*365*3600)
	require.NoError(t, err)
	require.Equal(t, certPath, returnedCertPath)
	require.Equal(t, keyPath, returnedKeyPath)

	// Verify files were created at specified paths
	require.FileExists(t, certPath)
	require.FileExists(t, keyPath)

	// Verify file permissions (key should be 0600)
	// Skip permission check on Windows as Unix-style permissions don't apply
	if runtime.GOOS != "windows" {
		info, err := os.Stat(keyPath)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
}

func TestGenerateCertCreatesParentDirectory(t *testing.T) {
	// Parent directory doesn't exist - should be created by os.MkdirAll
	tmpDir := t.TempDir()
	parentDir := filepath.Join(tmpDir, "new", "nested", "dir")
	certPath := filepath.Join(parentDir, "cert.pem")
	keyPath := filepath.Join(parentDir, "key.pem")

	certPathResult, keyPathResult, err := generateCert("urn:test:client", 2048, certPath, keyPath, 24*365*3600)
	require.NoError(t, err)
	require.Equal(t, certPath, certPathResult)
	require.Equal(t, keyPath, keyPathResult)

	// Verify files were created
	require.FileExists(t, certPath)
	require.FileExists(t, keyPath)

	// Verify parent directory was created
	info, err := os.Stat(parentDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestGenerateCertMissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	_, _, err := generateCert("", 2048, certPath, keyPath, 24*365*3600)
	require.ErrorIs(t, err, ErrCertificateGeneration)
}

func TestGenerateCertDifferentKeySize(t *testing.T) {
	// Test with different RSA key size to validate rsaBits parameter
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Use 4096 bit RSA key instead of default 2048
	returnedCertPath, returnedKeyPath, err := generateCert("urn:test:client", 4096, certPath, keyPath, 24*365*3600)
	require.NoError(t, err)
	require.Equal(t, certPath, returnedCertPath)
	require.Equal(t, keyPath, returnedKeyPath)

	// Verify files were created
	require.FileExists(t, certPath)
	require.FileExists(t, keyPath)
}

func TestGenerateClientOptsPrefersUserTokenMatchingChannelPolicy(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	_, _, err := generateCert("urn:telegraf:gopcua:client", 2048, certFile, keyFile, time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name     string
		tokens   []*ua.UserTokenPolicy
		expected []string
	}{
		{
			name: "legacy policy listed first",
			tokens: []*ua.UserTokenPolicy{
				{PolicyID: "anonymous", TokenType: ua.UserTokenTypeAnonymous},
				{PolicyID: "user_basic128rsa15", TokenType: ua.UserTokenTypeUserName, SecurityPolicyURI: ua.SecurityPolicyURIPrefix + "Basic128Rsa15"},
				{PolicyID: "user_basic256", TokenType: ua.UserTokenTypeUserName, SecurityPolicyURI: ua.SecurityPolicyURIPrefix + "Basic256"},
				{PolicyID: "user_basic256sha256", TokenType: ua.UserTokenTypeUserName, SecurityPolicyURI: ua.SecurityPolicyURIPrefix + "Basic256Sha256"},
			},
			expected: []string{"user_basic256sha256", "anonymous", "user_basic128rsa15", "user_basic256"},
		},
		{
			name: "empty token policy inherits the channel policy",
			tokens: []*ua.UserTokenPolicy{
				{PolicyID: "user_basic128rsa15", TokenType: ua.UserTokenTypeUserName, SecurityPolicyURI: ua.SecurityPolicyURIPrefix + "Basic128Rsa15"},
				{PolicyID: "user_inherit", TokenType: ua.UserTokenTypeUserName},
			},
			expected: []string{"user_inherit", "user_basic128rsa15"},
		},
		{
			name: "no matching token keeps the server order",
			tokens: []*ua.UserTokenPolicy{
				{PolicyID: "user_basic128rsa15", TokenType: ua.UserTokenTypeUserName, SecurityPolicyURI: ua.SecurityPolicyURIPrefix + "Basic128Rsa15"},
				{PolicyID: "user_basic256", TokenType: ua.UserTokenTypeUserName, SecurityPolicyURI: ua.SecurityPolicyURIPrefix + "Basic256"},
			},
			expected: []string{"user_basic128rsa15", "user_basic256"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := []*ua.EndpointDescription{{
				EndpointURL:        "opc.tcp://localhost:4840",
				SecurityPolicyURI:  ua.SecurityPolicyURIPrefix + "Basic256Sha256",
				SecurityMode:       ua.MessageSecurityModeSignAndEncrypt,
				SecurityLevel:      100,
				UserIdentityTokens: tt.tokens,
			}}
			client := &OpcUAClient{
				Config: &OpcUAClientConfig{
					Endpoint:       "opc.tcp://localhost:4840",
					SecurityPolicy: "Basic256Sha256",
					SecurityMode:   "SignAndEncrypt",
					AuthMethod:     "UserName",
					Username:       config.NewSecret([]byte("user")),
					Password:       config.NewSecret([]byte("pass")),
					Certificate:    certFile,
					PrivateKey:     keyFile,
				},
				Log: &testutil.Logger{},
			}
			_, err := client.generateClientOpts(endpoints)
			require.NoError(t, err)

			actual := make([]string, 0, len(tt.tokens))
			for _, token := range endpoints[0].UserIdentityTokens {
				actual = append(actual, token.PolicyID)
			}
			require.Equal(t, tt.expected, actual)
		})
	}
}
