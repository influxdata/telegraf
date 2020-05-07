package opcua_client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// OpcUA type
type OpcUA struct {
	Name           string   `toml:"name"`
	Endpoint       string   `toml:"endpoint"`
	SecurityPolicy string   `toml:"security_policy"`
	SecurityMode   string   `toml:"security_mode"`
	Certificate    string   `toml:"certificate"`
	PrivateKey     string   `toml:"private_key"`
	Interval       string   `toml:"time_interval"`
	TimeOut        int      `toml:"timeout"`
	NodeList       []OPCTag `toml:"nodes"`
	Nodes          []string
	NodeData       []OPCData
	NodeIDs        []*ua.NodeID
	NodeIDerror    []error
	state          ConnectionState

	// status
	ReadSuccess  int
	ReadError    int
	NumberOfTags int

	// internal values
	client *opcua.Client
	// nodeMonitor       *monitor.NodeMonitor
	// nodesData         *monitor.DataChangeMessage
	// nodesSubscription *monitor.Subscription

	ctx context.Context
	// cancel context.CancelFunc
}

// OPCTag type
type OPCTag struct {
	Name           string `toml:"name"`
	Namespace      string `toml:"namespace"`
	IdentifierType string `toml:"identifier_type"`
	Identifier     string `toml:"identifier"`
	DataType       string `toml:"data_type"`
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

const description = `Retrieve data from OPCUA slave devices`
const sampleConfig = `
# ## Connection Configuration
#  ##
#  ## The plugin supports connections to PLCs via OPCUA
#  ##
#  ## Device name
name = "opcua_rocks"
#
#  # OPC UA Endpoint URL
endpoint = "opc.tcp://opcua.rocks:4840"
#
#  ## Read Timeout
#  ## add an arbitrary timeout (seconds) to demonstrate how to stop a subscription
#  ## with a context.
timeout = 30
#
#  # Time Inteval, default = 100 * time.Millisecond
#  # interval = "10000000"
#
#  # Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto
security_policy = "None"
#
#  # Security mode: None, Sign, SignAndEncrypt. Default: auto
security_mode = "None"
#
#  # Path to cert.pem. Required for security mode/policy != None
#  # cert = "/etc/telegraf/cert.pem"
#
#  # Path to private key.pem. Required for security mode/policy != None
#  # key = "/etc/telegraf/key.pem"
#
#  ## Measurements
#  ## node id to subscribe to
#  ## name       			- the variable name
#  ## namespace  			- integer value 0 thru 3
#  ## identifier_type		- s=string, i=numeric, g=guid, b=opaque
#  ## identifier			- tag as shown in opcua browser
#  ## data_type  			- boolean, byte, short, int, uint, uint16, int16, uint32, int32, float, double, string, datetime, number
#  ## Template 			- {name="", namespace="", identifier_type="", identifier="", data_type="", description=""},
nodes = [
		{name="ProductName", namespace="0", identifier_type="i", identifier="2261", data_type="string", description="open62541 OPC UA Server"},
		{name="ProductUri", namespace="0", identifier_type="i", identifier="2262", data_type="string", description="http://open62541.org"},
		{name="ManufacturerName", namespace="0", identifier_type="i", identifier="2263", data_type="string", description="open62541"},
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
	return nil

}

func (o *OpcUA) validateEndpoint() error {
	//check device name
	if o.Name == "" {
		return fmt.Errorf("device name is empty")
	}
	//check device name
	if o.Endpoint == "" {
		return fmt.Errorf("device name is empty")
	}

	_, err := url.Parse(o.Endpoint)
	if err != nil {
		return fmt.Errorf("endpoint url is invalid")
	}

	if o.Interval == "" {
		o.Interval = opcua.DefaultSubscriptionInterval.String()
	}

	_, err = time.ParseDuration(o.Interval)
	if err != nil {
		return fmt.Errorf("fatal error with time interval")
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
		// search data type
		switch item.DataType {
		case "boolean", "byte", "short", "int", "uint", "uint16", "int16", "uint32", "int32", "float", "double", "string", "datetime", "number":
			break
		default:
			return fmt.Errorf("invalid data type '%s' in '%s'", item.DataType, item.Name)
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

// Connect to a OPCUA Slave device
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

		endpoints, err := opcua.GetEndpoints(o.Endpoint)
		if err != nil {
			return fmt.Errorf("Endpoint Error: %s", err)
		}

		ep := opcua.SelectEndpoint(endpoints, o.SecurityPolicy, ua.MessageSecurityModeFromString(o.SecurityMode))
		if ep == nil {
			return fmt.Errorf("Failed to find suitable endpoint")
		}
		opts := []opcua.Option{
			opcua.SecurityPolicy(o.SecurityPolicy),
			opcua.SecurityModeString(o.SecurityMode),
			opcua.CertificateFile(o.Certificate),
			opcua.PrivateKeyFile(o.PrivateKey),
			opcua.AuthAnonymous(),
			opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
		}

		o.client = opcua.NewClient(o.Endpoint, opts...)
		if err := o.client.Connect(o.ctx); err != nil {
			return fmt.Errorf("Error in Client Connection: %s", err)
		}

		regResp, err := o.client.RegisterNodes(&ua.RegisterNodesRequest{
			NodesToRegister: o.NodeIDs,
		})
		if err != nil {
			return fmt.Errorf("RegisterNodes failed: %v", err)
		}

		req := &ua.ReadRequest{
			MaxAge:             2000,
			NodesToRead:        readvalues(regResp.RegisteredNodeIDs),
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		}
		resp, err := o.client.Read(req)
		if err != nil {
			o.ReadError++
			return fmt.Errorf("RegisterNodes failed: %v", err)
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

	default:
		return fmt.Errorf("invalid opc.tcp endpoint")
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
		//o.Log.Debugf("Connecting %v", o.Servers)
		err := Connect(o)
		if err != nil {
			o.state = Disconnected
			return err
		}
	}

	o.state = Connected

	for i, n := range o.NodeList {
		fields := make(map[string]interface{})
		tags := map[string]string{
			"name": n.Name,
			"type": n.DataType,
		}

		fields[o.NodeData[i].TagName] = o.NodeData[i].Value
		acc.AddFields(o.Name, fields, tags)
	}
	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("opcua_client", func() telegraf.Input { return &OpcUA{} })
}
