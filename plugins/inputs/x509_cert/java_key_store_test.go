package x509_cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

type selfSignedCert struct {
	certPEM []byte
	keyPEM  []byte
	certDER []byte
}

// generateTestKeystores creates temporary JKS & PKCS#12 keystores for testing
func generateTestKeystores(t *testing.T) (pkcs12Path, jksPath string) {
	t.Helper()

	// Generate a test certificate
	selfSigned := generateselfSignedCert(t)

	pkcs12Path = createTestPKCS12(t, selfSigned.certPEM, selfSigned.keyPEM)
	jksPath = createTestJKS(t, selfSigned.certDER)

	return pkcs12Path, jksPath
}

// generateselfSignedCert generates a dummy self-signed certificate
func generateselfSignedCert(t *testing.T) selfSignedCert {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Certificate",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	return selfSignedCert{
		certPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
		keyPEM:  pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)}),
		certDER: certDER,
	}
}

// Helper function to set file permissions on non-Windows OS
func setFilePermissions(t *testing.T, path string, mode os.FileMode) {
	if runtime.GOOS != "windows" {
		path = strings.TrimPrefix(path, "pkcs12://")
		path = strings.TrimPrefix(path, "jks://")
		require.NoError(t, os.Chmod(path, mode))
	}
}

// createTestPKCS12 creates a temporary PKCS#12 keystore
func createTestPKCS12(t *testing.T, certPEM, keyPEM []byte) string {
	t.Helper()

	// Decode certificate
	block, _ := pem.Decode(certPEM)
	require.NotNil(t, block, "failed to parse certificate PEM")

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	// Decode private key
	block, _ = pem.Decode(keyPEM)
	require.NotNil(t, block, "failed to parse private key PEM")

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	// Encode PKCS#12 keystore
	pfxData, err := pkcs12.Modern.Encode(privKey, cert, nil, "test-password")
	require.NoError(t, err)

	// Use `t.TempDir()` to ensure cleanup
	tempDir := t.TempDir()
	pkcs12Path := filepath.Join(tempDir, "test-keystore.p12")

	err = os.WriteFile(pkcs12Path, pfxData, 0600)
	require.NoError(t, err)

	pkcs12Path = filepath.ToSlash(pkcs12Path)
	if runtime.GOOS == "windows" {
		pkcs12Path = strings.TrimPrefix(pkcs12Path, "/")
	}

	return "pkcs12://" + pkcs12Path
}

// createTestJKS creates a temporary JKS keystore
func createTestJKS(t *testing.T, certDER []byte) string {
	t.Helper()

	// Use `t.TempDir()` to ensure cleanup
	tempDir := t.TempDir()
	jksPath := filepath.Join(tempDir, "test-keystore.jks")

	// Create JKS keystore and add a trusted certificate
	jks := keystore.New()
	err := jks.SetTrustedCertificateEntry("test-alias", keystore.TrustedCertificateEntry{
		Certificate: keystore.Certificate{
			Type:    "X.509",
			Content: certDER,
		},
	})
	require.NoError(t, err)

	// Write keystore to file
	output, err := os.Create(jksPath)
	require.NoError(t, err)
	defer output.Close()

	require.NoError(t, jks.Store(output, []byte("test-password")))

	jksPath = filepath.ToSlash(jksPath)
	if runtime.GOOS == "windows" {
		jksPath = strings.TrimPrefix(jksPath, "/")
	}

	return "jks://" + jksPath
}

func TestGatherKeystores_Success(t *testing.T) {
	pkcs12Path, jksPath := generateTestKeystores(t)

	tests := []struct {
		name     string
		mode     os.FileMode
		content  string
		password string
	}{
		{name: "valid PKCS12 keystore", mode: 0640, content: pkcs12Path, password: "test-password"},
		{name: "valid JKS keystore", mode: 0640, content: jksPath, password: "test-password"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setFilePermissions(t, test.content, test.mode)

			sc := X509Cert{
				Sources: []string{test.content},
				Log:     testutil.Logger{},
			}

			if test.password != "" {
				sc.Password = config.NewSecret([]byte(test.password))
			} else {
				sc.Password = config.NewSecret(nil)
			}

			require.NoError(t, sc.Init())

			acc := testutil.Accumulator{}
			err := sc.Gather(&acc)

			require.NoError(t, err)
		})
	}
}

func TestGatherKeystores_Failures(t *testing.T) {
	pkcs12Path, jksPath := generateTestKeystores(t)

	tests := []struct {
		name     string
		mode     os.FileMode
		content  string
		password string
		expected string
	}{
		{name: "missing password PKCS12", mode: 0640, content: pkcs12Path, expected: "decryption password incorrect"},
		{name: "missing password JKS", mode: 0640, content: jksPath, expected: "got invalid digest"},
		{name: "wrong password PKCS12", mode: 0640, content: pkcs12Path, password: "wrong-password", expected: "decryption password incorrect"},
		{name: "wrong password JKS", mode: 0640, content: jksPath, password: "wrong-password", expected: "got invalid digest"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setFilePermissions(t, test.content, test.mode)

			sc := X509Cert{
				Sources: []string{test.content},
				Log:     testutil.Logger{},
			}

			if test.password != "" {
				sc.Password = config.NewSecret([]byte(test.password))
			} else {
				sc.Password = config.NewSecret(nil)
			}

			require.NoError(t, sc.Init())

			acc := testutil.Accumulator{}
			err := sc.Gather(&acc)

			require.NoError(t, err)
			require.ErrorContains(t, acc.Errors[0], test.expected)
		})
	}
}
