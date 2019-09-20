package sflow

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"runtime"

	d "github.com/influxdata/telegraf/plugins/parsers/sflow/decoder"
)

func thisFileColonLine() string {
	_, _, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%d", line)
}

// Line 1383 of SFlow v5 specification
var ipvMap = map[uint32]string{
	1: "IPV4",
	2: "IPV6",
}

// Line 1920 of SFlow v5 specfication
var headerProtocolMap = map[uint32]string{
	1:  "ETHERNET-ISO88023",
	2:  "ISO88024-TOKENBUS",
	3:  "ISO88025-TOKENRING",
	4:  "FDDI",
	5:  "FRAME-RELAY",
	6:  "X25",
	7:  "PPP",
	8:  "SMDS",
	9:  "AAL5",
	10: "AAL5-IP",
	11: "IPv4",
	12: "IPv6",
	13: "MPLS",
	14: "POS",
}

// The values here are scattered throughout the SFlow v5 specification - they are brought here in a single place for clarity
var formatMap = map[uint32]string{
	1:    "rawPacketHeaderFlowData",
	2:    "ethFrameFlowData",
	3:    "packetIPV4FlowData",
	4:    "packetIPV6FlowData",
	1001: "extendedSwitchFlowData",
	1002: "extendedRouterFlowData",
	1003: "extendedGatewayFlowData",
	1004: "extendedUserFlowData",
	1005: "extendedURLFlowData",
	1006: "extendedMPLSFlowData",
	1007: "extendedNATFlowData",
	1008: "extendedMPLSTunnelFlowData",
	1009: "extendedMPLSVCFlowData",
	1010: "extendedMPLSFECFlowData",
	1011: "extendedMPLSLDPFECFlowData",
	1012: "extendedVlanTunnelFlowData",
}

// V5FormatOptions captures configuration for controlling the processing of an SFlow V5 packet.
type V5FormatOptions struct {
	MaxFlowsPerSample      uint32
	MaxCountersPerSample   uint32
	MaxSamplesPerPacket    uint32
	MaxFlowHeaderLength    uint32
	MaxCounterHeaderLength uint32
	MaxSampleLength        uint32
	IncludeHeaders         bool
}

func (o V5FormatOptions) logIfNotDefault() {
	if o.MaxFlowsPerSample < math.MaxUint32 {
		log.Printf("D! [parser.sflow] !default maxFlowsPerSample: %d", o.MaxFlowsPerSample)
	}
	if o.MaxCountersPerSample < math.MaxUint32 {
		log.Printf("D! [parser.sflow] !default maxCountersPerSample: %d", o.MaxCountersPerSample)
	}
	if o.MaxSamplesPerPacket < math.MaxUint32 {
		log.Printf("D! [parser.sflow] !default maxSamplesPerPacket: %d", o.MaxSamplesPerPacket)
	}
	if o.MaxFlowHeaderLength < math.MaxUint32 {
		log.Printf("D! [parser.sflow] !default maxFlowHeaderLength: %d", o.MaxFlowHeaderLength)
	}
	if o.MaxCounterHeaderLength < math.MaxUint32 {
		log.Printf("D! [parser.sflow] !default maxCounterHeaderLength: %d", o.MaxCounterHeaderLength)
	}
	if o.MaxSampleLength < math.MaxUint32 {
		log.Printf("D! [parser.sflow] !default maxSampleLength: %d", o.MaxSampleLength)
	}
}

// NewDefaultV5FormatOptions answers a new V5FormatOptions with default values initialised
func NewDefaultV5FormatOptions() V5FormatOptions {
	return V5FormatOptions{math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, math.MaxUint32, true}
}

