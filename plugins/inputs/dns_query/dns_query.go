package dns_query

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/miekg/dns"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type DnsQuery struct {
	// Domains or subdomains to query
	Domains []string

	// Network protocol name
	Network string

	// Server to query
	Servers []string

	// Record type
	RecordType string `toml:"record_type"`

	// DNS server port number
	Port int

	// Dns query timeout in seconds. 0 means no timeout
	Timeout int
}

var sampleConfig = `
  ## servers to query
  servers = ["8.8.8.8"]

  ## Network is the network protocol name.
  # network = "udp"

  ## Domains or subdomains to query.
  # domains = ["."]

  ## Query record type.
  ## Posible values: A, AAAA, CNAME, MX, NS, PTR, TXT, SOA, SPF, SRV.
  # record_type = "A"

  ## Dns server port.
  # port = 53

  ## Query timeout in seconds.
  # timeout = 2
`

func (d *DnsQuery) SampleConfig() string {
	return sampleConfig
}

func (d *DnsQuery) Description() string {
	return "Query given DNS server and gives statistics"
}
func (d *DnsQuery) Gather(acc telegraf.Accumulator) error {
	d.setDefaultValues()

	for _, domain := range d.Domains {
		for _, server := range d.Servers {
			dnsQueryTime, err := d.getDnsQueryTime(domain, server)
			acc.AddError(err)
			tags := map[string]string{
				"server":      server,
				"domain":      domain,
				"record_type": d.RecordType,
			}

			fields := map[string]interface{}{"query_time_ms": dnsQueryTime}
			acc.AddFields("dns_query", fields, tags)
		}
	}

	return nil
}

func (d *DnsQuery) setDefaultValues() {
	if d.Network == "" {
		d.Network = "udp"
	}

	if len(d.RecordType) == 0 {
		d.RecordType = "NS"
	}

	if len(d.Domains) == 0 {
		d.Domains = []string{"."}
		d.RecordType = "NS"
	}

	if d.Port == 0 {
		d.Port = 53
	}

	if d.Timeout == 0 {
		d.Timeout = 2
	}
}

func (d *DnsQuery) getDnsQueryTime(domain string, server string) (float64, error) {
	dnsQueryTime := float64(0)

	c := new(dns.Client)
	c.ReadTimeout = time.Duration(d.Timeout) * time.Second
	c.Net = d.Network

	m := new(dns.Msg)
	recordType, err := d.parseRecordType()
	if err != nil {
		return dnsQueryTime, err
	}
	m.SetQuestion(dns.Fqdn(domain), recordType)
	m.RecursionDesired = true

	r, rtt, err := c.Exchange(m, net.JoinHostPort(server, strconv.Itoa(d.Port)))
	if err != nil {
		return dnsQueryTime, err
	}
	if r.Rcode != dns.RcodeSuccess {
		return dnsQueryTime, errors.New(fmt.Sprintf("Invalid answer name %s after %s query for %s\n", domain, d.RecordType, domain))
	}
	dnsQueryTime = float64(rtt.Nanoseconds()) / 1e6
	return dnsQueryTime, nil
}

func (d *DnsQuery) parseRecordType() (uint16, error) {
	var recordType uint16
	var error error

	switch d.RecordType {
	case "A":
		recordType = dns.TypeA
	case "AAAA":
		recordType = dns.TypeAAAA
	case "ANY":
		recordType = dns.TypeANY
	case "CNAME":
		recordType = dns.TypeCNAME
	case "MX":
		recordType = dns.TypeMX
	case "NS":
		recordType = dns.TypeNS
	case "PTR":
		recordType = dns.TypePTR
	case "SOA":
		recordType = dns.TypeSOA
	case "SPF":
		recordType = dns.TypeSPF
	case "SRV":
		recordType = dns.TypeSRV
	case "TXT":
		recordType = dns.TypeTXT
	default:
		error = errors.New(fmt.Sprintf("Record type %s not recognized", d.RecordType))
	}

	return recordType, error
}

func init() {
	inputs.Add("dns_query", func() telegraf.Input {
		return &DnsQuery{}
	})
}
