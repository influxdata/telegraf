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

func TestDuration_Marshal(t *testing.T) {
	var d config.Duration
	var p *config.Duration

	b, err := p.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"0s"`, string(b))

	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"0s"`, string(b))

	d = config.Duration(1 * time.Millisecond)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"1ms"`, string(b))

	d = config.Duration(1 * time.Second)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"1s"`, string(b))

	d = config.Duration(1 * time.Minute)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"1m"`, string(b))

	d = config.Duration(1 * time.Hour)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"1h"`, string(b))

	d = config.Duration(36 * time.Hour)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"36h"`, string(b))

	d = config.Duration(6*time.Hour + 12*time.Minute)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"6h12m"`, string(b))

	d = config.Duration(6*time.Hour + 12*time.Minute + 5*time.Second)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"6h12m5s"`, string(b))

	d = config.Duration(6*time.Hour + 12*time.Minute + 5*time.Second + 39*time.Millisecond)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"6h12m5.039s"`, string(b))

	d = config.Duration(6*time.Hour + 12*time.Minute + 5*time.Second + 39*time.Millisecond + 23*time.Microsecond)
	b, err = d.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, `"6h12m5.039023s"`, string(b))
}

func TestDuration_Unmarshal(t *testing.T) {
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
}

func TestSize_Marshal(t *testing.T) {
	var s config.Size
	var p *config.Size

	b, err := p.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, "0", string(b))

	b, err = s.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, "0", string(b))

	s = config.Size(11)
	b, err = s.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, "11", string(b))

	s = config.Size(11 * 1024)
	b, err = s.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, "11264", string(b))

	s = config.Size(11 * 1024 * 1024)
	b, err = s.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, "11534336", string(b))

	s = config.Size(454867870564751864)
	b, err = s.MarshalTOML()
	require.NoError(t, err)
	require.Equal(t, "454867870564751864", string(b))
}

func TestSize_Unmarshal(t *testing.T) {
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