// V5Format answers and ItemDecoder capable of decoding sFlow v5 packets in accordance
// with SFlow v5 specification at https://sflow.org/sflow_version_5.txt
func V5Format(options V5FormatOptions) d.ItemDecoder {

	options.logIfNotDefault()

	// The numbers on comments are line number references to the sflow v5 specification at

	sourceIDTypeFn := func(v uint32) (string, uint32) { return "sourceIdType", v >> 24 }
	sourceIDValueFn := func(v uint32) (string, uint32) { return "sourceIdValue", v & 0x00ffffff }

	inputFormatFn := func(v uint32) (string, uint32) { return "inputFormat", v >> 30 }
	inputValueFn := func(v uint32) (string, uint32) { return "inputValue", v & 0x0fffffff }

	outputFormatFn := func(v uint32) (string, uint32) { return "outputFormat", v >> 30 }
	outputValueFn := func(v uint32) (string, uint32) { return "outputValue", v & 0x0fffffff }

	ipv4Fn := func(key string) d.ItemDecoder { return d.Bin(key, 4) }
	ipv6Fn := func(key string) d.ItemDecoder { return d.Bin(key, 16) }

	ethFrameFlowData := d.Seq( // 1992
		d.UI32("length"),
		d.Bin("srcMac", 6),
		d.Bin("dstMac", 6),
		d.UI32("type"),
	)

	packetIPV4FlowData := d.Seq( // 2004
		d.UI32("length"),
		d.UI32("protocol"),
		ipv4Fn("srcIP"),
		ipv4Fn("dstIP"),
		d.UI32("srcPort"),
		d.UI32("dstPort"),
		d.UI32("tcpFlags"),
		d.UI32("tos"),
	)

	packetIPV6FlowData := d.Seq( // 2027
		d.UI32("length"),
		d.UI32("protocol"),
		ipv6Fn("srcIP"),
		ipv6Fn("dstIP"),
		d.UI32("srcPort"),
		d.UI32("dstPort"),
		d.UI32("tcpFlags"),
		d.UI32("priority"),
	)

	extendedSwitchFlowData := d.Seq( //2059
		d.UI32("srcVlan"),
		d.UI32("srcPriority"),
		d.UI32("dstVlan"),
		d.UI32("dstPriority"),
	)

	extendedRouterFlowData := d.Seq( //  2083
		d.UI32Mapped("nextHop.addressType", ipvMap),
		d.Alt("nextHop.addressType",
			d.Eql("nextHop.addressType", "IPV4", d.Bin("nextHop.address", 4)),
			d.Eql("nextHop.addressType", "IPV6", d.Bin("nextHop.address", 16)),
			// TO DO ALTERNATIVE
		),
		d.UI32("srcMaskLen"),
		d.UI32("dstMaskLen"),
	)

	extendedGatewayFlowData := d.Seq( // 2104
		d.UI32Mapped("nextHop.addressType", ipvMap),
		d.Alt("nextHop.addressType",
			d.Eql("nextHop.addressType", "IPV4", d.Bin("nextHop.address", 4)),
			d.Eql("nextHop.addressType", "IPV6", d.Bin("nextHop.address", 16)),
		),
		d.UI32("as"),
		d.UI32("srcAs"),
		d.UI32("srcPeerAs"),
		d.WarnAndBreak("WARN", "unimplemented support for extendedGateway", ""),
		// 2112 ui32 array
		// 2113 ui32 communites array
		d.UI32("localpref"),
	)

	extendedUserFlowData := d.Seq( // 2124
		d.WarnAndBreak("WARN", "unimplemented support for extendedUserFlowData", ""),
	)

	extendedURLFlowData := d.Seq( // 2147
		d.WarnAndBreak("WARN", "unimplemented support for extendedURLFlowData", ""),
	)

	extendedMPLSFlowData := d.Seq( // 2164
		d.WarnAndBreak("WARN", "unimplemented support for extendedMPLSFlowData", ""),
	)

	extendedNATFlowData := d.Seq( // 2177
		d.WarnAndBreak("WARN", "unimplemented support for extendedNATFlowData", ""),
	)

	extendedMPLSTunnelFlowData := d.Seq( // 2193
		d.WarnAndBreak("WARN", "unimplemented support for extendedMPLSTunnelFlowData", ""),
	)

	extendedMPLSVCFlowData := d.Seq( // 2202
		d.WarnAndBreak("WARN", "unimplemented support for extendedMPLSVCFlowData", ""),
	)

	extendedMPLSFECFlowData := d.Seq( // 2212
		d.WarnAndBreak("WARN", "unimplemented support for extendedMPLSFECFlowData", ""),
	)

	extendedMPLSLDPFECFlowData := d.Seq( // 2223
		d.WarnAndBreak("WARN", "unimplemented support for extendedMPLSLDPFECFlowData", ""),
	)

	extendedVlanTunnelFlowData := d.Seq( // 2253
		d.WarnAndBreak("WARN", "unimplemented support for extendedVlanTunnelFlowData", ""),
	)

	var headerDecoder d.ItemDecoder
	if options.IncludeHeaders {
		headerDecoder = d.Alt("protocol",
			d.Eql("protocol", "ETHERNET-ISO88023", ethHeader("header", "header.length")),
			d.AltDefault(d.WarnAndBreak("WARN", "unimplemented support for header.protocol %d", "protocol")),
		)
	}

	rawPacketHeaderFlowData := d.Seq(
		d.UI32Mapped("protocol", headerProtocolMap), // 1942 of type headerProtocolMap
		d.UI32("frameLength"),
		d.UI32("stripped"),
		d.UI32("header.length"),
		d.AsrtMax("header.length", options.MaxFlowHeaderLength, thisFileColonLine(), false),
		d.Sub("header.length", headerDecoder),
	)

	flowData := d.Alt("flowFormat",
		d.Eql("flowFormat", "rawPacketHeaderFlowData", rawPacketHeaderFlowData), // 1939
		d.Eql("flowFormat", "ethFrameFlowData", ethFrameFlowData),
		d.Eql("flowFormat", "packetIPV4FlowData", packetIPV4FlowData),
		d.Eql("flowFormat", "packetIPV6FlowData", packetIPV6FlowData),
		d.Eql("flowFormat", "extendedSwitchFlowData", extendedSwitchFlowData),
		d.Eql("flowFormat", "extendedRouterFlowData", extendedRouterFlowData),
		d.Eql("flowFormat", "extendedGatewayFlowData", extendedGatewayFlowData),
		d.Eql("flowFormat", "extendedUserFlowData", extendedUserFlowData),
		d.Eql("flowFormat", "extendedUserFlowData", extendedURLFlowData),
		d.Eql("flowFormat", "extendedMPLSFlowData", extendedMPLSFlowData),
		d.Eql("flowFormat", "extendedNATFlowData", extendedNATFlowData),
		d.Eql("flowFormat", "extendedMPLSTunnelFlowData", extendedMPLSTunnelFlowData),
		d.Eql("flowFormat", "extendedMPLSVCFlowData", extendedMPLSVCFlowData),
		d.Eql("flowFormat", "extendedMPLSFECFlowData", extendedMPLSFECFlowData),
		d.Eql("flowFormat", "extendedMPLSLDPFECFlowData", extendedMPLSLDPFECFlowData),
		d.Eql("flowFormat", "extendedVlanTunnelFlowData", extendedVlanTunnelFlowData),
	)

	flowRecord := d.Seq(
		d.UI32Mapped("flowFormat", formatMap), // 1599
		d.UI32("flowData.length"),
		d.AsrtMax("flowData.length", options.MaxFlowHeaderLength, thisFileColonLine(), false),
		d.Sub("flowData.length", flowData), // 1600 // TODO, put a max on this
	)

	flowSample := d.Seq( // 1617
		d.UI32("sequenceNumber"),
		d.UI32("", sourceIDTypeFn, sourceIDValueFn), // "sourceId") 1623
		// 1465 mist significant byte 0=ifIndex, 1=smonVlanDaaSource, 2=entPhysicalEntry
		// lower 3 bytes contain the 'relevant index value'
		// which is consistent with what f1 and f2 do
		d.UI32("samplingRate"),
		d.UI32("samplePool"),
		d.UI32("drops"),
		d.UI32("", inputFormatFn, inputValueFn),   // 1652
		d.UI32("", outputFormatFn, outputValueFn), // 1653
		d.UI32("flowRecords.length"),              // 1655
		d.AsrtMax("flowRecords.length", options.MaxFlowsPerSample, thisFileColonLine(), true),
		d.Iter("flowRecords", "flowRecords.length", flowRecord), //1655
	)

	flowSampleExpanded := d.Seq(
		d.UI32("sequenceNumber"),
		// sflow data source expanded 1707
		d.UI32("sourceIdType"),  // sFlowDataSource type
		d.UI32("sourceIdValue"), // sFlowDataSource index
		d.UI32("samplingRate"),
		d.UI32("samplePool"),
		d.UI32("drops"),
		d.UI32("inputFormat"),        // 1728
		d.UI32("inputValue"),         // 1728
		d.UI32("outputFormat"),       // 1729
		d.UI32("outputValue"),        // 1729
		d.UI32("flowRecords.length"), // 1731
		d.AsrtMax("flowRecords.length", options.MaxFlowsPerSample, thisFileColonLine(), true),
		d.Iter("flowRecords", "flowRecords.length", flowRecord), //1731
	)

	ifCounter := d.Seq( // 2267
		d.UI32("ifIndex"),
		d.UI32("ifType"),
		d.UI64("ifSpeed"),
		d.UI32("ifDirection"),
		d.UI32("ifStatus"),
		d.UI64("ifInOctets"),
		d.UI32("ifInUcastPkts"),
		d.UI32("ifInMulticastPkts"),
		d.UI32("ifInBroadcastPkts"),
		d.UI32("ifInDiscards"),
		d.UI32("ifInErrors"),
		d.UI32("ifInUnknownProtos"),
		d.UI64("ifOutOctets"),
		d.UI32("ifOutUcastPkts"),
		d.UI32("ifOutMulticastPkts"),
		d.UI32("ifOutBroadcastPkts"),
		d.UI32("ifOutDiscards"),
		d.UI32("ifOutErrors"),
		d.UI32("ifPromiscuousMode"),
	)

	ethernetCounter := d.Seq( // 2306
		d.UI32("dot3StatsAlignmentErrors"),
		d.UI32("dot3StatsFCSErrors"),
		d.UI32("dot3StatsSingleCollisionFrames"),
		d.UI32("dot3StatsMultipleCollisionFrames"),
		d.UI32("dot3StatsSQETestErrors"),
		d.UI32("dot3StatsDeferredTransmissions"),
		d.UI32("dot3StatsLateCollisions"),
		d.UI32("dot3StatsExcessiveCollisions"),
		d.UI32("dot3StatsInternalMacTransmitErrors"),
		d.UI32("dot3StatsCarrierSenseErrors"),
		d.UI32("dot3StatsFrameTooLongs"),
		d.UI32("dot3StatsInternalMacReceiveErrors"),
		d.UI32("dot3StatsSymbolErrors"),
	)

	tokenringCounter := d.Seq( // 2325
		d.UI32("dot5StatsLineErrors"),
		d.UI32("dot5StatsBurstErrors"),
		d.UI32("dot5StatsACErrors"),
		d.UI32("dot5StatsAbortTransErrors"),
		d.UI32("dot5StatsInternalErrors"),
		d.UI32("dot5StatsLostFrameErrors"),
		d.UI32("dot5StatsReceiveCongestions"),
		d.UI32("dot5StatsFrameCopiedErrors"),
		d.UI32("dot5StatsTokenErrors"),
		d.UI32("dot5StatsSoftErrors"),
		d.UI32("dot5StatsHardErrors"),
		d.UI32("dot5StatsSignalLoss"),
		d.UI32("dot5StatsTransmitBeacons"),
		d.UI32("dot5StatsRecoverys"),
		d.UI32("dot5StatsLobeWires"),
		d.UI32("dot5StatsRemoves"),
		d.UI32("dot5StatsSingles"),
		d.UI32("dot5StatsFreqErrors"),
	)

	vgCounter := d.Seq( // 2347
		d.UI32("dot12InHighPriorityFrames"),
		d.UI64("dot12InHighPriorityOctets"),
		d.UI32("dot12InNormPriorityFrames"),
		d.UI64("dot12InNormPriorityOctets"),
		d.UI32("dot12InIPMErrors"),
		d.UI32("dot12InOversizeFrameErrors"),
		d.UI32("dot12InDataErrors"),
		d.UI32("dot12InNullAddressedFrames"),
		d.UI32("dot12OutHighPriorityFrames"),
		d.UI64("dot12OutHighPriorityOctets"),
		d.UI32("dot12TransitionIntoTrainings"),
		d.UI64("dot12HCInHighPriorityOctets"),
		d.UI64("dot12HCInNormPriorityOctets"),
		d.UI64("dot12HCOutHighPriorityOctets"),
	)

	vlanCounter := d.Seq( // 2377
		d.UI32("vlan_id"),
		d.UI64("octets"),
		d.UI32("ucastPkts"),
		d.UI32("multicastPkts"),
		d.UI32("broadcastPkts"),
		d.UI32("discards"),
	)

	processorCounter := d.Seq( // 2395
		d.I32("5s_cpu"),
		d.I32("1m_cpu"),
		d.I32("5m_cpu"),
		d.UI64("total_memory"),
		d.UI64("free_memory"),
	)

	counterDataAlts := d.Alt("counterFormat",
		d.Eql("counterFormat", uint32(1), ifCounter),           // 2267
		d.Eql("counterFormat", uint32(2), ethernetCounter),     // 2304
		d.Eql("counterFormat", uint32(3), tokenringCounter),    // 2327
		d.Eql("counterFormat", uint32(4), vgCounter),           // 2347
		d.Eql("counterFormat", uint32(5), vlanCounter),         // 2375
		d.Eql("counterFormat", uint32(1001), processorCounter), // 2393
		// format 7 is showing up which isn't in original v5 spec but is now .... https://sflow.org/developers/structures.php
		d.AltDefault(d.WarnAndBreak("WARN", "unhandled counterFormat %d", "counterFormat")),
	)

	counterRecord := d.Seq( // 1604
		d.UI32("counterFormat"),
		d.UI32("counterData.length"),
		d.AsrtMax("counterData.length", options.MaxCounterHeaderLength, thisFileColonLine(), false),
		d.Sub("counterData.length", counterDataAlts), // TODO, put a max on this
	)

	countersSample := d.Seq( // 1661
		d.UI32("sequenceNumber"),
		d.UI32("", sourceIDTypeFn, sourceIDValueFn), // "sourceId") 1672
		d.UI32("counters.length"),
		d.AsrtMax("counters.length", options.MaxCountersPerSample, thisFileColonLine(), true),
		d.Iter("counters", "counters.length", counterRecord),
	)

	countersSampleExpanded := d.Seq( // 1744
		d.UI32("sequenceNumber"),
		d.UI32("sourceIdType"),  // 1689
		d.UI32("sourceIdValue"), // 1690
		d.UI32("counters.length"),
		d.AsrtMax("counters.length", options.MaxCountersPerSample, thisFileColonLine(), true),
		d.Iter("counters", "counters.length", counterRecord),
	)

	sampleRecord := d.Seq( // 1761
		d.UI32("sampleType"), // 1762
		d.UI32("sampleData.length"),
		d.AsrtMax("sampleData.length", options.MaxSampleLength, thisFileColonLine(), false),
		d.Sub("sampleData.length", // // TODO, put a max on this
			d.Alt("sampleType",
				d.Eql("sampleType", uint32(1), flowSample),             // 1 = flowSample 1615
				d.Eql("sampleType", uint32(2), countersSample),         // 2 = countersSample 1659
				d.Eql("sampleType", uint32(3), flowSampleExpanded),     // 3 = flowSampleExpanded 1701
				d.Eql("sampleType", uint32(4), countersSampleExpanded), // 4 = countersSampleExpanded 1744
				// TODO: default
			),
		),
	)

	result := d.Seq(
		d.UI32("version"),
		d.UI32Mapped("addressType", ipvMap), // 1388
		d.Alt("addressType", // 1788
			d.Eql("addressType", "IPV4", ipv4Fn("agentAddress")),
			d.Eql("addressType", "IPV6", ipv4Fn("agentAddress")),
			// TODO
		),
		d.UI32("subAgentId"),     // 1790
		d.UI32("sequenceNumber"), // 1801
		d.UI32("uptime"),         // 1804
		d.UI32("samples.length"), // 1812 - array of sample_record
		d.AsrtMax("samples.length", options.MaxSamplesPerPacket, thisFileColonLine(), false),
		d.Iter("samples", "samples.length", sampleRecord),
	)
	return result
}

