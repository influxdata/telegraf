package x509_cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
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

type SelfSignedCert struct {
	CertPEM []byte
	KeyPEM  []byte
	CertDER []byte
}

func normalizePathForURL(path string) string {
	// Convert Windows-style paths `C:\Users\test\file.p12` to `C:/Users/test/file.p12`
	path = filepath.ToSlash(path)

	// Ensure the correct prefix for Windows paths
	if runtime.GOOS == "windows" {
		// Remove extra leading slash if present
		if strings.HasPrefix(path, "/") {
			path = strings.TrimPrefix(path, "/")
		}
	}

	return path
}

// generateTestKeystores creates temporary JKS & PKCS#12 keystores for testing
func generateTestKeystores(t *testing.T) (pkcs12Path, jksPath string) {
	t.Helper()

	// Generate a test certificate
	selfSigned := generateSelfSignedCert(t)

	pkcs12Path = createTestPKCS12(t, selfSigned.CertPEM, selfSigned.KeyPEM)
	jksPath = createTestJKS(t, selfSigned.CertDER)

	return pkcs12Path, jksPath
}

// generateSelfSignedCert generates a dummy self-signed certificate
func generateSelfSignedCert(t *testing.T) SelfSignedCert {
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

	return SelfSignedCert{
		CertPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
		KeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)}),
		CertDER: certDER,
	}
}

// createTestPKCS12 creates a temporary PKCS#12 keystore
func createTestPKCS12(t *testing.T, certPEM, keyPEM []byte) string {
	t.Helper()

	// Decode certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	// Decode private key
	block, _ = pem.Decode(keyPEM)
	if block == nil {
		t.Fatal("failed to parse private key PEM")
	}
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

	// Normalize path for URL usage
	pkcs12Path = normalizePathForURL(pkcs12Path)

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

	// Normalize path for URL usage
	jksPath = normalizePathForURL(jksPath)

	return "jks://" + jksPath
}

func TestGatherKeystores(t *testing.T) {
	pkcs12Path, jksPath := generateTestKeystores(t)

	tests := []struct {
		name     string
		mode     os.FileMode
		content  string
		password string
		error    bool
	}{
		{name: "valid PKCS12 keystore", mode: 0640, content: pkcs12Path, password: "test-password"},
		{name: "valid JKS keystore", mode: 0640, content: jksPath, password: "test-password"},
		{name: "missing password PKCS12", mode: 0640, content: pkcs12Path, error: true},
		{name: "missing password JKS", mode: 0640, content: jksPath, error: true},
		{name: "wrong password PKCS12", mode: 0640, content: pkcs12Path, password: "wrong-password", error: true},
		{name: "wrong password JKS", mode: 0640, content: jksPath, password: "wrong-password", error: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if runtime.GOOS != "windows" {
				// To be Reviewed
				path := strings.TrimPrefix(test.content, "pkcs12://")
				path = strings.TrimPrefix(path, "jks://")
				require.NoError(t, os.Chmod(path, test.mode))
			}

			fmt.Println("DEBUG PATH:", test.content)
			sc := X509Cert{
				Sources: []string{test.content},
				Log:     testutil.Logger{},
			}

			// Set password if provided
			if test.password != "" {
				sc.Password = config.NewSecret([]byte(test.password))
			} else {
				sc.Password = config.NewSecret(nil)
			}

			require.NoError(t, sc.Init())

			acc := testutil.Accumulator{}
			err := sc.Gather(&acc)

			if (len(acc.Errors) > 0) != test.error {
				t.Errorf("%s", err)
			}
		})
	}
}
