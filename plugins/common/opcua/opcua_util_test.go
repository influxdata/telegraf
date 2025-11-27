package opcua

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
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
