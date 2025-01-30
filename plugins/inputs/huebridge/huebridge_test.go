package huebridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/tdrn-org/go-hue/mock"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	conf := config.NewConfig()
	err := conf.LoadConfig("testdata/conf/huebridge.conf")
	require.NoError(t, err)
	require.Len(t, conf.Inputs, 1)
	plugin, ok := conf.Inputs[0].Input.(*HueBridge)
	err = plugin.Init()
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, plugin.Bridges)
	require.Len(t, plugin.Bridges, 4)
	require.NotEmpty(t, plugin.RemoteClientId)
	require.NotEmpty(t, plugin.RemoteClientSecret)
	require.NotEmpty(t, plugin.RemoteCallbackUrl)
	require.NotEmpty(t, plugin.RemoteTokenDir)
	require.NotNil(t, plugin.RoomAssignments)
	require.Len(t, plugin.RoomAssignments, 2)
	require.Equal(t, config.Duration(60*time.Second), plugin.Timeout)
	require.Equal(t, "secret", plugin.TLSKeyPwd)
	require.True(t, plugin.InsecureSkipVerify)
}

func TestInitBridges(t *testing.T) {
	// Start mock server
	bridgeMock := mock.Start()
	require.NotNil(t, bridgeMock)
	defer bridgeMock.Shutdown()
	// Prepare token dir
	tokenDir, err := os.MkdirTemp("", "TestInitBridges")
	require.NoError(t, err)
	tokenFile := filepath.Join(tokenDir, mock.MockClientId, strings.ToUpper(mock.MockBridgeId)+".json")
	bridgeMock.WriteTokenFile(tokenFile)
	// Actual test
	plugin := &HueBridge{Timeout: config.Duration(10 * time.Second)}
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("address://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("cloud://%s:%s@%s/discovery", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("mdns://%s:%s@/", mock.MockBridgeId, mock.MockBridgeUsername))
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("remote://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.RemoteClientId = mock.MockClientId
	plugin.RemoteClientSecret = mock.MockClientSecret
	plugin.RemoteTokenDir = tokenDir
	plugin.InsecureSkipVerify = true
	plugin.Log = testutil.Logger{Name: "huebridge"}
	err = plugin.Init()
	require.NoError(t, err)
}

func TestGatherLocal(t *testing.T) {
	// Start mock server
	bridgeMock := mock.Start()
	require.NotNil(t, bridgeMock)
	defer bridgeMock.Shutdown()
	// Actual test
	plugin := &HueBridge{Timeout: config.Duration(10 * time.Second)}
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("address://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.RoomAssignments = map[string]string{"Name#7": "Name#15"}
	plugin.Log = testutil.Logger{Name: "huebridge"}
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/huebridge.txt", telegraf.Gauge)
}

func testMetric(t *testing.T, acc *testutil.Accumulator, file string, vt telegraf.ValueType) {
	parser := &influx.Parser{}
	err := parser.Init()
	require.NoError(t, err)
	expectedMetrics, err := testutil.ParseMetricsFromFile(file, parser)
	for index := range expectedMetrics {
		expectedMetrics[index].SetType(vt)
	}
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}
