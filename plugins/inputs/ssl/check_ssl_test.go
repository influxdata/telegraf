package ssl

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var servers = []string{"github.com:443"}
var server_of_timeout = []string{"8.8.8.8:443"}
var tout = "1s"

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var sslConfig = CheckExpire{
		Servers: servers,
		Timeout: tout,
	}
	var acc testutil.Accumulator
	tags := map[string]string{
		"server": "github.com:443",
	}
	fields := map[string]interface{}{}

	err := sslConfig.Gather(&acc)
	assert.NoError(t, err)
	metric, ok := acc.Get("ssl_cert")
	require.True(t, ok)
	expireTime, _ := metric.Fields["time_to_expire"]

	assert.NotEqual(t, 0, expireTime)
	fields["time_to_expire"] = expireTime
	acc.AssertContainsTaggedFields(t, "ssl_cert", fields, tags)
}

func TestGatheringTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var sslConfig = CheckExpire{
		Servers: server_of_timeout,
		Timeout: tout,
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
