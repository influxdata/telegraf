package csgo

import (
	"testing"
	"time"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon/rcontest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestCPUStats(t *testing.T) {
	// Define the input
	const input = `CPU   NetIn   NetOut    Uptime  Maps   FPS   Players  Svms    +-ms   ~tick
10.0      1.2      3.4   100     1   120.20       15    5.23    0.01    0.02`

	// Start the mockup server
	server := rcontest.NewUnstartedServer()
	server.Settings.Password = "password"
	server.SetAuthHandler(func(c *rcontest.Context) {
		if c.Request().Body() == c.Server().Settings.Password {
			pkg := rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, c.Request().ID, "")
			_, err := pkg.WriteTo(c.Conn())
			require.NoError(t, err)
		} else {
			pkg := rcon.NewPacket(rcon.SERVERDATA_AUTH_RESPONSE, -1, string([]byte{0x00}))
			_, err := pkg.WriteTo(c.Conn())
			require.NoError(t, err)
		}
	})
	server.SetCommandHandler(func(c *rcontest.Context) {
		pkg := rcon.NewPacket(rcon.SERVERDATA_RESPONSE_VALUE, c.Request().ID, input)
		_, err := pkg.WriteTo(c.Conn())
		require.NoError(t, err)
	})
	server.Start()
	defer server.Close()

	// Setup the plugin
	plugin := &CSGO{
		Servers: [][]string{
			{server.Addr(), "password"},
		},
	}
	require.NoError(t, plugin.Init())

	// Define expected result
	expected := []telegraf.Metric{
		metric.New(
			"csgo",
			map[string]string{
				"host": server.Addr(),
			},
			map[string]interface{}{
				"cpu":            10.0,
				"fps":            120.2,
				"maps":           1.0,
				"net_in":         1.2,
				"net_out":        3.4,
				"players":        15.0,
				"sv_ms":          5.23,
				"tick_ms":        0.02,
				"uptime_minutes": 100.0,
				"variance_ms":    0.01,
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	// Gather data
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	// Test the result
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}
