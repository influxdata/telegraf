package dns_query

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var servers = []string{"8.8.8.8"}
var domains = []string{"google.com"}

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DnsQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator

	err := dnsConfig.Gather(&acc)
	assert.NoError(t, err)
	metric, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)

	assert.NotEqual(t, 0, queryTime)
}

func TestGatheringMxRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DnsQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	dnsConfig.RecordType = "MX"

	err := dnsConfig.Gather(&acc)
	assert.NoError(t, err)
	metric, ok := acc.Get("dns_query")
	require.True(t, ok)
	queryTime, _ := metric.Fields["query_time_ms"].(float64)

	assert.NotEqual(t, 0, queryTime)
}

func TestGatheringRootDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	var dnsConfig = DnsQuery{
		Servers:    servers,
		Domains:    []string{"."},
		RecordType: "MX",
	}
	var acc testutil.Accumulator
	tags := map[string]string{
		"server":      "8.8.8.8",
		"domain":      ".",
		"record_type": "MX",
	}
	fields := map[string]interface{}{}

	err := dnsConfig.Gather(&acc)
	assert.NoError(t, err)
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
	var dnsConfig = DnsQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	tags := map[string]string{
		"server":      "8.8.8.8",
		"domain":      "google.com",
		"record_type": "NS",
	}
	fields := map[string]interface{}{}

	err := dnsConfig.Gather(&acc)
	assert.NoError(t, err)
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
	var dnsConfig = DnsQuery{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	dnsConfig.Port = 60054
	dnsConfig.Timeout = 1
	var err error

	channel := make(chan error, 1)
	go func() {
		channel <- dnsConfig.Gather(&acc)
	}()
	select {
	case res := <-channel:
		err = res
	case <-time.After(time.Second * 2):
		err = nil
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "i/o timeout")
}

func TestSettingDefaultValues(t *testing.T) {
	dnsConfig := DnsQuery{}

	dnsConfig.setDefaultValues()

	assert.Equal(t, []string{"."}, dnsConfig.Domains, "Default domain not equal \".\"")
	assert.Equal(t, "NS", dnsConfig.RecordType, "Default record type not equal 'NS'")
	assert.Equal(t, 53, dnsConfig.Port, "Default port number not equal 53")
	assert.Equal(t, 2, dnsConfig.Timeout, "Default timeout not equal 2")

	dnsConfig = DnsQuery{Domains: []string{"."}}

	dnsConfig.setDefaultValues()

	assert.Equal(t, "NS", dnsConfig.RecordType, "Default record type not equal 'NS'")
}

func TestRecordTypeParser(t *testing.T) {
	var dnsConfig = DnsQuery{}
	var recordType uint16

	dnsConfig.RecordType = "A"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeA, recordType)

	dnsConfig.RecordType = "AAAA"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeAAAA, recordType)

	dnsConfig.RecordType = "ANY"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeANY, recordType)

	dnsConfig.RecordType = "CNAME"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeCNAME, recordType)

	dnsConfig.RecordType = "MX"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeMX, recordType)

	dnsConfig.RecordType = "NS"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeNS, recordType)

	dnsConfig.RecordType = "PTR"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypePTR, recordType)

	dnsConfig.RecordType = "SOA"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeSOA, recordType)

	dnsConfig.RecordType = "SPF"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeSPF, recordType)

	dnsConfig.RecordType = "SRV"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeSRV, recordType)

	dnsConfig.RecordType = "TXT"
	recordType, _ = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeTXT, recordType)
}

func TestRecordTypeParserError(t *testing.T) {
	var dnsConfig = DnsQuery{}
	var err error

	dnsConfig.RecordType = "nil"
	_, err = dnsConfig.parseRecordType()
	assert.Error(t, err)
}