func ethHeader(fieldlName string, lenFieldName string) d.ItemDecoder {
	return d.Nest("header", d.Sub(lenFieldName, // TODO, put a max on this
		d.Seq(
			d.Bin("dstMac", 6, func(b []byte) interface{} { return binary.BigEndian.Uint64(append([]byte{0, 0}, b...)) }),
			d.Bin("srcMac", 6, func(b []byte) interface{} { return binary.BigEndian.Uint64(append([]byte{0, 0}, b...)) }),
			d.UI16("tagOrEType"),
			d.Alt("",
				d.Eql("tagOrEType", uint16(0x8100),
					d.Seq(
						d.UI16("", func(v uint16) (string, uint16) { return "vlanID", v & 0x0FFF }), // last 12 bits of it are the vlanid
						d.UI16("etype"), // just follows on from vlan id
					),
				),
				d.AltDefault( // Not an 802.1Q VLAN Tag, just treat as an ether type
					d.Asgn("tagOrEType", "etype"),
				),
			),
			d.Alt("etype",
				d.Eql("etype", uint16(0x0800), ipv4Header()),
				d.Eql("etype", uint16(0x86DD), ipv6Header()),
				d.AltDefault(d.WarnAndBreak("WARN", "unimplemented support for Ether Type %d", "etherType")),
			),
		),
	))
}

