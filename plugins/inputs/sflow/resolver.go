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
	start(dnsTTL time.Duration, snmpTTL time.Duration)
	resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric))
	stop()
}

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
}

func newAsyncResolver(dnsResolve bool, dnsMultiProcessor string, snmpResolve bool, snmpCommunity string) resolver {
	log.Printf("I! [inputs.sflow] dns cache = %t", dnsResolve)
	log.Printf("I! [inputs.sflow] snmp cache = %t", snmpResolve)
	log.Printf("I! [inputs.sflow] snmp community = %s", snmpCommunity)
	return &asyncResolver{
		dns:               dnsResolve,
		snmpIfaces:        snmpResolve,
		snmpCommunity:     snmpCommunity,
		dnsCache:          newCache(),
		ifaceCache:        newCache(),
		dnsp:              newDNSProcessor(dnsMultiProcessor),
		ipToFqdnFn:        ipToFqdn,
		ifIndexToIfNameFn: ifIndexToIfName,
	}
}

type asyncJob func()

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
	dnsCompletelyResolved := r.resolveDNSFromCache(m, dnsToResolve)
	ifaceCompletelyResolved := r.resolveIFaceFromCache(m, ifaceToResolve)
	if dnsCompletelyResolved && ifaceCompletelyResolved {
		onResolveFn(m)
	} else {
		agentIP, _ := m.GetTag("agent_ip")
		r.fnWorkerChannel <- func() {
			r.resolveAsyncDNS(m, dnsToResolve)
			r.resolveAsyncIFace(agentIP, m, ifaceToResolve)
			onResolveFn(m)
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

func (r *asyncResolver) resolveIFaceFromCache(m telegraf.Metric, tags map[string]string) bool {
	if !r.snmpIfaces {
		return true
	}
	result := true
	for srcTag, dstTag := range tags {
		srcTagValue, ok := m.GetTag(srcTag)
		if ok {
			located := r.ifaceCache.get(srcTagValue)
			if located != "" {
				m.AddField(dstTag, located)
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

func (r *asyncResolver) start(dnsTTL time.Duration, snmpTTL time.Duration) {
	dnsTTLStr := "(never)"
	if dnsTTL != 0 {
		dnsTTLStr = ""
		r.dnsTTLTicker = time.NewTicker(dnsTTL)
		go func() {
			for range r.dnsTTLTicker.C {
				log.Println("I! [inputs.sflow] clearing DNS cache")
				r.dnsCache.clear()
			}
		}()
	}
	snmpTTLStr := "(never)"
	if snmpTTL != 0 {
		snmpTTLStr = ""
		r.ifaceTTLTicker = time.NewTicker(snmpTTL)
		go func() {
			for range r.ifaceTTLTicker.C {
				log.Println("I! [inputs.sflow] clearing IFace cache")
				r.ifaceCache.clear()
			}
		}()
	}

	r.fnWorkerChannel = make(chan asyncJob)
	go func() {
		for {
			fn := <-r.fnWorkerChannel
			fn()
		}
	}()

	log.Printf("I! [inputs.sflow] dbs cache ttl = %d %s\n", dnsTTL, dnsTTLStr)
	log.Printf("I! [inputs.sflow] snmp cache ttl = %d %s\n", snmpTTL, snmpTTLStr)

}

func (r *asyncResolver) stop() {
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
		fqdn = r.dnsp.transform(name)
		if fqdn != name {
			log.Printf("D! [input.sflow] transformed dns[0] %s=>%s", name, fqdn)
		}
		log.Printf("D! [input.sflow] async resolve of %s=>%s", ipAddress, fqdn)
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
		log.Println("W! [inputs.sflow] err on snmp.Connect", err)
		return ifIndex
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
				log.Printf("D! [inputs.sflow] snmp bulk walk (%s) found %s as %s\n", snmpAgentIP, pdu.Name, string(b))
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
