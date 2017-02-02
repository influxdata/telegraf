// +build linux

package dns_activity

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"net"
	"strconv"
	"testing"
	"time"
)

var testServer = "8.8.8.8"
var testDomain = "google.com"

func TestDNS_Activity_Gather(t *testing.T) {
	var acc testutil.Accumulator

	var dnsactivity = DNS_Activity{
		Port:   53,
		Device: "",
	}
	dnsactivity.Start(&acc)
	c := new(dns.Client)
	c.ReadTimeout = time.Duration(30) * time.Second
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(testDomain), dns.TypeA)
	m.RecursionDesired = true

	_, _, err := c.Exchange(m, net.JoinHostPort(testServer, strconv.Itoa(dnsactivity.Port)))

	assert.NoError(t, err) // Make sure we got a response

	time.Sleep(time.Duration(500) * time.Millisecond) // Pause while the packet trickles through the kernel and libpcap

	dnsactivity.Gather(&acc)

	metric, success := acc.Get("dns_activity_type")

	assert.True(t, success)                           // Do we have the metric
	assert.True(t, metric.Fields["count"].(int) >= 1) // Did we count at least one DNS answer
	assert.True(t, metric.Fields["bytes"].(int) >= 4) // Did we count at least 4 bytes (A response size)

	metric, success = acc.Get("dns_activity_error")

	assert.True(t, success)                           // Do we have the metric
	assert.True(t, metric.Fields["count"].(int) >= 1) // Did we count at least one response

	dnsactivity.Stop()
}
