package sflow

import (
	"fmt"
	"math"
	"net"

	"github.com/influxdata/telegraf/plugins/inputs/sflow/decoder"
)

const (
	addressTypeIPv4 = uint32(1) // line: 1383
	addressTypeIPv6 = uint32(2) // line: 1384

	sampleTypeFlowSample         = uint32(1) // line: 1614
	sampleTypeFlowSampleExpanded = uint32(3) // line: 1698

	flowDataRawPacketHeaderFormat = uint32(1) // line: 1938

	headerProtocolEthernetIso88023 = uint32(1) // line: 1920

	ipProtocolTCP = byte(6)
	ipProtocolUDP = byte(17)

	metricName = "sflow"
)

var headerProtocolMap = map[uint32]string{
	headerProtocolEthernetIso88023: "ETHERNET-ISO88023", // line: 1920
}

var etypeMap = map[uint16]string{
	0x0800: "IPv4",
	0x86DD: "IPv6",
}

func bytesToIPStr(b []byte) string {
	return net.IP(b).String()
}

func bytesToMACStr(b []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}

var ipvMap = map[uint32]string{
	1: "IPV4", // line: 1383
	2: "IPV6", // line: 1384
}

// V5FormatOptions captures configuration for controlling the processing of an SFlow V5 packet.
type V5FormatOptions struct {
	MaxFlowsPerSample   uint32
	MaxSamplesPerPacket uint32
	MaxFlowHeaderLength uint32
	MaxSampleLength     uint32
}

// NewDefaultV5FormatOptions answers a new V5FormatOptions with default values initialised
func NewDefaultV5FormatOptions() V5FormatOptions {
	return V5FormatOptions{
		MaxFlowsPerSample:   math.MaxUint32,
		MaxSamplesPerPacket: math.MaxUint32,
		MaxFlowHeaderLength: math.MaxUint32,
		MaxSampleLength:     math.MaxUint32,
	}
}

// V5Format answers and decoder.Directive capable of decoding sFlow v5 packets in accordance
// with SFlow v5 specification at https://sflow.org/sflow_version_5.txt
func V5Format(options V5FormatOptions) decoder.Directive {
	return decoder.Seq( // line: 1823
		decoder.U32().Do(decoder.U32Assert(func(v uint32) bool { return v == 5 }, "Version %d not supported, only version 5")),
		decoder.U32().Switch( // agent_address line: 1787
			decoder.Case(addressTypeIPv4, decoder.Bytes(4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("agent_address"))),   // line: 1390
			decoder.Case(addressTypeIPv6, decoder.Bytes(16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("agent_address"))), // line: 1393
		),
		decoder.U32(), // sub_agent_id line: 1790
		decoder.U32(), // sequence_number line: 1801
		decoder.U32(), // uptime line: 1804
		decoder.U32().Iter(options.MaxSamplesPerPacket, sampleRecord(options)), // samples line: 1812
	)
}

func sampleRecord(options V5FormatOptions) decoder.Directive {
	var sampleType interface{}
	return decoder.Seq( // line: 1760
		decoder.U32().Ref(&sampleType), // sample_type line: 1761
		decoder.U32().Encapsulated(options.MaxSampleLength, // sample_data line: 1762
			decoder.Ref(sampleType).Switch(
				decoder.Case(sampleTypeFlowSample, flowSample(sampleType, options)),                 // line: 1614
				decoder.Case(sampleTypeFlowSampleExpanded, flowSampleExpanded(sampleType, options)), // line: 1698
				decoder.DefaultCase(nil), // this allows other cases to just be ignored rather than cause an error
			),
		),
	)
}

func flowSample(sampleType interface{}, options V5FormatOptions) decoder.Directive {
	var samplingRate = new(uint32)
	var sourceIDIndex = new(uint32)
	return decoder.Seq( // line: 1616
		decoder.U32(), // sequence_number line: 1617
		decoder.U32(). // source_id line: 1622
				Do(decoder.U32ToU32(func(v uint32) uint32 { return v >> 24 }).AsT("source_id_type")).                            // source_id_type Line 1465
				Do(decoder.U32ToU32(func(v uint32) uint32 { return v & 0x00ffffff }).Set(sourceIDIndex).AsT("source_id_index")), // line: 1468
		decoder.U32().Do(decoder.Set(samplingRate).AsF("sampling_rate")), // line: 1631
		decoder.U32(),                          // samplePool: Line 1632
		decoder.U32().Do(decoder.AsF("drops")), // Line 1636
		decoder.U32(). // line: 1651
				Do(decoder.U32ToU32(func(v uint32) uint32 { return v & 0x3fffffff }).AsT("input_ifindex")). // line: 1477
				Do(decoder.U32ToU32(func(v uint32) uint32 { return v & 0x3fffffff }).
					ToString(func(v uint32) string {
					if v == *sourceIDIndex {
						return "ingress"
					}
					return ""
				}).
				BreakIf("").
				AsT("sample_direction")),
		decoder.U32(). // line: 1652
				Do(decoder.U32ToU32(func(v uint32) uint32 { return v & 0x3fffffff }).AsT("output_ifindex")). // line: 1477
				Do(decoder.U32ToU32(func(v uint32) uint32 { return v & 0x3fffffff }).
					ToString(func(v uint32) string {
					if v == *sourceIDIndex {
						return "egress"
					}
					return ""
				}).
				BreakIf("").
				AsT("sample_direction")),
		decoder.U32().Iter(options.MaxFlowsPerSample, flowRecord(samplingRate, options)), // line: 1654
	)
}

