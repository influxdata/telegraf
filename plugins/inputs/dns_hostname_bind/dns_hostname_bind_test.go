package dns_hostname_bind

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var servers = []string{
	"198.41.0.4",     // a.root-servers.net.
	"192.228.79.201", // b.root-servers.net.
	"192.33.4.12",    // c.root-servers.net.
	"199.7.91.13",    // d.root-servers.net.
}

func TestSampleConfig(t *testing.T) {
	var dnsConfig = DnsHostnameBind{
		Servers: servers,
	}
	sample := dnsConfig.SampleConfig()
	assert.NotEqual(t, "", sample)
}

func TestDescription(t *testing.T) {
	var dnsConfig = DnsHostnameBind{
		Servers: servers,
	}
	desc := dnsConfig.Description()
	assert.NotEqual(t, "", desc)
}

func TestGathering(t *testing.T) {
	if testing.Short() {
		//t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DnsHostnameBind{
		Servers: servers,
	}
	var acc testutil.Accumulator

	err := dnsConfig.Gather(&acc)
	assert.NoError(t, err)
	metric, ok := acc.Get("dns_hostname_bind")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)
	hostname, _ := metric.Fields["hostname"].(string)

	assert.NotEqual(t, 0, queryTime)
	assert.NotEqual(t, "", hostname)
}

func TestSetDefaultValues(t *testing.T) {
	dnsConfig := DnsHostnameBind{}

	dnsConfig.setDefaultValues()

	assert.Equal(t, 53, dnsConfig.Port, "Default port number not equal 53")
	assert.Equal(t, 2, dnsConfig.Timeout, "Default timeout not equal 2")
}
