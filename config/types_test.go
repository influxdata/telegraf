package config_test

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors/reverse_dns"
	"github.com/stretchr/testify/require"
)

func TestConfigDuration(t *testing.T) {
	c := config.NewConfig()
	err := c.LoadConfigData([]byte(`
[[processors.reverse_dns]]
  cache_ttl = "3h"
  lookup_timeout = "17s"
  max_parallel_lookups = 13
  ordered = true
  [[processors.reverse_dns.lookup]]
    field = "source_ip"
    dest = "source_name"
`))
	require.NoError(t, err)
	require.Len(t, c.Processors, 1)
	p := c.Processors[0].Processor.(*reverse_dns.ReverseDNS)
	require.EqualValues(t, p.CacheTTL, 3*time.Hour)
	require.EqualValues(t, p.LookupTimeout, 17*time.Second)
	require.Equal(t, p.MaxParallelLookups, 13)
	require.Equal(t, p.Ordered, true)
}
