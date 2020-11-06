package opcua_client

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// OpcUA type
type OpcUA struct {
	Name           string          `toml:"name"`
	Endpoint       string          `toml:"endpoint"`
	SecurityPolicy string          `toml:"security_policy"`
	SecurityMode   string          `toml:"security_mode"`
	Certificate    string          `toml:"certificate"`
	PrivateKey     string          `toml:"private_key"`
	Username       string          `toml:"username"`
	Password       string          `toml:"password"`
	AuthMethod     string          `toml:"auth_method"`
	ConnectTimeout config.Duration `toml:"connect_timeout"`
	RequestTimeout config.Duration `toml:"request_timeout"`
	NodeList       []OPCTag        `toml:"nodes"`

	Nodes       []string     `toml:"-"`
	NodeData    []OPCData    `toml:"-"`
	NodeIDs     []*ua.NodeID `toml:"-"`
	NodeIDerror []error      `toml:"-"`
	state       ConnectionState

	// status
	ReadSuccess  int `toml:"-"`
	ReadError    int `toml:"-"`
	NumberOfTags int `toml:"-"`

	// internal values
	client *opcua.Client
	req    *ua.ReadRequest
	opts   []opcua.Option
}

// OPCTag type
type OPCTag struct {
	Name           string `toml:"name"`
	Namespace      string `toml:"namespace"`
	IdentifierType string `toml:"identifier_type"`
	Identifier     string `toml:"identifier"`
	Description    string `toml:"description"`
}

// OPCData type
type OPCData struct {
	TagName   string
	Value     interface{}
	Quality   ua.StatusCode
	TimeStamp string
	Time      string
	DataType  ua.TypeID
}

// ConnectionState used for constants
type ConnectionState int

const (
	//Disconnected constant state 0
	Disconnected ConnectionState = iota
	//Connecting constant state 1
	Connecting
	//Connected constant state 2
	Connected
)

const description = `Retrieve data from OPCUA devices`
const sampleConfig = `
  ## Device name
  # name = "localhost"
  #
  ## OPC UA Endpoint URL
  # endpoint = "opc.tcp://localhost:4840"
  #
  ## Maximum time allowed to establish a connect to the endpoint.
  # connect_timeout = "10s"
  #
  ## Maximum time allowed for a request over the estabilished connection.
  # request_timeout = "5s"
  #
  ## Security policy, one of "None", "Basic128Rsa15", "Basic256",
  ## "Basic256Sha256", or "auto"
  # security_policy = "auto"
  #
  ## Security mode, one of "None", "Sign", "SignAndEncrypt", or "auto"
  # security_mode = "auto"
  #
  ## Path to cert.pem. Required when security mode or policy isn't "None".
  ## If cert path is not supplied, self-signed cert and key will be generated.
  # certificate = "/etc/telegraf/cert.pem"
  #
  ## Path to private key.pem. Required when security mode or policy isn't "None".
  ## If key path is not supplied, self-signed cert and key will be generated.
  # private_key = "/etc/telegraf/key.pem"
  #
  ## Authentication Method, one of "Certificate", "UserName", or "Anonymous".  To
  ## authenticate using a specific ID, select 'Certificate' or 'UserName'
  # auth_method = "Anonymous"
  #
  ## Username. Required for auth_method = "UserName"
  # username = ""
  #
  ## Password. Required for auth_method = "UserName"
  # password = ""
  #
  ## Node ID configuration
  ## name       			- the variable name
  ## namespace  			- integer value 0 thru 3
  ## identifier_type		- s=string, i=numeric, g=guid, b=opaque
  ## identifier			- tag as shown in opcua browser
  ## data_type  			- boolean, byte, short, int, uint, uint16, int16,
  ##                        uint32, int32, float, double, string, datetime, number
  ## Example:
  ## {name="ProductUri", namespace="0", identifier_type="i", identifier="2262", data_type="string", description="http://open62541.org"}
  nodes = [
    {name="", namespace="", identifier_type="", identifier="", data_type="", description=""},
    {name="", namespace="", identifier_type="", identifier="", data_type="", description=""},
  ]
`

