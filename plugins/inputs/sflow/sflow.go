package socket_listener

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/sflow"
	"github.com/soniah/gosnmp"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type packetSFlowListener struct {
	net.PacketConn
	*SFlowListener
	*resolver
}

func (psl *packetSFlowListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.AddError(err)
			}
			break
		}

		metrics, err := psl.Parse(buf[:n])
		if err != nil {
			psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))
			// TODO rate limit
			continue
		}
		for _, m := range metrics {
			psl.resolve(m, func(resolvedM telegraf.Metric) {
				psl.AddMetric(m)
			})
		}
	}
}

type SFlowListener struct {
	ServiceAddress    string        `toml:"service_address"`
	ReadBufferSize    internal.Size `toml:"read_buffer_size"`
	SNMPCommunity     string        `toml:"snmp_community"`
	SNMPIfaceResolve  bool          `toml:"snmp_iface_resolve"`
	SNMPIfaceCacheTTL int           `toml:"snmp_iface_cache_ttl"`
	DNSFQDNResolve    bool          `toml:"dns_fqdn_resolve"`
	DNSFQDNCacheTTL   int           `toml:"dns_fqdn_cache_ttl"`
	dnsTTLTicker      *time.Ticker
	ifaceTTLTicker    *time.Ticker
	tlsint.ServerConfig
	parsers.Parser
	telegraf.Accumulator
	io.Closer
	resolver
}

func (sl *SFlowListener) Description() string {
	return "SFlow protocol listener"
}

func (sl *SFlowListener) SampleConfig() string {
	return `
  ## URL to listen on
  # service_address = "udp://:6343"
  # service_address = "udp4://:6343"
  # service_address = "udp6://:6343"
  
  ## Maximum socket buffer size (in bytes when no unit specified).
  ## For stream sockets, once the buffer fills up, the sender will start backing up.
  ## For datagram sockets, once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = "64KiB"

  ## Whether interface indexes should be turned into interface names via use of sn,p
  # snmp_iface_resolve = false

  ## The SNMP community string to use for access SNMP on the agents in order to resolve interface names
  # snmp_community = "public"

  ## The length of time the interface names are cached
  # snmp_iface_cache_ttl = 3600

  ## Should IP addresses be resolved to host names through DNS lookup
  # dns_fqdn_resolve = false

  ## The length of time the FWDNs are cached
  # dns_fqdn_cache_ttl = 3600
`
}

func (sl *SFlowListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (sl *SFlowListener) Start(acc telegraf.Accumulator) error {

	sl.resolver.dns = sl.DNSFQDNResolve
	sl.resolver.snmpIfaces = sl.SNMPIfaceResolve
	sl.resolver.snmpCommunity = sl.SNMPCommunity
	sl.dnsCache = make(map[string]string)
	sl.ifaceCache = make(map[string]string)

	log.Printf("I! [inputs.sflow] dbs cache = %t", sl.resolver.dns)
	log.Printf("I! [inputs.sflow] dbs cache ttl = %d seconds", sl.DNSFQDNCacheTTL)
	log.Printf("I! [inputs.sflow] snmp cache = %t", sl.resolver.snmpIfaces)
	log.Printf("I! [inputs.sflow] snmp cache ttl = %d seconds", sl.SNMPIfaceCacheTTL)
	log.Printf("I! [inputs.sflow] snmp community = %s", sl.resolver.snmpCommunity)

	sl.dnsTTLTicker = time.NewTicker(time.Duration(sl.DNSFQDNCacheTTL) * time.Second)
	go func() {
		for range sl.dnsTTLTicker.C {
			sl.resolver.mux.Lock()
			sl.resolver.mux.Unlock()
			log.Println("D! [inputs.sflow] clearing DNS cache")
			sl.dnsCache = make(map[string]string)
		}
	}()
	sl.ifaceTTLTicker = time.NewTicker(time.Duration(sl.SNMPIfaceCacheTTL) * time.Second)
	go func() {
		for range sl.ifaceTTLTicker.C {
			sl.resolver.mux.Lock()
			sl.resolver.mux.Unlock()
			log.Println("D! [inputs.sflow] clearing IFace cache")
			sl.ifaceCache = make(map[string]string)
		}
	}()

	parser, err := sflow.NewParser("sflow", sl.SNMPCommunity, make(map[string]string)) // TODO
	if err != nil {
		return err
	}
	sl.Parser = parser

	sl.Accumulator = acc
	spl := strings.SplitN(sl.ServiceAddress, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid service address: %s", sl.ServiceAddress)
	}

	protocol := spl[0]
	addr := spl[1]

	switch protocol {
	case "udp", "udp4", "udp6":
		pc, err := udpListen(protocol, addr)
		if err != nil {
			return err
		}
		if sl.ReadBufferSize.Size > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				srb.SetReadBuffer(int(sl.ReadBufferSize.Size))
			} else {
				log.Printf("W! Unable to set read buffer on a %s socket", protocol)
			}
		}

		log.Printf("I! [inputs.sflow] Listening on %s://%s", protocol, pc.LocalAddr())

		psl := &packetSFlowListener{
			PacketConn:    pc,
			SFlowListener: sl,
			resolver:      &sl.resolver,
		}

		sl.Closer = psl
		go psl.listen()
	default:
		return fmt.Errorf("unsupported protocol '%s' in '%s'", protocol, sl.ServiceAddress)
	}

	return nil
}

