package input

import (
	"errors"
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/testutil"
)

func TestValidateOPCTags(t *testing.T) {
	tests := []struct {
		name   string
		config InputClientConfig
		err    error
	}{
		{
			"duplicates",
			InputClientConfig{
				MetricName: "mn",
				RootNodes: []NodeSettings{
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
					},
				},
				Groups: []NodeGroupSettings{
					{
						Nodes: []NodeSettings{
							{
								FieldName:      "fn",
								Namespace:      "2",
								IdentifierType: "s",
								Identifier:     "i1",
							},
						},
						DefaultTags: map[string]string{"t1": "v1", "t2": "v2"},
					},
				},
			},
			errors.New(`name "fn" is duplicated (metric name "mn", tags "t1=v1, t2=v2")`),
		},
		{
			"empty tag value not allowed",
			InputClientConfig{
				MetricName: "mn",
				RootNodes: []NodeSettings{
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": ""},
					},
				},
			},
			errors.New(`empty tag value for tag "t1" in "fn"`),
		},
		{
			"empty tag name not allowed",
			InputClientConfig{
				MetricName: "mn",
				RootNodes: []NodeSettings{
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"": "1"},
					},
				},
			},
			errors.New(`empty tag name in tags for "fn"`),
		},
		{
			"different metric tag names",
			InputClientConfig{
				MetricName: "mn",
				RootNodes: []NodeSettings{
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
					},
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "v1", "t3": "v2"},
					},
				},
			},
			nil,
		},
		{
			"different metric tag values",
			InputClientConfig{
				MetricName: "mn",
				RootNodes: []NodeSettings{
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "foo", "t2": "v2"},
					},
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "bar", "t2": "v2"},
					},
				},
			},
			nil,
		},
		{
			"different metric names",
			InputClientConfig{
				MetricName: "mn",
				Groups: []NodeGroupSettings{
					{
						MetricName: "mn",
						Namespace:  "2",
						Nodes: []NodeSettings{
							{
								FieldName:      "fn",
								IdentifierType: "s",
								Identifier:     "i1",
								DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
							},
						},
					},
					{
						MetricName: "mn2",
						Namespace:  "2",
						Nodes: []NodeSettings{
							{
								FieldName:      "fn",
								IdentifierType: "s",
								Identifier:     "i1",
								DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
							},
						},
					},
				},
			},
			nil,
		},
		{
			"different field names",
			InputClientConfig{
				MetricName: "mn",
				RootNodes: []NodeSettings{
					{
						FieldName:      "fn",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
					},
					{
						FieldName:      "fn2",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "i1",
						DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
					},
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OpcUAInputClient{
				Config: tt.config,
				Log:    testutil.Logger{},
			}
			require.Equal(t, tt.err, o.InitNodeMetricMapping())
		})
	}
}

func TestNewNodeMetricMappingTags(t *testing.T) {
	tests := []struct {
		name         string
		settings     NodeSettings
		groupTags    map[string]string
		expectedTags map[string]string
		err          error
	}{
		{
			name: "empty tags",
			settings: NodeSettings{
				FieldName:      "f",
				Namespace:      "2",
				IdentifierType: "s",
				Identifier:     "h",
			},
			groupTags:    map[string]string{},
			expectedTags: map[string]string{},
			err:          nil,
		},
		{
			name: "node tags only",
			settings: NodeSettings{
				FieldName:      "f",
				Namespace:      "2",
				IdentifierType: "s",
				Identifier:     "h",
				DefaultTags:    map[string]string{"t1": "v1"},
			},
			groupTags:    map[string]string{},
			expectedTags: map[string]string{"t1": "v1"},
			err:          nil,
		},
		{
			name: "group tags only",
			settings: NodeSettings{
				FieldName:      "f",
				Namespace:      "2",
				IdentifierType: "s",
				Identifier:     "h",
			},
			groupTags:    map[string]string{"t1": "v1"},
			expectedTags: map[string]string{"t1": "v1"},
			err:          nil,
		},
		{
			name: "node tag overrides group tags",
			settings: NodeSettings{
				FieldName:      "f",
				Namespace:      "2",
				IdentifierType: "s",
				Identifier:     "h",
				DefaultTags:    map[string]string{"t1": "v2"},
			},
			groupTags:    map[string]string{"t1": "v1"},
			expectedTags: map[string]string{"t1": "v2"},
			err:          nil,
		},
		{
			name: "node tag merged with group tags",
			settings: NodeSettings{
				FieldName:      "f",
				Namespace:      "2",
				IdentifierType: "s",
				Identifier:     "h",
				DefaultTags:    map[string]string{"t2": "v2"},
			},
			groupTags:    map[string]string{"t1": "v1"},
			expectedTags: map[string]string{"t1": "v1", "t2": "v2"},
			err:          nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nmm, err := NewNodeMetricMapping("testmetric", tt.settings, tt.groupTags)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.expectedTags, nmm.MetricTags)
		})
	}
}

