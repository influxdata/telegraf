// +build linux

package dns_activity

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/miekg/dns"
	"github.com/miekg/pcap"
	"log"
	"os"
	"strings"
	"sync"
)

type DNS_Activity struct {
	Port   int
	Device string
	handle *pcap.Pcap
}

var (
	snaplen    = 65536
	countermap = struct {
		sync.RWMutex
		m map[string]int
	}{m: make(map[string]int)}
	bytesmap = struct {
		sync.RWMutex
		m map[string]int
	}{m: make(map[string]int)}
	errormap = struct {
		sync.RWMutex
		m map[string]int
	}{m: make(map[string]int)}
)

func (s *DNS_Activity) Description() string {
	return "Gathers statistics on the type of DNS queries being made"
}

const sampleConfig = `
  ## port to listen for DNS queries on
  port = 53
  ## interface to capture DNS queries on. Leave blank for all
  device = "eth0"
  `

func (s *DNS_Activity) SampleConfig() string {
	return sampleConfig
}

func (s *DNS_Activity) Gather(acc telegraf.Accumulator) error {
	countersnapshot := make(map[string]int)
	bytessnapshot := make(map[string]int)
	errorsnapshot := make(map[string]int)
	countermap.Lock()
	bytesmap.Lock()
	errormap.Lock()
	for k, v := range countermap.m {
		countersnapshot[k] = v
	}
	countermap.m = make(map[string]int)

	for k, v := range bytesmap.m {
		bytessnapshot[k] = v
	}
	bytesmap.m = make(map[string]int)

	for k, v := range errormap.m {
		errorsnapshot[k] = v
	}
	errormap.m = make(map[string]int)

	countermap.Unlock()
	bytesmap.Unlock()
	errormap.Unlock()
	var record string

	for record = range countersnapshot {
		tags := map[string]string{"query_type": record}
		acc.AddFields("dns_activity_type", map[string]interface{}{"count": countersnapshot[record], "bytes": bytessnapshot[record]}, tags)
	}
	for record = range errorsnapshot {
		tags := map[string]string{"error_type": record}
		acc.AddFields("dns_activity_error", map[string]interface{}{"count": errorsnapshot[record]}, tags)
	}

	return nil
}

func (d *DNS_Activity) Start(_ telegraf.Accumulator) error {
	expr := fmt.Sprintf("port %d", d.Port)
	device := d.Device
	if device == "" {
		devs, err := pcap.FindAllDevs()
		if err != nil {
			fmt.Fprintln(os.Stderr, "tcpdump: couldn't find any devices:", err)
			return err
		}
		if len(devs) == 0 {
			fmt.Fprintln(os.Stderr, "tcpdump: couldn't find any devices")
			return errors.New("tcpdump: couldn't find any devices")
		}

		device = devs[0].Name

	}
	fmt.Printf("Using device: %v\n", device)
	handle, err := pcap.OpenLive(device, int32(snaplen), true, 500)
	if handle == nil {
		log.Fatal(fmt.Sprintf("tcpdump fatal error: %s", err))
		return err
	}

	ferr := handle.SetFilter(expr)
	if ferr != nil {
		fmt.Println("Error setting PCAP filter:", ferr)
		return ferr
	}
	d.handle = handle
	go d.captureloop()
	return nil
}

func (d *DNS_Activity) Stop() {
	if d.handle != nil {
		d.handle.Close()
	}
}

func init() {
	inputs.Add("dns_activity", func() telegraf.Input {
		return &DNS_Activity{Port: 53, Device: ""}
	})
}

func (d *DNS_Activity) captureloop() {
	for pkt, r := d.handle.NextEx(); r >= 0; pkt, r = d.handle.NextEx() {
		if r == 0 {
			// timeout, continue
			continue
		}
		pkt.Decode()
		msg := new(dns.Msg)
		err := msg.Unpack(pkt.Payload)
		// We only want packets which have been successfully unpacked
		// and have at least one answer, or NXDOMAIN
		if err == nil && (len(msg.Answer) > 0 || msg.Rcode != dns.RcodeSuccess) {
			for a := range msg.Answer {
				dnstype := strings.ToUpper(dns.TypeToString[msg.Answer[a].Header().Rrtype])
				countermap.Lock()
				countermap.m[dnstype]++
				countermap.Unlock()
				bytesmap.Lock()
				bytesmap.m[dnstype] += int(msg.Answer[a].Header().Rdlength) // Add up the data values
				bytesmap.Unlock()
			}
			errormap.Lock()
			errormap.m[dns.RcodeToString[msg.Rcode]]++
			errormap.Unlock()
		}
	}
}
