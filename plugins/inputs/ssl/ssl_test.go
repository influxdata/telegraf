package ssl

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var servers = []Server{
		{
			Host:    "github.com:443",
			Timeout: 5,
		},
		{
			Host:    "github.com",
			Timeout: 5,
		},
	}
	var sslConfig = Ssl{
		Servers: servers,
	}
	var acc testutil.Accumulator

	err := acc.GatherError(sslConfig.Gather)
	assert.NoError(t, err)
	metric, ok := acc.Get("ssl")
	require.True(t, ok)
	timeToExp, _ := metric.Fields["time_to_expiration"].(float64)

	assert.NotEqual(t, 0, timeToExp)
}
