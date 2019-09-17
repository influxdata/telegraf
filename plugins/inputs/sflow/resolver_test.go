package sflow

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func Test_goodDNSProcessor(t *testing.T) {
	defer testEmptyLog(t)()

	str := map[string]string{
		"fi-es-he6-z4-e02-qfx03-dev.netdevice.nesc.nokia.net":     "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
		"fi-es-he6-z4-e02-qfx03-dev-em0-0.transit.nesc.nokia.net": "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
		"fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net":               "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
	}

	p := newDNSProcessor(`(.*)(?:(?:-e.[0-9]-[0-9]\.transit)|(?:\.netdevice))(.*)`)

	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor(`s/(.*)(?:(?:-e.[0-9]-[0-9]\.transit)|(?:\.netdevice))(.*)/$1$2`)
	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor("")
	for k := range str {
		transformed := p.transform(k)
		if transformed != k {
			t.Fatalf("actual %s != expected %s", transformed, k)
		}
	}
}

func Test_badDNSProcessor(t *testing.T) {
	defer testEmptyLog(t)()

	str := map[string]string{
		"fi-es-he6-z4-e02-qfx03-dev.netdevice.nesc.nokia.net":     "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
		"fi-es-he6-z4-e02-qfx03-dev-em0-0.transit.nesc.nokia.net": "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
	}

	p := newDNSProcessor(`(.*)`)

	for k, v := range str {
		transformed := p.transform(k)
		if transformed == v {
			t.Fatalf("actual %s == expected %s and that is a surprise", transformed, v)
		}
	}
}

func Test_deanDNSProcessor(t *testing.T) {
	defer testEmptyLog(t)()

	str := map[string]string{
		"192.168.0.49": "deans-laptop",
		"192.168.0.50": "192.168.0.50",
	}

	p := newDNSProcessor(`s/192.168.0.49/deans-laptop`)
	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor("")
	for k := range str {
		transformed := p.transform(k)
		if transformed != k {
			t.Fatalf("actual %s != expected %s", transformed, k)
		}
	}
}

func Test_exampleDNSProcessor(t *testing.T) {
	defer testEmptyLog(t)()

	str := map[string]string{
		"hostx-net1": "hostx",
		"hostx-net2": "hostx",
		"hostx":      "hostx",
	}

	p := newDNSProcessor(`s/(.*)(?:-net[0-9])/$1`)
	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor("")
	for k := range str {
		transformed := p.transform(k)
		if transformed != k {
			t.Fatalf("actual %s != expected %s", transformed, k)
		}
	}
}

// dnsResolveTest is a helper for the dns resolution tests below
func dnsResolveTest(t *testing.T, srcTagName, srcTagValue, resolvedValue, dstTagName string) {
	defer testEmptyLog(t)()

	// Create a resole and replace its lowest level dns lookup function with something we can
	// control via channel reads and writes
	resolver := newAsyncResolver(true, "", true, "")
	resolver.start(time.Duration(30)*time.Second, time.Duration(30)*time.Second)
	asyncResolver, ok := resolver.(*asyncResolver)
	if !ok {
		t.Errorf("resolve not an asyncResolver but a %T", asyncResolver)
	}
	ipToResolveCh := make(chan string)
	resolvedNameCh := make(chan string)
	asyncResolver.ipToFqdnFn = func(ip string) string {
		ipToResolveCh <- ip
		result := <-resolvedNameCh
		return result
	}

	// channel and goroutine to support metrics to resolve on a channel and results on another channel
	metricToResolveCh := make(chan telegraf.Metric)
	resolvedMetricCh := make(chan telegraf.Metric)
	go func() {
		for toResolve := range metricToResolveCh {
			resolver.resolve(toResolve, func(m telegraf.Metric) {
				resolvedMetricCh <- m
			})
		}
	}()

	// srcTagName of srcTagValue should resolve to resolvedValue via async lookup (go routine / channel) use of our replacement ipToFqdnFn
	lpIn, _ := metric.New("metric", map[string]string{srcTagName: srcTagValue}, nil, time.Now())
	metricToResolveCh <- lpIn
	ip := <-ipToResolveCh
	if ip != srcTagValue {
		t.Errorf("ip to resolve != %s but %s", srcTagValue, ip)
	}
	resolvedNameCh <- resolvedValue
	lpOut := <-resolvedMetricCh
	aip, ok := lpOut.GetTag(dstTagName)
	if ok {
		if aip != resolvedValue {
			t.Errorf("%s != %s but %s", dstTagName, resolvedValue, aip)
		}
	} else {
		t.Errorf("not ok getting %s tag", dstTagName)
	}

	// srcTagName of srcTagValue should resolve to resolvedValue from cache rather than lookup (go routine / channel) use of our replacement ipToFqdnFn
	lpIn, _ = metric.New("metric", map[string]string{srcTagName: srcTagValue}, nil, time.Now())
	metricToResolveCh <- lpIn
	lpOut = <-resolvedMetricCh
	aip, ok = lpOut.GetTag(dstTagName)
	if ok {
		if aip != resolvedValue {
			t.Errorf("%s != %s but %s", dstTagName, resolvedValue, aip)
		}
	} else {
		t.Error("not ok getting host tag")
	}

	// if we clear the cache then it should resolve to a lookup from via our replacement fn
	asyncResolver.dnsCache.clear()
	lpIn, _ = metric.New("metric", map[string]string{srcTagName: srcTagValue}, nil, time.Now())
	metricToResolveCh <- lpIn
	ip = <-ipToResolveCh
	if ip != srcTagValue {
		t.Errorf("ip to resolve != %s but %s", resolvedValue, ip)
	}
	resolvedNameCh <- resolvedValue
	lpOut = <-resolvedMetricCh
	aip, ok = lpOut.GetTag(dstTagName)
	if ok {
		if aip != resolvedValue {
			t.Errorf("%s != %s but %s", dstTagName, resolvedValue, aip)
		}
	} else {
		t.Errorf("not ok getting %s tag", dstTagName)
	}
}

