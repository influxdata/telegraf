// Package sflow contains a Telegraf input plugin that listens for SFLow V5 network flow sample monitoring packets, parses them to extract flow
// samples which it turns into Metrics for output
package sflow

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/network_flow"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/network_flow/netflow"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type packetListener struct {
	net.PacketConn
	*Listener
	network_flow.Resolver
}

func (psl *packetListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, a, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.AddError(err)
			}
			break
		}
		psl.process(a, buf[:n])
	}
}

func (psl *packetListener) process(addr net.Addr, buf []byte) {
	fmt.Println("netflow received len(buf)", len(buf))
	metrics, err := psl.Parse(buf)
	if err != nil {
		psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))

	}
	fmt.Println("netflow resulted in len(metrisc), err", len(metrics), err)

	for _, m := range metrics {
		if h, _, e := net.SplitHostPort(addr.String()); e == nil {
			m.AddTag("agentAddress", h)
		}
		psl.Resolver.Resolve(m, func(resolvedM telegraf.Metric) {
			psl.AddMetric(resolvedM)
		})
	}
}

// Listener configuration structure
type Listener struct {
	ServiceAddress string        `toml:"service_address"`
	ReadBufferSize internal.Size `toml:"read_buffer_size"`

	SNMPCommunity     string `toml:"snmp_community"`
	SNMPIfaceResolve  bool   `toml:"snmp_iface_resolve"`
	SNMPIfaceCacheTTL int    `toml:"snmp_iface_cache_ttl"`

	DNSFQDNResolve        bool   `toml:"dns_fqdn_resolve"`
	DNSFQDNCacheTTL       int    `toml:"dns_fqdn_cache_ttl"`
	DNSMultiNameProcessor string `toml:"dns_multi_name_processor"`

	nameResolver network_flow.Resolver
	parsers.Parser
	telegraf.Accumulator
	io.Closer
}

// Description answers a description of this input plugin
func (sl *Listener) Description() string {
	return "Netflow v9/v10 Protocol Listener"
}

// SampleConfig answers a sample configuration
func (sl *Listener) SampleConfig() string {
	return `
	## URL to listen on
	# service_address = "udp://:2055"
	# service_address = "udp4://:2055"
	# service_address = "udp6://:2055"
    
	## Maximum socket buffer size (in bytes when no unit specified).
	## For stream sockets, once the buffer fills up, the sender will start backing up.
	## For datagram sockets, once the buffer fills up, metrics will start dropping.
	## Defaults to the OS default.
	# read_buffer_size = "64KiB"

	# Whether IP addresses should be resolved to host names
	# dns_fqdn_resolve = true

	# How long should resolved IP->Hostnames be cached (in seconds)
	# dns_fqdn_cache_ttl = 3600
	
	# Optional processing instructions for transforming DNS resolve host names
	# dns_multi_name_processor = "s/(.*)(?:-net[0-9])/$1"

	# Whether Interface Indexes should be resolved to Interface Names via SNMP
	# snmp_iface_resolve = true
	
	# SNMP Community string to use when resolving Interface Names
	# snmp_community = "public"

	# How long should resolved Iface Index->Iface Name be cached (in seconds)
	# snmp_iface_cache_ttl = 3600
	`
}

// Gather is a NOP for sFlow as it receives, asynchronously, sFlow network packets
func (sl *Listener) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start starts this sFlow listener listening on the configured network for sFlow packets
func (sl *Listener) Start(acc telegraf.Accumulator) error {

	dnsToResolve := map[string]string{
		"agentAddress":           "agentHost",
		"sourceIPv4Address":      "sourceIPv4Host",
		"destinationIPv4Address": "sourceIPv4Host",
		"sourceIPv6Address":      "sourceIPv6Host",
		"destinationIPv6Address": "destinationIPv6Host",
		"exporterIPv4Address":    "exporterIPv4Host",
		"exporterIPv6Address":    "exporterIPv6Host",
	}

	sl.Accumulator = acc
	sl.nameResolver = network_flow.NewAsyncResolver(sl.DNSFQDNResolve, time.Duration(sl.DNSFQDNCacheTTL)*time.Second, sl.DNSMultiNameProcessor, sl.SNMPIfaceResolve, time.Duration(sl.SNMPIfaceCacheTTL)*time.Second, sl.SNMPCommunity, "netflow", dnsToResolve)
	sl.nameResolver.Start()

	parser, err := netflow.NewParser("netflow", make(map[string]string))
	if err != nil {
		return err
	}
	sl.Parser = parser

	spl := strings.SplitN(sl.ServiceAddress, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid service address: %s", sl.ServiceAddress)
	}

	protocol := spl[0]
	addr := spl[1]

	pc, err := newUDPListener(protocol, addr)
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

	log.Printf("I! [inputs.netflow] Listening on %s://%s", protocol, pc.LocalAddr())

	psl := &packetListener{
		PacketConn: pc,
		Listener:   sl,
		Resolver:   sl.nameResolver,
	}

	sl.Closer = psl
	go psl.listen()

	return nil
}

// Stop this Listener
func (sl *Listener) Stop() {
	if sl.Closer != nil {
		sl.Close()
		sl.Closer = nil
	}
	sl.nameResolver.Stop()
}

// newListener constructs a new vanilla, unconfigured, listener and returns it
func newListener() *Listener {
	p, _ := netflow.NewParser("netflow", make(map[string]string))
	return &Listener{Parser: p}
}

// newUDPListener answers a net.PacketConn for the expected UDP network and address passed in
func newUDPListener(network string, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return net.ListenUDP(network, addr)
	default:
		return nil, fmt.Errorf("unsupported network type %s", network)
	}
}

// init registers this SFflow input plug in with the Telegraf framework
func init() {
	inputs.Add("netflow", func() telegraf.Input { return newListener() })
}
