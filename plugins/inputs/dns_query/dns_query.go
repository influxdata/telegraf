//go:generate ../../../tools/readme_config_includer/generator
package dns_query

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type ResultType uint64

const (
	Success ResultType = iota
	Timeout
	Error
)

type DNSQuery struct {
	Domains       []string        `toml:"domains"`
	Network       string          `toml:"network"`
	Servers       []string        `toml:"servers"`
	RecordType    string          `toml:"record_type"`
	Port          int             `toml:"port"`
	Timeout       config.Duration `toml:"timeout"`
	IncludeFields []string        `toml:"include_fields"`

	fieldEnabled map[string]bool
}

func (*DNSQuery) SampleConfig() string {
	return sampleConfig
}

func (d *DNSQuery) Init() error {
	// Convert the included fields into a lookup-table
	d.fieldEnabled = make(map[string]bool, len(d.IncludeFields))
	for _, f := range d.IncludeFields {
		switch f {
		case "first_ip", "all_ips":
		default:
			return fmt.Errorf("invalid field %q included", f)
		}
		d.fieldEnabled[f] = true
	}

	// Set defaults
	if d.Network == "" {
		d.Network = "udp"
	}

	if d.RecordType == "" {
		d.RecordType = "NS"
	}

	if len(d.Domains) == 0 {
		d.Domains = []string{"."}
		d.RecordType = "NS"
	}

	if d.Port < 1 {
		d.Port = 53
	}

	return nil
}

func (d *DNSQuery) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, domain := range d.Domains {
		for _, server := range d.Servers {
			wg.Add(1)
			go func(domain, server string) {
				defer wg.Done()

				fields, tags, err := d.query(domain, server)
				if err != nil {
					var opErr *net.OpError
					if !errors.As(err, &opErr) || !opErr.Timeout() {
						acc.AddError(err)
					}
				}
				acc.AddFields("dns_query", fields, tags)
			}(domain, server)
		}
	}
	wg.Wait()

	return nil
}

func (d *DNSQuery) query(domain string, server string) (map[string]interface{}, map[string]string, error) {
	tags := map[string]string{
		"server":      server,
		"domain":      domain,
		"record_type": d.RecordType,
		"result":      "error",
	}

	fields := map[string]interface{}{
		"query_time_ms": float64(0),
		"result_code":   uint64(Error),
	}

	c := dns.Client{
		ReadTimeout: time.Duration(d.Timeout),
		Net:         d.Network,
	}

	recordType, err := d.parseRecordType()
	if err != nil {
		return fields, tags, err
	}

	var msg dns.Msg
	msg.SetQuestion(dns.Fqdn(domain), recordType)
	msg.RecursionDesired = true

	addr := net.JoinHostPort(server, strconv.Itoa(d.Port))
	r, rtt, err := c.Exchange(&msg, addr)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) && opErr.Timeout() {
			tags["result"] = "timeout"
			fields["result_code"] = uint64(Timeout)
			return fields, tags, err
		}
		return fields, tags, err
	}

	// Fill valid fields
	tags["rcode"] = dns.RcodeToString[r.Rcode]
	fields["rcode_value"] = r.Rcode
	fields["query_time_ms"] = float64(rtt.Nanoseconds()) / 1e6

	// Handle the failure case
	if r.Rcode != dns.RcodeSuccess {
		return fields, tags, fmt.Errorf("invalid answer (%s) from %s after %s query for %s", dns.RcodeToString[r.Rcode], server, d.RecordType, domain)
	}

	// Success
	tags["result"] = "success"
	fields["result_code"] = uint64(Success)

	if d.fieldEnabled["first_ip"] {
		for _, record := range r.Answer {
			if ip, found := extractIP(record); found {
				fields["ip"] = ip
				break
			}
		}
	}
	if d.fieldEnabled["all_ips"] {
		for i, record := range r.Answer {
			if ip, found := extractIP(record); found {
				fields["ip_"+strconv.Itoa(i)] = ip
			}
		}
	}

	return fields, tags, nil
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
		err = fmt.Errorf("record type %s not recognized", d.RecordType)
	}

	return recordType, err
}

func extractIP(record dns.RR) (string, bool) {
	if r, ok := record.(*dns.A); ok {
		return r.A.String(), true
	}
	if r, ok := record.(*dns.AAAA); ok {
		return r.AAAA.String(), true
	}
	return "", false
}

func init() {
	inputs.Add("dns_query", func() telegraf.Input {
		return &DNSQuery{
			Timeout: config.Duration(2 * time.Second),
		}
	})
}