func TestNewNodeMetricMappingIdStrInstantiated(t *testing.T) {
	nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
		FieldName:      "f",
		Namespace:      "2",
		IdentifierType: "s",
		Identifier:     "h",
	}, map[string]string{})
	require.NoError(t, err)
	require.Equal(t, "ns=2;s=h", nmm.idStr)
}

func TestValidateNodeToAdd(t *testing.T) {
	tests := []struct {
		name     string
		existing map[metricParts]struct{}
		nmm      *NodeMetricMapping
		err      error
	}{
		{
			name:     "valid",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "2",
					IdentifierType: "s",
					Identifier:     "hf",
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: nil,
		},
		{
			name:     "empty field name not allowed",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "",
					Namespace:      "2",
					IdentifierType: "s",
					Identifier:     "hf",
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: errors.New(`empty name in ""`),
		},
		{
			name:     "empty namespace not allowed",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "",
					IdentifierType: "s",
					Identifier:     "hf",
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: errors.New("node \"f\": must specify either 'namespace' or 'namespace_uri'"),
		},
		{
			name:     "empty identifier type not allowed",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "2",
					IdentifierType: "",
					Identifier:     "hf",
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: errors.New(`invalid identifier type "" in "f"`),
		},
		{
			name:     "invalid identifier type not allowed",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "2",
					IdentifierType: "j",
					Identifier:     "hf",
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: errors.New(`invalid identifier type "j" in "f"`),
		},
		{
			name: "duplicate metric not allowed",
			existing: map[metricParts]struct{}{
				{metricName: "testmetric", fieldName: "f", tags: "t1=v1, t2=v2"}: {},
			},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "2",
					IdentifierType: "s",
					Identifier:     "hf",
					DefaultTags:    map[string]string{"t1": "v1", "t2": "v2"},
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: errors.New(`name "f" is duplicated (metric name "testmetric", tags "t1=v1, t2=v2")`),
		},
		{
			name:     "identifier type mismatch",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "2",
					IdentifierType: "i",
					Identifier:     "hf",
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: errors.New(`identifier type "i" does not match the type of identifier "hf"`),
		},
	}

	for idT, idV := range map[string]string{
		"s": "hf",
		"i": "1",
		"g": "849683f0-ce92-4fa2-836f-a02cde61d75d",
		"b": "aGVsbG8gSSBhbSBhIHRlc3QgaWRlbnRpZmllcg=="} {
		tests = append(tests, struct {
			name     string
			existing map[metricParts]struct{}
			nmm      *NodeMetricMapping
			err      error
		}{
			name:     "identifier type " + idT + " allowed",
			existing: map[metricParts]struct{}{},
			nmm: func() *NodeMetricMapping {
				nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
					FieldName:      "f",
					Namespace:      "2",
					IdentifierType: idT,
					Identifier:     idV,
				}, map[string]string{})
				require.NoError(t, err)
				return nmm
			}(),
			err: nil,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNodeToAdd(tt.existing, tt.nmm)
			require.Equal(t, tt.err, err)
		})
	}
}

