package dns_query

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *DNSQuery
		expected *DNSQuery
	}{
		{
			name:   "empty",
			plugin: &DNSQuery{},
			expected: &DNSQuery{
				Network:    "udp",
				RecordType: "NS",
				Domains:    []string{"."},
				Port:       53,
			},
		},
		{
			name: "record type",
			plugin: &DNSQuery{
				RecordType: "A",
			},
			expected: &DNSQuery{
				Network:    "udp",
				RecordType: "A",
				Domains:    []string{"."},
				Port:       53,
			},
		},
		{
			name: "domain",
			plugin: &DNSQuery{
				Domains: []string{"google.com"},
			},
			expected: &DNSQuery{
				Network:    "udp",
				RecordType: "NS",
				Domains:    []string{"google.com"},
				Port:       53,
			},
		},
		{
			name: "record type and domain",
			plugin: &DNSQuery{
				RecordType: "A",
				Domains:    []string{"google.com"},
			},
			expected: &DNSQuery{
				Network:    "udp",
				RecordType: "A",
				Domains:    []string{"google.com"},
				Port:       53,
			},
		},
		{
			name: "timeout",
			plugin: &DNSQuery{
				Timeout: config.Duration(2 * time.Second),
			},
			expected: &DNSQuery{
				Network:    "udp",
				RecordType: "NS",
				Domains:    []string{"."},
				Port:       53,
				Timeout:    config.Duration(2 * time.Second),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := tt.plugin
			require.NoError(t, plugin.Init())
			require.EqualExportedValues(t, tt.expected, plugin)
		})
	}
}

func TestRecordTypeParser(t *testing.T) {
	tests := []struct {
		record   string
		expected uint16
	}{
		{
			record:   "A",
			expected: dns.TypeA,
		},
		{
			record:   "AAAA",
			expected: dns.TypeAAAA,
		},
		{
			record:   "ANY",
			expected: dns.TypeANY,
		},
		{
			record:   "CNAME",
			expected: dns.TypeCNAME,
		},
		{
			record:   "MX",
			expected: dns.TypeMX,
		},
		{
			record:   "NS",
			expected: dns.TypeNS,
		},
		{
			record:   "PTR",
			expected: dns.TypePTR,
		},
		{
			record:   "SOA",
			expected: dns.TypeSOA,
		},
		{
			record:   "SPF",
			expected: dns.TypeSPF,
		},
		{
			record:   "SRV",
			expected: dns.TypeSRV,
		},
		{
			record:   "TXT",
			expected: dns.TypeTXT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.record, func(t *testing.T) {
			plugin := &DNSQuery{
				Timeout:    config.Duration(2 * time.Second),
				Domains:    []string{"example.com"},
				RecordType: tt.record,
			}
			require.NoError(t, plugin.Init())
			require.Equal(t, tt.expected, plugin.record)
		})
	}
}

func TestRecordTypeParserError(t *testing.T) {
	plugin := &DNSQuery{
		Timeout:    config.Duration(2 * time.Second),
		RecordType: "nil",
	}
	require.ErrorContains(t, plugin.Init(), "record type \"nil\" not recognized")
}

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := &DNSQuery{
		Servers: []string{"8.8.8.8"},
		Domains: []string{"google.com"},
		Timeout: config.Duration(2 * time.Second),
	}

	var acc testutil.Accumulator
	require.NoError(t, dnsConfig.Init())
	require.NoError(t, acc.GatherError(dnsConfig.Gather))
	m, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, ok := m.Fields["query_time_ms"].(float64)
	require.True(t, ok)
	require.NotEqual(t, float64(0), queryTime)
}

func TestGatherInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := &DNSQuery{
		Servers: []string{"8.8.8.8"},
		Domains: []string{"qwerty123.example.com"},
		Timeout: config.Duration(1 * time.Second),
	}

	var acc testutil.Accumulator
	require.NoError(t, dnsConfig.Init())
	require.NoError(t, dnsConfig.Gather(&acc))
	require.Empty(t, acc.Errors)
}

func TestGatheringMxRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := &DNSQuery{
		Servers:    []string{"8.8.8.8"},
		Domains:    []string{"google.com"},
		RecordType: "MX",
		Timeout:    config.Duration(2 * time.Second),
	}
	var acc testutil.Accumulator

	require.NoError(t, dnsConfig.Init())
	require.NoError(t, acc.GatherError(dnsConfig.Gather))
	m, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, ok := m.Fields["query_time_ms"].(float64)
	require.True(t, ok)
	require.NotEqual(t, float64(0), queryTime)
	preference, ok := m.Fields["preference"].(uint16)
	require.True(t, ok)
	require.NotEqual(t, 0, preference)
}

func TestGatheringRootDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := &DNSQuery{
		Servers:    []string{"8.8.8.8"},
		Domains:    []string{"."},
		RecordType: "MX",
		Timeout:    config.Duration(2 * time.Second),
	}
	require.NoError(t, dnsConfig.Init())

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(dnsConfig.Gather))

	m, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, ok := m.Fields["query_time_ms"].(float64)
	require.True(t, ok)

	expected := []telegraf.Metric{
		metric.New(
			"dns_query",
			map[string]string{
				"server":      "8.8.8.8",
				"domain":      ".",
				"record_type": "MX",
				"rcode":       "NOERROR",
				"result":      "success",
			},
			map[string]interface{}{
				"rcode_value":   0,
				"result_code":   uint64(0),
				"query_time_ms": queryTime,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestMetricContainsServerAndDomainAndRecordTypeTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := &DNSQuery{
		Servers: []string{"8.8.8.8"},
		Domains: []string{"google.com"},
		Timeout: config.Duration(2 * time.Second),
	}
	require.NoError(t, dnsConfig.Init())

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(dnsConfig.Gather))

	m, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, ok := m.Fields["query_time_ms"].(float64)
	require.True(t, ok)
	expected := []telegraf.Metric{
		metric.New(
			"dns_query",
			map[string]string{
				"server":      "8.8.8.8",
				"domain":      "google.com",
				"record_type": "NS",
				"rcode":       "NOERROR",
				"result":      "success",
			},
			map[string]interface{}{
				"rcode_value":   0,
				"result_code":   uint64(0),
				"query_time_ms": queryTime,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestGatheringTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := &DNSQuery{
		Servers: []string{"8.8.8.8"},
		Domains: []string{"google.com"},
		Timeout: config.Duration(1 * time.Second),
		Port:    60054,
	}
	require.NoError(t, dnsConfig.Init())

	var acc testutil.Accumulator
	channel := make(chan error, 1)
	go func() {
		channel <- acc.GatherError(dnsConfig.Gather)
	}()
	select {
	case err := <-channel:
		require.NoError(t, err)
	case <-time.After(time.Second * 2):
		require.Fail(t, "DNS query did not timeout")
	}
}
