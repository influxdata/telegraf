package sflow

import (
	"encoding/binary"
	"net"
)

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
	MaxFlowsPerSample    uint32
	MaxCountersPerSample uint32
	MaxSamplesPerPacket  uint32
	MaxFlowHeaderLength  uint32
	IncludeHeaders       bool
}

// NewDefaultV5FormatOptions answers a new V5FormatOptions with default values initialised
func NewDefaultV5FormatOptions() V5FormatOptions {
	return V5FormatOptions{10, 10, 10, 2048, true}
}

// SFlowFormat answers and ItemDecoder capable of decoding sFlow v5 packets in accordance
// with SFlow v5 specification at https://sflow.org/sflow_version_5.txt
func V5Format(options V5FormatOptions) ItemDecoder {

	// The numbers on comments are line number references to the sflow v5 specification at

	sourceIDTypeFn := func(v uint32) (string, uint32) { return "sourceIdType", v >> 24 }
	sourceIDValueFn := func(v uint32) (string, uint32) { return "sourceIdValue", v & 0x00ffffff }

	inputFormatFn := func(v uint32) (string, uint32) { return "inputFormat", v >> 30 }
	inputValueFn := func(v uint32) (string, uint32) { return "inputValue", v & 0x0fffffff }

	outputFormatFn := func(v uint32) (string, uint32) { return "outputFormat", v >> 30 }
	outputValueFn := func(v uint32) (string, uint32) { return "outputValue", v & 0x0fffffff }

	ipv4Fn := func(key string) ItemDecoder { return bin(key, 4) }
	ipv6Fn := func(key string) ItemDecoder { return bin(key, 16) }

	ethFrameFlowData := seq( // 1992
		ui32("length"),
		bin("srcMac", 6),
		bin("dstMac", 6),
		ui32("type"),
	)

	packetIPV4FlowData := seq( // 2004
		ui32("length"),
		ui32("protocol"),
		ipv4Fn("srcIP"),
		ipv4Fn("dstIP"),
		ui32("srcPort"),
		ui32("dstPort"),
		ui32("tcpFlags"),
		ui32("tos"),
	)

	packetIPV6FlowData := seq( // 2027
		ui32("length"),
		ui32("protocol"),
		ipv6Fn("srcIP"),
		ipv6Fn("dstIP"),
		ui32("srcPort"),
		ui32("dstPort"),
		ui32("tcpFlags"),
		ui32("priority"),
	)

	extendedSwitchFlowData := seq( //2059
		ui32("srcVlan"),
		ui32("srcPriority"),
		ui32("dstVlan"),
		ui32("dstPriority"),
	)

	extendedRouterFlowData := seq( //  2083
		ui32Mapped("nextHop.addressType", ipvMap),
		alt("nextHop.addressType",
			eql("nextHop.addressType", "IPV4", bin("nextHop.address", 4)),
			eql("nextHop.addressType", "IPV6", bin("nextHop.address", 16)),
		),
		ui32("srcMaskLen"),
		ui32("dstMaskLen"),
	)

	extendedGatewayFlowData := seq( // 2104
		ui32Mapped("nextHop.addressType", ipvMap),
		alt("nextHop.addressType",
			eql("nextHop.addressType", "IPV4", bin("nextHop.address", 4)),
			eql("nextHop.addressType", "IPV6", bin("nextHop.address", 16)),
		),
		ui32("as"),
		ui32("srcAs"),
		ui32("srcPeerAs"),
		warnAndBreak("WARN", "unimplemented support for extendedGateway", ""),
		// 2112 ui32 array
		// 2113 ui32 communites array
		ui32("localpref"),
	)

	extendedUserFlowData := seq( // 2124
		warnAndBreak("WARN", "unimplemented support for extendedUserFlowData", ""),
	)

	extendedURLFlowData := seq( // 2147
		warnAndBreak("WARN", "unimplemented support for extendedURLFlowData", ""),
	)

	extendedMPLSFlowData := seq( // 2164
		warnAndBreak("WARN", "unimplemented support for extendedMPLSFlowData", ""),
	)

	extendedNATFlowData := seq( // 2177
		warnAndBreak("WARN", "unimplemented support for extendedNATFlowData", ""),
	)

	extendedMPLSTunnelFlowData := seq( // 2193
		warnAndBreak("WARN", "unimplemented support for extendedMPLSTunnelFlowData", ""),
	)

	extendedMPLSVCFlowData := seq( // 2202
		warnAndBreak("WARN", "unimplemented support for extendedMPLSVCFlowData", ""),
	)

	extendedMPLSFECFlowData := seq( // 2212
		warnAndBreak("WARN", "unimplemented support for extendedMPLSFECFlowData", ""),
	)

	extendedMPLSLDPFECFlowData := seq( // 2223
		warnAndBreak("WARN", "unimplemented support for extendedMPLSLDPFECFlowData", ""),
	)

	extendedVlanTunnelFlowData := seq( // 2253
		warnAndBreak("WARN", "unimplemented support for extendedVlanTunnelFlowData", ""),
	)

	var headerDecoder ItemDecoder
	if options.IncludeHeaders {
		headerDecoder = alt("protocol",
			eql("protocol", "ETHERNET-ISO88023", ethHeader("header", "header.length")),
			altDefault(warnAndBreak("WARN", "unimplemented support for header.protocol %d", "protocol")),
		)
	}

	rawPacketHeaderFlowData := seq(
		ui32Mapped("protocol", headerProtocolMap), // 1942 of type headerProtocolMap
		ui32("frameLength"),
		ui32("stripped"),
		ui32("header.length"),
		asrtMax("header.length", options.MaxFlowHeaderLength),
		sub("header.length", headerDecoder),
	)

	flowData := alt("flowFormat",
		eql("flowFormat", "rawPacketHeaderFlowData", rawPacketHeaderFlowData), // 1939
		eql("flowFormat", "ethFrameFlowData", ethFrameFlowData),
		eql("flowFormat", "packetIPV4FlowData", packetIPV4FlowData),
		eql("flowFormat", "packetIPV6FlowData", packetIPV6FlowData),
		eql("flowFormat", "extendedSwitchFlowData", extendedSwitchFlowData),
		eql("flowFormat", "extendedRouterFlowData", extendedRouterFlowData),
		eql("flowFormat", "extendedGatewayFlowData", extendedGatewayFlowData),
		eql("flowFormat", "extendedUserFlowData", extendedUserFlowData),
		eql("flowFormat", "extendedUserFlowData", extendedURLFlowData),
		eql("flowFormat", "extendedMPLSFlowData", extendedMPLSFlowData),
		eql("flowFormat", "extendedNATFlowData", extendedNATFlowData),
		eql("flowFormat", "extendedMPLSTunnelFlowData", extendedMPLSTunnelFlowData),
		eql("flowFormat", "extendedMPLSVCFlowData", extendedMPLSVCFlowData),
		eql("flowFormat", "extendedMPLSFECFlowData", extendedMPLSFECFlowData),
		eql("flowFormat", "extendedMPLSLDPFECFlowData", extendedMPLSLDPFECFlowData),
		eql("flowFormat", "extendedVlanTunnelFlowData", extendedVlanTunnelFlowData),
	)

	flowRecord := seq(
		ui32Mapped("flowFormat", formatMap), // 1599
		ui32("flowData.length"),
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		sub("flowData.length", flowData), // 1600 // TODO, put a max on this
	)

	flowSample := seq( // 1617
		ui32("sequenceNumber"),
		ui32("", sourceIDTypeFn, sourceIDValueFn), // "sourceId") 1623
		// 1465 mist significant byte 0=ifIndex, 1=smonVlanDaaSource, 2=entPhysicalEntry
		// lower 3 bytes contain the 'relevant index value'
		// which is consistent with what f1 and f2 do
		ui32("samplingRate"),
		ui32("samplePool"),
		ui32("drops"),
		ui32("", inputFormatFn, inputValueFn),   // 1652
		ui32("", outputFormatFn, outputValueFn), // 1653
		ui32("flowRecords.length"),              // 1655
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		iter("flowRecords", "flowRecords.length", options.MaxFlowsPerSample, flowRecord), //1655
	)

	flowSampleExpanded := seq(
		ui32("sequenceNumber"),
		// sflow data source expanded 1707
		ui32("sourceIdType"),  // sFlowDataSource type
		ui32("sourceIdValue"), // sFlowDataSource index
		ui32("samplingRate"),
		ui32("samplePool"),
		ui32("drops"),
		ui32("inputFormat"),        // 1728
		ui32("inputValue"),         // 1728
		ui32("outputFormat"),       // 1729
		ui32("outputValue"),        // 1729
		ui32("flowRecords.length"), // 1731
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		iter("flowRecords", "flowRecords.length", options.MaxFlowsPerSample, flowRecord), //1731
	)

	ifCounter := seq( // 2267
		ui32("ifIndex"),
		ui32("ifType"),
		ui64("ifSpeed"),
		ui32("ifDirection"),
		ui32("ifStatus"),
		ui64("ifInOctets"),
		ui32("ifInUcastPkts"),
		ui32("ifInMulticastPkts"),
		ui32("ifInBroadcastPkts"),
		ui32("ifInDiscards"),
		ui32("ifInErrors"),
		ui32("ifInUnknownProtos"),
		ui64("ifOutOctets"),
		ui32("ifOutUcastPkts"),
		ui32("ifOutMulticastPkts"),
		ui32("ifOutBroadcastPkts"),
		ui32("ifOutDiscards"),
		ui32("ifOutErrors"),
		ui32("ifPromiscuousMode"),
	)

	ethernetCounter := seq( // 2306
		ui32("dot3StatsAlignmentErrors"),
		ui32("dot3StatsFCSErrors"),
		ui32("dot3StatsSingleCollisionFrames"),
		ui32("dot3StatsMultipleCollisionFrames"),
		ui32("dot3StatsSQETestErrors"),
		ui32("dot3StatsDeferredTransmissions"),
		ui32("dot3StatsLateCollisions"),
		ui32("dot3StatsExcessiveCollisions"),
		ui32("dot3StatsInternalMacTransmitErrors"),
		ui32("dot3StatsCarrierSenseErrors"),
		ui32("dot3StatsFrameTooLongs"),
		ui32("dot3StatsInternalMacReceiveErrors"),
		ui32("dot3StatsSymbolErrors"),
	)

	tokenringCounter := seq( // 2325
		ui32("dot5StatsLineErrors"),
		ui32("dot5StatsBurstErrors"),
		ui32("dot5StatsACErrors"),
		ui32("dot5StatsAbortTransErrors"),
		ui32("dot5StatsInternalErrors"),
		ui32("dot5StatsLostFrameErrors"),
		ui32("dot5StatsReceiveCongestions"),
		ui32("dot5StatsFrameCopiedErrors"),
		ui32("dot5StatsTokenErrors"),
		ui32("dot5StatsSoftErrors"),
		ui32("dot5StatsHardErrors"),
		ui32("dot5StatsSignalLoss"),
		ui32("dot5StatsTransmitBeacons"),
		ui32("dot5StatsRecoverys"),
		ui32("dot5StatsLobeWires"),
		ui32("dot5StatsRemoves"),
		ui32("dot5StatsSingles"),
		ui32("dot5StatsFreqErrors"),
	)

	vgCounter := seq( // 2347
		ui32("dot12InHighPriorityFrames"),
		ui64("dot12InHighPriorityOctets"),
		ui32("dot12InNormPriorityFrames"),
		ui64("dot12InNormPriorityOctets"),
		ui32("dot12InIPMErrors"),
		ui32("dot12InOversizeFrameErrors"),
		ui32("dot12InDataErrors"),
		ui32("dot12InNullAddressedFrames"),
		ui32("dot12OutHighPriorityFrames"),
		ui64("dot12OutHighPriorityOctets"),
		ui32("dot12TransitionIntoTrainings"),
		ui64("dot12HCInHighPriorityOctets"),
		ui64("dot12HCInNormPriorityOctets"),
		ui64("dot12HCOutHighPriorityOctets"),
	)

	vlanCounter := seq( // 2377
		ui32("vlan_id"),
		ui64("octets"),
		ui32("ucastPkts"),
		ui32("multicastPkts"),
		ui32("broadcastPkts"),
		ui32("discards"),
	)

	processorCounter := seq( // 2395
		i32("5s_cpu"),
		i32("1m_cpu"),
		i32("5m_cpu"),
		ui64("total_memory"),
		ui64("free_memory"),
	)

	counterDataAlts := alt("counterFormat",
		eql("counterFormat", uint32(1), ifCounter),           // 2267
		eql("counterFormat", uint32(2), ethernetCounter),     // 2304
		eql("counterFormat", uint32(3), tokenringCounter),    // 2327
		eql("counterFormat", uint32(4), vgCounter),           // 2347
		eql("counterFormat", uint32(5), vlanCounter),         // 2375
		eql("counterFormat", uint32(1001), processorCounter), // 2393
		// format 7 is showing up which isn't in original v5 spec but is now .... https://sflow.org/developers/structures.php
		altDefault(warnAndBreak("WARN", "unhandled counterFormat %d", "counterFormat")),
	)

	counterRecord := seq( // 1604
		ui32("counterFormat"),
		ui32("counterData.length"),
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		sub("counterData.length", counterDataAlts), // TODO, put a max on this
	)

	countersSample := seq( // 1661
		ui32("sequenceNumber"),
		ui32("", sourceIDTypeFn, sourceIDValueFn), // "sourceId") 1672
		ui32("counters.length"),
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		iter("counters", "counters.length", options.MaxCountersPerSample, counterRecord),
	)

	countersSampleExpanded := seq( // 1744
		ui32("sequenceNumber"),
		ui32("sourceIdType"),  // 1689
		ui32("sourceIdValue"), // 1690
		ui32("counters.length"),
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		iter("counters", "counters.length", options.MaxCountersPerSample, counterRecord),
	)

	sampleRecord := seq( // 1761
		ui32("sampleType"), // 1762
		ui32("sampleData.length"),
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		sub("sampleData.length", // // TODO, put a max on this
			alt("sampleType",
				eql("sampleType", uint32(1), flowSample),             // 1 = flowSample 1615
				eql("sampleType", uint32(2), countersSample),         // 2 = countersSample 1659
				eql("sampleType", uint32(3), flowSampleExpanded),     // 3 = flowSampleExpanded 1701
				eql("sampleType", uint32(4), countersSampleExpanded), // 4 = countersSampleExpanded 1744
			),
		),
	)

	result := seq(
		ui32("version"),
		ui32Mapped("addressType", ipvMap), // 1388
		alt("addressType", // 1788
			eql("addressType", "IPV4", ipv4Fn("agentAddress")),
			eql("addressType", "IPV6", ipv4Fn("agentAddress")),
		),
		ui32("subAgentId"),     // 1790
		ui32("sequenceNumber"), // 1801
		ui32("uptime"),         // 1804
		ui32("samples.length"), // 1812 - array of sample_record
		//asrtMax("header.length", options.maxFlowHeaderLength), // what should it do if we breach? breakAndWarn?
		iter("samples", "samples.length", options.MaxSamplesPerPacket, sampleRecord),
	)
	return result
}