func flowSampleExpanded(sampleType interface{}, options V5FormatOptions) decoder.Directive {
	var samplingRate = new(uint32)
	var sourceIDIndex = new(uint32)
	return decoder.Seq( // line: 1700
		decoder.U32(), // sequence_number line: 1701
		decoder.U32().Do(decoder.AsT("source_id_type")),                     // line: 1706 + 16878
		decoder.U32().Do(decoder.Set(sourceIDIndex).AsT("source_id_index")), // line 1689
		decoder.U32().Do(decoder.Set(samplingRate).AsF("sampling_rate")),    // sample_rate line: 1707
		decoder.U32(),                          // saple_pool line: 1708
		decoder.U32().Do(decoder.AsF("drops")), // line: 1712
		decoder.U32(),                          // inputt line: 1727
		decoder.U32(). // input line: 1727
				Do(decoder.AsT("input_ifindex")). // line: 1728
				Do(decoder.U32ToStr(func(v uint32) string {
				if v == *sourceIDIndex {
					return "ingress"
				}
				return ""
			}).
				BreakIf("").
				AsT("sample_direction")),
		decoder.U32(), // output line: 1728
		decoder.U32(). // outpuit line: 1728
				Do(decoder.AsT("output_ifindex")). // line: 1729 CHANFE AS FOR NON EXPANDED
				Do(decoder.U32ToStr(func(v uint32) string {
				if v == *sourceIDIndex {
					return "egress"
				}
				return ""
			}).
				BreakIf("").
				AsT("sample_direction")),
		decoder.U32().Iter(options.MaxFlowsPerSample, flowRecord(samplingRate, options)), // line: 1730
	)
}

func flowRecord(samplingRate *uint32, options V5FormatOptions) decoder.Directive {
	var flowFormat interface{}
	return decoder.Seq( // line: 1597
		decoder.U32().Ref(&flowFormat), // line 1598
		decoder.U32().Encapsulated(options.MaxFlowHeaderLength, // line 1599
			decoder.Ref(flowFormat).Switch(
				decoder.Case(flowDataRawPacketHeaderFormat, rawPacketHeaderFlowData(samplingRate, options)), // line: 1938
				decoder.DefaultCase(nil),
			),
		),
	)
}

func rawPacketHeaderFlowData(samplingRate *uint32, options V5FormatOptions) decoder.Directive {
	var protocol interface{}
	var headerLength interface{}
	return decoder.Seq( // line: 1940
		decoder.U32().Ref(&protocol).Do(decoder.MapU32ToStr(headerProtocolMap).AsT("header_protocol")), // line: 1941
		decoder.U32(). // line: 1942
				Do(decoder.AsF("frame_length")).
				Do(decoder.U32ToU32(func(in uint32) uint32 {
				return in * (*samplingRate)
			}).AsF("bytes")),
		decoder.U32(), // stripped line: 1967
		decoder.U32().Ref(&headerLength).Do(decoder.AsF("header_length")),
		decoder.Ref(headerLength).Encapsulated(options.MaxFlowHeaderLength,
			decoder.Ref(protocol).Switch(
				decoder.Case(headerProtocolEthernetIso88023, ethHeader(options)),
				decoder.DefaultCase(nil),
			)),
	)
}

