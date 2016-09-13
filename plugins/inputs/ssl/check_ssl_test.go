package ssl

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var servers = []string{"github.com:443"}

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var sslConfig = CheckExpire{
		Servers: servers,
	}
	var acc testutil.Accumulator

	err := sslConfig.Gather(&acc)
	assert.NoError(t, err)
	metric, ok := acc.Get("check_ssl")
	require.True(t, ok)
	expireTime, _ := metric.Fields["time_to_expire"].(float64)

	assert.NotEqual(t, 0, expireTime)
}

func TestGatheringTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var sslConfig = CheckExpire{
		Servers: servers,
	}
	var acc testutil.Accumulator
	var err error

	channel := make(chan error, 1)
	go func() {
		channel <- sslConfig.Gather(&acc)
	}()
	select {
	case res := <-channel:
		err = res
	case <-time.After(time.Second * 5):
		err = nil
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "i/o timeout")
}