func ethHeader(fieldlName string, lenFieldName string) ItemDecoder {
	return nest("header", sub(lenFieldName, // TODO, put a max on this
		seq(
			bin("dstMac", 6, func(b []byte) interface{} { return binary.BigEndian.Uint64(append([]byte{0, 0}, b...)) }),
			bin("srcMac", 6, func(b []byte) interface{} { return binary.BigEndian.Uint64(append([]byte{0, 0}, b...)) }),
			ui16("tagOrEType"),
			alt("",
				eql("tagOrEType", uint16(0x8100),
					seq(
						ui16("", func(v uint16) (string, uint16) { return "vlanID", v & 0x0FFF }), // last 12 bits of it are the vlanid
						ui16("etype"), // just follows on from vlan id
					),
				),
				altDefault( // Not an 802.1Q VLAN Tag, just treat as an ether type
					asgn("tagOrEType", "etype"),
				),
			),
			alt("etype",
				eql("etype", uint16(0x0800), ipv4Header()),
				eql("etype", uint16(0x86DD), ipv6Header()),
				altDefault(warnAndBreak("WARN", "unimplemented support for Ether Type %d", "etherType")),
			),
		),
	))
}

func ipv4Header() ItemDecoder {
	return seq(
		ui16("",
			func(v uint16) (string, uint16) { return "IPversion", (v & 0xF000) >> 12 },
			//func(v uint16) (string, uint16) { return "ihl", (v & 0x7000) >> 8 }, ignore
			func(v uint16) (string, uint16) { return "dscp", (v & 0xFC) >> 2 },
			func(v uint16) (string, uint16) { return "ecn", v & 0x3 },
		),
		ui16("total_length"),
		ui16("fragmentId"), // identification
		ui16("",
			func(v uint16) (string, uint16) { return "flags", (v & 0xE000) >> 13 },
			func(v uint16) (string, uint16) { return "fragmentOffset", v & 0x1FFF },
		),
		bin("IPTTL", 1, func(b []byte) interface{} { return uint8(b[0]) }),
		bin("proto", 1, func(b []byte) interface{} { return uint16(b[0]) }),
		ui16(""), // ugnoreheader_checksum"),
		bin("srcIP", 4, func(b []byte) interface{} { return net.IP(b) }),
		bin("dstIP", 4, func(b []byte) interface{} { return net.IP(b) }),
		// TODO, I'm assuming no options
		alt("",
			eql("proto", uint16(6), tcpHeader()),
			eql("proto", uint16(17), udpHeader()),
			altDefault(warnAndBreak("WARN", "unimplemented support for protol %d", "proto")),
		),
	)
}