// ifaceResolveTest is a helper for the interface name resolution tests below
func ifaceResolveTest(t *testing.T, srcTagName, srcTagValue, resolvedValue, dstTagName string) {
	defer testEmptyLog(t)()

	// Create a resole and replace its lowest level dns lookup function with something we can
	// control via channel reads and writes
	resolver := newAsyncResolver(true, "", true, "")
	resolver.start(time.Duration(30)*time.Second, time.Duration(30)*time.Second)
	asyncResolver, ok := resolver.(*asyncResolver)
	if !ok {
		t.Errorf("resolve not an asyncResolver but a %T", asyncResolver)
	}
	indexToResolveCh := make(chan string)
	resolvedNameCh := make(chan string)
	asyncResolver.ifIndexToIfNameFn = func(id uint64, _ string, _ string, index string) string {
		indexToResolveCh <- index
		result := <-resolvedNameCh
		return result
	}
	// need to put this in to stop it trying to resolve the agent_ip
	asyncResolver.ipToFqdnFn = func(ip string) string {
		return ip
	}

	// channel and goroutine to support metrics to resolve on a channel and results on another channel
	metricToResolveCh := make(chan telegraf.Metric)
	resolvedMetricCh := make(chan telegraf.Metric)
	go func() {
		for toResolve := range metricToResolveCh {
			resolver.resolve(toResolve, func(m telegraf.Metric) {
				resolvedMetricCh <- m
			})
		}
	}()

	// srcTagName of srcTagValue should resolve to resolvedValue via async lookup (go routine / channel) use of our replacement ipToFqdnFn
	lpIn, _ := metric.New("metric", map[string]string{srcTagName: srcTagValue, "agent_ip": "192.168.0.1"}, nil, time.Now())
	metricToResolveCh <- lpIn
	index := <-indexToResolveCh
	if index != srcTagValue {
		t.Errorf("index to resolve != %s but %s", srcTagValue, index)
	}
	resolvedNameCh <- resolvedValue
	lpOut := <-resolvedMetricCh
	name, ok := lpOut.GetTag(dstTagName)
	if ok {
		if name != resolvedValue {
			t.Errorf("%s != %s but %s", dstTagName, resolvedValue, name)
		}
	} else {
		t.Errorf("not ok getting %s tag", dstTagName)
	}

	// srcTagName of srcTagValue should resolve to resolvedValue from cache rather than lookup (go routine / channel) use of our replacement ipToFqdnFn
	lpIn, _ = metric.New("metric", map[string]string{srcTagName: srcTagValue, "agent_ip": "192.168.0.1"}, nil, time.Now())
	metricToResolveCh <- lpIn
	lpOut = <-resolvedMetricCh
	name, ok = lpOut.GetTag(dstTagName)
	if ok {
		if name != resolvedValue {
			t.Errorf("%s != %s but %s", dstTagName, resolvedValue, name)
		}
	} else {
		t.Error("not ok getting host tag")
	}

	// if we clear the cache then it should resolve to a lookup from via our replacement fn
	asyncResolver.ifaceCache.clear()
	lpIn, _ = metric.New("metric", map[string]string{srcTagName: srcTagValue, "agent_ip": "192.168.0.1"}, nil, time.Now())
	metricToResolveCh <- lpIn
	name = <-indexToResolveCh
	if name != srcTagValue {
		t.Errorf("ip to resolve != %s but %s", resolvedValue, name)
	}
	resolvedNameCh <- resolvedValue
	lpOut = <-resolvedMetricCh
	name, ok = lpOut.GetTag(dstTagName)
	if ok {
		if name != resolvedValue {
			t.Errorf("%s != %s but %s", dstTagName, resolvedValue, name)
		}
	} else {
		t.Errorf("not ok getting %s tag", dstTagName)
	}
}

func Test_agent_ip_to_host_resolve(t *testing.T) {
	defer testEmptyLog(t)()
	dnsResolveTest(t, "agent_ip", "192.168.0.1", "localhost", "host")
}

func Test_src_ip_to_src_host_resolve(t *testing.T) {
	defer testEmptyLog(t)()
	dnsResolveTest(t, "src_ip", "192.168.0.1", "localhost", "src_host")
}

func Test_dst_ip_to_dst_host_resolve(t *testing.T) {
	defer testEmptyLog(t)()
	dnsResolveTest(t, "dst_ip", "192.168.0.1", "localhost", "dst_host")
}

func Test_source_id_index_to_source_id_name_resolve(t *testing.T) {
	defer testEmptyLog(t)()
	ifaceResolveTest(t, "source_id_index", "5", "eth0", "source_id_name")
}

func Test_netif_index_out_to_netif_name_out_resolve(t *testing.T) {
	defer testEmptyLog(t)()
	ifaceResolveTest(t, "netif_index_out", "6", "eth1", "netif_name_out")
}

func Test_netif_index_in_to_netif_name_in_resolve(t *testing.T) {
	defer testEmptyLog(t)()
	ifaceResolveTest(t, "netif_index_in", "7", "eth2", "netif_name_in")
}
