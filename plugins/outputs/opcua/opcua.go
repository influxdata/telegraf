package opcua

import (
	"time"

	"github.com/gopcua/opcua"
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
	CertFile                   string         `toml:"certFile"`       //defaults to "None"
	KeyFile                    string         `toml:"keyFile"`        //defaults to "None"
	Debug                      bool           `toml:"debug"`          //defaults to false
	SelfSignedCert             bool           `toml:"selfSignedCert"` //defaults to false
	SelfSignedCertExpiresAfter time.Duration  `toml:"selfSignedCert"` //defaults to 1 year
	selfSignedCert             selfSignedCert //internally created
}

type selfSignedCert struct {
	Cert       []byte
	Key        []byte
	Expiration time.Date
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

func init() {
	outputs.Add("opcua", func() telegraf.Output {
		return &Opcua{
			Endpoint:                   "opc.tcp://localhost:50000",
			Policy:                     "Auto",
			Mode:                       "Auto",
			CertFile:                   "None",
			KeyFile:                    "None",
			Debug:                      false,
			SelfSignedCert:             false,
			SelfSignedCertExpiresAfter: (365 * 24 * time.Hour),
		}
	})
}
