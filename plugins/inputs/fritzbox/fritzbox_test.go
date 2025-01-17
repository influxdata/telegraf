package fritzbox

import (
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-tr064/mock"
)

func TestInitDefaults(t *testing.T) {
	plugin := defaultFritzbox()
	err := plugin.Init()
	require.NoError(t, err)
	require.Len(t, plugin.URLs, 0)
	require.True(t, plugin.DeviceInfo)
	require.True(t, plugin.WanInfo)
	require.True(t, plugin.PppInfo)
	require.True(t, plugin.DslInfo)
	require.True(t, plugin.WlanInfo)
	require.False(t, plugin.HostsInfo)
	require.Equal(t, 30, plugin.FullQueryCycle)
	require.Equal(t, defaultTimeout, plugin.Timeout)
	require.False(t, plugin.TlsSkipVerify)
	require.NotNil(t, plugin.Log)
}

func TestConfig(t *testing.T) {
	conf, err := os.ReadFile("testdata/fritzbox.conf")
	require.NoError(t, err)
	var plugin = defaultFritzbox()
	err = toml.Unmarshal(conf, plugin)
	require.NoError(t, err)
	err = plugin.Init()
	require.NoError(t, err)
	require.Len(t, plugin.URLs, 2)
	require.False(t, plugin.DeviceInfo)
	require.False(t, plugin.WanInfo)
	require.False(t, plugin.PppInfo)
	require.False(t, plugin.DslInfo)
	require.False(t, plugin.WlanInfo)
	require.True(t, plugin.HostsInfo)
	require.Equal(t, 6, plugin.FullQueryCycle)
	require.Equal(t, config.Duration(60*time.Second), plugin.Timeout)
	require.True(t, plugin.TlsSkipVerify)
}

var testMocks = []*mock.ServiceMock{
	mock.ServiceMockFromFile("/deviceinfo", "testdata/DeviceInfo.xml"),
	mock.ServiceMockFromFile("/wancommonifconfig", "testdata/WANCommonInterfaceConfig1.xml"),
	mock.ServiceMockFromFile("/WANCommonIFC1", "testdata/WANCommonInterfaceConfig2.xml"),
	mock.ServiceMockFromFile("/wanpppconn", "testdata/WANPPPConnection.xml"),
	mock.ServiceMockFromFile("/wandslifconfig", "testdata/WANDSLInterfaceConfig.xml"),
	mock.ServiceMockFromFile("/wlanconfig", "testdata/WLANConfiguration.xml"),
	mock.ServiceMockFromFile("/hosts", "testdata/Hosts.xml"),
}

func TestGatherDeviceInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultFritzbox()
	plugin.URLs = append(plugin.URLs, tr064Server.Server().String())
	plugin.DeviceInfo = true
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/fritzbox_device.txt", telegraf.Untyped)
}

func TestGatherWanInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultFritzbox()
	plugin.URLs = append(plugin.URLs, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = true
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/fritzbox_wan.txt", telegraf.Untyped)
}

func TestGatherPppInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultFritzbox()
	plugin.URLs = append(plugin.URLs, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = true
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/fritzbox_ppp.txt", telegraf.Untyped)
}

func TestGatherDslInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultFritzbox()
	plugin.URLs = append(plugin.URLs, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = true
	plugin.WlanInfo = false
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/fritzbox_dsl.txt", telegraf.Untyped)
}

func TestGatherWlanInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultFritzbox()
	plugin.URLs = append(plugin.URLs, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = true
	plugin.HostsInfo = false
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/fritzbox_wlan.txt", telegraf.Gauge)
}

func TestGatherHostsInfo(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata", testMocks...)
	defer tr064Server.Shutdown()
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Actual test
	plugin := defaultFritzbox()
	plugin.URLs = append(plugin.URLs, tr064Server.Server().String())
	plugin.DeviceInfo = false
	plugin.WanInfo = false
	plugin.PppInfo = false
	plugin.DslInfo = false
	plugin.WlanInfo = false
	plugin.HostsInfo = true
	err := plugin.Init()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/fritzbox_host.txt", telegraf.Gauge)
}

func testMetric(t *testing.T, acc *testutil.Accumulator, file string, vt telegraf.ValueType) {
	testutil.PrintMetrics(acc.GetTelegrafMetrics())
	parser := &influx.Parser{}
	err := parser.Init()
	require.NoError(t, err)
	expectedMetrics, err := testutil.ParseMetricsFromFile(file, parser)
	for index := range expectedMetrics {
		expectedMetrics[index].SetType(vt)
	}
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