func ipv4Header() d.ItemDecoder {
	return d.Seq(
		d.UI16("",
			func(v uint16) (string, uint16) { return "IPversion", (v & 0xF000) >> 12 },
			//func(v uint16) (string, uint16) { return "ihl", (v & 0x7000) >> 8 }, ignore
			func(v uint16) (string, uint16) { return "dscp", (v & 0xFC) >> 2 },
			func(v uint16) (string, uint16) { return "ecn", v & 0x3 },
		),
		d.UI16("total_length"),
		d.UI16("fragmentId"), // identification
		d.UI16("",
			func(v uint16) (string, uint16) { return "flags", (v & 0xE000) >> 13 },
			func(v uint16) (string, uint16) { return "fragmentOffset", v & 0x1FFF },
		),
		d.Bin("IPTTL", 1, func(b []byte) interface{} { return uint8(b[0]) }),
		d.Bin("proto", 1, func(b []byte) interface{} { return uint16(b[0]) }),
		d.UI16(""), // ugnoreheader_checksum"),
		d.Bin("srcIP", 4, func(b []byte) interface{} { return net.IP(b) }),
		d.Bin("dstIP", 4, func(b []byte) interface{} { return net.IP(b) }),
		// TODO, I'm assuming no options
		d.Alt("",
			d.Eql("proto", uint16(6), tcpHeader()),
			d.Eql("proto", uint16(17), udpHeader()),
			d.AltDefault(d.WarnAndBreak("WARN", "unimplemented support for protol %d", "proto")),
		),
	)
}

