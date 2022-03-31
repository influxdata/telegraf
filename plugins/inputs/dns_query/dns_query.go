package dns_query

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ResultType uint64

const (
	Success ResultType = iota
	Timeout
	Error
)

type DNSQuery struct {
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

func (d *DNSQuery) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	d.setDefaultValues()

	for _, domain := range d.Domains {
		for _, server := range d.Servers {
			wg.Add(1)
			go func(domain, server string) {
				fields := make(map[string]interface{}, 2)
				tags := map[string]string{
					"server":      server,
					"domain":      domain,
					"record_type": d.RecordType,
				}

				dnsQueryTime, rcode, err := d.getDNSQueryTime(domain, server)
				if rcode >= 0 {
					tags["rcode"] = dns.RcodeToString[rcode]
					fields["rcode_value"] = rcode
				}
				if err == nil {
					setResult(Success, fields, tags)
					fields["query_time_ms"] = dnsQueryTime
				} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					setResult(Timeout, fields, tags)
				} else if err != nil {
					setResult(Error, fields, tags)
					acc.AddError(err)
				}

				acc.AddFields("dns_query", fields, tags)

				wg.Done()
			}(domain, server)
		}
	}

	wg.Wait()
	return nil
}

func (d *DNSQuery) setDefaultValues() {
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

func (d *DNSQuery) getDNSQueryTime(domain string, server string) (float64, int, error) {
	dnsQueryTime := float64(0)

	c := new(dns.Client)
	c.ReadTimeout = time.Duration(d.Timeout) * time.Second
	c.Net = d.Network

	m := new(dns.Msg)
	recordType, err := d.parseRecordType()
	if err != nil {
		return dnsQueryTime, -1, err
	}
	m.SetQuestion(dns.Fqdn(domain), recordType)
	m.RecursionDesired = true

	r, rtt, err := c.Exchange(m, net.JoinHostPort(server, strconv.Itoa(d.Port)))
	if err != nil {
		return dnsQueryTime, -1, err
	}
	if r.Rcode != dns.RcodeSuccess {
		return dnsQueryTime, r.Rcode, fmt.Errorf("Invalid answer (%s) from %s after %s query for %s", dns.RcodeToString[r.Rcode], server, d.RecordType, domain)
	}
	dnsQueryTime = float64(rtt.Nanoseconds()) / 1e6
	return dnsQueryTime, r.Rcode, nil
}

func (d *DNSQuery) parseRecordType() (uint16, error) {
	var recordType uint16
	var err error

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
		err = fmt.Errorf("Record type %s not recognized", d.RecordType)
	}

	return recordType, err
}

func setResult(result ResultType, fields map[string]interface{}, tags map[string]string) {
	var tag string
	switch result {
	case Success:
		tag = "success"
	case Timeout:
		tag = "timeout"
	case Error:
		tag = "error"
	}

	tags["result"] = tag
	fields["result_code"] = uint64(result)
}

func init() {
	inputs.Add("dns_query", func() telegraf.Input {
		return &DNSQuery{}
	})
}
