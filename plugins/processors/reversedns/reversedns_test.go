package reversedns

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSimpleReverseLookup(t *testing.T) {
	now := time.Now()
	m, _ := metric.New("name", map[string]string{
		"dest_ip": "8.8.8.8",
	}, map[string]interface{}{
		"source_ip": "127.0.0.1",
	}, now)

	dns := newReverseDNS()
	dns.Lookups = []lookupEntry{
		{
			Field: "source_ip",
			Dest:  "source_name",
		},
		{
			Tag:  "dest_ip",
			Dest: "dest_name",
		},
	}
	acc := testutil.Accumulator{}
	dns.Init()
	dns.Start(&acc)
	dns.Add(m)
	dns.Stop()
	// should be processed now.

	require.Len(t, acc.GetTelegrafMetrics(), 1)
	processedMetric := acc.GetTelegrafMetrics()[0]
	f, ok := processedMetric.GetField("source_name")
	require.True(t, ok)
	require.EqualValues(t, "localhost", f)

	tag, ok := processedMetric.GetTag("dest_name")
	require.True(t, ok)
	require.EqualValues(t, "dns.google.", tag)
}

func TestLoadingConfig(t *testing.T) {
	c := config.NewConfig()
	err := c.LoadConfigData([]byte("[[processors.reverse_dns]]\n" + sampleConfig))
	require.NoError(t, err)

	require.Len(t, c.Processors, 1)
}
