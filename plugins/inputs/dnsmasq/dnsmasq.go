package dnsmasq

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miekg/dns"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ResultType uint64

const (
	Success ResultType = 0
	Timeout            = 1
	Error              = 2
)

type Dnsmasq struct {
	c *dns.Client

	// Dnsmasq server IP address
	Server string

	// Dnsmasq server port
	Port int
}

var sampleConfig = `
  ## Dnsmasq server IP address.
  # server = "127.0.0.1"
  #
  ## Dnsmasq server port.
  # port = 53
`

func (d *Dnsmasq) SampleConfig() string {
	return sampleConfig
}

func (d *Dnsmasq) Description() string {
	return "Read Dnsmasq metrics by dns query"
}

func (d *Dnsmasq) Gather(acc telegraf.Accumulator) error {
	d.setDefaultValues()
	fields := make(map[string]interface{}, 2)
	tags := map[string]string{
		"server": d.Server,
		"port":   fmt.Sprint(d.Port),
	}
	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{
			question("cachesize.bind."),
			question("insertions.bind."),
			question("evictions.bind."),
			question("misses.bind."),
			question("hits.bind."),
			question("auth.bind."),
			question("servers.bind."),
		},
	}
	in, _, err := d.c.Exchange(msg, d.Server)
	if err != nil {
		return err
	}
	for _, a := range in.Answer {
		txt, ok := a.(*dns.TXT)
		if !ok {
			continue
		}
		switch txt.Hdr.Name {
		case "servers.bind.":
			for _, str := range txt.Txt {
				arr := strings.Fields(str)
				if got, want := len(arr), 3; got != want {
					return fmt.Errorf("stats DNS record servers.bind.: unexpeced number of argument in record: got %d, want %d", got, want)
				}
				queries, err := strconv.ParseFloat(arr[1], 64)
				if err != nil {
					return err
				}
				failedQueries, err := strconv.ParseFloat(arr[2], 64)
				if err != nil {
					return err
				}
				fields["queries"] = queries
				fields["queries_failed"] = failedQueries
			}
		default:
			if got, want := len(txt.Txt), 1; got != want {
				return fmt.Errorf("stats DNS record %q: unexpected number of replies: got %d, want %d", txt.Hdr.Name, got, want)
			}
			f, err := strconv.ParseFloat(txt.Txt[0], 64)
			if err != nil {
				return err
			}
			names := strings.Split(txt.Hdr.Name, ".")
			if len(names) > 0 {
				fields[names[0]] = f
			}
		}
	}

	acc.AddFields("dnsmasq", fields, tags)
	return nil
}

func (d *Dnsmasq) setDefaultValues() {
	if d.Server == "" {
		d.Server = "127.0.0.1"
	}
	if d.Port == 0 {
		d.Port = 53
	}
}

func question(name string) dns.Question {
	return dns.Question{
		Name:   name,
		Qtype:  dns.TypeTXT,
		Qclass: dns.ClassCHAOS,
	}
}

func init() {
	dnsClient := &dns.Client{
		SingleInflight: true,
	}
	inputs.Add("dnsmasq", func() telegraf.Input {
		return &Dnsmasq{c: dnsClient}
	})
}