// Description will appear directly above the plugin definition in the config file
func (o *OpcUA) Description() string {
	return description
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (o *OpcUA) SampleConfig() string {
	return sampleConfig
}

// Init will initialize all tags
func (o *OpcUA) Init() error {
	o.state = Disconnected

	err := o.validateEndpoint()
	if err != nil {
		return err
	}

	err = o.InitNodes()
	if err != nil {
		return err
	}
	o.NumberOfTags = len(o.NodeList)

	o.setupOptions()

	return nil

}

func (o *OpcUA) validateEndpoint() error {
	if o.Name == "" {
		return fmt.Errorf("device name is empty")
	}

	if o.Endpoint == "" {
		return fmt.Errorf("endpoint url is empty")
	}

	_, err := url.Parse(o.Endpoint)
	if err != nil {
		return fmt.Errorf("endpoint url is invalid")
	}

	//search security policy type
	switch o.SecurityPolicy {
	case "None", "Basic128Rsa15", "Basic256", "Basic256Sha256", "auto":
		break
	default:
		return fmt.Errorf("invalid security type '%s' in '%s'", o.SecurityPolicy, o.Name)
	}
	//search security mode type
	switch o.SecurityMode {
	case "None", "Sign", "SignAndEncrypt", "auto":
		break
	default:
		return fmt.Errorf("invalid security type '%s' in '%s'", o.SecurityMode, o.Name)
	}
	return nil
}

//InitNodes Method on OpcUA
func (o *OpcUA) InitNodes() error {
	if len(o.NodeList) == 0 {
		return nil
	}

	err := o.validateOPCTags()
	if err != nil {
		return err
	}

	return nil
}

func (o *OpcUA) validateOPCTags() error {
	nameEncountered := map[string]bool{}
	for i, item := range o.NodeList {
		//check empty name
		if item.Name == "" {
			return fmt.Errorf("empty name in '%s'", item.Name)
		}
		//search name duplicate
		if nameEncountered[item.Name] {
			return fmt.Errorf("name '%s' is duplicated in '%s'", item.Name, item.Name)
		} else {
			nameEncountered[item.Name] = true
		}
		//search identifier type
		switch item.IdentifierType {
		case "s", "i", "g", "b":
			break
		default:
			return fmt.Errorf("invalid identifier type '%s' in '%s'", item.IdentifierType, item.Name)
		}

		// build nodeid
		o.Nodes = append(o.Nodes, BuildNodeID(item))

		//parse NodeIds and NodeIds errors
		nid, niderr := ua.ParseNodeID(o.Nodes[i])
		// build NodeIds and Errors
		o.NodeIDs = append(o.NodeIDs, nid)
		o.NodeIDerror = append(o.NodeIDerror, niderr)
		// Grow NodeData for later input
		o.NodeData = append(o.NodeData, OPCData{})
	}
	return nil
}

// BuildNodeID build node ID from OPC tag
func BuildNodeID(tag OPCTag) string {
	return "ns=" + tag.Namespace + ";" + tag.IdentifierType + "=" + tag.Identifier
}

// Connect to a OPCUA device
func Connect(o *OpcUA) error {
	u, err := url.Parse(o.Endpoint)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "opc.tcp":
		o.state = Connecting

		if o.client != nil {
			o.client.CloseSession()
		}

		o.client = opcua.NewClient(o.Endpoint, o.opts...)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.ConnectTimeout))
		defer cancel()
		if err := o.client.Connect(ctx); err != nil {
			return fmt.Errorf("Error in Client Connection: %s", err)
		}

		regResp, err := o.client.RegisterNodes(&ua.RegisterNodesRequest{
			NodesToRegister: o.NodeIDs,
		})
		if err != nil {
			return fmt.Errorf("RegisterNodes failed: %v", err)
		}

		o.req = &ua.ReadRequest{
			MaxAge:             2000,
			NodesToRead:        readvalues(regResp.RegisteredNodeIDs),
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		}

		err = o.getData()
		if err != nil {
			return fmt.Errorf("Get Data Failed: %v", err)
		}

	default:
		return fmt.Errorf("unsupported scheme %q in endpoint. Expected opc.tcp", u.Scheme)
	}
	return nil
}

