package csgo

import (
	"github.com/influxdata/telegraf/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testInput = `CPU   NetIn   NetOut    Uptime  Maps   FPS   Players  Svms    +-ms   ~tick
10.0      1.2      3.4   100     1   120.20       15    5.23    0.01    0.02`

var (
	expectedOutput = statsData{
		10.0, 1.2, 3.4, 100.0, 1, 120.20, 15, 5.23, 0.01, 0.02,
	}
)

func TestCPUStats(t *testing.T) {
	c := NewCSGOStats()
	var acc testutil.Accumulator
	err := c.gatherServer(c.Servers[0], requestMock, &acc)
	if err != nil {
		t.Error(err)
	}

	if !acc.HasMeasurement("csgo") {
		t.Errorf("acc.HasMeasurement: expected csgo")
	}

	assert.Equal(t, "1.2.3.4:1234", acc.Metrics[0].Tags["host"])
	assert.Equal(t, expectedOutput.CPU, acc.Metrics[0].Fields["cpu"])
	assert.Equal(t, expectedOutput.NetIn, acc.Metrics[0].Fields["net_in"])
	assert.Equal(t, expectedOutput.NetOut, acc.Metrics[0].Fields["net_out"])
	assert.Equal(t, expectedOutput.UptimeMinutes, acc.Metrics[0].Fields["uptime_minutes"])
	assert.Equal(t, expectedOutput.Maps, acc.Metrics[0].Fields["maps"])
	assert.Equal(t, expectedOutput.FPS, acc.Metrics[0].Fields["fps"])
	assert.Equal(t, expectedOutput.Players, acc.Metrics[0].Fields["players"])
	assert.Equal(t, expectedOutput.Sim, acc.Metrics[0].Fields["sv_ms"])
	assert.Equal(t, expectedOutput.Variance, acc.Metrics[0].Fields["variance_ms"])
	assert.Equal(t, expectedOutput.Tick, acc.Metrics[0].Fields["tick_ms"])
}

func requestMock(_ string, _ string) (string, error) {
	return testInput, nil
}

func NewCSGOStats() *CSGO {
	return &CSGO{
		Servers: [][]string{
			{"1.2.3.4:1234", "password"},
		},
	}
}
