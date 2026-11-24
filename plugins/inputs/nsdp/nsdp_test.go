package nsdp

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-nsdp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestLoadConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/nsdp.conf"))
	require.Len(t, conf.Inputs, 1)
	plugin, ok := conf.Inputs[0].Input.(*NSDP)
	require.True(t, ok)

	// Verify successful Init
	require.NoError(t, plugin.Init())

	// Verify everything is setup according to config file
	require.Equal(t, "127.0.0.1:63322", plugin.Address)
	require.Equal(t, uint(1), plugin.DeviceLimit)
	require.Equal(t, config.Duration(5*time.Second), plugin.Timeout)
}

func TestInvalidTimeoutConfig(t *testing.T) {
	plugin := &NSDP{
		Timeout: config.Duration(0 * time.Second),
	}

	// Verify failing Init
	require.EqualError(t, plugin.Init(), "timeout must be greater than zero")
}

func TestGather(t *testing.T) {
	// Responses of the test responder
	responses := []string{
		"0102000000000000bcd07432b8dc123456789abc000037b94e53445000000000" +
			"0001000847533130384576330003000773776974636832000600040a01000410" +
			"0000310100000000e73b5f1a000000001e31523c000000000000000000000000" +
			"0000000000000000000000000000000000000000100000310200000000152d5e" +
			"ae0000000052ea11ea0000000000000000000000000000000000000000000000" +
			"000000000000000000100000310300000000068561aa00000000bcc8cb350000" +
			"0000000000000000000000000000000000000000000000000000000000001000" +
			"0031040000000002d5fe00000000002b37dad900000000000000000000000000" +
			"0000000000000000000000000000000000000010000031050000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000100000310600000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000100000" +
			"3107000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000001000003108000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000ffff0000",
		"0102000000000000bcd07432b8dccba987654321000037b94e53445000000000" +
			"0001000847533130384576330003000773776974636831000600040a01000310" +
			"00003101000000059a9d833200000000303e8eb5000000000000000000000000" +
			"0000000000000000000000000000000000000000100000310200000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000001000003103000000000d9a35e4000000026523c6660000" +
			"0000000000000000000000000000000000000000000000000000000000001000" +
			"003104000000000041c7530000000002cd94ba00000000000000000000000000" +
			"0000000000000000000000000000000000000010000031050000000021b9ca41" +
			"000000031a9bff61000000000000000000000000000000000000000000000000" +
			"0000000000000000100000310600000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000100000" +
			"3107000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000001000003108000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000ffff0000",
	}

	// Setup and start test responder and plugin. The NSDP client always
	// listens on the target port minus one which might be in use, so retry
	// with a different port pair for a limited time.
	var responder *nsdp.TestResponder
	var plugin *NSDP
	var err error
	require.Eventually(t, func() bool {
		// Start the responder which should always succeed
		if responder, err = nsdp.NewTestResponder("localhost:0"); err != nil {
			return true
		}
		responder.AddResponses(responses...)
		if err = responder.Start(); err != nil {
			return true
		}

		// Setup the plugin to target the test responder
		plugin = &NSDP{
			Address:     responder.Target(),
			DeviceLimit: 2,
			Timeout:     config.Duration(2 * time.Second),
			Log:         testutil.Logger{Name: "nsdp"},
		}
		if err = plugin.Init(); err != nil {
			if serr := responder.Stop(); serr != nil {
				t.Logf("stopping responder failed: %v", serr)
			}
			return true
		}

		// Try to start the plugin which might fail with "address already in use".
		// In this case we try again. All other errors must fail the test.
		if err = plugin.Start(nil); err != nil {
			if serr := responder.Stop(); serr != nil {
				t.Logf("stopping responder failed: %v", serr)
			}
			return !strings.Contains(err.Error(), "bind:")
		}
		return true
	}, 5*time.Second, 100*time.Millisecond, "unable to find a free port pair for the NSDP responder")
	require.NoError(t, err)
	defer responder.Stop() //nolint:errcheck // ignore error
	defer plugin.Stop()

	// Verify successfull Gather
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	// Verify collected metrics are as expected
	expectedMetrics := loadExpectedMetrics(t, "testdata/metrics/nsdp_device_port.txt", telegraf.Counter)
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func loadExpectedMetrics(t *testing.T, file string, vt telegraf.ValueType) []telegraf.Metric {
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	expectedMetrics, err := testutil.ParseMetricsFromFile(file, parser)
	require.NoError(t, err)
	for index := range expectedMetrics {
		expectedMetrics[index].SetType(vt)
	}
	return expectedMetrics
}
