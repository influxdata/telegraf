package nsdp

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
	"github.com/tdrn-org/go-nsdp"
)

func TestInitDefaults(t *testing.T) {
	plugin := defaultNSDP()
	err := plugin.Init()
	require.NoError(t, err)
	require.Equal(t, nsdp.IPv4BroadcastTarget, plugin.Target)
	require.Equal(t, uint(0), plugin.DeviceLimit)
	require.Equal(t, defaultTimeout, plugin.Timeout)
	require.NotNil(t, plugin.Log)
}

func TestConfig(t *testing.T) {
	conf, err := os.ReadFile("testdata/conf/nsdp.conf")
	require.NoError(t, err)
	var plugin = defaultNSDP()
	err = toml.Unmarshal(conf, plugin)
	require.NoError(t, err)
	err = plugin.Init()
	require.NoError(t, err)
	require.Equal(t, pluginTestResponderTarget, plugin.Target)
	require.Equal(t, uint(1), plugin.DeviceLimit)
	require.Equal(t, config.Duration(5*time.Second), plugin.Timeout)
}

const pluginTestResponderTarget = "127.0.0.1:63322"

func TestGather(t *testing.T) {
	// Enable debug logging
	logger.SetupLogging(&logger.Config{Debug: true})
	// Setup test responder
	responder, err := nsdp.NewTestResponder(pluginTestResponderTarget)
	require.Nil(t, err)
	defer responder.Stop()
	responder.AddResponses(
		"0102000000000000bcd07432b8dc123456789abc000037b94e534450000000000001000847533130384576330003000773776974636832000600040a010004100000310100000000e73b5f1a000000001e31523c0000000000000000000000000000000000000000000000000000000000000000100000310200000000152d5eae0000000052ea11ea0000000000000000000000000000000000000000000000000000000000000000100000310300000000068561aa00000000bcc8cb35000000000000000000000000000000000000000000000000000000000000000010000031040000000002d5fe00000000002b37dad900000000000000000000000000000000000000000000000000000000000000001000003105000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000310600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000031070000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000003108000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffff0000",
		"0102000000000000bcd07432b8dccba987654321000037b94e534450000000000001000847533130384576330003000773776974636831000600040a0100031000003101000000059a9d833200000000303e8eb5000000000000000000000000000000000000000000000000000000000000000010000031020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000003103000000000d9a35e4000000026523c66600000000000000000000000000000000000000000000000000000000000000001000003104000000000041c7530000000002cd94ba000000000000000000000000000000000000000000000000000000000000000010000031050000000021b9ca41000000031a9bff610000000000000000000000000000000000000000000000000000000000000000100000310600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000031070000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000003108000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffff0000")
	err = responder.Start()
	require.Nil(t, err)
	// Actual test
	plugin := defaultNSDP()
	plugin.Target = pluginTestResponderTarget
	plugin.DeviceLimit = 2
	err = plugin.Init()
	require.Nil(t, err)
	acc := &testutil.Accumulator{}
	err = acc.GatherError(plugin.Gather)
	require.NoError(t, err)
	testMetric(t, acc, "testdata/metrics/nsdp_device_port.txt", telegraf.Counter)
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
