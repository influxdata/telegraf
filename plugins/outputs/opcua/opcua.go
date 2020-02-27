package opcua

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// Opcua struct to configure client.
type Opcua struct {
	Client                     opcua.Client   //internally created
	Endpoint                   string         `toml:"endpoint"`       //defaults to "opc.tcp://localhost:50000"
	NodeID                     string         `toml:"nodeId"`         //required
	Policy                     string         `toml:"policy"`         //defaults to "Auto"
	Mode                       string         `toml:"mode"`           //defaults to "Auto"
	Username                   string         `toml:"username"`       //defaults to nil
	Password                   string         `toml:"password"`       //defaults to nil
	CertFile                   string         `toml:"certFile"`       //defaults to "None"
	KeyFile                    string         `toml:"keyFile"`        //defaults to "None"
	AuthMethod string `toml:"authMethod"` //defaults to "Anonymous" - accepts Anonymous, Username, Certificate
	Debug                      bool           `toml:"debug"`          //defaults to false
	CreateSelfSignedCert       bool           `toml:"SelfSignedCert"` //defaults to false
	SelfSignedCertExpiresAfter time.Duration  `toml:"selfSignedCert"` //defaults to 1 year
	selfSignedCertNextExpires  time.Date      //internally created
	opts                       []opcua.Option //internally created
	cleanupCerts bool //internally created
}

var sampleConfig = `
  #########

  Sample Config Here

  #########

`

// SampleConfig returns a sample config
func (o *Opcua) SampleConfig() string {
	return sampleConfig
}

// connect, write, description, close

// Connect Opcua client
func (o *Opcua) Connect() error {
	return nil
}

// Write new value to node
func (o *Opcua) Write(metrics []telegraf.Metric) error {
	return nil
}

// Description - Opcua description
func (o *Opcua) Description() string {
	return ""
}

// Close Opcua connection
func (o *Opcua) Close() error {
	return nil
}

// Close Opcua connection
func (o *Opcua) parseOptions() error {
	// opts := []opcua.Option{
	// 	opcua.SecurityPolicy(*policy),
	// 	opcua.SecurityModeString(*mode),
	// 	opcua.CertificateFile(*certFile),
	// 	opcua.PrivateKeyFile(*keyFile),
	// 	opcua.AuthAnonymous(),
	// 	opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	// }

	//ua.UserTokenTypeAnonymous - opcua.AuthAnonymous()
	//ua.UserTokenTypeUserName - opcua.AuthUsername(*o.Username, *o.Password)
	//ua.UserTokenTypeCertificate - ua.UserTokenTypeCertificate(cert)

	var err error{}

	// Set Policy
	if len(o.Policy) > 0 {
		o.opts = append(o.opts, opcua.SecurityPolicy(*o.Policy))
	}

	// Set Mode
	if len(o.Mode) > 0 {
		o.opts = append(o.opts, opcua.SecurityModeString(*o.Mode))
	}

	// Set Auth
	if len(o.Username) > 0 {
		if len(o.Password) > 0 {
			o.opts = append(o.opts, opcua.AuthUsername(*o.Username, *o.Password))
		} else {
			return fmt.Errorf("username supplied for auth without supplying a password")
		}
	} else {
		o.opts = append(o.opts, opcua.AuthAnonymous())
	}

	// Set Certs
	if o.CreateSelfSignedCert {

		// if no cert file path specified
		if !len(o.CertFile) > 0 {
			tempDir, err := newTempDir()
			o.CertFile = filepath.Join(tempDir, "cert.pem")
			log.Printf("creating file %s", o.CertFile)
			o.KeyFile = filepath.Join(tempDir, "key.pem")
			log.Printf("creating file %s", o.KeyFile)
			o.cleanupCerts = true
		}

		generate_cert(o.Endpoint, 2048, o.CertFile, o.KeyFile, o.SelfSignedCertExpiresAfter)

	}

	return err
}

func getEndpointDescription(endpoint, policy, mode string) (ua.EndpointDescription, error) {
	endpoints, err := opcua.GetEndpoints(*endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		err = fmt.Errorf("Failed to find suitable endpoint")
	}

	return ep,err
}

func init() {
	outputs.Add("opcua", func() telegraf.Output {
		return &Opcua{
			Endpoint:                   "opc.tcp://localhost:50000",
			Policy:                     "Auto",
			Mode:                       "Auto",
			CertFile:                   "None",
			KeyFile:                    "None",
			AuthMethod: "Anonymous",
			Debug:                      false,
			SelfSignedCert:             false,
			SelfSignedCertExpiresAfter: (365 * 24 * time.Hour),
			cleanupCerts: false,
		}
	})
}


// SELF SIGNED CERT FUNCTIONS

func newTempDir() (string, error){
	dir, err := ioutil.TempDir("", "ssc")
	return dir, err
}

func generate_cert(host string, rsaBits int, certFile, keyFile string, dur time.Duration) {

	if len(host) == 0 {
		log.Fatalf("Missing required host parameter")
	}
	if rsaBits == 0 {
		rsaBits = 2048
	}
	if len(certFile) == 0 {
		certFile = "./trusted/cert.pem"
	}
	if len(keyFile) == 0 {
		keyFile = "./trusted/key.pem"
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(dur)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Gopcua Test Client"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageContentCommitment | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
		if uri, err := url.Parse(h); err == nil {
			template.URIs = append(template.URIs, uri)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", certFile, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("failed to write data to %s: %s", certFile, err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("error closing %s: %s", certFile, err)
	}
	log.Printf("wrote %s\n", certFile)

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("failed to open %s for writing: %s", keyFile, err)
		return
	}
	if err := pem.Encode(keyOut, pemBlockForKey(priv)); err != nil {
		log.Fatalf("failed to write data to %s: %s", keyFile, err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("error closing %s: %s", keyFile, err)
	}
	log.Printf("wrote %s\n", keyFile)
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}