// ethHeader answers a decode Directive that will decode an ethernet frame header
// according to https://en.wikipedia.org/wiki/Ethernet_frame
func ethHeader(options V5FormatOptions) decoder.Directive {
	var tagOrEType interface{}
	etype := new(uint16)
	return decoder.Seq(
		decoder.OpenMetric(metricName),
		decoder.Bytes(6).Do(decoder.BytesToStr(6, bytesToMACStr).AsT("dst_mac")),
		decoder.Bytes(6).Do(decoder.BytesToStr(6, bytesToMACStr).AsT("src_mac")),
		decoder.U16().Ref(&tagOrEType).Switch(
			decoder.Case(uint16(0x8100),
				decoder.Seq(
					decoder.U16(),
					decoder.U16().Do(decoder.Set(etype)), // just follows on from vlan id
				),
			),
			decoder.DefaultCase( // Not an 802.1Q VLAN Tag, just treat as an ether type
				decoder.Ref(tagOrEType).Do(decoder.Set(etype)),
			),
		),
		decoder.U16Value(etype).Do(decoder.MapU16ToStr(etypeMap).AsT("ether_type")),
		decoder.U16Value(etype).Switch(
			decoder.Case(uint16(0x0800), ipv4Header(options)),
			decoder.Case(uint16(0x86DD), ipv6Header(options)),
			decoder.DefaultCase(nil),
		),
		decoder.CloseMetric(),
	)

}

// ipv4Header answers a decode Directive that decode an IPv4 header according to
// https://en.wikipedia.org/wiki/IPv4
func ipv4Header(options V5FormatOptions) decoder.Directive {
	var proto interface{}
	return decoder.Seq(
		decoder.U16().
			Do(decoder.U16ToU16(func(in uint16) uint16 { return (in & 0xFC) >> 2 }).AsT("ip_dscp")).
			Do(decoder.U16ToU16(func(in uint16) uint16 { return in & 0x3 }).AsT("ip_ecn")),
		decoder.U16().Do(decoder.AsF("ip_total_length")),
		decoder.U16(),
		decoder.U16().
			Do(decoder.U16ToU16(func(v uint16) uint16 { return (v & 0xE000) >> 13 }).AsF("ip_flags")).
			Do(decoder.U16ToU16(func(v uint16) uint16 { return v & 0x1FFF }).AsF("ip_fragment_offset")),
		decoder.Bytes(1).Do(decoder.BytesTo(1, func(b []byte) interface{} { return uint8(b[0]) }).AsF("ip_ttl")),
		decoder.Bytes(1).Ref(&proto),
		decoder.U16(),
		decoder.Bytes(4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("src_ip")),
		decoder.Bytes(4).Do(decoder.BytesToStr(4, bytesToIPStr).AsT("dst_ip")),
		decoder.Ref(proto).Switch( // Does not consider IHL and Options
			decoder.Case(ipProtocolTCP, tcpHeader(options)),
			decoder.Case(ipProtocolUDP, udpHeader(options)),
			decoder.DefaultCase(nil),
		),
	)
}

// ipv6Header answers a decode Directive that decode an IPv6 header according to
// https://en.wikipedia.org/wiki/IPv6_packet
func ipv6Header(options V5FormatOptions) decoder.Directive {
	nextHeader := new(uint16)
	return decoder.Seq(
		decoder.U32().
			Do(decoder.U32ToU32(func(in uint32) uint32 { return (in & 0xFC00000) >> 22 }).AsF("ip_dscp")).
			Do(decoder.U32ToU32(func(in uint32) uint32 { return (in & 0x300000) >> 20 }).AsF("ip_ecn")),
		decoder.U16(),
		decoder.U16().
			Do(decoder.U16ToU16(func(in uint16) uint16 { return (in & 0xFF00) >> 8 }).Set(nextHeader)),
		decoder.Bytes(16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("src_ip")),
		decoder.Bytes(16).Do(decoder.BytesToStr(16, bytesToIPStr).AsT("dst_ip")),
		decoder.U16Value(nextHeader).Switch(
			decoder.Case(uint16(ipProtocolTCP), tcpHeader(options)),
			decoder.Case(uint16(ipProtocolUDP), udpHeader(options)),
			decoder.DefaultCase(nil),
		),
	)
}

func tcpHeader(options V5FormatOptions) decoder.Directive {
	return decoder.Seq(
		decoder.U16().
			Do(decoder.AsT("src_port")),
		decoder.U16().
			Do(decoder.AsT("dst_port")),
		decoder.U32(), //"sequence"),
		decoder.U32(), //"ack_number"),
		decoder.Bytes(2).
			Do(decoder.BytesToU32(2, func(b []byte) uint32 { return uint32((b[0] & 0xF0) * 4) }).AsF("tcp_header_length")),
		decoder.U16().Do(decoder.AsF("tcp_window_size")),
		decoder.U16(), // "checksum"),
		decoder.U16().Do(decoder.AsF("tcp_urgent_pointer")),
	)
}

func udpHeader(options V5FormatOptions) decoder.Directive {
	return decoder.Seq(
		decoder.U16().
			Do(decoder.AsT("src_port")),
		decoder.U16().
			Do(decoder.AsT("dst_port")),
		decoder.U16().Do(decoder.AsF("udp_length")),
	)
}
