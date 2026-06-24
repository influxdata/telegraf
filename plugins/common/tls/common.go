package tls

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"

	"go.step.sm/crypto/pemutil"
)

const TLSMinVersionDefault = tls.VersionTLS12

var tlsVersionMap = map[string]uint16{
	"TLS10": tls.VersionTLS10,
	"TLS11": tls.VersionTLS11,
	"TLS12": tls.VersionTLS12,
	"TLS13": tls.VersionTLS13,
}

var tlsCipherMapInit sync.Once
var tlsCipherMapSecure map[string]uint16
var tlsCipherMapInsecure map[string]uint16

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		cert, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("could not read certificate %q: %w", certFile, err)
		}
		if !pool.AppendCertsFromPEM(cert) {
			return nil, fmt.Errorf("could not parse any PEM certificates %q: %w", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile, privateKeyPassphrase string) error {
	certBytes, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("could not load certificate %q: %w", certFile, err)
	}

	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("could not load private key %q: %w", keyFile, err)
	}

	keyPEMBlock, _ := pem.Decode(keyBytes)
	if keyPEMBlock == nil {
		return errors.New("failed to decode private key: no PEM data found")
	}

	var cert tls.Certificate
	if keyPEMBlock.Type == "ENCRYPTED PRIVATE KEY" {
		if privateKeyPassphrase == "" {
			return errors.New("missing password for PKCS#8 encrypted private key")
		}
		rawDecryptedKey, err := pemutil.DecryptPKCS8PrivateKey(keyPEMBlock.Bytes, []byte(privateKeyPassphrase))
		if err != nil {
			return fmt.Errorf("failed to decrypt PKCS#8 private key: %w", err)
		}
		decryptedKey, err := x509.ParsePKCS8PrivateKey(rawDecryptedKey)
		if err != nil {
			return fmt.Errorf("failed to parse decrypted PKCS#8 private key: %w", err)
		}
		privateKey, ok := decryptedKey.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("decrypted key is not a RSA private key: %T", decryptedKey)
		}
		cert, err = tls.X509KeyPair(certBytes, pem.EncodeToMemory(&pem.Block{Type: keyPEMBlock.Type, Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))
		if err != nil {
			return fmt.Errorf("failed to load cert/key pair: %w", err)
		}
	} else if keyPEMBlock.Headers["Proc-Type"] == "4,ENCRYPTED" {
		// The key is an encrypted private key with the DEK-Info header.
		// This is currently unsupported because of the deprecation of x509.IsEncryptedPEMBlock and x509.DecryptPEMBlock.
		return errors.New("password-protected keys in pkcs#1 format are not supported")
	} else {
		cert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return fmt.Errorf("failed to load cert/key pair: %w", err)
		}
	}
	config.Certificates = []tls.Certificate{cert}
	return nil
}

func init() {
	tlsCipherMapInit.Do(func() {
		// Initialize the secure suites
		suites := tls.CipherSuites()
		tlsCipherMapSecure = make(map[string]uint16, len(suites))
		for _, s := range suites {
			tlsCipherMapSecure[s.Name] = s.ID
		}

		suites = tls.InsecureCipherSuites()
		tlsCipherMapInsecure = make(map[string]uint16, len(suites))
		for _, s := range suites {
			tlsCipherMapInsecure[s.Name] = s.ID
		}
	})
}
