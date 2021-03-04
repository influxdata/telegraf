package opcua_client

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type OPCTags struct {
	Name           string
	Namespace      string
	IdentifierType string
	Identifier     string
	Want           string
}

func TestClient1Integration(t *testing.T) {
	t.Skip("Skipping due to dial tcp 195.254.227.245:4840: connect: connection refused")

	var testopctags = []OPCTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
	}

	var o OpcUA
	var err error

	o.MetricName = "testing"
	o.Endpoint = "opc.tcp://opcua.rocks:4840"
	o.AuthMethod = "Anonymous"
	o.ConnectTimeout = config.Duration(10 * time.Second)
	o.RequestTimeout = config.Duration(1 * time.Second)
	o.SecurityPolicy = "None"
	o.SecurityMode = "None"
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
		} else {
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
}

func TestTagsSliceToMap(t *testing.T) {
	m, err := tagsSliceToMap([][]string{{"foo", "bar"}, {"baz", "bat"}})
	assert.NoError(t, err)
	assert.Len(t, m, 2)
	assert.Equal(t, m["foo"], "bar")
	assert.Equal(t, m["baz"], "bat")
}

func TestTagsSliceToMap_twoStrings(t *testing.T) {
	var err error
	_, err = tagsSliceToMap([][]string{{"foo", "bar", "baz"}})
	assert.Error(t, err)
	_, err = tagsSliceToMap([][]string{{"foo"}})
	assert.Error(t, err)
}

func TestTagsSliceToMap_dupeKey(t *testing.T) {
	_, err := tagsSliceToMap([][]string{{"foo", "bar"}, {"foo", "bat"}})
	assert.Error(t, err)
}

func TestTagsSliceToMap_empty(t *testing.T) {
	_, err := tagsSliceToMap([][]string{{"foo", ""}})
	assert.Equal(t, fmt.Errorf("tag 1 has empty value"), err)
	_, err = tagsSliceToMap([][]string{{"", "bar"}})
	assert.Equal(t, fmt.Errorf("tag 1 has empty name"), err)
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
			}
			require.Equal(t, tt.err, o.validateOPCTags())
		})
	}
}
