package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type PKIPaths struct {
	ServerPem  string
	ServerCert string
	ServerKey  string
	ClientCert string
}

type pki struct {
	keyPath string
}

func NewPKI(keyPath string) *pki {
	return &pki{keyPath: keyPath}
}

func (p *pki) TLSClientConfig() *tls.ClientConfig {
	return &tls.ClientConfig{
		TLSCA:   p.CACertPath(),
		TLSCert: p.ClientCertPath(),
		TLSKey:  p.ClientKeyPath(),
	}
}

func (p *pki) TLSServerConfig() *tls.ServerConfig {
	return &tls.ServerConfig{
		TLSAllowedCACerts: []string{p.CACertPath()},
		TLSCert:           p.ServerCertPath(),
		TLSKey:            p.ServerKeyPath(),
		TLSCipherSuites:   []string{p.CipherSuite()},
		TLSMinVersion:     p.TLSMinVersion(),
		TLSMaxVersion:     p.TLSMaxVersion(),
	}
}

func (p *pki) ReadCACert() string {
	return readCertificate(p.CACertPath())
}

func (p *pki) CACertPath() string {
	return path.Join(p.keyPath, "cacert.pem")
}

func (p *pki) CipherSuite() string {
	return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
}

func (p *pki) TLSMinVersion() string {
	return "TLS11"
}

func (p *pki) TLSMaxVersion() string {
	return "TLS13"
}

func (p *pki) ReadClientCert() string {
	return readCertificate(p.ClientCertPath())
}

func (p *pki) ClientCertPath() string {
	return path.Join(p.keyPath, "clientcert.pem")
}

func (p *pki) ReadClientKey() string {
	return readCertificate(p.ClientKeyPath())
}

func (p *pki) ClientKeyPath() string {
	return path.Join(p.keyPath, "clientkey.pem")
}

func (p *pki) ClientCertAndKeyPath() string {
	return path.Join(p.keyPath, "client.pem")
}

func (p *pki) ClientEncKeyPath() string {
	return path.Join(p.keyPath, "clientenckey.pem")
}

func (p *pki) ClientPKCS8KeyPath() string {
	return path.Join(p.keyPath, "clientkey.pkcs8.pem")
}

func (p *pki) ClientEncPKCS8KeyPath() string {
	return path.Join(p.keyPath, "clientenckey.pkcs8.pem")
}

func (p *pki) ClientCertAndEncKeyPath() string {
	return path.Join(p.keyPath, "clientenc.pem")
}

func (p *pki) ReadServerCert() string {
	return readCertificate(p.ServerCertPath())
}

func (p *pki) ServerCertPath() string {
	return path.Join(p.keyPath, "servercert.pem")
}

func (p *pki) ReadServerKey() string {
	return readCertificate(p.ServerKeyPath())
}

func (p *pki) ServerKeyPath() string {
	return path.Join(p.keyPath, "serverkey.pem")
}

func (p *pki) ServerCertAndKeyPath() string {
	return path.Join(p.keyPath, "server.pem")
}

func (p *pki) AbsolutePaths() (*PKIPaths, error) {
	tlsPem, err := filepath.Abs(p.ServerCertAndKeyPath())
	if err != nil {
		return nil, err
	}
	tlsCert, err := filepath.Abs(p.ServerCertPath())
	if err != nil {
		return nil, err
	}
	tlsKey, err := filepath.Abs(p.ServerKeyPath())
	if err != nil {
		return nil, err
	}
	cert, err := filepath.Abs(p.ClientCertPath())
	if err != nil {
		return nil, err
	}

	return &PKIPaths{
		ServerPem:  tlsPem,
		ServerCert: tlsCert,
		ServerKey:  tlsKey,
		ClientCert: cert,
	}, nil
}

func readCertificate(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("opening %q: %v", filename, err))
	}
	octets, err := io.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("reading %q: %v", filename, err))
	}
	return string(octets)
}

type PKICertificates struct {
	CAPrivate []byte
	CAPublic  []byte
	Private   []byte
	Public    []byte
}

func GenerateCertificatesRSA(common string, addresses []string, notAfter time.Time) (*PKICertificates, error) {
	notBefore := time.Now()
	if notAfter.Before(notBefore) {
		notBefore = notAfter.Add(1 * time.Minute)
	}

	// Create the CA certificate
	caPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("generating private RSA key failed: %w", err)
	}

	serialCA, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, fmt.Errorf("generating CA serial number failed: %w", err)
	}

	caCert := &x509.Certificate{
		SerialNumber: serialCA,
		Subject: pkix.Name{
			Organization: []string{"Telegraf Testing Inc."},
			Country:      []string{"US"},
			CommonName:   "Root CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caPriv.PublicKey, caPriv)
	if err != nil {
		return nil, fmt.Errorf("generating CA certificate failed: %w", err)
	}

	// Create a leaf certificate
	leafPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating private key failed: %w", err)
	}

	serialLeaf, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, fmt.Errorf("generating leaf serial number failed: %w", err)
	}

	ips := make([]net.IP, 0, len(addresses))
	for _, addr := range addresses {
		ips = append(ips, net.ParseIP(addr))
	}

	leaf := &x509.Certificate{
		SerialNumber: serialLeaf,
		Subject: pkix.Name{
			Organization: []string{"Telegraf Testing Inc."},
			Country:      []string{"US"},
			CommonName:   common,
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		IsCA:        false,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: ips,
	}
	leafBytes, err := x509.CreateCertificate(rand.Reader, leaf, caCert, &leafPriv.PublicKey, caPriv)
	if err != nil {
		return nil, fmt.Errorf("generating leaf certificate failed: %w", err)
	}

	return &PKICertificates{
		CAPrivate: pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(caPriv),
		}),
		CAPublic: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caBytes}),
		Public:   pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafBytes}),
		Private: pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(leafPriv),
		}),
	}, nil
}