func TestInitNodeMetricMapping(t *testing.T) {
	tests := []struct {
		testname string
		config   InputClientConfig
		expected []NodeMetricMapping
		err      error
	}{
		{
			testname: "only root node",
			config: InputClientConfig{
				MetricName: "testmetric",
				Timestamp:  TimestampSourceTelegraf,
				RootNodes: []NodeSettings{
					{
						FieldName:      "f",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "id1",
						DefaultTags:    map[string]string{"t1": "v1"},
					},
				},
			},
			expected: []NodeMetricMapping{
				{
					Tag: NodeSettings{
						FieldName:      "f",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "id1",
						DefaultTags:    map[string]string{"t1": "v1"},
					},
					idStr:      "ns=2;s=id1",
					metricName: "testmetric",
					MetricTags: map[string]string{"t1": "v1"},
				},
			},
			err: nil,
		},
		{
			testname: "root node and group node",
			config: InputClientConfig{
				MetricName: "testmetric",
				Timestamp:  TimestampSourceTelegraf,
				RootNodes: []NodeSettings{
					{
						FieldName:      "f",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "id1",
						DefaultTags:    map[string]string{"t1": "v1"},
					},
				},
				Groups: []NodeGroupSettings{
					{
						MetricName:     "groupmetric",
						Namespace:      "3",
						IdentifierType: "s",
						Nodes: []NodeSettings{
							{
								FieldName:   "f",
								Identifier:  "id2",
								DefaultTags: map[string]string{"t2": "v2"},
							},
						},
					},
				},
			},
			expected: []NodeMetricMapping{
				{
					Tag: NodeSettings{
						FieldName:      "f",
						Namespace:      "2",
						IdentifierType: "s",
						Identifier:     "id1",
						DefaultTags:    map[string]string{"t1": "v1"},
					},
					idStr:      "ns=2;s=id1",
					metricName: "testmetric",
					MetricTags: map[string]string{"t1": "v1"},
				},
				{
					Tag: NodeSettings{
						FieldName:      "f",
						Namespace:      "3",
						IdentifierType: "s",
						Identifier:     "id2",
						DefaultTags:    map[string]string{"t2": "v2"},
					},
					idStr:      "ns=3;s=id2",
					metricName: "groupmetric",
					MetricTags: map[string]string{"t2": "v2"},
				},
			},
			err: nil,
		},
		{
			testname: "only group node",
			config: InputClientConfig{
				MetricName: "testmetric",
				Timestamp:  TimestampSourceTelegraf,
				Groups: []NodeGroupSettings{
					{
						MetricName:     "groupmetric",
						Namespace:      "3",
						IdentifierType: "s",
						Nodes: []NodeSettings{
							{
								FieldName:   "f",
								Identifier:  "id2",
								DefaultTags: map[string]string{"t2": "v2"},
							},
						},
					},
				},
			},
			expected: []NodeMetricMapping{
				{
					Tag: NodeSettings{
						FieldName:      "f",
						Namespace:      "3",
						IdentifierType: "s",
						Identifier:     "id2",
						DefaultTags:    map[string]string{"t2": "v2"},
					},
					idStr:      "ns=3;s=id2",
					metricName: "groupmetric",
					MetricTags: map[string]string{"t2": "v2"},
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			o := OpcUAInputClient{Config: tt.config}
			err := o.InitNodeMetricMapping()
			require.NoError(t, err)
			require.Equal(t, tt.expected, o.NodeMetricMapping)
		})
	}
}

func TestUpdateNodeValue(t *testing.T) {
	type testStep struct {
		nodeIdx  int
		value    interface{}
		status   ua.StatusCode
		expected interface{}
	}
	tests := []struct {
		testname string
		steps    []testStep
	}{
		{
			"value should update when code ok",
			[]testStep{
				{
					0,
					"Harmony",
					ua.StatusOK,
					"Harmony",
				},
			},
		},
		{
			"value should not update when code bad",
			[]testStep{
				{
					0,
					"Harmony",
					ua.StatusOK,
					"Harmony",
				},
				{
					0,
					"Odium",
					ua.StatusBad,
					"Harmony",
				},
				{
					0,
					"Ati",
					ua.StatusOK,
					"Ati",
				},
			},
		},
	}

	conf := &opcua.OpcUAClientConfig{
		Endpoint:       "opc.tcp://localhost:4930",
		SecurityPolicy: "None",
		SecurityMode:   "None",
		AuthMethod:     "",
		ConnectTimeout: config.Duration(2 * time.Second),
		RequestTimeout: config.Duration(2 * time.Second),
		Workarounds:    opcua.OpcUAWorkarounds{},
	}
	c, err := conf.CreateClient(testutil.Logger{})
	require.NoError(t, err)
	o := OpcUAInputClient{
		OpcUAClient: c,
		Log:         testutil.Logger{},
		NodeMetricMapping: []NodeMetricMapping{
			{
				Tag: NodeSettings{
					FieldName: "f",
				},
			},
			{
				Tag: NodeSettings{
					FieldName: "f2",
				},
			},
		},
		LastReceivedData: make([]NodeValue, 2),
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			o.LastReceivedData = make([]NodeValue, 2)
			for i, step := range tt.steps {
				v, err := ua.NewVariant(step.value)
				require.NoError(t, err)
				o.UpdateNodeValue(0, &ua.DataValue{
					Value:             v,
					Status:            step.status,
					SourceTimestamp:   time.Date(2022, 03, 17, 8, 33, 00, 00, &time.Location{}).Add(time.Duration(i) * time.Second),
					SourcePicoseconds: 0,
					ServerTimestamp:   time.Date(2022, 03, 17, 8, 33, 00, 500, &time.Location{}).Add(time.Duration(i) * time.Second),
					ServerPicoseconds: 0,
				})
				require.Equal(t, step.expected, o.LastReceivedData[0].Value)
			}
		})
	}
}

