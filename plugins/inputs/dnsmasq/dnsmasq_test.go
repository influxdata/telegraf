package dnsmasq

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test should with dnsmasq server,or run container with
// `docker run -p 53:53/tcp -p 53:53/udp --cap-add=NET_ADMIN andyshinn/dnsmasq:2.75`
var server = "127.0.0.1:53"
var port = 53

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	dnsClient := &dns.Client{
		SingleInflight: true,
	}
	var dnsmasqConfig = Dnsmasq{
		c:      dnsClient,
		Server: server,
		Port:   port,
	}
	var acc testutil.Accumulator

	err := acc.GatherError(dnsmasqConfig.Gather)
	assert.NoError(t, err)
	metric, ok := acc.Get("dnsmasq")
	require.True(t, ok)

	metricNames := []string{
		"auth",
		"cachesize",
		"evictions",
		"hits",
		"insertions",
		"misses",
		"queries",
		"queries_failed",
	}
	for _, metricName := range metricNames {
		_, ok := metric.Fields[metricName].(float64)
		assert.True(t, true, ok)
	}
}

func TestSettingDefaultValues(t *testing.T) {
	dnsmasqConfig := Dnsmasq{}

	dnsmasqConfig.setDefaultValues()

	assert.Equal(t, "127.0.0.1", dnsmasqConfig.Server, "Default dnsmasq server ip not equal \"127.0.0.1\"")
	assert.Equal(t, 53, dnsmasqConfig.Port, "Default dnsmasq server port not equal 53")
}
