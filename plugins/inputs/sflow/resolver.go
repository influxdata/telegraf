package sflow

import (
	"context"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
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
	start()
	resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric))
	stop()
}

// asyncJob is a function type to be placed into a worker channel
type asyncJob func() error

type asyncResolver struct {
	dns               bool
	snmpIfaces        bool
	snmpCommunity     string
	dnsCache          *cache
	ifaceCache        *cache
	dnsTTLTicker      *time.Ticker
	ifaceTTLTicker    *time.Ticker
	fnWorkerChannel   chan asyncJob
	dnsp              *dnsProcessor
	ipToFqdnFn        func(ip string) string
	ifIndexToIfNameFn func(community string, snmpAgentIP string, ifIndex string) string
	snmpTTL           time.Duration
	dnsTTL            time.Duration
}

func newAsyncResolver(dnsResolve bool, dnsTTL time.Duration, dnsMultiProcessor string, snmpResolve bool, snmpTTL time.Duration, snmpCommunity string) resolver {
	log.Printf("I! [inputs.sflow] dns cache = %t", dnsResolve)
	log.Printf("I! [inputs.sflow] dbs cache ttl = %d\n", dnsTTL)
	log.Printf("I! [inputs.sflow] snmp cache = %t", snmpResolve)
	log.Printf("I! [inputs.sflow] snmp community = %s", snmpCommunity)
	log.Printf("I! [inputs.sflow] snmp cache ttl = %d\n", snmpTTL)
	return &asyncResolver{
		dns:               dnsResolve,
		snmpIfaces:        snmpResolve,
		snmpCommunity:     snmpCommunity,
		dnsCache:          newCache(),
		ifaceCache:        newCache(),
		dnsp:              newDNSProcessor(dnsMultiProcessor),
		ipToFqdnFn:        ipToFqdn,
		snmpTTL:           snmpTTL,
		dnsTTL:            dnsTTL,
		ifIndexToIfNameFn: ifIndexToIfName,
	}
}

func (r *asyncResolver) resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric)) {
	dnsToResolve := map[string]string{
		"agent_ip": "host",
		"src_ip":   "src_host",
		"dst_ip":   "dst_host",
	}
	ifaceToResolve := map[string]string{
		"source_id_index": "source_id_name",
		"netif_index_out": "netif_name_out",
		"netif_index_in":  "netif_name_in",
	}
	agentIP, _ := m.GetTag("agent_ip")
	dnsCompletelyResolved := r.resolveDNSFromCache(m, dnsToResolve)
	ifaceCompletelyResolved := r.resolveIFaceFromCache(agentIP, m, ifaceToResolve)
	if dnsCompletelyResolved && ifaceCompletelyResolved {
		onResolveFn(m)
	} else {
		// place this function into the channel for the async worker to process
		r.fnWorkerChannel <- func() error {
			r.resolveAsyncDNS(m, dnsToResolve)
			r.resolveAsyncIFace(agentIP, m, ifaceToResolve)
			onResolveFn(m)
			return nil
		}
	}
}

func (r *asyncResolver) resolveDNSFromCache(m telegraf.Metric, tags map[string]string) bool {
	if !r.dns {
		return true
	}
	result := true
	for k, v := range tags {
		tagValue, ok := m.GetTag(k)
		if ok {
			located := r.dnsCache.get(tagValue)
			if located != "" {
				m.AddTag(v, located)
			} else {
				result = false
			}
		}
	}
	return result
}

func (r *asyncResolver) resolveAsyncDNS(m telegraf.Metric, tags map[string]string) {
	for k, v := range tags {
		tagValue, ok := m.GetTag(k)
		if ok {
			r.resolveDNS(tagValue, func(fqdn string) {
				m.AddTag(v, fqdn)
			})
		}
	}
}

func (r *asyncResolver) resolveIFaceFromCache(agentIP string, m telegraf.Metric, tags map[string]string) bool {
	if !r.snmpIfaces {
		return true
	}
	result := true
	for srcTag, dstTag := range tags {
		srcTagValue, ok := m.GetTag(srcTag)
		if ok {
			keyToLookupInCache := fmt.Sprintf("%s-%s", agentIP, srcTagValue)
			located := r.ifaceCache.get(keyToLookupInCache)
			if located != "" {
				m.AddTag(dstTag, located)
			} else {
				result = false
			}
		}
	}
	return result
}

func (r *asyncResolver) resolveAsyncIFace(agentIP string, m telegraf.Metric, tags map[string]string) {
	for srcTag, dstTag := range tags {
		srcTagValue, ok := m.GetTag(srcTag)
		if ok {
			r.resolveIFace(srcTagValue, agentIP, func(name string) {
				m.AddTag(dstTag, name)
			})
		}
	}
}

func (r *asyncResolver) start() {
	if r.dnsTTL != 0 {
		r.dnsTTLTicker = time.NewTicker(r.dnsTTL)
		go func() {
			for range r.dnsTTLTicker.C {
				log.Println("I! [inputs.sflow] clearing DNS cache")
				r.dnsCache.clear()
			}
		}()
	}
	if r.snmpTTL != 0 {
		r.ifaceTTLTicker = time.NewTicker(r.snmpTTL)
		go func() {
			for range r.ifaceTTLTicker.C {
				log.Println("I! [inputs.sflow] clearing IFace cache")
				r.ifaceCache.clear()
			}
		}()
	}

	// our worker goroutine just takes a function from the worker channel and executes it
	r.fnWorkerChannel = make(chan asyncJob)
	go func() {
		for {
			fn := <-r.fnWorkerChannel
			if fn() != nil {
				return // terminates the goroutine if the function pulled from the channel returns an error
			}
		}
	}()
}

