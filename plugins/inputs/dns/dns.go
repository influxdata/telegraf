package dns

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/miekg/dns"
	"net"
	"strconv"
	"time"
)

type Dns struct {
	// Domains or subdomains to query
	Domains []string

	// Server to query
	Servers []string

	// Record type
	RecordType string

	// DNS server port number
	Port int

	// Dns query timeout in seconds. 0 means no timeout
	Timeout int
}

var sampleConfig = `
  ### Domains or subdomains to query
  domains = ["mjasion.pl"] # required

  ### servers to query
  servers = ["8.8.8.8"] # required

  ### Query record type. Posible values: A, CNAME, MX, TXT, NS. Default is "A"
  recordType = "A" # optional

  ### Dns server port. 53 is default
  port = 53 # optional

  ### Query timeout in seconds. Default is 2 seconds
  timeout = 2 # optional
`

func (d *Dns) SampleConfig() string {
	return sampleConfig
}

func (d *Dns) Description() string {
	return "Query given DNS server and gives statistics"
}
func (d *Dns) Gather(acc telegraf.Accumulator) error {
	d.setDefaultValues()
	for _, domain := range d.Domains {
		for _, server := range d.Servers {
			dnsQueryTime, err := d.getDnsQueryTime(domain, server)
			if err != nil {
				return err
			}
			tags := map[string]string{
				"server":     server,
				"domain":     domain,
				"recordType": d.RecordType,
			}

			acc.Add("dns", dnsQueryTime, tags)
		}
	}

	return nil
}

func (d *Dns) setDefaultValues() {
	if len(d.RecordType) == 0 {
		d.RecordType = "A"
	}
	if d.Port == 0 {
		d.Port = 53
	}
	if d.Timeout == 0 {
		d.Timeout = 2
	}
}

func (d *Dns) getDnsQueryTime(domain string, server string) (float64, error) {
	dnsQueryTime := float64(0)

	c := new(dns.Client)
	c.ReadTimeout = time.Duration(d.Timeout) * time.Second

	m := new(dns.Msg)
	recordType, err := d.parseRecordType()
	if err != nil {
		return dnsQueryTime, err
	}
	m.SetQuestion(dns.Fqdn(domain), recordType)
	m.RecursionDesired = true

	start_time := time.Now()
	r, _, err := c.Exchange(m, net.JoinHostPort(server, strconv.Itoa(d.Port)))
	queryDuration := time.Since(start_time)

	if err != nil {
		return dnsQueryTime, err
	}
	if r.Rcode != dns.RcodeSuccess {
		return dnsQueryTime, errors.New(fmt.Sprintf("Invalid answer name %s after %s query for %s\n", domain, d.RecordType, domain))
	}

	dnsQueryTime = float64(queryDuration.Nanoseconds()) / 1e6
	return dnsQueryTime, nil
}

func (d *Dns) parseRecordType() (uint16, error) {
	var recordType uint16
	var error error

	switch d.RecordType {
	case "A":
		recordType = dns.TypeA
	case "CNAME":
		recordType = dns.TypeCNAME
	case "MX":
		recordType = dns.TypeMX
	case "NS":
		recordType = dns.TypeNS
	case "TXT":
		recordType = dns.TypeTXT
	default:
		error = errors.New(fmt.Sprintf("Record type %s not recognized", d.RecordType))
	}

	return recordType, error
}

func init() {
	inputs.Add("dns", func() telegraf.Input {
		return &Dns{}
	})
}
