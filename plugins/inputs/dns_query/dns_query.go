//go:generate ../../../tools/readme_config_includer/generator
package dns_query

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"slices"
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

var ignoredErrors = []string{
	"NXDOMAIN",
}

type resultType uint64

const (
	successResult resultType = iota
	timeoutResult
	errorResult
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
	client       *dns.Client
	record       uint16
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

	if len(d.Domains) == 0 {
		d.Domains = []string{"."}
	}

	if d.Port < 1 {
		d.Port = 53
	}

	// Convert the record type
	switch d.RecordType {
	case "":
		d.RecordType = "NS"
	case "A":
		d.record = dns.TypeA
	case "AAAA":
		d.record = dns.TypeAAAA
	case "ANY":
		d.record = dns.TypeANY
	case "CNAME":
		d.record = dns.TypeCNAME
	case "MX":
		d.record = dns.TypeMX
	case "NS":
		d.record = dns.TypeNS
	case "PTR":
		d.record = dns.TypePTR
	case "SOA":
		d.record = dns.TypeSOA
	case "SPF":
		d.record = dns.TypeSPF
	case "SRV":
		d.record = dns.TypeSRV
	case "TXT":
		d.record = dns.TypeTXT
	default:
		return fmt.Errorf("record type %q not recognized", d.RecordType)
	}

	// Create a client for querying
	d.client = &dns.Client{
		ReadTimeout: time.Duration(d.Timeout),
		Net:         d.Network,
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
				if err != nil && !slices.Contains(ignoredErrors, tags["rcode"]) {
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

func (d *DNSQuery) query(domain, server string) (map[string]interface{}, map[string]string, error) {
	tags := map[string]string{
		"server":      server,
		"domain":      domain,
		"record_type": d.RecordType,
		"result":      "error",
	}

	fields := map[string]interface{}{
		"query_time_ms": float64(0),
		"result_code":   uint64(errorResult),
	}

	var msg dns.Msg
	msg.SetQuestion(dns.Fqdn(domain), d.record)
	msg.RecursionDesired = true

	addr := net.JoinHostPort(server, strconv.Itoa(d.Port))
	r, rtt, err := d.client.Exchange(&msg, addr)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) && opErr.Timeout() {
			tags["result"] = "timeout"
			fields["result_code"] = uint64(timeoutResult)
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
	fields["result_code"] = uint64(successResult)

	// Fill out custom fields for specific record types
	for _, record := range r.Answer {
		switch x := record.(type) {
		case *dns.A:
			fields["name"] = x.Hdr.Name
		case *dns.AAAA:
			fields["name"] = x.Hdr.Name
		case *dns.CNAME:
			fields["name"] = x.Hdr.Name
		case *dns.MX:
			fields["name"] = x.Hdr.Name
			fields["preference"] = x.Preference
		case *dns.SOA:
			fields["expire"] = x.Expire
			fields["minttl"] = x.Minttl
			fields["name"] = x.Hdr.Name
			fields["refresh"] = x.Refresh
			fields["retry"] = x.Retry
			fields["serial"] = x.Serial
		}
	}

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
