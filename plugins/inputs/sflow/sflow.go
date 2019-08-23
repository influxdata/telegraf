package socket_listener

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
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
		psl.process(buf[:n])
	}
}

func (psl *packetSFlowListener) process(buf []byte) {
	metrics, err := psl.Parse(buf)
	if err != nil {
		psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))

	}
	//fmt.Println("ranging over parsed metrics to resolve", len(metrics))
	//showNext := false
	for _, m := range metrics {
		//fmt.Println("passed metric to resolve", i)
		psl.resolve(m, func(resolvedM telegraf.Metric) {
			//fmt.Println("resolved m", i)
			/*
				tagList := resolvedM.TagList()
				if len(tagList) == 24 {
					if tagList[23] == nil || showNext {
						if tagList[23] == nil {
							fmt.Printf("yep, we seem to have a nil tag at 23\n")
						} else {
							fmt.Println("showing next")
						}
						for _, t := range tagList {
							if t != nil {
								fmt.Printf("%s -> %s\n", t.Key, t.Value)
							}
						}
						showNext = tagList[23] == nil
					}
				}
			*/
			skip := false
			for i, t := range resolvedM.TagList() {
				if t == nil {
					log.Printf("E! [inputs.sflow] masking nil tag error (index:%d)\n", i)
					skip = true
					break
				}
			}
			if !skip {
				psl.AddMetric(resolvedM)
			}
		})
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
	TESTDriver        bool          `toml:"test_stochastic"`
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
	sl.resolver.dnsLookup = ipToFqdn
	sl.resolver.ifaceLookup = ifIndexToIfName
	sl.dnsCache = make(map[string]string)
	sl.ifaceCache = make(map[string]string)

	if sl.TESTDriver {
		sl.resolver.dnsLookup = testIPtoFqdn
		sl.resolver.ifaceLookup = testIFaceLookup
	}

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
		if sl.TESTDriver {
			testStochasticDriver(psl)
		}
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

	ifaceLookup func(id uint64, community string, snmpAgentIP string, ifIndex string) string
	dnsLookup   func(ipAddress string) string
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
		fqdn = r.dnsLookup(ipAddress)
		r.mux.Lock()
		defer r.mux.Unlock()
		log.Printf("D! [input.sflow] async resolve of %s=>%s", ipAddress, fqdn)
		r.dnsCache[ipAddress] = fqdn
		resolved(fqdn)
	}
}