func bytesToNetIP(b []byte) interface{} {
	return net.IP(b)
}

func ipv6Header() d.ItemDecoder {

	// TODO: consider options offset

	return d.Seq(
		d.UI32("",
			func(v uint32) (string, uint32) { return "IPversion", (v & 0xF000) >> 28 },
			//func(v uint32) (string, uint32) { return "ds", (v & 0xFC00000) >> 22 }, UNUSED
			//func(v uint32) (string, uint32) { return "ecn", (v & 0x300000) >> 20 },
			func(v uint32) (string, uint32) { return "IPv6FlowLabel", v & 0xFFFFF }),
		d.UI16("paylloadLength"),
		d.UI16("",
			func(v uint16) (string, uint16) { return "nextHeader", (v & 0xFF00) >> 8 },
			func(v uint16) (string, uint16) { return "hopLimit", (v & 0xFF) }),
		d.Bin("srcIP", 16, bytesToNetIP),
		d.Bin("dstIP", 16, bytesToNetIP),
	)
}

func tcpHeader() d.ItemDecoder {
	return d.Seq(
		d.UI16("srcPort"),
		d.UI16("dstPort"),
		d.UI32("sequence"),
		d.UI32("ack_number"),
		d.Bin("tcp_header_length", 2, func(b []byte) interface{} { return uint32((b[0] & 0xF0) * 4) }), // ignore other pieces
		d.UI16("tcp_window_size"),
		d.UI16("checksum"),
		d.UI16("urgent_pointer"),
	)
}

func udpHeader() d.ItemDecoder {
	return d.Seq(
		d.UI16("srcPort"),
		d.UI16("dstPort"),
		d.UI16("udp_length"),
	)
}
