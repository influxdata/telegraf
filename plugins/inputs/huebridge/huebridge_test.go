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
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/tdrn-org/go-hue/mock"

	"github.com/stretchr/testify/require"
)

func TestInitDefaults(t *testing.T) {
	plugin := defaultHueBridge()
	err := plugin.Init()
	require.NoError(t, err)
	require.NotNil(t, plugin.Bridges)
	require.Len(t, plugin.Bridges, 0)
	require.Empty(t, plugin.RemoteClientId)
	require.Empty(t, plugin.RemoteClientSecret)
	require.Empty(t, plugin.RemoteCallbackUrl)
	require.Empty(t, plugin.RemoteTokenDir)
	require.NotNil(t, plugin.RoomAssignments)
	require.Len(t, plugin.RoomAssignments, 0)
	require.Equal(t, defaultTimeout, plugin.Timeout)
	require.Empty(t, plugin.TLSCA)
	require.Empty(t, plugin.TLSCert)
	require.Empty(t, plugin.TLSKey)
	require.Empty(t, plugin.TLSKeyPwd)
	require.False(t, plugin.InsecureSkipVerify)
	require.NotNil(t, plugin.Log)
}

func TestConfig(t *testing.T) {
	conf, err := os.ReadFile("testdata/conf/huebridge.conf")
	require.NoError(t, err)
	plugin := defaultHueBridge()
	err = toml.Unmarshal(conf, plugin)
	require.NoError(t, err)
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
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultHueBridge()
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("address://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("cloud://%s:%s@%s/discovery", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("mdns://%s:%s@/", mock.MockBridgeId, mock.MockBridgeUsername))
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("remote://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.RemoteClientId = mock.MockClientId
	plugin.RemoteClientSecret = mock.MockClientSecret
	plugin.RemoteTokenDir = tokenDir
	plugin.InsecureSkipVerify = true
	err = plugin.Init()
	require.NoError(t, err)
}

func TestGatherLocal(t *testing.T) {
	// Start mock server
	bridgeMock := mock.Start()
	require.NotNil(t, bridgeMock)
	defer bridgeMock.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultHueBridge()
	plugin.Bridges = append(plugin.Bridges, fmt.Sprintf("address://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host))
	plugin.RoomAssignments = [][]string{{"Name#7", "Name#15"}}
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