func ipToFqdn(ipAddress string) string {
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
		name = r.ifaceLookup(id, r.snmpCommunity, agentIP, ifaceIndex)
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

func testStochasticDriver(psl *packetSFlowListener) {
	r := rand.New(rand.NewSource(0))
	ticker := time.NewTicker(5 * time.Millisecond)
	go func() {
		packet := testSelectPacket(r)
		packetBytes := make([]byte, hex.DecodedLen(len(packet)))
		_, err := hex.Decode(packetBytes, packet)
		if err != nil {
			log.Panicln(err)
		}
		for range ticker.C {
			if r.Intn(20) <= 18 {
				packet = testSelectPacket(r)
				packetBytes = make([]byte, hex.DecodedLen(len(packet)))
				_, err := hex.Decode(packetBytes, packet)
				if err != nil {
					log.Panicln(err)
				}
			} else {
				byteToMessWith := r.Intn(len(packetBytes) - 1)
				packetBytes[byteToMessWith] = uint8(r.Intn(255))
			}
			psl.process(packetBytes)
		}
	}()
}

func testSelectPacket(r *rand.Rand) []byte {
	var src []byte
	switch r.Intn(6) {
	case 0:
		src = []byte("0000000500000001c0a80102000000100000f3d40bfa047f0000000200000001000000d00001210a000001fe000004000484240000000000000001fe00000200000000020000000100000090000000010000010b0000000400000080000c2936d3d694c691aa97600800450000f9f19040004011b4f5c0a80913c0a8090a00a1ba0500e5641f3081da02010104066d6f746f6770a281cc02047b46462e0201000201003081bd3012060d2b06010201190501010281dc710201003013060d2b06010201190501010281e66802025acc3012060d2b0601020119050101000003e9000000100000000900000000000000090000000000000001000000d00000e3cc000002100000400048eb740000000000000002100000020000000002000000010000009000000001000000970000000400000080000c2936d3d6fcecda44008f81000009080045000081186440003f119098c0a80815c0a8090a9a690202006d23083c33303e4170722031312030393a33333a3031206b6e6f64653120736e6d70645b313039385d3a20436f6e6e656374696f6e2066726f6d205544503a205b3139322e3136382e392e31305d3a34393233362d000003e90000001000000009000000000000000900000000")
	case 1:
		src = []byte("00000005000000010a00015000000000000f58998ae119780000000300000003000000c4000b62a90000000000100c840000040024fb7e1e0000000000000000001017840000000000100c8400000001000000010000009000000001000005bc0000000400000080001b17000130001201f58d44810023710800450205a6305440007e06ee92ac100016d94d52f505997e701fa1e17aff62574a50100200355f000000ffff00000b004175746f72697a7a6174610400008040ffff000400008040050031303030320500313030302004000000000868a200000000000000000860a200000000000000000003000000c40003cecf000000000010170400004000a168ac1c000000000000000000101784000000000010170400000001000000010000009000000001000005f200000004000000800024e8324338d4ae52aa0b54810020060800450005dc5420400080061397c0a8060cc0a806080050efcfbb25bad9a21c839a501000fff54000008a55f70975a0ff88b05735597ae274bd81fcba17e6e9206b8ea0fb07d05fc27dad06cfe3fdba5d2fc4d057b0add711e596cbe5e9b4bbe8be59cd77537b7a89f7414a628b736d00000003000000c0000c547a0000000000100c04000004005bc3c3b50000000000000000001017840000000000100c0400000001000000010000008c000000010000007e000000040000007a001b17000130001201f58d448100237108004500006824ea4000ff32c326d94d5105501018f02e88d003000001dd39b1d025d1c68689583b2ab21522d5b5a959642243804f6d51e63323091cc04544285433eb3f6b29e1046a6a2fa7806319d62041d8fa4bd25b7cd85b8db54202054a077ac11de84acbe37a550004")
	case 2:
		src = []byte("000000050000000189dd4f010000000000003d4f21151ad40000000600000001000000bc354b97090000020c000013b175792bea000000000000028f0000020c0000000300000001000000640000000100000058000000040000005408b2587a57624c16fc0b61a5080045000046c3e440003a1118a0052aada7569e5ab367a6e35b0032d7bbf1f2fb2eb2490a97f87abc31e135834be367000002590000ffffffffffffffff02add830d51e0aec14cf000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e32a000000160000000b00000001000000a88b8ffb57000002a2000013b12e344fd800000000000002a20000028f0000000300000001000000500000000100000042000000040000003e4c16fc0b6202c03e0fdecafe080045000030108000007d11fe45575185a718693996f0570e8c001c20614ad602003fd6d4afa6a6d18207324000271169b00000000003e90000001000000000000000000000000000000000000003ea000000100000000189dd4f210000000f0000001800000001000000e8354b970a0000020c000013b175793f9b000000000000028f0000020c00000003000000010000009000000001000001a500000004000000800231466d0b2c4c16fc0b61a5080045000193198f40003a114b75052aae1f5f94c778678ef24d017f50ea7622287c30799e1f7d45932d01ca92c46d930000927c0000ffffffffffffffff02ad0eea6498953d1c7ebb6dbdf0525c80e1a9a62bacfea92f69b7336c2f2f60eba0593509e14eef167eb37449f05ad70b8241c1a46d000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000160000001000000001000000e8354b970b0000020c000013b17579534c000000000000028f0000020c00000003000000010000009000000001000000b500000004000000800231466d0b2c4c16fc0b61a50800450000a327c240003606fd67b93c706a021ff365045fe8a0976d624df8207083501800edb31b0000485454502f312e3120323030204f4b0d0a5365727665723a2050726f746f636f6c20485454500d0a436f6e74656e742d4c656e6774683a20313430340d0a436f6e6e656374696f6e3a20000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000170000001000000001000000e8354b970c0000020c000013b1757966fd000000000000028f0000020c000000030000000100000090000000010000018e00000004000000800231466d0b2c4c16fc0b61a508004500017c7d2c40003a116963052abd8d021c940e67e7e0d501682342dbe7936bd47ef487dee5591ec1b24d83622e000072250000ffffffffffffffff02ad0039d8ba86a90017071d76b177de4d8c4e23bcaaaf4d795f77b032f959e0fb70234d4c28922d4e08dd3330c66e34bff51cc8ade5000003e90000001000000000000000000000000000000000000003ea0000001000000001c342e1fd000000160000001000000001000000e80d6146ac000002a1000013b17880b49d00000000000002a10000028f00000003000000010000009000000001000005ee00000004000000804c16fc0b6201d8b122766a2c0800450005dc04574000770623a11fcd80a218691d4cf2fe01bbd4f47482065fd63a5010fabd7987000052a20002c8c43ea91ca1eaa115663f5218a37fbb409dfbbedff54731ef41199b35535905ac2366a05a803146ced544abf45597f3714327d59f99e30c899c39fc5a4b67d12087bf8db2bc000003e90000001000000000000000000000000000000000000003ea000000100000000189dd4f210000001000000018")
	case 3:
		src = []byte("00000005000000010ae0648100000002000093d824ac82340000000100000001000000d000019f94000001010000100019f94000000000000000010100000000000000020000000100000090000000010000058c00000008000000800008e3fffc10d4f4be04612486dd60000000054e113a2607f8b0400200140000000000000008262000edc000e804a25e30c581af36fa01bbfa6f054e249810b584bcbf12926c2e29a779c26c72db483e8191524fe2288bfdaceaf9d2e724d04305706efcfdef70db86873bbacf29698affe4e7d6faa21d302f9b4b023291a05a000003e90000001000000001000000000000000100000000")
	case 4:
		src = []byte("00000005000000010a00015000000000000f58898ae0fa380000000700000004000000ec00006ece0000000000101784000000030000000200000034000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000058001017840000000600000002540be400000000010000000300007b8ebd37b97e61ff94860803e8e908ffb2b500000000000000000000000000018e7c31ee7ba4195f041874579ff021ba936300000000000000000000000100000007000000380011223344550003f8b15645e7e7d6960000002fe2fc02fc01edbf580000000000000000000000000000000001dcb9cf000000000000000000000004000000ec00006ece0000000000100184000000030000000200000034000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000058001001840000000600000002540be400000000010000000300000841131d1fd9f850bfb103617cb401e6598900000000000000000000000000000bec1902e5da9212e3e96d7996e922513250000000000000000000000001000000070000003800112233445500005c260acbddb3000100000003e2fc02fc01ee414f0000000000000000000000000000000001dccdd30000000000000000000000030000008400004606000000000010030400004000ad9dc19b0000000000000000001017840000000000100304000000010000000100000050000000010000004400000004000000400012815116c4001517cf426d8100200608004500002895da40008006d74bc0a8060ac0a8064f04ef04aab1797122cf7eaf4f5010ffff7727000000000000000000000003000000b0001bd698000000000010148400000400700b180f000000000000000000101504000000000010148400000001000000010000007c000000010000006f000000040000006b001b17000131f0f755b9afc081000439080045000059045340005206920c1f0d4703d94d52e201bbf14977d1e9f15498af36801800417f1100000101080afdf3c70400e043871503010020ff268cfe2e2fd5fffe1d3d704a91d57b895f174c4b4428c66679d80a307294303f00000003000000c40003ceca000000000010170400004000a166aa7a000000000000000000101784000000000010170400000001000000010000009000000001000005f200000004000000800024e8369e2bd4ae52aa0b54810020060800450005dc4c71400080061b45c0a8060cc0a806090050f855692a7a94a1154ae1801001046b6a00000101080a6869a48d151016d046a84a7aa1c6743fa05179f7ecbd4e567150cb6f2077ff89480ae730637d26d2237c08548806f672c7476eb1b5a447b42cb9ce405994d152fa3e000000030000008c001bd699000000000010148400000400700b180f0000000000000000001015040000000000101484000000010000000100000058000000010000004a0000000400000046001b17000131f0f755b9afc0810004390800450000340ce040003a06bea5c1ce8793d94d528f00504c3b08b18f275b83d5df8010054586ad00000101050a5b83d5de5b83d5df11d800000003000000c400004e07000000000010028400004000c7ec97f2000000000000000000100784000000000010028400000001000000010000009000000001000005f2000000040000008000005e0001ff005056800dd18100000a0800450005dc5a42400040066ef70a000ac8c0a8967201bbe17c81597908caf8a05f5010010328610000f172263da0ba5d6223c079b8238bc841256bf17c4ffb08ad11c4fbff6f87ae1624a6b057b8baa9342114e5f5b46179083020cb560c4e9eadcec6dfd83e102ddbc27024803eb5")
	case 5:
		src = []byte("00000005000000010a000150000000000006d14d8ae0fe200000000200000004000000ac00006d15000000004b00ca000000000200000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000584b00ca0000000001000000000000000000000001000000010000308ae33bb950eb92a8a3004d0bb406899571000000000000000000000000000012f7ed9c9db8c24ed90604eaf0bd04636edb00000000000000000000000100000004000000ac00006d15000000004b0054000000000200000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000584b00540000000001000000003b9aca000000000100000003000067ba8e64fd23fa65f26d0215ec4a0021086600000000000000000000000000002002c3b21045c2378ad3001fb2f300061872000000000000000000000001")
	case 6:
		src = []byte("0000000500000001c0a80102000000100000f3e70bfb3f590000000400000002000000a800000005000001fc000000020000000100000058000001fc00000006000000003b9aca000000000100000003000000035cfc18b203042a08000000120000004900000000000000000000000000000000c818b33e018afb7d00176fa30021698f00000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001fa000000020000000100000058000001fa00000006000000003b9aca00000000010000000300000132e5eee21da6c2e42d000003fa0000001500000000000000000000000000000100abe764d694ed34b100176bca0021697f00000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001f8000000020000000100000058000001f800000006000000003b9aca000000000100000003000001302c8b23eab41128d2000003e5000000120000000000000000000000000000019abd2b695de4797c3400176c3a0021699100000000000000000000000000000002000000340000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000a800000005000001f6000000020000000100000058000001f600000006000000003b9aca0000000001000000030000011dbd163e689cba2cc7000003e5000000520000000000000000000000000000010348cead1888a1e1ae00176c4f00216b89000000000000000000000000000000020000003400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	}
	result := make([]byte, len(src))
	copy(result[:], src)

	//fmt.Printf("%v\n", result)

	return result
}

func testIPtoFqdn(ipAddress string) string {
	return ipAddress

}
func testIFaceLookup(id uint64, community string, snmpAgentIP string, ifIndex string) string {
	return ifIndex
}