func TestMetricForNode(t *testing.T) {
	conf := &opcua.OpcUAClientConfig{
		Endpoint:       "opc.tcp://localhost:4930",
		SecurityPolicy: "None",
		SecurityMode:   "None",
		AuthMethod:     "",
		ConnectTimeout: config.Duration(2 * time.Second),
		RequestTimeout: config.Duration(2 * time.Second),
		Workarounds:    opcua.OpcUAWorkarounds{},
	}
	c, err := conf.CreateClient(testutil.Logger{})
	require.NoError(t, err)
	o := OpcUAInputClient{
		Config: InputClientConfig{
			Timestamp: TimestampSourceSource,
		},
		OpcUAClient:      c,
		Log:              testutil.Logger{},
		LastReceivedData: make([]NodeValue, 2),
	}

	tests := []struct {
		testname string
		nmm      []NodeMetricMapping
		v        interface{}
		isArray  bool
		dataType ua.TypeID
		time     time.Time
		status   ua.StatusCode
		expected telegraf.Metric
	}{
		{
			testname: "metric build correctly",
			nmm: []NodeMetricMapping{
				{
					Tag: NodeSettings{
						FieldName: "fn",
					},
					idStr:      "ns=3;s=hi",
					metricName: "testingmetric",
					MetricTags: map[string]string{"t1": "v1"},
				},
			},
			v:        16,
			isArray:  false,
			dataType: ua.TypeIDInt32,
			time:     time.Date(2022, 03, 17, 8, 55, 00, 00, &time.Location{}),
			status:   ua.StatusOK,
			expected: metric.New("testingmetric",
				map[string]string{"t1": "v1", "id": "ns=3;s=hi"},
				map[string]interface{}{"Quality": "The operation succeeded. StatusGood (0x0)", "fn": 16},
				time.Date(2022, 03, 17, 8, 55, 00, 00, &time.Location{})),
		},
		{
			testname: "array-like metric build correctly",
			nmm: []NodeMetricMapping{
				{
					Tag: NodeSettings{
						FieldName: "fn",
					},
					idStr:      "ns=3;s=hi",
					metricName: "testingmetric",
					MetricTags: map[string]string{"t1": "v1"},
				},
			},
			v:        []int32{16, 17},
			isArray:  true,
			dataType: ua.TypeIDInt32,
			time:     time.Date(2022, 03, 17, 8, 55, 00, 00, &time.Location{}),
			status:   ua.StatusOK,
			expected: metric.New("testingmetric",
				map[string]string{"t1": "v1", "id": "ns=3;s=hi"},
				map[string]interface{}{"Quality": "The operation succeeded. StatusGood (0x0)", "fn[0]": 16, "fn[1]": 17},
				time.Date(2022, 03, 17, 8, 55, 00, 00, &time.Location{})),
		},
		{
			testname: "nil does not panic",
			nmm: []NodeMetricMapping{
				{
					Tag: NodeSettings{
						FieldName: "fn",
					},
					idStr:      "ns=3;s=hi",
					metricName: "testingmetric",
					MetricTags: map[string]string{"t1": "v1"},
				},
			},
			v:        nil,
			isArray:  false,
			dataType: ua.TypeIDNull,
			time:     time.Date(2022, 03, 17, 8, 55, 00, 00, &time.Location{}),
			status:   ua.StatusOK,
			expected: metric.New("testingmetric",
				map[string]string{"t1": "v1", "id": "ns=3;s=hi"},
				map[string]interface{}{"Quality": "The operation succeeded. StatusGood (0x0)"},
				time.Date(2022, 03, 17, 8, 55, 00, 00, &time.Location{})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			o.NodeMetricMapping = tt.nmm
			o.LastReceivedData[0].SourceTime = tt.time
			o.LastReceivedData[0].Quality = tt.status
			o.LastReceivedData[0].Value = tt.v
			o.LastReceivedData[0].DataType = tt.dataType
			o.LastReceivedData[0].IsArray = tt.isArray
			actual := o.MetricForNode(0)
			require.Equal(t, tt.expected.Tags(), actual.Tags())
			require.Equal(t, tt.expected.Fields(), actual.Fields())
			require.Equal(t, tt.expected.Time(), actual.Time())
		})
	}
}