func bytesToNetIP(b []byte) interface{} {
	return net.IP(b)
}

func ipv6Header() ItemDecoder {
	return seq(
		ui32("",
			func(v uint32) (string, uint32) { return "IPversion", (v & 0xF000) >> 28 },
			//func(v uint32) (string, uint32) { return "ds", (v & 0xFC00000) >> 22 }, UNUSED
			//func(v uint32) (string, uint32) { return "ecn", (v & 0x300000) >> 20 },
			func(v uint32) (string, uint32) { return "IPv6FlowLabel", v & 0xFFFFF }),
		ui16("paylloadLength"),
		ui16("",
			func(v uint16) (string, uint16) { return "nextHeader", (v & 0xFF00) >> 8 },
			func(v uint16) (string, uint16) { return "hopLimit", (v & 0xFF) }),
		bin("srcIP", 16, bytesToNetIP),
		bin("dstIP", 16, bytesToNetIP),
	)
}

func tcpHeader() ItemDecoder {
	return seq(
		ui16("srcPort"),
		ui16("dstPort"),
		ui32("sequence"),
		ui32("ack_number"),
		bin("tcp_header_length", 2, func(b []byte) interface{} { return uint32((b[0] & 0xF0) * 4) }), // ignore other pieces
		ui16("tcp_window_size"),
		ui16("checksum"),
		ui16("urgent_pointer"),
	)
}

func udpHeader() ItemDecoder {
	return seq(
		ui16("srcPort"),
		ui16("dstPort"),
		ui16("udp_length"),
	)
}
