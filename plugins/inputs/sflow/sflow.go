// Package sflow contains a Telegraf input plugin that listens for SFLow V5 network flow sample monitoring packets, parses them to extract flow
// samples which it turns into Metrics for output
package sflow

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"text/scanner"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/sflow"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type packetListener struct {
	net.PacketConn
	*Listener
	resolver
}

func (psl *packetListener) listen() {
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

func (psl *packetListener) process(buf []byte) {
	metrics, err := psl.Parse(buf)
	if err != nil {
		psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))

	}
	for _, m := range metrics {
		psl.resolver.resolve(m, func(resolvedM telegraf.Metric) {
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

	TagsAsFields string `toml:"as_fields"`

	MaxFlowsPerSample      uint32 `toml:"max_flows_per_sample"`
	MaxCountersPerSample   uint32 `toml:"max_counters_per_sample"`
	MaxSamplesPerPacket    uint32 `toml:"max_samples_per_packet"`
	MaxSampleLength        uint32 `toml:"max_sample_length"`
	MaxFlowHeaderLength    uint32 `toml:"max_flow_header_length"`
	MaxCounterHeaderLength uint32 `toml:"max_counter_header_length"`

	nameResolver resolver
	parsers.Parser
	telegraf.Accumulator
	io.Closer
}

// Description answers a description of this input plugin
func (sl *Listener) Description() string {
	return "SFlow Protocol Listener"
}

// SampleConfig answers a sample configuration
func (sl *Listener) SampleConfig() string {
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

	# Comma separated list of tags, by name, to forced back to fields
	# as_fields = "src_port,src_port_name"
	`
}

// Gather is a NOP for sFlow as it receives, asynchronously, sFlow network packets
func (sl *Listener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (sl *Listener) getSflowConfig() sflow.V5FormatOptions {
	sflowConfig := sflow.NewDefaultV5FormatOptions()
	if sl.MaxFlowsPerSample > 0 {
		sflowConfig.MaxFlowsPerSample = sl.MaxFlowsPerSample
	}
	if sl.MaxCountersPerSample > 0 {
		sflowConfig.MaxCountersPerSample = sl.MaxCountersPerSample
	}
	if sl.MaxSamplesPerPacket > 0 {
		sflowConfig.MaxSamplesPerPacket = sl.MaxSamplesPerPacket
	}
	if sl.MaxSampleLength > 0 {
		sflowConfig.MaxSampleLength = sl.MaxSampleLength
	}
	if sl.MaxFlowHeaderLength > 0 {
		sflowConfig.MaxFlowHeaderLength = sl.MaxFlowHeaderLength
	}
	if sl.MaxCounterHeaderLength > 0 {
		sflowConfig.MaxCounterHeaderLength = sl.MaxCounterHeaderLength
	}
	return sflowConfig
}

// Start starts this sFlow listener listening on the configured network for sFlow packets
func (sl *Listener) Start(acc telegraf.Accumulator) error {
	sl.Accumulator = acc
	sl.nameResolver = newAsyncResolver(sl.DNSFQDNResolve, time.Duration(sl.DNSFQDNCacheTTL)*time.Second, sl.DNSMultiNameProcessor, sl.SNMPIfaceResolve, time.Duration(sl.SNMPIfaceCacheTTL)*time.Second, sl.SNMPCommunity)
	sl.nameResolver.start()

	parser, err := sflow.NewParser("sflow", make(map[string]string), sl.getSflowConfig())
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

	log.Printf("I! [inputs.sflow] Listening on %s://%s", protocol, pc.LocalAddr())

	psl := &packetListener{
		PacketConn: pc,
		Listener:   sl,
		resolver:   sl.nameResolver,
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
	sl.nameResolver.stop()
}

// newListener constructs a new vanilla, unconfigured, listener and returns it
func newListener() *Listener {
	p, _ := sflow.NewParser("sflow", make(map[string]string), sflow.NewDefaultV5FormatOptions())
	return &Listener{Parser: p}
}

// mapTagsAsFieldsCommaString taks a comma separated string and turns it into a map of those strings v boolean 'true'
func mapTagsAsFieldsCommaString(input string) map[string]bool {
	var s scanner.Scanner
	s.Init(strings.NewReader(input))
	s.Whitespace |= 1 << ','
	asFields := make(map[string]bool)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		asFields[s.TokenText()] = true
	}
	return asFields
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
	inputs.Add("sflow", func() telegraf.Input { return newListener() })
}
