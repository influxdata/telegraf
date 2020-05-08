package opcua

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// Opcua struct to configure client.
type Opcua struct {
	Client                     *opcua.Client     //internally created
	Endpoint                   string            `toml:"endpoint"`                       //defaults to "opc.tcp://localhost:50000"
	NodeIDMap                  map[string]string `toml:"node_id_map"`                    //required
	Policy                     string            `toml:"policy"`                         //defaults to "Auto"
	Mode                       string            `toml:"mode"`                           //defaults to "Auto"
	Username                   string            `toml:"username"`                       //defaults to nil
	Password                   string            `toml:"password"`                       //defaults to nil
	CertFile                   string            `toml:"cert_file"`                      //defaults to ""
	KeyFile                    string            `toml:"key_file"`                       //defaults to ""
	AuthMethod                 string            `toml:"auth_method"`                    //defaults to "Anonymous" - accepts Anonymous, Username, Certificate
	Debug                      bool              `toml:"debug"`                          //defaults to false
	CreateSelfSignedCert       bool              `toml:"self_signed_cert"`               //defaults to false
	SelfSignedCertExpiresAfter time.Duration     `toml:"self_signed_cert_expires_after"` //defaults to 1 year
	selfSignedCertNextExpires  time.Time         //internally created
	opts                       []opcua.Option    //internally created
}

var sampleConfig = `
  #########

  Sample Config Here

  #########

  #TODO: UPDATE THIS
  [[[node_id_map]]]
  height="ns=2;s=HeightData"
  weight="ns=2;s=WeightData"
  age="ns=2;s=AgeData"
`

// SampleConfig returns a sample config
func (o *Opcua) SampleConfig() string {
	return sampleConfig
}

// connect, write, description, close

// Connect Opcua client
func (o *Opcua) Connect() error {

	o.setupOptions()

	c := opcua.NewClient(o.Endpoint, o.opts...)
	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		o.Client = nil
		return err
	}

	o.Client = c

	return nil
}

func (o *Opcua) clearOptions() error {
	o.opts = []opcua.Option{}
	return nil
}

func (o *Opcua) setupOptions() error {

	// Get a list of the endpoints for our target server
	endpoints, err := opcua.GetEndpoints(o.Endpoint)
	if err != nil {
		log.Fatal(err)
	}
	o.opts = generateClientOpts(endpoints, o.CertFile, o.KeyFile, o.Policy, o.Mode, o.AuthMethod, o.Username, o.Password, o.CreateSelfSignedCert, o.SelfSignedCertExpiresAfter)

	return nil
}

// Write new value to node
func (o *Opcua) Write(metrics []telegraf.Metric) error {

	allErrs := map[string]error{}

	for _, metric := range metrics {
		for key, value := range metric.Fields() {

			_nodeID := o.NodeIDMap[key]

			if _nodeID != "" {
				err := o.updateNode(_nodeID, value)
				if err != nil {
					log.Printf("Error writing to '%s' (value '%s')\n error:%s", _nodeID, value, err)
					allErrs[key] = err
				}
			} else {
				log.Printf("No mapping found for field '%s' (value '%s')", key, &value)
			}
		}
	}

	if len(allErrs) > 0 {
		message := "Errors during write:"
		for node, e := range allErrs {
			message = fmt.Sprintf("%s,\n%s: %s", message, node, e.Error())
		}
		log.Printf("All errs:\n '%s'", message)
		return fmt.Errorf("ERROR: %s", message)
	}

	return nil
}

func (o *Opcua) updateNode(nodeID string, newValue interface{}) error {

	id, err := ua.ParseNodeID(nodeID)
	v, err := ua.NewVariant(newValue)

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			&ua.WriteValue{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        v,
				},
			},
		},
	}

	resp, err := o.Client.Write(req)
	if err != nil {
		return err
	}
	for _, code := range resp.Results {
		log.Printf("Result:\n '%s'\n", code.Error())
	}

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

func (o *Opcua) Init() error {

	if o.CreateSelfSignedCert && o.CertFile == "" && o.KeyFile == "" {
		directory, err := newTempDir()

		o.CertFile = path.Join(directory, "cert.pem")
		o.KeyFile = path.Join(directory, "key.pem")

		return err
	}

	return nil
}

func init() {
	outputs.Add("opcua", func() telegraf.Output {
		return &Opcua{
			Endpoint:                   "opc.tcp://localhost:50000",
			Policy:                     "Auto",
			Mode:                       "Auto",
			CertFile:                   "",
			KeyFile:                    "",
			AuthMethod:                 "Anonymous",
			Debug:                      false,
			CreateSelfSignedCert:       false,
			SelfSignedCertExpiresAfter: (365 * 24 * time.Hour),
		}
	})
}
