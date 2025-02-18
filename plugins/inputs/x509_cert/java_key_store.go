package x509_cert

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"software.sslmate.com/src/go-pkcs12"
)

func (c *X509Cert) processPKCS12(certPath string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PKCS#12 file: %w", err)
	}

	// Get the password string from config.Secret
	password, err := c.Password.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}
	passwordStr := password.String()
	password.Destroy()

	_, cert, caCerts, err := pkcs12.DecodeChain(data, passwordStr)
	if err != nil {
		_, cert, caCerts, err = pkcs12.DecodeChain(data, "") // Retry without password
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode PKCS#12 keystore: %w", err)
	}

	// Ensure Root CA pool exists
	if c.tlsCfg.RootCAs == nil {
		c.tlsCfg.RootCAs = x509.NewCertPool()
	}

	// Add CA certificates to RootCAs
	for _, caCert := range caCerts {
		c.tlsCfg.RootCAs.AddCert(caCert)
	}

	return append([]*x509.Certificate{cert}, caCerts...), nil
}

func (c *X509Cert) processJKS(certPath string) ([]*x509.Certificate, error) {
	file, err := os.Open(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JKS file: %w", err)
	}
	defer file.Close()

	// Get the password string from config.Secret
	password, err := c.Password.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}
	defer password.Destroy()

	ks := keystore.New()
	if err := ks.Load(file, password.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to decode JKS: %w", err)
	}

	// Ensure Root CA pool exists
	if c.tlsCfg.RootCAs == nil {
		c.tlsCfg.RootCAs = x509.NewCertPool()
	}

	certs := make([]*x509.Certificate, 0, len(ks.Aliases()))

	for _, alias := range ks.Aliases() {
		// Check for both trusted certificates and private key entries
		if entry, err := ks.GetTrustedCertificateEntry(alias); err == nil {
			cert, err := x509.ParseCertificate(entry.Certificate.Content)
			if err == nil {
				c.tlsCfg.RootCAs.AddCert(cert)
				certs = append(certs, cert)
			}
		} else if entry, err := ks.GetPrivateKeyEntry(alias, password.Bytes()); err == nil {
			for _, certData := range entry.CertificateChain {
				cert, err := x509.ParseCertificate(certData.Content)
				if err == nil {
					c.tlsCfg.RootCAs.AddCert(cert)
					certs = append(certs, cert)
				}
			}
		}
	}

	return certs, nil
}
