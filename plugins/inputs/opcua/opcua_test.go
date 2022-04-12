package opcua

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

type OPCTags struct {
	Name           string
	Namespace      string
	IdentifierType string
	Identifier     string
	Want           interface{}
}

func TestGetDataBadNodeContainerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Spin-up the container
	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "open62541/open62541:1.0",
			ExposedPorts: []string{"4840/tcp"},
			WaitingFor:   wait.ForListeningPort("4840/tcp"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err, "starting container failed")
	defer func() {
		require.NoError(t, container.Terminate(ctx), "terminating container failed")
	}()

	// Get the connection details from the container
	addr, err := container.Host(ctx)
	require.NoError(t, err, "getting container host address failed")
	p, err := container.MappedPort(ctx, "4840/tcp")
	require.NoError(t, err, "getting container host port failed")
	port := p.Port()

	var testopctags = []OPCTags{
		{"ProductName", "1", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
	}

	var o OpcUA
	o.MetricName = "testing"
	o.Endpoint = fmt.Sprintf("opc.tcp://%s:%s", addr, port)
	fmt.Println(o.Endpoint)
	o.AuthMethod = "Anonymous"
	o.ConnectTimeout = config.Duration(10 * time.Second)
	o.RequestTimeout = config.Duration(1 * time.Second)
	o.SecurityPolicy = "None"
	o.SecurityMode = "None"
	o.codes = []ua.StatusCode{ua.StatusOK}
	logger := &testutil.CaptureLogger{}
	o.Log = logger

	g := GroupSettings{
		MetricName: "anodic_current",
		TagsSlice: [][]string{
			{"pot", "2002"},
		},
	}

	for _, tags := range testopctags {
		g.Nodes = append(g.Nodes, MapOPCTag(tags))
	}
	o.Groups = append(o.Groups, g)
	err = o.Init()
	require.NoError(t, err)
	err = Connect(&o)
	require.NoError(t, err)
	require.Contains(t, logger.LastError, "E! [] status not OK for node 'ProductName'(metric name 'anodic_current', tags 'pot=2002')")
}

func TestClient1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var testopctags = []OPCTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
		{"badnode", "1", "i", "1337", nil},
		{"goodnode", "1", "s", "the.answer", "42"},
	}

	var o OpcUA
	var err error

	o.MetricName = "testing"
	o.Endpoint = "opc.tcp://localhost:4840"
	o.AuthMethod = "Anonymous"
	o.ConnectTimeout = config.Duration(10 * time.Second)
	o.RequestTimeout = config.Duration(1 * time.Second)
	o.SecurityPolicy = "None"
	o.SecurityMode = "None"
	o.codes = []ua.StatusCode{ua.StatusOK}
	o.Log = testutil.Logger{}
	for _, tags := range testopctags {
		o.RootNodes = append(o.RootNodes, MapOPCTag(tags))
	}
	err = o.Init()
	if err != nil {
		t.Errorf("Initialize Error: %s", err)
	}
	err = Connect(&o)
	if err != nil {
		t.Fatalf("Connect Error: %s", err)
	}

	for i, v := range o.nodeData {
		if v.Value != nil {
			types := reflect.TypeOf(v.Value)
			value := reflect.ValueOf(v.Value)
			compare := fmt.Sprintf("%v", value.Interface())
			if compare != testopctags[i].Want {
				t.Errorf("Tag %s: Values %v for type %s  does not match record", o.nodes[i].tag.FieldName, value.Interface(), types)
			}
		} else if testopctags[i].Want != nil {
			t.Errorf("Tag: %s has value: %v", o.nodes[i].tag.FieldName, v.Value)
		}
	}
}

func MapOPCTag(tags OPCTags) (out NodeSettings) {
	out.FieldName = tags.Name
	out.Namespace = tags.Namespace
	out.IdentifierType = tags.IdentifierType
	out.Identifier = tags.Identifier
	return out
}

