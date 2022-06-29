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

func TestDuration(t *testing.T) {
	var d config.Duration

	require.NoError(t, d.UnmarshalTOML([]byte(`"1s"`)))
	require.Equal(t, time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalTOML([]byte(`1s`)))
	require.Equal(t, time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalTOML([]byte(`'1s'`)))
	require.Equal(t, time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalTOML([]byte(`10`)))
	require.Equal(t, 10*time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalTOML([]byte(`1.5`)))
	require.Equal(t, time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalTOML([]byte(``)))
	require.Equal(t, 0*time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalTOML([]byte(`""`)))
	require.Equal(t, 0*time.Second, time.Duration(d))

	require.Error(t, d.UnmarshalTOML([]byte(`"1"`)))  // string missing unit
	require.Error(t, d.UnmarshalTOML([]byte(`'2'`)))  // string missing unit
	require.Error(t, d.UnmarshalTOML([]byte(`'ns'`))) // string missing time
	require.Error(t, d.UnmarshalTOML([]byte(`'us'`))) // string missing time
}

func TestSize(t *testing.T) {
	var s config.Size

	require.NoError(t, s.UnmarshalTOML([]byte(`"1B"`)))
	require.Equal(t, int64(1), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalTOML([]byte(`1`)))
	require.Equal(t, int64(1), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalTOML([]byte(`'1'`)))
	require.Equal(t, int64(1), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalTOML([]byte(`"1GB"`)))
	require.Equal(t, int64(1000*1000*1000), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalTOML([]byte(`"12GiB"`)))
	require.Equal(t, int64(12*1024*1024*1024), int64(s))
}
