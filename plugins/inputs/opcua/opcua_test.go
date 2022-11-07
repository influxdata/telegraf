package opcua

import (
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

const servicePort = "4840"

type OPCTags struct {
	Name           string
	Namespace      string
	IdentifierType string
	Identifier     string
	Want           interface{}
}

func MapOPCTag(tags OPCTags) (out input.NodeSettings) {
	out.FieldName = tags.Name
	out.Namespace = tags.Namespace
	out.IdentifierType = tags.IdentifierType
	out.Identifier = tags.Identifier
	return out
}

func TestGetDataBadNodeContainerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "open62541/open62541",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("TCP network layer listening on opc.tcp://"),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	var testopctags = []OPCTags{
		{"ProductName", "1", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
	}

	readConfig := ReadClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
				Certificate:    "",
				PrivateKey:     "",
				Username:       "",
				Password:       "",
				AuthMethod:     "Anonymous",
				ConnectTimeout: config.Duration(10 * time.Second),
				RequestTimeout: config.Duration(1 * time.Second),
				Workarounds:    opcua.OpcUAWorkarounds{},
			},
			MetricName: "testing",
			RootNodes:  make([]input.NodeSettings, 0),
			Groups:     make([]input.NodeGroupSettings, 0),
		},
	}

	g := input.NodeGroupSettings{
		MetricName: "anodic_current",
		TagsSlice: [][]string{
			{"pot", "2002"},
		},
	}

	for _, tags := range testopctags {
		g.Nodes = append(g.Nodes, MapOPCTag(tags))
	}
	readConfig.Groups = append(readConfig.Groups, g)

	logger := &testutil.CaptureLogger{}
	readClient, err := readConfig.CreateReadClient(logger)
	require.NoError(t, err)
	err = readClient.Init()
	require.NoError(t, err)

	err = readClient.Connect()
	require.NoError(t, err)
}

func TestReadClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "open62541/open62541",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("TCP network layer listening on opc.tcp://"),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	var testopctags = []OPCTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
		{"badnode", "1", "i", "1337", nil},
		{"goodnode", "1", "s", "the.answer", int32(42)},
		{"DateTime", "1", "i", "51037", "0001-01-01T00:00:00Z"},
	}

	readConfig := ReadClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
				Certificate:    "",
				PrivateKey:     "",
				Username:       "",
				Password:       "",
				AuthMethod:     "Anonymous",
				ConnectTimeout: config.Duration(10 * time.Second),
				RequestTimeout: config.Duration(1 * time.Second),
				Workarounds:    opcua.OpcUAWorkarounds{},
			},
			MetricName: "testing",
			RootNodes:  make([]input.NodeSettings, 0),
			Groups:     make([]input.NodeGroupSettings, 0),
		},
	}

	for _, tags := range testopctags {
		readConfig.RootNodes = append(readConfig.RootNodes, MapOPCTag(tags))
	}

	client, err := readConfig.CreateReadClient(testutil.Logger{})
	require.NoError(t, err)

	err = client.Init()
	require.NoError(t, err, "Initialization")
	err = client.Connect()
	require.NoError(t, err, "Connect")

	for i, v := range client.LastReceivedData {
		require.Equal(t, testopctags[i].Want, v.Value)
	}
}