// TestNodeIDGeneration tests that NodeID() generates correct node ID strings
func TestNodeIDGeneration(t *testing.T) {
	tests := []struct {
		name     string
		node     NodeSettings
		expected string
	}{
		{
			name: "namespace index format",
			node: NodeSettings{
				Namespace:      "3",
				IdentifierType: "s",
				Identifier:     "Temperature",
			},
			expected: "ns=3;s=Temperature",
		},
		{
			name: "namespace URI format",
			node: NodeSettings{
				NamespaceURI:   "http://opcfoundation.org/UA/",
				IdentifierType: "i",
				Identifier:     "2255",
			},
			expected: "nsu=http://opcfoundation.org/UA/;i=2255",
		},
		{
			name: "namespace index with numeric identifier",
			node: NodeSettings{
				Namespace:      "0",
				IdentifierType: "i",
				Identifier:     "2256",
			},
			expected: "ns=0;i=2256",
		},
		{
			name: "namespace URI with string identifier",
			node: NodeSettings{
				NamespaceURI:   "http://example.com/MyNamespace",
				IdentifierType: "s",
				Identifier:     "MyVariable",
			},
			expected: "nsu=http://example.com/MyNamespace;s=MyVariable",
		},
		{
			name: "namespace URI with GUID identifier",
			node: NodeSettings{
				NamespaceURI:   "http://vendor.com/",
				IdentifierType: "g",
				Identifier:     "12345678-1234-1234-1234-123456789012",
			},
			expected: "nsu=http://vendor.com/;g=12345678-1234-1234-1234-123456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.node.NodeID()
			require.Equal(t, tt.expected, actual)
		})
	}
}

// TestEventNodeIDGeneration tests that EventNodeSettings.NodeID() generates correct node ID strings
func TestEventNodeIDGeneration(t *testing.T) {
	tests := []struct {
		name     string
		node     EventNodeSettings
		expected string
	}{
		{
			name: "event node with namespace index",
			node: EventNodeSettings{
				Namespace:      "1",
				IdentifierType: "i",
				Identifier:     "2041",
			},
			expected: "ns=1;i=2041",
		},
		{
			name: "event node with namespace URI",
			node: EventNodeSettings{
				NamespaceURI:   "http://opcfoundation.org/UA/",
				IdentifierType: "i",
				Identifier:     "2253",
			},
			expected: "nsu=http://opcfoundation.org/UA/;i=2253",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.node.NodeID()
			require.Equal(t, tt.expected, actual)
		})
	}
}

// TestNodeValidationBothNamespaces tests that validation fails when both namespace and namespace_uri are set
func TestNodeValidationBothNamespaces(t *testing.T) {
	existing := make(map[metricParts]struct{})
	nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
		FieldName:      "test",
		Namespace:      "3",
		NamespaceURI:   "http://opcfoundation.org/UA/",
		IdentifierType: "s",
		Identifier:     "Temperature",
	}, map[string]string{})
	require.NoError(t, err)

	err = validateNodeToAdd(existing, nmm)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot specify both 'namespace' and 'namespace_uri'")
}

