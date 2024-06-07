package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/processors/reverse_dns"
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
	require.EqualValues(t, 3*time.Hour, p.CacheTTL)
	require.EqualValues(t, 17*time.Second, p.LookupTimeout)
	require.Equal(t, 13, p.MaxParallelLookups)
	require.True(t, p.Ordered)
}

func TestDuration(t *testing.T) {
	var d config.Duration

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalText([]byte(`1s`)))
	require.Equal(t, time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalText([]byte(`10`)))
	require.Equal(t, 10*time.Second, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalText([]byte(`1.5`)))
	require.Equal(t, 1500*time.Millisecond, time.Duration(d))

	d = config.Duration(0)
	require.NoError(t, d.UnmarshalText([]byte(``)))
	require.Equal(t, 0*time.Second, time.Duration(d))

	require.Error(t, d.UnmarshalText([]byte(`"1"`)))  // string missing unit
	require.Error(t, d.UnmarshalText([]byte(`'2'`)))  // string missing unit
	require.Error(t, d.UnmarshalText([]byte(`'ns'`))) // string missing time
	require.Error(t, d.UnmarshalText([]byte(`'us'`))) // string missing time
}

func TestSize(t *testing.T) {
	var s config.Size

	require.NoError(t, s.UnmarshalText([]byte(`1B`)))
	require.Equal(t, int64(1), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalText([]byte(`1`)))
	require.Equal(t, int64(1), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalText([]byte(`1GB`)))
	require.Equal(t, int64(1000*1000*1000), int64(s))

	s = config.Size(0)
	require.NoError(t, s.UnmarshalText([]byte(`12GiB`)))
	require.Equal(t, int64(12*1024*1024*1024), int64(s))
}

func TestTOMLParsingStringDurations(t *testing.T) {
	cfg := []byte(`
[[inputs.typesmockup]]
	durations = [
		"1s",
		'''1s''',
		'1s',
		"1.5s",
		"",
		'',
		"2h",
		"42m",
		"100ms",
		"100us",
		"100ns",
		"1d",
		"7.5d",
		"7d8h15m",
		"3d7d",
		"15m8h3.5d"
	]
`)

	expected := []time.Duration{
		1 * time.Second,
		1 * time.Second,
		1 * time.Second,
		1500 * time.Millisecond,
		0,
		0,
		2 * time.Hour,
		42 * time.Minute,
		100 * time.Millisecond,
		100 * time.Microsecond,
		100 * time.Nanosecond,
		24 * time.Hour,
		7*24*time.Hour + 12*time.Hour,
		7*24*time.Hour + 8*time.Hour + 15*time.Minute,
		10 * 24 * time.Hour,
		3*24*time.Hour + 12*time.Hour + 8*time.Hour + 15*time.Minute,
	}

	// Load the data
	c := config.NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)
	plugin := c.Inputs[0].Input.(*MockupTypesPlugin)

	require.Empty(t, plugin.Sizes)
	require.Len(t, plugin.Durations, len(expected))
	for i, actual := range plugin.Durations {
		require.EqualValuesf(t, expected[i], actual, "case %d failed", i)
	}
}

func TestTOMLParsingIntegerDurations(t *testing.T) {
	cfg := []byte(`
[[inputs.typesmockup]]
	durations = [
		1,
		10,
		3601
	]
`)

	expected := []time.Duration{
		1 * time.Second,
		10 * time.Second,
		3601 * time.Second,
	}

	// Load the data
	c := config.NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)
	plugin := c.Inputs[0].Input.(*MockupTypesPlugin)

	require.Empty(t, plugin.Sizes)
	require.Len(t, plugin.Durations, len(expected))
	for i, actual := range plugin.Durations {
		require.EqualValuesf(t, expected[i], actual, "case %d failed", i)
	}
}

func TestTOMLParsingFloatDurations(t *testing.T) {
	cfg := []byte(`
[[inputs.typesmockup]]
	durations = [
		42.0,
		1.5
	]
`)

	expected := []time.Duration{
		42 * time.Second,
		1500 * time.Millisecond,
	}

	// Load the data
	c := config.NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)
	plugin := c.Inputs[0].Input.(*MockupTypesPlugin)

	require.Empty(t, plugin.Sizes)
	require.Len(t, plugin.Durations, len(expected))
	for i, actual := range plugin.Durations {
		require.EqualValuesf(t, expected[i], actual, "case %d failed", i)
	}
}

func TestTOMLParsingStringSizes(t *testing.T) {
	cfg := []byte(`
[[inputs.typesmockup]]
	sizes = [
		"1B",
		"1",
		'1',
		'''15kB''',
		"""15KiB""",
		"1GB",
		"12GiB"
	]
`)

	expected := []int64{
		1,
		1,
		1,
		15 * 1000,
		15 * 1024,
		1000 * 1000 * 1000,
		12 * 1024 * 1024 * 1024,
	}

	// Load the data
	c := config.NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)
	plugin := c.Inputs[0].Input.(*MockupTypesPlugin)

	require.Empty(t, plugin.Durations)
	require.Len(t, plugin.Sizes, len(expected))
	for i, actual := range plugin.Sizes {
		require.EqualValuesf(t, expected[i], actual, "case %d failed", i)
	}
}

func TestTOMLParsingIntegerSizes(t *testing.T) {
	cfg := []byte(`
[[inputs.typesmockup]]
	sizes = [
		0,
		1,
		1000,
		1024
	]
`)

	expected := []int64{
		0,
		1,
		1000,
		1024,
	}

	// Load the data
	c := config.NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)
	plugin := c.Inputs[0].Input.(*MockupTypesPlugin)

	require.Empty(t, plugin.Durations)
	require.Len(t, plugin.Sizes, len(expected))
	for i, actual := range plugin.Sizes {
		require.EqualValuesf(t, expected[i], actual, "case %d failed", i)
	}
}

/*** Mockup (input) plugin for testing to avoid cyclic dependencies ***/
type MockupTypesPlugin struct {
	Durations []config.Duration `toml:"durations"`
	Sizes     []config.Size     `toml:"sizes"`
}

func (*MockupTypesPlugin) SampleConfig() string                { return "Mockup test types plugin" }
func (*MockupTypesPlugin) Gather(_ telegraf.Accumulator) error { return nil }

// Register the mockup plugin on loading
func init() {
	// Register the mockup input plugin for the required names
	inputs.Add("typesmockup", func() telegraf.Input { return &MockupTypesPlugin{} })
}
