package dns

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var servers = []string{"8.8.8.8"}
var domains = []string{"mjasion.pl"}

func TestGathering(t *testing.T) {
	var dnsConfig = Dns{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator

	dnsConfig.Gather(&acc)
	metric, _ := acc.Get("dns")
	queryTime, _ := metric.Fields["value"].(float64)

	assert.NotEqual(t, 0, queryTime)
}

func TestGatheringMxRecord(t *testing.T) {
	var dnsConfig = Dns{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	dnsConfig.RecordType = "MX"

	dnsConfig.Gather(&acc)
	metric, _ := acc.Get("dns")
	queryTime, _ := metric.Fields["value"].(float64)

	assert.NotEqual(t, 0, queryTime)
}

func TestMetricContainsServerAndDomainAndRecordTypeTags(t *testing.T) {
	var dnsConfig = Dns{
		Servers: servers,
		Domains: domains,
	}
	var acc testutil.Accumulator
	tags := map[string]string{
		"server":     "8.8.8.8",
		"domain":     "mjasion.pl",
		"recordType": "A",
	}
	fields := map[string]interface{}{}

	dnsConfig.Gather(&acc)
	metric, _ := acc.Get("dns")
	queryTime, _ := metric.Fields["value"].(float64)

	fields["value"] = queryTime
	acc.AssertContainsTaggedFields(t, "dns", fields, tags)
}

func TestGatheringTimeout(t *testing.T) {
	var dnsConfig = Dns{
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
	dnsConfig := Dns{}

	dnsConfig.setDefaultValues()

	assert.Equal(t, "A", dnsConfig.RecordType, "Default record type not equal 'A'")
	assert.Equal(t, 53, dnsConfig.Port, "Default port number not equal 53")
	assert.Equal(t, 2, dnsConfig.Timeout, "Default timeout not equal 2")
}

func TestRecordTypeParser(t *testing.T) {
	var dnsConfig = Dns{}
	var recordType uint16
	var err error

	dnsConfig.RecordType = "A"
	recordType, err = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeA, recordType)

	dnsConfig.RecordType = "CNAME"
	recordType, err = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeCNAME, recordType)

	dnsConfig.RecordType = "MX"
	recordType, err = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeMX, recordType)

	dnsConfig.RecordType = "NS"
	recordType, err = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeNS, recordType)

	dnsConfig.RecordType = "TXT"
	recordType, err = dnsConfig.parseRecordType()
	assert.Equal(t, dns.TypeTXT, recordType)

	dnsConfig.RecordType = "nil"
	recordType, err = dnsConfig.parseRecordType()
	assert.Error(t, err)
}
