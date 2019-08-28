package sflow

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/soniah/gosnmp"
)

type cache struct {
	data map[string]string
	mux  sync.Mutex
}

func (c *cache) get(key string) string {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.data[key]
}

func (c *cache) set(key string, value string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.data[key] = value
}

func (c *cache) clear() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.data = make(map[string]string)
}

func newCache() *cache {
	result := &cache{data: make(map[string]string)}
	return result

}

type resolver interface {
	resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric))
	start(dnsTTL time.Duration, snmpTTL time.Duration)
	stop()
}

type asyncResolver struct {
	dns             bool
	snmpIfaces      bool
	snmpCommunity   string
	dnsCache        *cache
	ifaceCache      *cache
	dnsTTLTicker    *time.Ticker
	ifaceTTLTicker  *time.Ticker
	fnWorkerChannel chan asyncJob
}

func newAsyncResolver(dnsResolve bool, snmpResolve bool, snmpCommunity string) resolver {
	log.Printf("I! [inputs.sflow] dbs cache = %t", dnsResolve)
	log.Printf("I! [inputs.sflow] snmp cache = %t", snmpResolve)
	log.Printf("I! [inputs.sflow] snmp community = %s", snmpCommunity)
	return &asyncResolver{
		dns:           dnsResolve,
		snmpIfaces:    snmpResolve,
		snmpCommunity: snmpCommunity,
		dnsCache:      newCache(),
		ifaceCache:    newCache(),
	}
}

type onResolve struct {
	done    func()
	counter int
	mux     sync.Mutex
}

func (or *onResolve) decrement() {
	or.mux.Lock()
	defer or.mux.Unlock()
	or.counter--
	if or.counter == 0 {
		or.done()
	}
}

func (or *onResolve) increment() {
	or.mux.Lock()
	defer or.mux.Unlock()
	or.counter++
}

var ops uint64

type asyncJob func()

func (r *asyncResolver) resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric)) {
	or := &onResolve{done: func() { onResolveFn(m) }, counter: 1}
	r.dnsResolve(m, "agent_ip", "host", or)
	r.dnsResolve(m, "src_ip", "src_host", or)
	r.dnsResolve(m, "dst_ip", "dst_host", or)
	agentIP, ok := m.GetTag("agent_ip")
	if ok {
		r.ifaceResolve(m, "source_id_index", "source_id_name", agentIP, or)
		r.ifaceResolve(m, "netif_index_out", "netif_name_out", agentIP, or)
		r.ifaceResolve(m, "netif_index_in", "netif_name_in", agentIP, or)
	}
	or.decrement() // this will do the resolve if there was nothing to resolve
}

func (r *asyncResolver) start(dnsTTL time.Duration, snmpTTL time.Duration) {
	r.dnsTTLTicker = time.NewTicker(dnsTTL)
	go func() {
		for range r.dnsTTLTicker.C {
			log.Println("D! [inputs.sflow] clearing DNS cache")
			r.dnsCache.clear()
		}
	}()
	r.ifaceTTLTicker = time.NewTicker(snmpTTL)
	go func() {
		for range r.ifaceTTLTicker.C {
			log.Println("D! [inputs.sflow] clearing IFace cache")
			r.ifaceCache.clear()
		}
	}()

	r.fnWorkerChannel = make(chan asyncJob)
	go func() {
		for {
			fn := <-r.fnWorkerChannel
			fn()
		}
	}()

	log.Printf("I! [inputs.sflow] dbs cache ttl = %d\n", dnsTTL)
	log.Printf("I! [inputs.sflow] snmp cache ttl = %d\n", snmpTTL)

}

func (r *asyncResolver) stop() {
	r.dnsTTLTicker.Stop()
	r.ifaceTTLTicker.Stop()
}

func (r *asyncResolver) dnsResolve(m telegraf.Metric, srcTag string, dstTag string, or *onResolve) {
	value, ok := m.GetTag(srcTag)
	if r.dns && ok {
		or.increment()
		fn := func() {
			r.resolveDNS(value, func(fqdn string) {
				m.AddTag(dstTag, fqdn)
				or.decrement()
			})
		}
		r.fnWorkerChannel <- fn
	}
}

func (r *asyncResolver) ifaceResolve(m telegraf.Metric, srcTag string, dstTag string, agentIP string, or *onResolve) {
	value, ok := m.GetTag(srcTag)
	if r.snmpIfaces && ok {
		or.increment()
		fn := func() {
			r.resolveIFace(value, agentIP, func(name string) {
				m.AddTag(dstTag, name)
				or.decrement()
			})
		}
		r.fnWorkerChannel <- fn
	}
}

func (r *asyncResolver) resolveDNS(ipAddress string, resolved func(fqdn string)) {
	fqdn := r.dnsCache.get(ipAddress)
	if fqdn != "" {
		log.Printf("D! [input.sflow] sync cache lookup %s=>%s", ipAddress, fqdn)
	} else {
		fqdn = r.ipToFqdn(ipAddress)
		log.Printf("D! [input.sflow] async resolve of %s=>%s", ipAddress, fqdn)
		r.dnsCache.set(ipAddress, fqdn)
	}
	resolved(fqdn)
}

