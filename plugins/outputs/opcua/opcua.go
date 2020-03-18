package opcua

import (
	"fmt"
	"log"
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
	Endpoint                   string         `toml:"endpoint"`                       //defaults to "opc.tcp://localhost:50000"
	NodeID                     string         `toml:"node_id"`                        //required
	Policy                     string         `toml:"policy"`                         //defaults to "Auto"
	Mode                       string         `toml:"mode"`                           //defaults to "Auto"
	Username                   string         `toml:"username"`                       //defaults to nil
	Password                   string         `toml:"password"`                       //defaults to nil
	CertFile                   string         `toml:"cert_file"`                      //defaults to "None"
	KeyFile                    string         `toml:"key_file"`                       //defaults to "None"
	AuthMethod                 string         `toml:"auth_method"`                    //defaults to "Anonymous" - accepts Anonymous, Username, Certificate
	Debug                      bool           `toml:"debug"`                          //defaults to false
	CreateSelfSignedCert       bool           `toml:"self_signed_cert"`               //defaults to false
	SelfSignedCertExpiresAfter time.Duration  `toml:"self_signed_cert_expires_after"` //defaults to 1 year
	selfSignedCertNextExpires  time.Time      //internally created
	opts                       []opcua.Option //internally created
	cleanupCerts               bool           //internally created
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

	var err error

	// Set Policy
	if len(o.Policy) > 0 {
		o.opts = append(o.opts, opcua.SecurityPolicy(o.Policy))
	}

	// Set Mode
	if len(o.Mode) > 0 {
		o.opts = append(o.opts, opcua.SecurityModeString(o.Mode))
	}

	// Set Auth
	if len(o.Username) > 0 {
		if len(o.Password) > 0 {
			o.opts = append(o.opts, opcua.AuthUsername(o.Username, o.Password))
		} else {
			return fmt.Errorf("username supplied for auth without supplying a password")
		}
	} else {
		o.opts = append(o.opts, opcua.AuthAnonymous())
	}

	// Set Certs
	if o.CreateSelfSignedCert {

		// if no cert file path specified
		if len(o.CertFile) < 1 {
			tempDir, err := newTempDir()
			if err != nil {

			}
			o.CertFile = filepath.Join(tempDir, "cert.pem")
			log.Printf("creating file %s", o.CertFile)
			o.KeyFile = filepath.Join(tempDir, "key.pem")
			log.Printf("creating file %s", o.KeyFile)
			o.cleanupCerts = true
		}

		generateCert(o.Endpoint, 2048, o.CertFile, o.KeyFile, o.SelfSignedCertExpiresAfter)

	}

	return err
}

func getEndpointDescription(endpoint, policy, mode string) (ua.EndpointDescription, error) {
	endpoints, err := opcua.GetEndpoints(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	if ep == nil {
		err = fmt.Errorf("Failed to find suitable endpoint")
	}

	return *ep, err
}

func init() {
	outputs.Add("opcua", func() telegraf.Output {
		return &Opcua{
			Endpoint:                   "opc.tcp://localhost:50000",
			Policy:                     "Auto",
			Mode:                       "Auto",
			CertFile:                   "None",
			KeyFile:                    "None",
			AuthMethod:                 "Anonymous",
			Debug:                      false,
			CreateSelfSignedCert:       false,
			SelfSignedCertExpiresAfter: (365 * 24 * time.Hour),
			cleanupCerts:               false,
		}
	})
}
