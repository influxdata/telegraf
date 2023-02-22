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

var servers = []string{"8.8.8.8"}
var domains = []string{"google.com"}

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := DNSQuery{
		Servers: servers,
		Domains: domains,
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

func TestGatheringMxRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := DNSQuery{
		Servers:    servers,
		Domains:    domains,
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
}

func TestGatheringRootDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	dnsConfig := DNSQuery{
		Servers:    servers,
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

	dnsConfig := DNSQuery{
		Servers: servers,
		Domains: domains,
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

	dnsConfig := DNSQuery{
		Servers: servers,
		Domains: domains,
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

func TestSettingDefaultValues(t *testing.T) {
	dnsConfig := DNSQuery{
		Timeout: config.Duration(2 * time.Second),
	}
	require.NoError(t, dnsConfig.Init())
	require.Equal(t, []string{"."}, dnsConfig.Domains, "Default domain not equal \".\"")
	require.Equal(t, "NS", dnsConfig.RecordType, "Default record type not equal 'NS'")
	require.Equal(t, 53, dnsConfig.Port, "Default port number not equal 53")
	require.Equal(t, config.Duration(2*time.Second), dnsConfig.Timeout, "Default timeout not equal 2s")

	dnsConfig = DNSQuery{
		Domains: []string{"."},
		Timeout: config.Duration(2 * time.Second),
	}
	require.NoError(t, dnsConfig.Init())
	require.Equal(t, "NS", dnsConfig.RecordType, "Default record type not equal 'NS'")
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
			plugin := DNSQuery{
				Timeout:    config.Duration(2 * time.Second),
				Domains:    []string{"example.com"},
				RecordType: tt.record,
			}
			require.NoError(t, plugin.Init())
			recordType, err := plugin.parseRecordType()
			require.NoError(t, err)
			require.Equal(t, tt.expected, recordType)
		})
	}
}

func TestRecordTypeParserError(t *testing.T) {
	plugin := DNSQuery{
		Timeout:    config.Duration(2 * time.Second),
		RecordType: "nil",
	}

	_, err := plugin.parseRecordType()
	require.Error(t, err)
}
