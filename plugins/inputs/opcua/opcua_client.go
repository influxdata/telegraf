package opcua_client

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

// OpcUA type
type OpcUA struct {
	MetricName     string          `toml:"name"`
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
	RootNodes      []NodeSettings  `toml:"nodes"`
	Groups         []GroupSettings `toml:"group"`

	nodes       []Node
	nodeData    []OPCData
	nodeIDs     []*ua.NodeID
	nodeIDerror []error
	state       ConnectionState

	// status
	ReadSuccess selfstat.Stat `toml:"-"`
	ReadError   selfstat.Stat `toml:"-"`

	// internal values
	client *opcua.Client
	req    *ua.ReadRequest
	opts   []opcua.Option
}

// OPCTag type
type NodeSettings struct {
	FieldName      string     `toml:"name"`
	Namespace      string     `toml:"namespace"`
	IdentifierType string     `toml:"identifier_type"`
	Identifier     string     `toml:"identifier"`
	DataType       string     `toml:"data_type"`   // Kept for backward compatibility but was never used.
	Description    string     `toml:"description"` // Kept for backward compatibility but was never used.
	TagsSlice      [][]string `toml:"tags"`
}

type Node struct {
	tag        NodeSettings
	idStr      string
	metricName string
	metricTags map[string]string
}

type GroupSettings struct {
	MetricName     string         `toml:"name"`            // Overrides plugin's setting
	Namespace      string         `toml:"namespace"`       // Can be overridden by node setting
	IdentifierType string         `toml:"identifier_type"` // Can be overridden by node setting
	Nodes          []NodeSettings `toml:"nodes"`
	TagsSlice      [][]string     `toml:"tags"`
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
  ## Metric name
  # name = "opcua"
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
  ## name              - field name to use in the output
  ## namespace         - OPC UA namespace of the node (integer value 0 thru 3)
  ## identifier_type   - OPC UA ID type (s=string, i=numeric, g=guid, b=opaque)
  ## identifier        - OPC UA ID (tag as shown in opcua browser)
  ## Example:
  ## {name="ProductUri", namespace="0", identifier_type="i", identifier="2262"}
  # nodes = [
  #  {name="", namespace="", identifier_type="", identifier=""},
  #  {name="", namespace="", identifier_type="", identifier=""},
  #]
  #
  ## Node Group
  ## Sets defaults for OPC UA namespace and ID type so they aren't required in
  ## every node.  A group can also have a metric name that overrides the main
  ## plugin metric name.
  ##
  ## Multiple node groups are allowed
  #[[inputs.opcua.group]]
  ## Group Metric name. Overrides the top level name.  If unset, the
  ## top level name is used.
  # name =
  #
  ## Group default namespace. If a node in the group doesn't set its
  ## namespace, this is used.
  # namespace =
  #
  ## Group default identifier type. If a node in the group doesn't set its
  ## namespace, this is used.
  # identifier_type =
  #
  ## Node ID Configuration.  Array of nodes with the same settings as above.
  # nodes = [
  #  {name="", namespace="", identifier_type="", identifier=""},
  #  {name="", namespace="", identifier_type="", identifier=""},
  #]
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

	err = o.setupOptions()
	if err != nil {
		return err
	}

	tags := map[string]string{
		"endpoint": o.Endpoint,
	}
	o.ReadError = selfstat.Register("opcua", "read_error", tags)
	o.ReadSuccess = selfstat.Register("opcua", "read_success", tags)

	return nil
}

func (o *OpcUA) validateEndpoint() error {
	if o.MetricName == "" {
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
		return fmt.Errorf("invalid security type '%s' in '%s'", o.SecurityPolicy, o.MetricName)
	}
	//search security mode type
	switch o.SecurityMode {
	case "None", "Sign", "SignAndEncrypt", "auto":
		break
	default:
		return fmt.Errorf("invalid security type '%s' in '%s'", o.SecurityMode, o.MetricName)
	}
	return nil
}

func tagsSliceToMap(tags [][]string) (map[string]string, error) {
	m := make(map[string]string)
	for i, tag := range tags {
		if len(tag) != 2 {
			return nil, fmt.Errorf("tag %d needs 2 values, has %d: %v", i+1, len(tag), tag)
		}
		if tag[0] == "" {
			return nil, fmt.Errorf("tag %d has empty name", i+1)
		}
		if tag[1] == "" {
			return nil, fmt.Errorf("tag %d has empty value", i+1)
		}
		if _, ok := m[tag[0]]; ok {
			return nil, fmt.Errorf("tag %d has duplicate key: %v", i+1, tag[0])
		}
		m[tag[0]] = tag[1]
	}
	return m, nil
}

