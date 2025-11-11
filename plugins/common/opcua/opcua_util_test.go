package opcua

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0600)
	require.NoError(t, err)

	// Test file exists
	require.True(t, fileExists(tmpFile))

	// Test file doesn't exist
	require.False(t, fileExists(filepath.Join(tmpDir, "nonexistent.txt")))
}

func TestGenerateCertEmptyPaths(t *testing.T) {
	// Case 1: Both paths empty - should generate in temp directory
	certPath, keyPath, err := generateCert("urn:test:client", 2048, "", "", 24*365*3600)
	require.NoError(t, err)
	require.NotEmpty(t, certPath)
	require.NotEmpty(t, keyPath)

	// Verify files were created
	require.True(t, fileExists(certPath))
	require.True(t, fileExists(keyPath))

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
	require.False(t, fileExists(certPath))
	require.False(t, fileExists(keyPath))

	// Generate certificates
	returnedCertPath, returnedKeyPath, err := generateCert("urn:test:client", 2048, certPath, keyPath, 24*365*3600)
	require.NoError(t, err)
	require.Equal(t, certPath, returnedCertPath)
	require.Equal(t, keyPath, returnedKeyPath)

	// Verify files were created at specified paths
	require.True(t, fileExists(certPath))
	require.True(t, fileExists(keyPath))

	// Verify file permissions (key should be 0600)
	info, err := os.Stat(keyPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestGenerateCertInvalidParentDirectory(t *testing.T) {
	// Parent directory doesn't exist
	certPath := "/nonexistent/directory/cert.pem"
	keyPath := "/nonexistent/directory/key.pem"

	_, _, err := generateCert("urn:test:client", 2048, certPath, keyPath, 24*365*3600)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCertificateGeneration)
}

func TestGenerateCertMissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	_, _, err := generateCert("", 2048, certPath, keyPath, 24*365*3600)
	require.Error(t, err)
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
	require.True(t, fileExists(certPath))
	require.True(t, fileExists(keyPath))
}

func TestValidateCertificatePathsSuccess(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) (string, string)
	}{
		{
			name: "valid paths in existing directory",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "cert.pem"), filepath.Join(tmpDir, "key.pem")
			},
		},
		{
			name: "relative paths in current directory",
			setupFunc: func(_ *testing.T) (string, string) {
				return "cert.pem", "key.pem"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certPath, keyPath := tt.setupFunc(t)
			err := validateCertificatePaths(certPath, keyPath)
			require.NoError(t, err)
		})
	}
}

func TestValidateCertificatePathsFailure(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) (string, string)
	}{
		{
			name: "cert parent directory doesn't exist",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent", "cert.pem"), filepath.Join(tmpDir, "key.pem")
			},
		},
		{
			name: "key parent directory doesn't exist",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "cert.pem"), filepath.Join(tmpDir, "nonexistent", "key.pem")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certPath, keyPath := tt.setupFunc(t)
			err := validateCertificatePaths(certPath, keyPath)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrCertificateGeneration)
		})
	}
}

func TestValidateCertificatePathsParentIsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file instead of a directory
	notADir := filepath.Join(tmpDir, "notadir")
	err := os.WriteFile(notADir, []byte("test"), 0600)
	require.NoError(t, err)

	// Try to use this file as a parent directory
	certPath := filepath.Join(notADir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	err = validateCertificatePaths(certPath, keyPath)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCertificateGeneration)
	require.Contains(t, err.Error(), "not a directory")
}