func TestReadClientConfig(t *testing.T) {
	toml := `
[[inputs.opcua]]
name = "localhost"
endpoint = "opc.tcp://localhost:4840"
connect_timeout = "10s"
request_timeout = "5s"
security_policy = "auto"
security_mode = "auto"
certificate = "/etc/telegraf/cert.pem"
private_key = "/etc/telegraf/key.pem"
auth_method = "Anonymous"
username = ""
password = ""

[[inputs.opcua.nodes]]
  name = "name"
  namespace = "1"
  identifier_type = "s"
  identifier="one"
  tags=[["tag0", "val0"]]

[[inputs.opcua.nodes]]
  name="name2"
  namespace="2"
  identifier_type="s"
  identifier="two"
  tags=[["tag0", "val0"], ["tag00", "val00"]]
  default_tags = {tag6 = "val6"}

[[inputs.opcua.group]]
name = "foo"
namespace = "3"
identifier_type = "i"
tags = [["tag1", "val1"], ["tag2", "val2"]]
nodes = [{name="name3", identifier="3000", tags=[["tag3", "val3"]]}]

[[inputs.opcua.group]]
name = "bar"
namespace = "0"
identifier_type = "i"
tags = [["tag1", "val1"], ["tag2", "val2"]]
[[inputs.opcua.group.nodes]]
  name = "name4"
  identifier = "4000"
  tags=[["tag4", "val4"]]
  default_tags = { tag1 = "override" }

[[inputs.opcua.group.nodes]]
  name = "name5"
  identifier = "4001"

[inputs.opcua.workarounds]
additional_valid_status_codes = ["0xC0"]

[inputs.opcua.request_workarounds]
use_unregistered_reads = true
`

	c := config.NewConfig()
	err := c.LoadConfigData([]byte(toml))
	require.NoError(t, err)

	require.Len(t, c.Inputs, 1)

	o, ok := c.Inputs[0].Input.(*OpcUA)
	require.True(t, ok)

	require.Equal(t, "localhost", o.ReadClientConfig.MetricName)
	require.Equal(t, "opc.tcp://localhost:4840", o.ReadClientConfig.Endpoint)
	require.Equal(t, config.Duration(10*time.Second), o.ReadClientConfig.ConnectTimeout)
	require.Equal(t, config.Duration(5*time.Second), o.ReadClientConfig.RequestTimeout)
	require.Equal(t, "auto", o.ReadClientConfig.SecurityPolicy)
	require.Equal(t, "auto", o.ReadClientConfig.SecurityMode)
	require.Equal(t, "/etc/telegraf/cert.pem", o.ReadClientConfig.Certificate)
	require.Equal(t, "/etc/telegraf/key.pem", o.ReadClientConfig.PrivateKey)
	require.Equal(t, "Anonymous", o.ReadClientConfig.AuthMethod)
	require.Equal(t, "", o.ReadClientConfig.Username)
	require.Equal(t, "", o.ReadClientConfig.Password)
	require.Equal(t, []input.NodeSettings{
		{
			FieldName:      "name",
			Namespace:      "1",
			IdentifierType: "s",
			Identifier:     "one",
			TagsSlice:      [][]string{{"tag0", "val0"}},
		},
		{
			FieldName:      "name2",
			Namespace:      "2",
			IdentifierType: "s",
			Identifier:     "two",
			TagsSlice:      [][]string{{"tag0", "val0"}, {"tag00", "val00"}},
			DefaultTags:    map[string]string{"tag6": "val6"},
		},
	}, o.ReadClientConfig.RootNodes)
	require.Equal(t, []input.NodeGroupSettings{
		{
			MetricName:     "foo",
			Namespace:      "3",
			IdentifierType: "i",
			TagsSlice:      [][]string{{"tag1", "val1"}, {"tag2", "val2"}},
			Nodes: []input.NodeSettings{{
				FieldName:  "name3",
				Identifier: "3000",
				TagsSlice:  [][]string{{"tag3", "val3"}},
			}},
		},
		{
			MetricName:     "bar",
			Namespace:      "0",
			IdentifierType: "i",
			TagsSlice:      [][]string{{"tag1", "val1"}, {"tag2", "val2"}},
			Nodes: []input.NodeSettings{{
				FieldName:   "name4",
				Identifier:  "4000",
				TagsSlice:   [][]string{{"tag4", "val4"}},
				DefaultTags: map[string]string{"tag1": "override"},
			}, {
				FieldName:  "name5",
				Identifier: "4001",
			}},
		},
	}, o.ReadClientConfig.Groups)
	require.Equal(t, opcua.OpcUAWorkarounds{AdditionalValidStatusCodes: []string{"0xC0"}}, o.ReadClientConfig.Workarounds)
	require.Equal(t, ReadClientWorkarounds{UseUnregisteredReads: true}, o.ReadClientConfig.ReadClientWorkarounds)
	err = o.Init()
	require.NoError(t, err)
	require.Len(t, o.client.NodeMetricMapping, 5, "incorrect number of nodes")
	require.EqualValues(t, o.client.NodeMetricMapping[0].MetricTags, map[string]string{"tag0": "val0"})
	require.EqualValues(t, o.client.NodeMetricMapping[1].MetricTags, map[string]string{"tag6": "val6"})
	require.EqualValues(t, o.client.NodeMetricMapping[2].MetricTags, map[string]string{"tag1": "val1", "tag2": "val2", "tag3": "val3"})
	require.EqualValues(t, o.client.NodeMetricMapping[3].MetricTags, map[string]string{"tag1": "override", "tag2": "val2"})
	require.EqualValues(t, o.client.NodeMetricMapping[4].MetricTags, map[string]string{"tag1": "val1", "tag2": "val2"})
}