//InitNodes Method on OpcUA
func (o *OpcUA) InitNodes() error {
	for _, node := range o.RootNodes {
		o.nodes = append(o.nodes, Node{
			metricName: o.MetricName,
			tag:        node,
		})
	}

	for _, group := range o.Groups {
		if group.MetricName == "" {
			group.MetricName = o.MetricName
		}
		groupTags, err := tagsSliceToMap(group.TagsSlice)
		if err != nil {
			return err
		}
		for _, node := range group.Nodes {
			if node.Namespace == "" {
				node.Namespace = group.Namespace
			}
			if node.IdentifierType == "" {
				node.IdentifierType = group.IdentifierType
			}
			nodeTags, err := tagsSliceToMap(node.TagsSlice)
			if err != nil {
				return err
			}
			mergedTags := make(map[string]string)
			for k, v := range groupTags {
				mergedTags[k] = v
			}
			for k, v := range nodeTags {
				mergedTags[k] = v
			}
			o.nodes = append(o.nodes, Node{
				metricName: group.MetricName,
				tag:        node,
				metricTags: mergedTags,
			})
		}
	}

	err := o.validateOPCTags()
	if err != nil {
		return err
	}

	return nil
}

type metricParts struct {
	metricName string
	fieldName  string
	tags       string // sorted by tag name and in format tag1=value1, tag2=value2
}

func newMP(n *Node) metricParts {
	var keys []string
	for key := range n.metricTags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, key := range keys {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(n.metricTags[key])
	}
	x := metricParts{
		metricName: n.metricName,
		fieldName:  n.tag.FieldName,
		tags:       sb.String(),
	}
	return x
}

func (o *OpcUA) validateOPCTags() error {
	nameEncountered := map[metricParts]struct{}{}
	for _, node := range o.nodes {
		mp := newMP(&node)
		//check empty name
		if node.tag.FieldName == "" {
			return fmt.Errorf("empty name in '%s'", node.tag.FieldName)
		}
		//search name duplicate
		if _, ok := nameEncountered[mp]; ok {
			return fmt.Errorf("name '%s' is duplicated (metric name '%s', tags '%s')",
				mp.fieldName, mp.metricName, mp.tags)
		}

		//add it to the set
		nameEncountered[mp] = struct{}{}

		//search identifier type
		switch node.tag.IdentifierType {
		case "s", "i", "g", "b":
			break
		default:
			return fmt.Errorf("invalid identifier type '%s' in '%s'", node.tag.IdentifierType, node.tag.FieldName)
		}

		node.idStr = BuildNodeID(node.tag)

		//parse NodeIds and NodeIds errors
		nid, niderr := ua.ParseNodeID(node.idStr)
		// build NodeIds and Errors
		o.nodeIDs = append(o.nodeIDs, nid)
		o.nodeIDerror = append(o.nodeIDerror, niderr)
		// Grow NodeData for later input
		o.nodeData = append(o.nodeData, OPCData{})
	}
	return nil
}

// BuildNodeID build node ID from OPC tag
func BuildNodeID(tag NodeSettings) string {
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
			return fmt.Errorf("error in Client Connection: %s", err)
		}

		regResp, err := o.client.RegisterNodes(&ua.RegisterNodesRequest{
			NodesToRegister: o.nodeIDs,
		})
		if err != nil {
			return fmt.Errorf("registerNodes failed: %v", err)
		}

		o.req = &ua.ReadRequest{
			MaxAge:             2000,
			NodesToRead:        readvalues(regResp.RegisteredNodeIDs),
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		}

		err = o.getData()
		if err != nil {
			return fmt.Errorf("get Data Failed: %v", err)
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
		return err
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
		o.ReadError.Incr(1)
		return fmt.Errorf("RegisterNodes Read failed: %v", err)
	}
	o.ReadSuccess.Incr(1)
	for i, d := range resp.Results {
		if d.Status != ua.StatusOK {
			return fmt.Errorf("status not OK: %v", d.Status)
		}
		o.nodeData[i].TagName = o.nodes[i].tag.FieldName
		if d.Value != nil {
			o.nodeData[i].Value = d.Value.Value()
			o.nodeData[i].DataType = d.Value.Type()
		}
		o.nodeData[i].Quality = d.Status
		o.nodeData[i].TimeStamp = d.ServerTimestamp.String()
		o.nodeData[i].Time = d.SourceTimestamp.String()
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

	for i, n := range o.nodes {
		fields := make(map[string]interface{})
		tags := map[string]string{
			"id": n.idStr,
		}
		for k, v := range n.metricTags {
			tags[k] = v
		}

		fields[o.nodeData[i].TagName] = o.nodeData[i].Value
		fields["Quality"] = strings.TrimSpace(fmt.Sprint(o.nodeData[i].Quality))
		acc.AddFields(n.metricName, fields, tags)
	}
	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("opcua", func() telegraf.Input {
		return &OpcUA{
			MetricName:     "opcua",
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
