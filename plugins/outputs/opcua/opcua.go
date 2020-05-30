package opcua

import (
	"context"
	"fmt"
	"log"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// Opcua struct to configure client.
type Opcua struct {
	Client     *opcua.Client     //internally created
	Endpoint   string            `toml:"endpoint"`    //defaults to "opc.tcp://localhost:50000"
	NodeIDMap  map[string]string `toml:"node_id_map"` //required
	Policy     string            `toml:"policy"`      //defaults to "Auto"
	Mode       string            `toml:"mode"`        //defaults to "Auto"
	Username   string            `toml:"username"`    //defaults to nil
	Password   string            `toml:"password"`    //defaults to nil
	CertFile   string            `toml:"cert_file"`   //defaults to ""
	KeyFile    string            `toml:"key_file"`    //defaults to ""
	AuthMethod string            `toml:"auth_method"` //defaults to "Anonymous" - accepts Anonymous, Username, Certificate
	opts       []opcua.Option    //internally created
}

var sampleConfig = `
  [[outputs.opcua]]
  endpoint = "opc.tcp://localhost:49320"
  policy = "None"
  mode = "None"
  [outputs.opcua.node_id_map]
	usage_idle = "ns=2;s=MyPLC.NodeToUpdate"
	
  # node_id_map takes the form:
  #
  #[outputs.opcua.node_id_map]
  #  meric_field_name1 = "opcua id 1"
  #  meric_field_name2 = "opcua id 2"
  # 
  # The OPC UA client will iterate over the fields in a receieved metric and update the corresponding opcua node id with the metric field's value.
  #
  #
  # Full list of options:
  #
  #
  #	endpoint = "" #defaults to "opc.tcp://localhost:50000"
  # [node_id_map] #required
  #   field = "node id"
  #
  # policy = "" #defaults to "Auto"
  # mode = "" #defaults to "Auto"
  # username = "" #defaults to nil
  # password = "" #defaults to nil
  # cert_file = "" #defaults to "" - path to cert file
  # key_file = "" #defaults to "" - path to key file
  # auth_method = "" #defaults to "Anonymous" - accepts Anonymous, Username, Certificate
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
	o.opts = generateClientOpts(endpoints, o.CertFile, o.KeyFile, o.Policy, o.Mode, o.AuthMethod, o.Username, o.Password)

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
				log.Printf("No mapping found for field '%s' (value '%s')", key, fmt.Sprint(&value))
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
			{
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

// Init intializes the client
func (o *Opcua) Init() error {

	// if o.CreateSelfSignedCert && o.CertFile == "" && o.KeyFile == "" {
	// 	directory, err := newTempDir()

	// 	o.CertFile = path.Join(directory, "cert.pem")
	// 	o.KeyFile = path.Join(directory, "key.pem")

	// 	return err
	// }

	return nil
}

func init() {
	outputs.Add("opcua", func() telegraf.Output {
		return &Opcua{
			Endpoint:   "opc.tcp://localhost:50000",
			Policy:     "Auto",
			Mode:       "Auto",
			CertFile:   "",
			KeyFile:    "",
			AuthMethod: "Anonymous",
		}
	})
}
