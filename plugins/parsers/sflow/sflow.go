package sflow

import (
	"runtime"
)

var ipvMap = map[uint32]string{
	1: "IPV4",
	2: "IPV6",
}

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

func file() string {
	_, file, _, _ := runtime.Caller(0)
	return file
}

func line() string {
	_, _, line, _ := runtime.Caller(0)
	return string(line)
}

func sourceIDTypeFn(v uint32) (string, uint32)  { return "sourceIdType", v >> 24 }
func sourceIDValueFn(v uint32) (string, uint32) { return "sourceIdValue", v & 0x00ffffff }

func inputFormatFn(v uint32) (string, uint32) { return "inputFormat", v >> 30 }
func inputValueFn(v uint32) (string, uint32)  { return "inputValue", v & 0x0fffffff }

func outputFormatFn(v uint32) (string, uint32) { return "outputFormat", v >> 30 }
func outputValueFn(v uint32) (string, uint32)  { return "outputValue", v & 0x0fffffff }

func ipv4Fn(key string) ItemDecoder { return bin(key, 4) }
func ipv6Fn(key string) ItemDecoder { return bin(key, 16) }

//     The most significant 2 bits are used to indicate the format of
//the 30 bit value.

// SFlowFormat answers and ItemDecoder capable of decoding sFlow v5 packets
func SFlowFormat(maxFlowsPerSample uint32, maxCountersPerSample uint32, maxSamplesPerPacket uint32) ItemDecoder {

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

	rawPacketHeaderFlowData := seq(
		ui32Mapped("protocol", headerProtocolMap), // 1942 of type headerProtocolMap
		ui32("frameLength"),
		ui32("stripped"),
		ui32("header.length"),
		sub("header.length", // TODO, put a max on this
			alt("protocol",
				eql("protocol", "ETHERNET-ISO88023", ethHeader("header", "header.length")),
				altDefault(warnAndBreak("WARN", "unimplemented support for header.protocol %d", "protocol")),
			),
		),
	)

	// not liking this below, too much use of duplicate infromation in this strings v seq processor names. Ummm!
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
		iter("flowRecords", "flowRecords.length", maxFlowsPerSample, flowRecord), //1655
	)

	flowSampleExpanded := seq(
		ui32("sequenceNumber"),
		// sflow data source expanded 1707
		ui32("sourceIdType"),  /* sFlowDataSource type */
		ui32("sourceIdValue"), /* sFlowDataSource index */
		ui32("samplingRate"),
		ui32("samplePool"),
		ui32("drops"),
		ui32("inputFormat"),        // 1728
		ui32("inputValue"),         // 1728
		ui32("outputFormat"),       // 1729
		ui32("outputValue"),        // 1729
		ui32("flowRecords.length"), // 1731
		iter("flowRecords", "flowRecords.length", maxFlowsPerSample, flowRecord), //1731
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

	counterDataAlts := alt("counterFormat", // in cloudflare (sflow.go:271) they are comparinsg the "format" which is not the counterFormat but sampleType
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
		sub("counterData.length", counterDataAlts), // TODO, put a max on this
	)

	countersSample := seq( // 1661
		ui32("sequenceNumber"),
		ui32("", sourceIDTypeFn, sourceIDValueFn), // "sourceId") 1672
		ui32("counters.length"),
		iter("counters", "counters.length", maxCountersPerSample, counterRecord),
	)

	countersSampleExpanded := seq( // 1744
		ui32("sequenceNumber"),
		ui32("sourceIdType"),  // 1689
		ui32("sourceIdValue"), // 1690
		ui32("counters.length"),
		iter("counters", "counters.length", maxCountersPerSample, counterRecord),
	)

	sampleRecord := seq( // 1761
		ui32("sampleType"), // 1762
		ui32("sampleData.length"),
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
		iter("samples", "samples.length", maxSamplesPerPacket, sampleRecord),
	)
	return result
}
