package dns_query

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

var servers = []string{"8.8.8.8"}
var domains = []string{"google.com"}

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DNSQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator

	err := acc.GatherError(dnsConfig.Gather)
	require.NoError(t, err)
	metric, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)

	require.NotEqual(t, 0, queryTime)
}

func TestGatheringMxRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DNSQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	dnsConfig.RecordType = "MX"

	err := acc.GatherError(dnsConfig.Gather)
	require.NoError(t, err)
	metric, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)

	require.NotEqual(t, 0, queryTime)
}

func TestGatheringRootDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DNSQuery{
		Servers:    servers,
		Domains:    []string{"."},
		RecordType: "MX",
	}
	var acc testutil.Accumulator
	tags := map[string]string{
		"server":      "8.8.8.8",
		"domain":      ".",
		"record_type": "MX",
		"rcode":       "NOERROR",
		"result":      "success",
	}
	fields := map[string]interface{}{
		"rcode_value": 0,
		"result_code": uint64(0),
	}

	err := acc.GatherError(dnsConfig.Gather)
	require.NoError(t, err)
	metric, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)

	fields["query_time_ms"] = queryTime
	acc.AssertContainsTaggedFields(t, "dns_query", fields, tags)
}

func TestMetricContainsServerAndDomainAndRecordTypeTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DNSQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	tags := map[string]string{
		"server":      "8.8.8.8",
		"domain":      "google.com",
		"record_type": "NS",
		"rcode":       "NOERROR",
		"result":      "success",
	}
	fields := map[string]interface{}{
		"rcode_value": 0,
		"result_code": uint64(0),
	}

	err := acc.GatherError(dnsConfig.Gather)
	require.NoError(t, err)
	metric, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)

	fields["query_time_ms"] = queryTime
	acc.AssertContainsTaggedFields(t, "dns_query", fields, tags)
}

func TestGatheringTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DNSQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	dnsConfig.Port = 60054
	dnsConfig.Timeout = 1

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
	dnsConfig := DNSQuery{}

	dnsConfig.setDefaultValues()

	require.Equal(t, []string{"."}, dnsConfig.Domains, "Default domain not equal \".\"")
	require.Equal(t, "NS", dnsConfig.RecordType, "Default record type not equal 'NS'")
	require.Equal(t, 53, dnsConfig.Port, "Default port number not equal 53")
	require.Equal(t, 2, dnsConfig.Timeout, "Default timeout not equal 2")

	dnsConfig = DNSQuery{Domains: []string{"."}}

	dnsConfig.setDefaultValues()

	require.Equal(t, "NS", dnsConfig.RecordType, "Default record type not equal 'NS'")
}

func TestRecordTypeParser(t *testing.T) {
	var dnsConfig = DNSQuery{}
	var recordType uint16

	dnsConfig.RecordType = "A"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeA, recordType)

	dnsConfig.RecordType = "AAAA"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeAAAA, recordType)

	dnsConfig.RecordType = "ANY"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeANY, recordType)

	dnsConfig.RecordType = "CNAME"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeCNAME, recordType)

	dnsConfig.RecordType = "MX"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeMX, recordType)

	dnsConfig.RecordType = "NS"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeNS, recordType)

	dnsConfig.RecordType = "PTR"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypePTR, recordType)

	dnsConfig.RecordType = "SOA"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeSOA, recordType)

	dnsConfig.RecordType = "SPF"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeSPF, recordType)

	dnsConfig.RecordType = "SRV"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeSRV, recordType)

	dnsConfig.RecordType = "TXT"
	recordType, _ = dnsConfig.parseRecordType()
	require.Equal(t, dns.TypeTXT, recordType)
}

func TestRecordTypeParserError(t *testing.T) {
	var dnsConfig = DNSQuery{}
	var err error

	dnsConfig.RecordType = "nil"
	_, err = dnsConfig.parseRecordType()
	require.Error(t, err)
}