func TestConfig(t *testing.T) {
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
nodes = [
  {name="name", namespace="1", identifier_type="s", identifier="one"},
  {name="name2", namespace="2", identifier_type="s", identifier="two"},
]
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
nodes = [{name="name4", identifier="4000", tags=[["tag1", "override"]]}]

[inputs.opcua.workarounds]
additional_valid_status_codes = ["0xC0"]
`

	c := config.NewConfig()
	err := c.LoadConfigData([]byte(toml))
	require.NoError(t, err)

	require.Len(t, c.Inputs, 1)

	o, ok := c.Inputs[0].Input.(*OpcUA)
	require.True(t, ok)

	require.Len(t, o.RootNodes, 2)
	require.Equal(t, o.RootNodes[0].FieldName, "name")
	require.Equal(t, o.RootNodes[1].FieldName, "name2")

	require.Len(t, o.Groups, 2)
	require.Equal(t, o.Groups[0].MetricName, "foo")
	require.Len(t, o.Groups[0].Nodes, 1)
	require.Equal(t, o.Groups[0].Nodes[0].Identifier, "3000")

	require.NoError(t, o.InitNodes())
	require.Len(t, o.nodes, 4)
	require.Len(t, o.nodes[2].metricTags, 3)
	require.Len(t, o.nodes[3].metricTags, 2)

	require.Len(t, o.Workarounds.AdditionalValidStatusCodes, 1)
	require.Equal(t, o.Workarounds.AdditionalValidStatusCodes[0], "0xC0")
}

func TestTagsSliceToMap(t *testing.T) {
	m, err := tagsSliceToMap([][]string{{"foo", "bar"}, {"baz", "bat"}})
	require.NoError(t, err)
	require.Len(t, m, 2)
	require.Equal(t, m["foo"], "bar")
	require.Equal(t, m["baz"], "bat")
}

func TestTagsSliceToMap_twoStrings(t *testing.T) {
	var err error
	_, err = tagsSliceToMap([][]string{{"foo", "bar", "baz"}})
	require.Error(t, err)
	_, err = tagsSliceToMap([][]string{{"foo"}})
	require.Error(t, err)
}

func TestTagsSliceToMap_dupeKey(t *testing.T) {
	_, err := tagsSliceToMap([][]string{{"foo", "bar"}, {"foo", "bat"}})
	require.Error(t, err)
}

func TestTagsSliceToMap_empty(t *testing.T) {
	_, err := tagsSliceToMap([][]string{{"foo", ""}})
	require.Equal(t, fmt.Errorf("tag 1 has empty value"), err)
	_, err = tagsSliceToMap([][]string{{"", "bar"}})
	require.Equal(t, fmt.Errorf("tag 1 has empty name"), err)
}

func TestValidateOPCTags(t *testing.T) {
	tests := []struct {
		name  string
		nodes []Node
		err   error
	}{
		{
			"same",
			[]Node{
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "v1", "t2": "v2"},
				},
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "v1", "t2": "v2"},
				},
			},
			fmt.Errorf("name 'fn' is duplicated (metric name 'mn', tags 't1=v1, t2=v2')"),
		},
		{
			"different metric tag names",
			[]Node{
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "", "t2": ""},
				},
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "", "t3": ""},
				},
			},
			nil,
		},
		{
			"different metric tag values",
			[]Node{
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "foo", "t2": ""},
				},
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "bar", "t2": ""},
				},
			},
			nil,
		},
		{
			"different metric names",
			[]Node{
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "", "t2": ""},
				},
				{
					metricName: "mn2",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "", "t2": ""},
				},
			},
			nil,
		},
		{
			"different field names",
			[]Node{
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "", "t2": ""},
				},
				{
					metricName: "mn",
					tag:        NodeSettings{FieldName: "fn2", IdentifierType: "s"},
					metricTags: map[string]string{"t1": "", "t2": ""},
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OpcUA{
				nodes: tt.nodes,
				Log:   testutil.Logger{},
			}
			require.Equal(t, tt.err, o.validateOPCTags())
		})
	}
}

func TestSetupWorkarounds(t *testing.T) {
	var o OpcUA
	o.codes = []ua.StatusCode{ua.StatusOK}

	o.Workarounds.AdditionalValidStatusCodes = []string{"0xC0", "0x00AA0000"}

	err := o.setupWorkarounds()
	require.NoError(t, err)

	require.Len(t, o.codes, 3)
	require.Equal(t, o.codes[0], ua.StatusCode(0))
	require.Equal(t, o.codes[1], ua.StatusCode(192))
	require.Equal(t, o.codes[2], ua.StatusCode(11141120))
}

func TestCheckStatusCode(t *testing.T) {
	var o OpcUA
	o.codes = []ua.StatusCode{ua.StatusCode(0), ua.StatusCode(192), ua.StatusCode(11141120)}
	require.Equal(t, o.checkStatusCode(ua.StatusCode(192)), true)
}