func udpListen(network string, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		var addr *net.UDPAddr
		var err error
		var ifi *net.Interface
		if spl := strings.SplitN(address, "%", 2); len(spl) == 2 {
			address = spl[0]
			ifi, err = net.InterfaceByName(spl[1])
			if err != nil {
				return nil, err
			}
		}
		addr, err = net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if addr.IP.IsMulticast() {
			return net.ListenMulticastUDP(network, ifi, addr)
		}
		return net.ListenUDP(network, addr)
	}
	return net.ListenPacket(network, address)
}

func (sl *SFlowListener) Stop() {
	if sl.Closer != nil {
		sl.Close()
		sl.Closer = nil
	}
	sl.dnsTTLTicker.Stop()
}

func newSFlowListener() *SFlowListener {
	parser, _ := sflow.NewParser("sflow", "public", make(map[string]string)) // TODO

	return &SFlowListener{
		Parser: parser,
	}
}

type unixCloser struct {
	path   string
	closer io.Closer
}

func (uc unixCloser) Close() error {
	err := uc.closer.Close()
	os.Remove(uc.path) // ignore error
	return err
}

func init() {
	inputs.Add("sflow", func() telegraf.Input { return newSFlowListener() })
}

type resolver struct {
	dns           bool
	snmpIfaces    bool
	snmpCommunity string
	dnsCache      map[string]string
	ifaceCache    map[string]string
	mux           sync.Mutex
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

// Request the asynchronous resolution of ip addresses and snmp interfaces names for the given metric
// and when fully resolved then execute the provided callback
func (r *resolver) resolve(m telegraf.Metric, onResolveFn func(resolved telegraf.Metric)) {
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

func (r *resolver) dnsResolve(m telegraf.Metric, srcTag string, dstTag string, or *onResolve) {
	value, ok := m.GetTag(srcTag)
	if r.dns && ok {
		or.increment()
		go r.resolveDNS(value, func(fqdn string) {
			m.AddTag(dstTag, fqdn)
			or.decrement()
		})
	}
}

func (r *resolver) ifaceResolve(m telegraf.Metric, srcTag string, dstTag string, agentIP string, or *onResolve) {
	value, ok := m.GetTag(srcTag)
	if r.snmpIfaces && ok {
		or.increment()
		go r.resolveIFace(value, agentIP, func(name string) {
			m.AddTag(dstTag, name)
			or.decrement()
		})
	}
}

func (r *resolver) resolveDNS(ipAddress string, resolved func(fqdn string)) {
	//r.mux.Lock()
	//defer r.mux.Unlock()

	fqdn, ok := r.lookupFromDNSCache(ipAddress)
	//fqdn, ok := r.lookupFromDNSCache(ipAddress)

	if ok {
		log.Printf("D! [input.sflow] sync cache lookup %s=>%s", ipAddress, fqdn)
		resolved(fqdn)
	} else {
		ctx, cancel := context.WithTimeout(context.TODO(), 10000*time.Millisecond)
		defer cancel()
		resolver := net.Resolver{}
		names, err := resolver.LookupAddr(ctx, ipAddress)
		fqdn = ipAddress
		if err == nil {
			if len(names) > 0 {
				fqdn = names[0]
			}
		} else {
			log.Printf("!E [input.sflow] dns lookup of %s resulted in error %s", ipAddress, err)
		}
		r.mux.Lock()
		defer r.mux.Unlock()
		log.Printf("D! [input.sflow] async resolve of %s=>%s", ipAddress, fqdn)
		r.dnsCache[ipAddress] = fqdn
		resolved(fqdn)
	}
}

func (r *resolver) lookupFromDNSCache(v string) (string, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	result, ok := r.dnsCache[v]
	return result, ok
}

func (r *resolver) resolveIFace(ifaceIndex string, agentIP string, resolved func(fqdn string)) {
	id := atomic.AddUint64(&ops, 1)
	name, ok := r.lookupFromIFaceCache(agentIP, ifaceIndex)
	if ok {
		log.Printf("D! [input.sflow] %d sync cache lookup (%s,%s)=>%s", id, agentIP, ifaceIndex, name)
		resolved(name)
	} else {
		// look it up
		name = ifIndexToIfName(id, r.snmpCommunity, agentIP, ifaceIndex)
		r.mux.Lock()
		defer r.mux.Unlock()
		log.Printf("D! [input.sflow] %d async resolve of (%s,%s)=>%s", id, agentIP, ifaceIndex, name)
		r.ifaceCache[fmt.Sprintf("%s-%s", agentIP, ifaceIndex)] = name
		resolved(name)
	}
}

func (r *resolver) lookupFromIFaceCache(agentIP string, ifaceIndex string) (string, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	result, ok := r.ifaceCache[fmt.Sprintf("%s-%s", agentIP, ifaceIndex)]
	return result, ok
}

// So, Ive established that this wasn't thread safe. Might be I need a differen COnnection object.
var ifIndexToIfNameMux sync.Mutex

func ifIndexToIfName(id uint64, community string, snmpAgentIP string, ifIndex string) string {
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