// TestNodeValidationNeitherNamespace tests that validation fails when neither namespace nor namespace_uri is set
func TestNodeValidationNeitherNamespace(t *testing.T) {
	existing := make(map[metricParts]struct{})
	nmm, err := NewNodeMetricMapping("testmetric", NodeSettings{
		FieldName:      "test",
		IdentifierType: "s",
		Identifier:     "Temperature",
	}, map[string]string{})
	require.NoError(t, err)

	err = validateNodeToAdd(existing, nmm)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must specify either 'namespace' or 'namespace_uri'")
}

// TestEventNodeValidationBothNamespaces tests event node validation with both namespace types
func TestEventNodeValidationBothNamespaces(t *testing.T) {
	node := EventNodeSettings{
		Namespace:      "1",
		NamespaceURI:   "http://opcfoundation.org/UA/",
		IdentifierType: "i",
		Identifier:     "2041",
	}

	err := node.validateEventNodeSettings()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot specify both 'namespace' and 'namespace_uri'")
}

// TestEventNodeValidationNeitherNamespace tests event node validation with neither namespace type
func TestEventNodeValidationNeitherNamespace(t *testing.T) {
	node := EventNodeSettings{
		IdentifierType: "i",
		Identifier:     "2041",
	}

	err := node.validateEventNodeSettings()
	require.Error(t, err)
	require.Contains(t, err.Error(), "must specify either 'namespace' or 'namespace_uri'")
}

// TestGroupNamespaceURIInheritance tests that nodes inherit namespace_uri from groups
func TestGroupNamespaceURIInheritance(t *testing.T) {
	client := &OpcUAInputClient{
		Config: InputClientConfig{
			MetricName: "opcua",
			Groups: []NodeGroupSettings{
				{
					Namespace:      "",
					NamespaceURI:   "http://opcfoundation.org/UA/",
					IdentifierType: "i",
					Nodes: []NodeSettings{
						{
							FieldName:  "node1",
							Identifier: "2255",
							// Should inherit namespace_uri from group
						},
						{
							FieldName:    "node2",
							Identifier:   "2256",
							NamespaceURI: "http://custom.org/UA/", // Override group default
						},
					},
				},
			},
		},
	}

	err := client.InitNodeMetricMapping()
	require.NoError(t, err)
	require.Len(t, client.NodeMetricMapping, 2)

	// First node should inherit from group
	require.Equal(t, "http://opcfoundation.org/UA/", client.NodeMetricMapping[0].Tag.NamespaceURI)
	require.Equal(t, "nsu=http://opcfoundation.org/UA/;i=2255", client.NodeMetricMapping[0].Tag.NodeID())

	// Second node should use its own namespace_uri
	require.Equal(t, "http://custom.org/UA/", client.NodeMetricMapping[1].Tag.NamespaceURI)
	require.Equal(t, "nsu=http://custom.org/UA/;i=2256", client.NodeMetricMapping[1].Tag.NodeID())
}

// TestEventGroupNamespaceURIInheritance tests that event nodes inherit namespace_uri from event groups
func TestEventGroupNamespaceURIInheritance(t *testing.T) {
	eventGroup := EventGroupSettings{
		NamespaceURI:   "http://opcfoundation.org/UA/",
		IdentifierType: "i",
		NodeIDSettings: []EventNodeSettings{
			{
				Identifier: "2253",
				// Should inherit namespace_uri from group
			},
			{
				Identifier:   "2254",
				NamespaceURI: "http://custom.org/UA/", // Override group default
			},
		},
	}

	eventGroup.UpdateNodeIDSettings()

	// First node should inherit from group
	require.Equal(t, "http://opcfoundation.org/UA/", eventGroup.NodeIDSettings[0].NamespaceURI)
	require.Equal(t, "nsu=http://opcfoundation.org/UA/;i=2253", eventGroup.NodeIDSettings[0].NodeID())

	// Second node should use its own namespace_uri
	require.Equal(t, "http://custom.org/UA/", eventGroup.NodeIDSettings[1].NamespaceURI)
	require.Equal(t, "nsu=http://custom.org/UA/;i=2254", eventGroup.NodeIDSettings[1].NodeID())
}
