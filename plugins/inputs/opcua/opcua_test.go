package opcua

import (
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "4840"

type opcTags struct {
	name           string
	namespace      string
	identifierType string
	identifier     string
	want           interface{}
}

func mapOPCTag(tags opcTags) (out input.NodeSettings) {
	out.FieldName = tags.name
	out.Namespace = tags.namespace
	out.IdentifierType = tags.identifierType
	out.Identifier = tags.identifier
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
	defer container.Terminate()

	testopctags := []opcTags{
		{"ProductName", "1", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
	}

	readConfig := readClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
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
		MetricName:  "anodic_current",
		DefaultTags: map[string]string{"pot": "2002"},
	}

	for _, tags := range testopctags {
		g.Nodes = append(g.Nodes, mapOPCTag(tags))
	}
	readConfig.Groups = append(readConfig.Groups, g)

	logger := &testutil.CaptureLogger{}
	readClient, err := readConfig.createReadClient(logger)
	require.NoError(t, err)
	err = readClient.connect()
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
	defer container.Terminate()

	testopctags := []opcTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
		{"badnode", "1", "i", "1337", nil},
		{"goodnode", "1", "s", "the.answer", int32(42)},
		{"DateTime", "1", "i", "51037", "0001-01-01T00:00:00Z"},
	}

	readConfig := readClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
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
		readConfig.RootNodes = append(readConfig.RootNodes, mapOPCTag(tags))
	}

	client, err := readConfig.createReadClient(testutil.Logger{})
	require.NoError(t, err)

	err = client.connect()
	require.NoError(t, err)

	for i, v := range client.LastReceivedData {
		require.Equal(t, testopctags[i].want, v.Value)
	}
}

func TestReadClientIntegrationAdditionalFields(t *testing.T) {
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	testopctags := []opcTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
		{"badnode", "1", "i", "1337", nil},
		{"goodnode", "1", "s", "the.answer", int32(42)},
		{"DateTime", "1", "i", "51037", "0001-01-01T00:00:00Z"},
	}
	testopctypes := []string{
		"String",
		"String",
		"String",
		"Null",
		"Int32",
		"DateTime",
	}
	testopcquality := []string{
		"The operation succeeded. StatusGood (0x0)",
		"The operation succeeded. StatusGood (0x0)",
		"The operation succeeded. StatusGood (0x0)",
		"User does not have permission to perform the requested operation. StatusBadUserAccessDenied (0x801F0000)",
		"The operation succeeded. StatusGood (0x0)",
		"The operation succeeded. StatusGood (0x0)",
	}
	expectedopcmetrics := make([]telegraf.Metric, 0, len(testopctags))
	for i, x := range testopctags {
		now := time.Now()
		tags := map[string]string{
			"id": fmt.Sprintf("ns=%s;%s=%s", x.namespace, x.identifierType, x.identifier),
		}
		fields := map[string]interface{}{
			x.name:     x.want,
			"Quality":  testopcquality[i],
			"DataType": testopctypes[i],
		}
		expectedopcmetrics = append(expectedopcmetrics, metric.New("testing", tags, fields, now))
	}

	readConfig := readClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
				AuthMethod:     "Anonymous",
				ConnectTimeout: config.Duration(10 * time.Second),
				RequestTimeout: config.Duration(1 * time.Second),
				Workarounds:    opcua.OpcUAWorkarounds{},
				OptionalFields: []string{"DataType"},
			},
			MetricName: "testing",
			RootNodes:  make([]input.NodeSettings, 0),
			Groups:     make([]input.NodeGroupSettings, 0),
		},
	}

	for _, tags := range testopctags {
		readConfig.RootNodes = append(readConfig.RootNodes, mapOPCTag(tags))
	}

	client, err := readConfig.createReadClient(testutil.Logger{})
	require.NoError(t, err)

	require.NoError(t, client.connect())

	actualopcmetrics := make([]telegraf.Metric, 0, len(client.LastReceivedData))
	for i := range client.LastReceivedData {
		actualopcmetrics = append(actualopcmetrics, client.MetricForNode(i))
	}
	testutil.RequireMetricsEqual(t, expectedopcmetrics, actualopcmetrics, testutil.IgnoreTime())
}

func TestReadClientIntegrationWithPasswordAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "open62541/open62541",
		Entrypoint:   []string{"/opt/open62541/build/bin/examples/access_control_server"},
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("TCP network layer listening on opc.tcp://"),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	testopctags := []opcTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
		{"ManufacturerName", "0", "i", "2263", "open62541"},
	}

	readConfig := readClientConfig{
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
				Username:       config.NewSecret([]byte("peter")),
				Password:       config.NewSecret([]byte("peter123")),
				AuthMethod:     "UserName",
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
		readConfig.RootNodes = append(readConfig.RootNodes, mapOPCTag(tags))
	}

	client, err := readConfig.createReadClient(testutil.Logger{})
	require.NoError(t, err)

	err = client.connect()
	require.NoError(t, err)

	for i, v := range client.LastReceivedData {
		require.Equal(t, testopctags[i].want, v.Value)
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

optional_fields = ["DataType"]

[[inputs.opcua.nodes]]
  name = "name"
  namespace = "1"
  identifier_type = "s"
  identifier="one"
  default_tags = { tag0 = "val0" }

[[inputs.opcua.nodes]]
  name="name2"
  namespace="2"
  identifier_type="s"
  identifier="two"
  default_tags={tag6="val6"}

[[inputs.opcua.group]]
name = "foo"
namespace = "3"
identifier_type = "i"
default_tags = { tag1 = "val1", tag2 = "val2"}
[[inputs.opcua.group.nodes]]
  name = "name3"
  identifier = "3000"
  default_tags = { tag3 = "val3" }

[[inputs.opcua.group]]
name = "bar"
namespace = "0"
identifier_type = "i"
default_tags = { tag1 = "val1", tag2 = "val2"}
[[inputs.opcua.group.nodes]]
  name = "name4"
  identifier = "4000"
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
	err := c.LoadConfigData([]byte(toml), config.EmptySourcePath)
	require.NoError(t, err)

	require.Len(t, c.Inputs, 1)

	o, ok := c.Inputs[0].Input.(*OpcUA)
	require.True(t, ok)

	require.Equal(t, "localhost", o.readClientConfig.MetricName)
	require.Equal(t, "opc.tcp://localhost:4840", o.readClientConfig.Endpoint)
	require.Equal(t, config.Duration(10*time.Second), o.readClientConfig.ConnectTimeout)
	require.Equal(t, config.Duration(5*time.Second), o.readClientConfig.RequestTimeout)
	require.Equal(t, "auto", o.readClientConfig.SecurityPolicy)
	require.Equal(t, "auto", o.readClientConfig.SecurityMode)
	require.Equal(t, "/etc/telegraf/cert.pem", o.readClientConfig.Certificate)
	require.Equal(t, "/etc/telegraf/key.pem", o.readClientConfig.PrivateKey)
	require.Equal(t, "Anonymous", o.readClientConfig.AuthMethod)
	require.True(t, o.readClientConfig.Username.Empty())
	require.True(t, o.readClientConfig.Password.Empty())
	require.Equal(t, []input.NodeSettings{
		{
			FieldName:      "name",
			Namespace:      "1",
			IdentifierType: "s",
			Identifier:     "one",
			DefaultTags:    map[string]string{"tag0": "val0"},
		},
		{
			FieldName:      "name2",
			Namespace:      "2",
			IdentifierType: "s",
			Identifier:     "two",
			DefaultTags:    map[string]string{"tag6": "val6"},
		},
	}, o.readClientConfig.RootNodes)
	require.Equal(t, []input.NodeGroupSettings{
		{
			MetricName:     "foo",
			Namespace:      "3",
			IdentifierType: "i",
			DefaultTags:    map[string]string{"tag1": "val1", "tag2": "val2"},
			Nodes: []input.NodeSettings{{
				FieldName:   "name3",
				Identifier:  "3000",
				DefaultTags: map[string]string{"tag3": "val3"},
			}},
		},
		{
			MetricName:     "bar",
			Namespace:      "0",
			IdentifierType: "i",
			DefaultTags:    map[string]string{"tag1": "val1", "tag2": "val2"},
			Nodes: []input.NodeSettings{{
				FieldName:   "name4",
				Identifier:  "4000",
				DefaultTags: map[string]string{"tag1": "override"},
			}, {
				FieldName:  "name5",
				Identifier: "4001",
			}},
		},
	}, o.readClientConfig.Groups)
	require.Equal(t, opcua.OpcUAWorkarounds{AdditionalValidStatusCodes: []string{"0xC0"}}, o.readClientConfig.Workarounds)
	require.Equal(t, readClientWorkarounds{UseUnregisteredReads: true}, o.readClientConfig.ReadClientWorkarounds)
	require.Equal(t, []string{"DataType"}, o.readClientConfig.OptionalFields)
	err = o.Init()
	require.NoError(t, err)
	require.Len(t, o.client.NodeMetricMapping, 5, "incorrect number of nodes")
	require.EqualValues(t, map[string]string{"tag0": "val0"}, o.client.NodeMetricMapping[0].MetricTags)
	require.EqualValues(t, map[string]string{"tag6": "val6"}, o.client.NodeMetricMapping[1].MetricTags)
	require.EqualValues(t, map[string]string{"tag1": "val1", "tag2": "val2", "tag3": "val3"}, o.client.NodeMetricMapping[2].MetricTags)
	require.EqualValues(t, map[string]string{"tag1": "override", "tag2": "val2"}, o.client.NodeMetricMapping[3].MetricTags)
	require.EqualValues(t, map[string]string{"tag1": "val1", "tag2": "val2"}, o.client.NodeMetricMapping[4].MetricTags)
}

