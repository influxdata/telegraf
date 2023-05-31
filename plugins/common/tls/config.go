package tls

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/influxdata/telegraf/internal/choice"
	"github.com/youmark/pkcs8"
)

const TLSMinVersionDefault = tls.VersionTLS12

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TLSCA               string `toml:"tls_ca"`
	TLSCert             string `toml:"tls_cert"`
	TLSKey              string `toml:"tls_key"`
	TLSKeyPwd           string `toml:"tls_key_pwd"`
	TLSMinVersion       string `toml:"tls_min_version"`
	InsecureSkipVerify  bool   `toml:"insecure_skip_verify"`
	ServerName          string `toml:"tls_server_name"`
	RenegotiationMethod string `toml:"tls_renegotiation_method"`
	Enable              *bool  `toml:"tls_enable"`

	SSLCA   string `toml:"ssl_ca" deprecated:"1.7.0;use 'tls_ca' instead"`
	SSLCert string `toml:"ssl_cert" deprecated:"1.7.0;use 'tls_cert' instead"`
	SSLKey  string `toml:"ssl_key" deprecated:"1.7.0;use 'tls_key' instead"`
}

// ServerConfig represents the standard server TLS config.
type ServerConfig struct {
	TLSCert            string   `toml:"tls_cert"`
	TLSKey             string   `toml:"tls_key"`
	TLSKeyPwd          string   `toml:"tls_key_pwd"`
	TLSAllowedCACerts  []string `toml:"tls_allowed_cacerts"`
	TLSCipherSuites    []string `toml:"tls_cipher_suites"`
	TLSMinVersion      string   `toml:"tls_min_version"`
	TLSMaxVersion      string   `toml:"tls_max_version"`
	TLSAllowedDNSNames []string `toml:"tls_allowed_dns_names"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ClientConfig) TLSConfig() (*tls.Config, error) {
	// Check if TLS config is forcefully disabled
	if c.Enable != nil && !*c.Enable {
		return nil, nil
	}

	// Support deprecated variable names
	if c.TLSCA == "" && c.SSLCA != "" {
		c.TLSCA = c.SSLCA
	}
	if c.TLSCert == "" && c.SSLCert != "" {
		c.TLSCert = c.SSLCert
	}
	if c.TLSKey == "" && c.SSLKey != "" {
		c.TLSKey = c.SSLKey
	}

	// This check returns a nil (aka "disabled") or an empty config
	// (aka, "use the default") if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security,
	//     * an SNI server name, or
	//     * empty/never renegotiation method
	empty := c.TLSCA == "" && c.TLSKey == "" && c.TLSCert == ""
	empty = empty && !c.InsecureSkipVerify && c.ServerName == ""
	empty = empty && (c.RenegotiationMethod == "" || c.RenegotiationMethod == "never")

	if empty {
		// Check if TLS config is forcefully enabled and supposed to
		// use the system defaults.
		if c.Enable != nil && *c.Enable {
			return &tls.Config{}, nil
		}

		return nil, nil
	}

	var renegotiationMethod tls.RenegotiationSupport
	switch c.RenegotiationMethod {
	case "", "never":
		renegotiationMethod = tls.RenegotiateNever
	case "once":
		renegotiationMethod = tls.RenegotiateOnceAsClient
	case "freely":
		renegotiationMethod = tls.RenegotiateFreelyAsClient
	default:
		return nil, fmt.Errorf("unrecognized renegotation method %q, choose from: 'never', 'once', 'freely'", c.RenegotiationMethod)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		Renegotiation:      renegotiationMethod,
	}

	if c.TLSCA != "" {
		pool, err := makeCertPool([]string{c.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey, c.TLSKeyPwd)
		if err != nil {
			return nil, err
		}
	}

	// Explicitly and consistently set the minimal accepted version using the
	// defined default. We use this setting for both clients and servers
	// instead of relying on Golang's default that is different for clients
	// and servers and might change over time.
	tlsConfig.MinVersion = TLSMinVersionDefault
	if c.TLSMinVersion != "" {
		version, err := ParseTLSVersion(c.TLSMinVersion)
		if err != nil {
			return nil, fmt.Errorf("could not parse tls min version %q: %w", c.TLSMinVersion, err)
		}
		tlsConfig.MinVersion = version
	}

	if c.ServerName != "" {
		tlsConfig.ServerName = c.ServerName
	}

	return tlsConfig, nil
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ServerConfig) TLSConfig() (*tls.Config, error) {
	if c.TLSCert == "" && c.TLSKey == "" && len(c.TLSAllowedCACerts) == 0 {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if len(c.TLSAllowedCACerts) != 0 {
		pool, err := makeCertPool(c.TLSAllowedCACerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey, c.TLSKeyPwd)
		if err != nil {
			return nil, err
		}
	}

	if len(c.TLSCipherSuites) != 0 {
		cipherSuites, err := ParseCiphers(c.TLSCipherSuites)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse server cipher suites %s: %w", strings.Join(c.TLSCipherSuites, ","), err)
		}
		tlsConfig.CipherSuites = cipherSuites
	}

	if c.TLSMaxVersion != "" {
		version, err := ParseTLSVersion(c.TLSMaxVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse tls max version %q: %w", c.TLSMaxVersion, err)
		}
		tlsConfig.MaxVersion = version
	}

	// Explicitly and consistently set the minimal accepted version using the
	// defined default. We use this setting for both clients and servers
	// instead of relying on Golang's default that is different for clients
	// and servers and might change over time.
	tlsConfig.MinVersion = TLSMinVersionDefault
	if c.TLSMinVersion != "" {
		version, err := ParseTLSVersion(c.TLSMinVersion)
		if err != nil {
			return nil, fmt.Errorf("could not parse tls min version %q: %w", c.TLSMinVersion, err)
		}
		tlsConfig.MinVersion = version
	}

	if tlsConfig.MinVersion != 0 && tlsConfig.MaxVersion != 0 && tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return nil, fmt.Errorf("tls min version %q can't be greater than tls max version %q", tlsConfig.MinVersion, tlsConfig.MaxVersion)
	}

	// Since clientAuth is tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	// there must be certs to validate.
	if len(c.TLSAllowedCACerts) > 0 && len(c.TLSAllowedDNSNames) > 0 {
		tlsConfig.VerifyPeerCertificate = c.verifyPeerCertificate
	}

	return tlsConfig, nil
}

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
		var decryptedKey *rsa.PrivateKey
		decryptedKey, err = pkcs8.ParsePKCS8PrivateKeyRSA(keyPEMBlock.Bytes, []byte(privateKeyPassphrase))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted PKCS#8 private key: %w", err)
		}
		cert, err = tls.X509KeyPair(certBytes, pem.EncodeToMemory(&pem.Block{Type: keyPEMBlock.Type, Bytes: x509.MarshalPKCS1PrivateKey(decryptedKey)}))
		if err != nil {
			return fmt.Errorf("failed to load cert/key pair: %w", err)
		}
	} else if keyPEMBlock.Headers["Proc-Type"] == "4,ENCRYPTED" {
		// The key is an encrypted private key with the DEK-Info header.
		// This is currently unsupported because of the deprecation of x509.IsEncryptedPEMBlock and x509.DecryptPEMBlock.
		return fmt.Errorf("password-protected keys in pkcs#1 format are not supported")
	} else {
		cert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return fmt.Errorf("failed to load cert/key pair: %w", err)
		}
	}
	config.Certificates = []tls.Certificate{cert}
	return nil
}

func (c *ServerConfig) verifyPeerCertificate(rawCerts [][]byte, _ [][]*x509.Certificate) error {
	// The certificate chain is client + intermediate + root.
	// Let's review the client certificate.
	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("could not validate peer certificate: %w", err)
	}

	for _, name := range cert.DNSNames {
		if choice.Contains(name, c.TLSAllowedDNSNames) {
			return nil
		}
	}

	return fmt.Errorf("peer certificate not in allowed DNS Name list: %v", cert.DNSNames)
}