func (r *asyncResolver) stop() {
	r.fnWorkerChannel <- func() error {
		return fmt.Errorf("Stop")
	}
	if r.dnsTTLTicker != nil {
		r.dnsTTLTicker.Stop()
	}
	if r.ifaceTTLTicker != nil {
		r.ifaceTTLTicker.Stop()
	}
}

func (r *asyncResolver) resolveDNS(ipAddress string, resolved func(fqdn string)) {
	fqdn := r.dnsCache.get(ipAddress)
	if fqdn != "" {
		log.Printf("D! [input.sflow] sync cache lookup %s=>%s", ipAddress, fqdn)
	} else {
		name := r.ipToFqdnFn(ipAddress)
		log.Printf("D! [input.sflow] async resolve of %s=>%s", ipAddress, name)
		fqdn = r.dnsp.transform(name)
		if fqdn != name {
			log.Printf("D! [input.sflow] transformed dns[0] %s=>%s", name, fqdn)
		}
		r.dnsCache.set(ipAddress, fqdn)
	}
	resolved(fqdn)
}

func ipToFqdn(ipAddress string) string {
	ctx, cancel := context.WithTimeout(context.TODO(), 10000*time.Millisecond)
	defer cancel()
	resolver := net.Resolver{}
	names, err := resolver.LookupAddr(ctx, ipAddress)
	fqdn := ipAddress
	if err == nil {
		if len(names) != 0 {
			fqdn = names[0]
		}
	} else {
		log.Printf("W! [input.sflow] dns lookup of %s resulted in error %s", ipAddress, err)
	}
	return fqdn
}

func (r *asyncResolver) resolveIFace(ifaceIndex string, agentIP string, resolved func(fqdn string)) {
	name := r.ifaceCache.get(fmt.Sprintf("%s-%s", agentIP, ifaceIndex))
	if name != "" {
		log.Printf("D! [input.sflow] sync cache lookup (%s,%s)=>%s", agentIP, ifaceIndex, name)
	} else {
		// look it up
		name = r.ifIndexToIfNameFn(r.snmpCommunity, agentIP, ifaceIndex)
		log.Printf("D! [input.sflow] async resolve of (%s,%s)=>%s", agentIP, ifaceIndex, name)
		r.ifaceCache.set(fmt.Sprintf("%s-%s", agentIP, ifaceIndex), name)
	}
	resolved(name)
}

func ifIndexToIfName(community string, snmpAgentIP string, ifIndex string) string {
	oid := "1.3.6.1.2.1.31.1.1.1.1"
	gosnmp.Default.Target = snmpAgentIP
	if community != "" {
		gosnmp.Default.Community = community
	}
	gosnmp.Default.Timeout = 2 * time.Second
	gosnmp.Default.Retries = 5
	err := gosnmp.Default.Connect()
	if err != nil {
		log.Println("W! [inputs.sflow] err on snmp.Connect", err)
		return ifIndex
	}
	defer gosnmp.Default.Conn.Close()
	result, found := ifIndex, false
	pduNameToFind := fmt.Sprintf(".%s.%s", oid, ifIndex)
	err = gosnmp.Default.BulkWalk(oid, func(pdu gosnmp.SnmpPDU) error {
		switch pdu.Type {
		case gosnmp.OctetString:
			b := pdu.Value.([]byte)
			if pdu.Name == pduNameToFind {
				log.Printf("D! [inputs.sflow] snmp bulk walk (%s) found %s as %s\n", snmpAgentIP, pdu.Name, string(b))
				found = true
				result = string(b)
			}
		default:
		}
		return nil
	})
	if err != nil {
		log.Printf("W! inputs.sflow] unable to find %s in smmp results due to error %s\n", pduNameToFind, err)
	} else {
		if !found {
			log.Printf("W! [inputs.sflow] unable to find %s in smmp results\n", pduNameToFind)
		} else {
			log.Printf("D! [inputs.sflow] found %s in snmp results as %s\n", pduNameToFind, result)
		}
	}
	return result
}

// dnsProcessor is capable of taking a processing instruction to convert an input DNS name to an alternative name
type dnsProcessor struct {
	rePattern *regexp.Regexp
	template  string
}

//_ := `s/(.*)(?:(?:-e.[0-9]-[0-9]\.transit)|(?:\.netdevice))(.*)/$1$2`
// if starts with s/ then look for trailing / and this is the separation of regexp and tremplate
// if no trailing / then error
// if no start with s/ then consider it just to be the regexp and a default template of $1$2$3$4$5 will be used
func newDNSProcessor(processString string) *dnsProcessor {
	if processString == "" {
		return &dnsProcessor{}
	}
	re := ""
	template := ""
	loc := strings.Index(processString, "s/")
	endLoc := strings.LastIndex(processString, "/")
	if loc == 0 && endLoc > (loc+1) {
		re = processString[loc+2 : endLoc]
		template = processString[endLoc+1:]
	} else {
		re = processString
		template = "$1$2$3$4$5"
	}

	return &dnsProcessor{rePattern: regexp.MustCompile(re), template: template}
}

func (p *dnsProcessor) transform(name string) string {
	if p.rePattern == nil {
		return name
	}
	result := []byte{}
	// For each match of the regex in the content.
	expanded := false
	for _, submatches := range p.rePattern.FindAllStringSubmatchIndex(name, -1) {
		// Apply the captured submatches to the template and append the output
		// to the result.
		//fmt.Println(i, submatches)
		result = p.rePattern.ExpandString(result, p.template, name, submatches)
		expanded = true
	}
	if !expanded {
		return name
	}
	return string(result)
}