func (o *OpcUA) setupOptions() error {

	// Get a list of the endpoints for our target server
	endpoints, err := opcua.GetEndpoints(o.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	if o.Certificate == "" && o.PrivateKey == "" {
		if o.SecurityPolicy != "None" || o.SecurityMode != "None" {
			o.Certificate, o.PrivateKey = generateCert("urn:telegraf:gopcua:client", 2048, o.Certificate, o.PrivateKey, (365 * 24 * time.Hour))
		}
	}

	o.opts = generateClientOpts(endpoints, o.Certificate, o.PrivateKey, o.SecurityPolicy, o.SecurityMode, o.AuthMethod, o.Username, o.Password, time.Duration(o.RequestTimeout))

	return nil
}

func (o *OpcUA) getData() error {
	resp, err := o.client.Read(o.req)
	if err != nil {
		o.ReadError++
		return fmt.Errorf("RegisterNodes Read failed: %v", err)
	}
	o.ReadSuccess++
	for i, d := range resp.Results {
		if d.Status != ua.StatusOK {
			return fmt.Errorf("Status not OK: %v", d.Status)
		}
		o.NodeData[i].TagName = o.NodeList[i].Name
		if d.Value != nil {
			o.NodeData[i].Value = d.Value.Value()
			o.NodeData[i].DataType = d.Value.Type()
		}
		o.NodeData[i].Quality = d.Status
		o.NodeData[i].TimeStamp = d.ServerTimestamp.String()
		o.NodeData[i].Time = d.SourceTimestamp.String()
	}
	return nil
}

func readvalues(ids []*ua.NodeID) []*ua.ReadValueID {
	rvids := make([]*ua.ReadValueID, len(ids))
	for i, v := range ids {
		rvids[i] = &ua.ReadValueID{NodeID: v}
	}
	return rvids
}

func disconnect(o *OpcUA) error {
	u, err := url.Parse(o.Endpoint)
	if err != nil {
		return err
	}

	o.ReadError = 0
	o.ReadSuccess = 0

	switch u.Scheme {
	case "opc.tcp":
		o.state = Disconnected
		o.client.Close()
		return nil
	default:
		return fmt.Errorf("invalid controller")
	}
}

// Gather defines what data the plugin will gather.
func (o *OpcUA) Gather(acc telegraf.Accumulator) error {
	if o.state == Disconnected {
		o.state = Connecting
		err := Connect(o)
		if err != nil {
			o.state = Disconnected
			return err
		}
	}

	o.state = Connected

	err := o.getData()
	if err != nil && o.state == Connected {
		o.state = Disconnected
		disconnect(o)
		return err
	}

	for i, n := range o.NodeList {
		fields := make(map[string]interface{})
		tags := map[string]string{
			"name": n.Name,
			"id":   BuildNodeID(n),
		}

		fields[o.NodeData[i].TagName] = o.NodeData[i].Value
		fields["Quality"] = strings.TrimSpace(fmt.Sprint(o.NodeData[i].Quality))
		acc.AddFields(o.Name, fields, tags)
	}
	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("opcua", func() telegraf.Input {
		return &OpcUA{
			Name:           "localhost",
			Endpoint:       "opc.tcp://localhost:4840",
			SecurityPolicy: "auto",
			SecurityMode:   "auto",
			RequestTimeout: config.Duration(5 * time.Second),
			ConnectTimeout: config.Duration(10 * time.Second),
			Certificate:    "/etc/telegraf/cert.pem",
			PrivateKey:     "/etc/telegraf/key.pem",
			AuthMethod:     "Anonymous",
		}
	})
}
