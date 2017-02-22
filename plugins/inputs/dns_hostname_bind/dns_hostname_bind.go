// Package dns_hostname_bind does collect dns timing and the returned hostname
package dns_hostname_bind

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DnsHostnameBind holds configuration data for
type DnsHostnameBind struct {
	// Server to query
	Servers []string

	// DNS server port number
	Port int

	// Dns query timeout in seconds. 0 means no timeout
	Timeout int
}

var sampleConfig = `
  ## servers to query
  servers = ["192.203.230.10", "2001:500:1::53"] # required

  ## Dns server port. 53 is default
  port = 53 # optional

  ## Query timeout in seconds. Default is 2 seconds
  timeout = 2 # optional
`

func (d *DnsHostnameBind) SampleConfig() string {
	return sampleConfig
}

func (d *DnsHostnameBind) Description() string {
	return "Query given DNS server for hostname.bind in class chaos and gives timing"
}
func (d *DnsHostnameBind) Gather(acc telegraf.Accumulator) error {
	d.setDefaultValues()

	errChan := errchan.New(len(d.Servers))
	for _, server := range d.Servers {
		dnsQueryTime, hostname, err := d.getDnsQueryTime(server)
		errChan.C <- err
		tags := map[string]string{
			"server": server,
		}

		fields := map[string]interface{}{"query_time_ms": dnsQueryTime, "hostname": hostname}
		acc.AddFields("dns_hostname_bind", fields, tags)
	}

	return errChan.Error()
}

func (d *DnsHostnameBind) setDefaultValues() {
	if d.Port == 0 {
		d.Port = 53
	}

	if d.Timeout == 0 {
		d.Timeout = 2
	}
}

func (d *DnsHostnameBind) getDnsQueryTime(server string) (dnsQueryTime float64, hostname string, err error) {
	dnsQueryTime = 0.0
	hostname = ""

	c := new(dns.Client)
	c.ReadTimeout = time.Duration(d.Timeout) * time.Second

	// build dns query
	query := new(dns.Msg)
	query.RecursionDesired = false
	query.AuthenticatedData = false
	query.Question = make([]dns.Question, 1)
	query.Question[0] = dns.Question{Name: "hostname.bind.", Qtype: dns.TypeTXT, Qclass: dns.ClassCHAOS}

	r, rtt, err := c.Exchange(query, net.JoinHostPort(server, strconv.Itoa(d.Port)))
	if err != nil {
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		return dnsQueryTime, hostname, errors.New(fmt.Sprintf("Invalid answer from %s\n", server))
	}
	dnsQueryTime = float64(rtt.Nanoseconds()) / 1e6
	for _, answer := range r.Answer {
		if answer.Header().Rrtype == dns.TypeTXT {
			hostname = strings.Join(answer.(*dns.TXT).Txt, "")
		}
	}
	return dnsQueryTime, hostname, nil
}

func init() {
	inputs.Add("dns_hostname_bind", func() telegraf.Input {
		return &DnsHostnameBind{}
	})
}