func (r *asyncResolver) ipToFqdn(ipAddress string) string {
	ctx, cancel := context.WithTimeout(context.TODO(), 10000*time.Millisecond)
	defer cancel()
	resolver := net.Resolver{}
	names, err := resolver.LookupAddr(ctx, ipAddress)
	fqdn := ipAddress
	if err == nil {
		if len(names) > 0 {
			fqdn = names[0]
		}
	} else {
		log.Printf("!E [input.sflow] dns lookup of %s resulted in error %s", ipAddress, err)
	}
	return fqdn
}

func (r *asyncResolver) resolveIFace(ifaceIndex string, agentIP string, resolved func(fqdn string)) {
	id := atomic.AddUint64(&ops, 1)
	name := r.ifaceCache.get(fmt.Sprintf("%s-%s", agentIP, ifaceIndex))
	if name != "" {
		log.Printf("D! [input.sflow] %d sync cache lookup (%s,%s)=>%s", id, agentIP, ifaceIndex, name)
	} else {
		// look it up
		name = r.ifIndexToIfName(id, r.snmpCommunity, agentIP, ifaceIndex)
		log.Printf("D! [input.sflow] %d async resolve of (%s,%s)=>%s", id, agentIP, ifaceIndex, name)
		r.ifaceCache.set(fmt.Sprintf("%s-%s", agentIP, ifaceIndex), name)
	}
	resolved(name)
}

// So, Ive established that this wasn't thread safe. Might be I need a differen COnnection object.
var ifIndexToIfNameMux sync.Mutex

func (r *asyncResolver) ifIndexToIfName(id uint64, community string, snmpAgentIP string, ifIndex string) string {
	ifIndexToIfNameMux.Lock()
	defer ifIndexToIfNameMux.Unlock()
	// This doesn't make the most of the fact we look up all interface names but only cache/use one of them :-()
	oid := "1.3.6.1.2.1.31.1.1.1.1"
	gosnmp.Default.Target = snmpAgentIP
	if community != "" {
		gosnmp.Default.Community = community
	}
	gosnmp.Default.Timeout = 20 * time.Second
	gosnmp.Default.Retries = 5
	err := gosnmp.Default.Connect()
	if err != nil {
		log.Println("E! [inputs.sflow] err on snmp.Connect", err)
	}
	defer gosnmp.Default.Conn.Close()
	//ifaceNames := make(map[string]string)
	result, found := ifIndex, false
	pduNameToFind := fmt.Sprintf(".%s.%s", oid, ifIndex)
	err = gosnmp.Default.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		switch pdu.Type {
		case gosnmp.OctetString:
			b := pdu.Value.([]byte)
			if pdu.Name == pduNameToFind {
				log.Printf("D! [inputs.sflow] %d snmp bulk walk (%s) found %s as %s\n", id, snmpAgentIP, pdu.Name, string(b))
				found = true
				result = string(b)
			} else {
				//log.Printf("D! [inputs.sflow] %d snmp bulk walk (%s) found different %s not %s\n", id, snmpAgentIP, pdu.Name, pduNameToFind)
			}
		default:
		}
		return nil
	})
	if err != nil {
		log.Printf("E! inputs.sflow] %d unable to find %s in smmp results due to error %s\n", id, pduNameToFind, err)
	} else {
		if !found {
			log.Printf("D! [inputs.sflow] %d unable to find %s in smmp results\n", id, pduNameToFind)
		} else {
			log.Printf("D! [inputs.sflow] %d found %s in snmp results as %s\n", id, pduNameToFind, result)
		}
	}
	return result
}

type testResolver struct{}

func newTestResolver(dnsResolve bool, snmpResolve bool, snmpCommunity string) resolver {
	log.Printf("I! [inputs.sflow] test resolver in use")
	return &testResolver{}
}

func (r *testResolver) resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric)) {
	if v, ok := m.GetTag("agent_ip"); ok {
		m.AddTag("host", v)
	}
	if v, ok := m.GetTag("src_ip"); ok {
		m.AddTag("src_host", v)
	}
	if v, ok := m.GetTag("dst_ip"); ok {
		m.AddTag("dst_host", v)
	}
	if v, ok := m.GetTag("source_id_index"); ok {
		m.AddTag("source_id_name", v)
	}
	if v, ok := m.GetTag("netif_index_out"); ok {
		m.AddTag("netif_name_out", v)
	}
	if v, ok := m.GetTag("netif_index_in"); ok {
		m.AddTag("netif_name_in", v)
	}
	onResolveFn(m)
}

func (r *testResolver) start(dnsTTL time.Duration, snmpTTL time.Duration) {
	// NOP
}

func (r *testResolver) stop() {
	// NOP
}