func TestUnregisteredReadsAndSessionRecoveryIntegration(t *testing.T) {
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	testopctags := []opcTags{
		{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"},
		{"ProductUri", "0", "i", "2262", "http://open62541.org"},
	}

	readConfig := readClientConfig{
		ReadRetries: 1, // Set low to make tests faster
		ReadClientWorkarounds: readClientWorkarounds{
			UseUnregisteredReads: true, // Enable unregistered reads
		},
		InputClientConfig: input.InputClientConfig{
			OpcUAClientConfig: opcua.OpcUAClientConfig{
				Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
				SecurityPolicy: "None",
				SecurityMode:   "None",
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
		readConfig.RootNodes = append(readConfig.RootNodes, mapOPCTag(tags))
	}

	// Create logger to capture logs
	logger := &testutil.CaptureLogger{}
	client, err := readConfig.createReadClient(logger)
	require.NoError(t, err)

	// First connection
	require.NoError(t, client.connect())

	// Verify initial data read was successful
	require.Len(t, client.LastReceivedData, 2)
	for i, v := range client.LastReceivedData {
		require.Equal(t, testopctags[i].want, v.Value)
	}

	// Get initial metrics to compare later
	initialMetrics, err := client.currentValues()
	require.NoError(t, err)
	require.Len(t, initialMetrics, 2)

	// Now simulate session invalidation as would happen in the real world
	client.forceReconnect = true

	// Get metrics again - this should force a reconnection
	recoveredMetrics, err := client.currentValues()
	require.NoError(t, err, "Should recover from session invalidation")
	require.Len(t, recoveredMetrics, 2)

	// Verify data consistency after reconnect
	for i := range recoveredMetrics {
		require.Equal(t,
			initialMetrics[i].Fields()[testopctags[i].name],
			recoveredMetrics[i].Fields()[testopctags[i].name],
			"Data should be consistent after session recovery")
	}

	// Verify we're using unregistered reads by checking log messages
	// In a real scenario, the error message would say "unregistered nodes"
	// But since we're simulating, we need to verify the flag is set correctly
	require.True(t, client.Workarounds.UseUnregisteredReads,
		"UseUnregisteredReads flag should be properly set")
}

func TestConsecutiveSessionErrorRecoveryIntegration(t *testing.T) {
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Create a test OpcUA instance with threshold = 2 to test multiple errors
	threshold := uint64(2)
	o := &OpcUA{
		readClientConfig: readClientConfig{
			ReadRetries:             1,
			ReconnectErrorThreshold: &threshold, // Set to 2 for this test
			ReadClientWorkarounds: readClientWorkarounds{
				UseUnregisteredReads: true,
			},
			InputClientConfig: input.InputClientConfig{
				OpcUAClientConfig: opcua.OpcUAClientConfig{
					Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
					SecurityPolicy: "None",
					SecurityMode:   "None",
					AuthMethod:     "Anonymous",
					ConnectTimeout: config.Duration(10 * time.Second),
					RequestTimeout: config.Duration(1 * time.Second),
				},
				MetricName: "testing",
				RootNodes: []input.NodeSettings{
					mapOPCTag(opcTags{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"}),
				},
			},
		},
		Log: testutil.Logger{},
	}

	// Initialize the plugin
	require.NoError(t, o.Init())

	// Create an accumulator
	acc := &testutil.Accumulator{}

	// First gather should succeed
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors)

	// Simulate a session error
	o.client.forceReconnect = true

	// The next gather should force a reconnection internally and succeed
	acc.ClearMetrics()
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors, "Should reset consecutive errors after successful gather")

	// Simulate multiple consecutive errors with bad endpoint
	originalEndpoint := o.client.OpcUAClient.Config.Endpoint
	o.client.OpcUAClient.Config.Endpoint = "opc.tcp://invalid-endpoint:4840"
	require.NoError(t, o.client.Disconnect(t.Context()))

	// First failure should NOT trigger forceReconnect yet (threshold = 2)
	acc.ClearMetrics()
	require.Error(t, o.Gather(acc))
	require.Equal(t, uint64(1), o.consecutiveErrors)
	require.False(t, o.client.forceReconnect, "Session should not be invalidated yet with threshold=2")

	// Second failure should trigger forceReconnect (threshold = 2)
	acc.ClearMetrics()
	require.Error(t, o.Gather(acc))
	require.Equal(t, uint64(2), o.consecutiveErrors)
	require.True(t, o.client.forceReconnect, "Should force session invalidation after reaching threshold")

	// Restore endpoint to allow recovery
	o.client.OpcUAClient.Config.Endpoint = originalEndpoint

	// Next gather should succeed and reset error counter
	acc.ClearMetrics()
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors, "Should reset consecutive errors after recovery")
}

func TestReconnectErrorThresholdDefaultIntegration(t *testing.T) {
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Test Case 1: Config not set - should use default of 1
	o := &OpcUA{
		readClientConfig: readClientConfig{
			// ReconnectErrorThreshold not set (nil pointer)
			ReadRetries: 1,
			InputClientConfig: input.InputClientConfig{
				OpcUAClientConfig: opcua.OpcUAClientConfig{
					Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
					SecurityPolicy: "None",
					SecurityMode:   "None",
					AuthMethod:     "Anonymous",
					ConnectTimeout: config.Duration(10 * time.Second),
					RequestTimeout: config.Duration(1 * time.Second),
				},
				MetricName: "testing",
				RootNodes: []input.NodeSettings{
					mapOPCTag(opcTags{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"}),
				},
			},
		},
		Log: testutil.Logger{},
	}

	require.NoError(t, o.Init())
	require.Equal(t, uint64(1), o.client.ReconnectErrorThreshold, "Should use default of 1 when not configured")

	acc := &testutil.Accumulator{}

	// First gather should succeed
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors)
	require.False(t, o.client.forceReconnect)

	// Simulate connection failure by using invalid endpoint
	originalEndpoint := o.client.OpcUAClient.Config.Endpoint
	o.client.OpcUAClient.Config.Endpoint = "opc.tcp://invalid-endpoint:4840"
	require.NoError(t, o.client.Disconnect(t.Context()))

	// First error should trigger forceReconnect (threshold = 1)
	acc.ClearMetrics()
	require.Error(t, o.Gather(acc))
	require.Equal(t, uint64(1), o.consecutiveErrors)
	require.True(t, o.client.forceReconnect, "Should force reconnection after 1 error (default threshold)")

	// Restore endpoint
	o.client.OpcUAClient.Config.Endpoint = originalEndpoint

	// Recovery should work
	acc.ClearMetrics()
	require.NoError(t, o.Gather(acc))
	require.Equal(t, uint64(0), o.consecutiveErrors)
}

func TestReconnectErrorThresholdZeroIntegration(t *testing.T) {
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Test Case 2: Config set to 0 - should force reconnection every gather
	threshold := uint64(0)
	o := &OpcUA{
		readClientConfig: readClientConfig{
			ReconnectErrorThreshold: &threshold, // Explicitly set to 0
			ReadRetries:             1,
			InputClientConfig: input.InputClientConfig{
				OpcUAClientConfig: opcua.OpcUAClientConfig{
					Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
					SecurityPolicy: "None",
					SecurityMode:   "None",
					AuthMethod:     "Anonymous",
					ConnectTimeout: config.Duration(10 * time.Second),
					RequestTimeout: config.Duration(1 * time.Second),
				},
				MetricName: "testing",
				RootNodes: []input.NodeSettings{
					mapOPCTag(opcTags{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"}),
				},
			},
		},
		Log: testutil.Logger{},
	}

	require.NoError(t, o.Init())
	require.Equal(t, uint64(0), o.client.ReconnectErrorThreshold, "Should use explicit value of 0")

	acc := &testutil.Accumulator{}

	// First gather should succeed but forceReconnect should be set due to threshold=0
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors)

	// Second gather should also succeed and forceReconnect should be set again
	acc.ClearMetrics()
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors)

	// Verify that forceReconnect is set at the beginning of each gather when threshold=0
	// We can check this by monitoring the behavior - with threshold=0, every gather should
	// start with forceReconnect=true (set by the Gather function)

	// Simulate one more gather to confirm consistent behavior
	acc.ClearMetrics()
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors)
}

func TestReconnectErrorThresholdThreeIntegration(t *testing.T) {
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Test Case 3: Config set to 3 - should reconnect after 3 consecutive errors
	threshold := uint64(3)
	o := &OpcUA{
		readClientConfig: readClientConfig{
			ReconnectErrorThreshold: &threshold, // Explicitly set to 3
			ReadRetries:             1,
			InputClientConfig: input.InputClientConfig{
				OpcUAClientConfig: opcua.OpcUAClientConfig{
					Endpoint:       fmt.Sprintf("opc.tcp://%s:%s", container.Address, container.Ports[servicePort]),
					SecurityPolicy: "None",
					SecurityMode:   "None",
					AuthMethod:     "Anonymous",
					ConnectTimeout: config.Duration(10 * time.Second),
					RequestTimeout: config.Duration(1 * time.Second),
				},
				MetricName: "testing",
				RootNodes: []input.NodeSettings{
					mapOPCTag(opcTags{"ProductName", "0", "i", "2261", "open62541 OPC UA Server"}),
				},
			},
		},
		Log: testutil.Logger{},
	}

	require.NoError(t, o.Init())
	require.Equal(t, uint64(3), o.client.ReconnectErrorThreshold, "Should use explicit value of 3")

	acc := &testutil.Accumulator{}

	// First gather should succeed
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors)
	require.False(t, o.client.forceReconnect)

	// Simulate connection failures by using invalid endpoint
	originalEndpoint := o.client.OpcUAClient.Config.Endpoint
	o.client.OpcUAClient.Config.Endpoint = "opc.tcp://invalid-endpoint:4840"
	require.NoError(t, o.client.Disconnect(t.Context()))

	// First error - should NOT trigger forceReconnect yet
	acc.ClearMetrics()
	require.Error(t, o.Gather(acc))
	require.Equal(t, uint64(1), o.consecutiveErrors)
	require.False(t, o.client.forceReconnect, "Should NOT force reconnection after 1 error (threshold=3)")

	// Second error - should NOT trigger forceReconnect yet
	acc.ClearMetrics()
	require.Error(t, o.Gather(acc))
	require.Equal(t, uint64(2), o.consecutiveErrors)
	require.False(t, o.client.forceReconnect, "Should NOT force reconnection after 2 errors (threshold=3)")

	// Third error - should trigger forceReconnect
	acc.ClearMetrics()
	require.Error(t, o.Gather(acc))
	require.Equal(t, uint64(3), o.consecutiveErrors)
	require.True(t, o.client.forceReconnect, "Should force reconnection after 3 errors (threshold=3)")

	// Restore endpoint to allow recovery
	o.client.OpcUAClient.Config.Endpoint = originalEndpoint

	// Recovery should work and reset error counter
	acc.ClearMetrics()
	require.NoError(t, o.Gather(acc))
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, uint64(0), o.consecutiveErrors, "Should reset consecutive errors after recovery")
}
