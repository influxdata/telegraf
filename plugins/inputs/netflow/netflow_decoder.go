package netflow

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/netsampler/goflow2/decoders/netflow"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type fieldMapping struct {
	name    string
	decoder func([]byte) interface{}
}

// Default field mappings common for Netflow version 9 and IPFIX
// From documentations at
// - https://www.cisco.com/en/US/technologies/tk648/tk362/technologies_white_paper09186a00800a3db9.html
// - https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-information-elements
var fieldMappingsNetflowCommon = map[uint16][]fieldMapping{
	// 0: reserved
	1:  {{"in_bytes", decodeUint}},          // IN_BYTES / octetDeltaCount
	2:  {{"in_packets", decodeUint}},        // IN_PKTS / packetDeltaCount
	3:  {{"flows", decodeUint}},             // FLOWS / deltaFlowCount
	4:  {{"protocol", decodeL4Proto}},       // PROTOCOL / protocolIdentifier
	5:  {{"src_tos", decodeHex}},            // SRC_TOS / ipClassOfService
	6:  {{"tcp_flags", decodeTCPFlags}},     // TCP_FLAGS / tcpControlBits
	7:  {{"src_port", decodeUint}},          // L4_SRC_PORT / sourceTransportPort
	8:  {{"src", decodeIP}},                 // IPV4_SRC_ADDR / sourceIPv4Address
	9:  {{"src_mask", decodeUint}},          // SRC_MASK / sourceIPv4PrefixLength
	10: {{"in_snmp", decodeUint}},           // INPUT_SNMP / ingressInterface
	11: {{"dst_port", decodeUint}},          // L4_DST_PORT / destinationTransportPort
	12: {{"dst", decodeIP}},                 // IPV4_DST_ADDR / destinationIPv4Address
	13: {{"dst_mask", decodeUint}},          // DST_MASK / destinationIPv4PrefixLength
	14: {{"out_snmp", decodeUint}},          // OUTPUT_SNMP / egressInterface
	15: {{"next_hop", decodeIP}},            // IPV4_NEXT_HOP / ipNextHopIPv4Address
	16: {{"bgp_src_as", decodeUint}},        // SRC_AS / bgpSourceAsNumber
	17: {{"bgp_dst_as", decodeUint}},        // DST_AS / bgpDestinationAsNumber
	18: {{"bgp_next_hop", decodeIP}},        // BGP_IPV4_NEXT_HOP / bgpNextHopIPv4Address
	19: {{"out_mcast_packets", decodeUint}}, // MUL_DST_PKTS / postMCastPacketDeltaCount
	20: {{"out_mcast_bytes", decodeUint}},   // MUL_DST_BYTES / postMCastOctetDeltaCount
	21: {{"last_switched", decodeUint}},     // LAST_SWITCHED / flowEndSysUpTime
	22: {{"first_switched", decodeUint}},    // FIRST_SWITCHED / flowStartSysUpTime
	23: {{"out_bytes", decodeUint}},         // OUT_BYTES / postOctetDeltaCount
	24: {{"out_packets", decodeUint}},       // OUT_PKTS / postPacketDeltaCount
	25: {{"min_packet_len", decodeUint}},    // MIN_PKT_LNGTH / minimumIpTotalLength
	26: {{"max_packet_len", decodeUint}},    // MAX_PKT_LNGTH / maximumIpTotalLength
	27: {{"src", decodeIP}},                 // IPV6_SRC_ADDR / sourceIPv6Address
	28: {{"dst", decodeIP}},                 // IPV6_DST_ADDR / destinationIPv6Address
	29: {{"src_mask", decodeUint}},          // IPV6_SRC_MASK / sourceIPv6PrefixLength
	30: {{"dst_mask", decodeUint}},          // IPV6_DST_MASK / destinationIPv6PrefixLength
	31: {{"flow_label", decodeHex}},         // IPV6_FLOW_LABEL / flowLabelIPv6
	32: {
		{"icmp_type", func(b []byte) interface{} { return b[0] }}, // ICMP_TYPE / icmpTypeCodeIPv4
		{"icmp_code", func(b []byte) interface{} { return b[1] }},
	},
	33: {{"igmp_type", decodeUint}},               // MUL_IGMP_TYPE / igmpType
	34: {{"sampling_interval", decodeUint}},       // SAMPLING_INTERVAL / samplingInterval (deprecated)
	35: {{"sampling_algo", decodeSampleAlgo}},     // SAMPLING_ALGORITHM / samplingAlgorithm (deprecated)
	36: {{"flow_active_timeout", decodeUint}},     // FLOW_ACTIVE_TIMEOUT / flowActiveTimeout
	37: {{"flow_inactive_timeout", decodeUint}},   // FLOW_INACTIVE_TIMEOUT / flowIdleTimeout
	38: {{"engine_type", decodeEngineType}},       // ENGINE_TYPE / engineType (deprecated)
	39: {{"engine_id", decodeHex}},                // ENGINE_ID / engineId (deprecated)
	40: {{"total_bytes_exported", decodeUint}},    // TOTAL_BYTES_EXP / exportedOctetTotalCount
	41: {{"total_messages_exported", decodeUint}}, // TOTAL_PKTS_EXP / exportedMessageTotalCount
	42: {{"total_flows_exported", decodeUint}},    // TOTAL_FLOWS_EXP / exportedFlowRecordTotalCount
	// 43: vendor proprietary / deprecated
	44: {{"ipv4_src_prefix", decodeIP}},           // IPV4_SRC_PREFIX / sourceIPv4Prefix
	45: {{"ipv4_dst_prefix", decodeIP}},           // IPV4_DST_PREFIX / destinationIPv4Prefix
	46: {{"mpls_top_label_type", decodeMPLSType}}, // MPLS_TOP_LABEL_TYPE / mplsTopLabelType
	47: {{"mpls_top_label_ip", decodeIP}},         // MPLS_TOP_LABEL_IP_ADDR / mplsTopLabelIPv4Address
	48: {{"flow_sampler_id", decodeUint}},         // FLOW_SAMPLER_ID / samplerId (deprecated)
	49: {{"flow_sampler_mode", decodeSampleAlgo}}, // FLOW_SAMPLER_MODE / samplerMode (deprecated)
	50: {{"flow_sampler_interval", decodeUint}},   // FLOW_SAMPLER_RANDOM_INTERVAL / samplerRandomInterval (deprecated)
	// 51: vendor proprietary / deprecated
	52: {{"min_ttl", decodeUint}},         // MIN_TTL / minimumTTL
	53: {{"max_ttl", decodeUint}},         // MAX_TTL / maximumTTL
	54: {{"fragment_id", decodeHex}},      // IPV4_IDENT / fragmentIdentification
	55: {{"dst_tos", decodeHex}},          // DST_TOS / postIpClassOfService
	56: {{"in_src_mac", decodeHex}},       // IN_SRC_MAC / sourceMacAddress
	57: {{"out_dst_mac", decodeHex}},      // OUT_DST_MAC / postDestinationMacAddress
	58: {{"vlan_src", decodeUint}},        // SRC_VLAN / vlanId
	59: {{"vlan_dst", decodeUint}},        // DST_VLAN / postVlanId
	60: {{"ip_version", decodeIPVersion}}, // IP_PROTOCOL_VERSION / ipVersion
	61: {{"direction", decodeDirection}},  // DIRECTION / flowDirection
	62: {{"next_hop", decodeIP}},          // IPV6_NEXT_HOP / ipNextHopIPv6Address
	63: {{"bgp_next_hop", decodeIP}},      // BPG_IPV6_NEXT_HOP / bgpNextHopIPv6Address
	64: {{"ipv6_extensions", decodeHex}},  // IPV6_OPTION_HEADERS / ipv6ExtensionHeaders
	// 65 - 69: vendor proprietary
	70: {{"mpls_label_1", decodeHex}},      // MPLS_LABEL_1 / mplsTopLabelStackSection
	71: {{"mpls_label_2", decodeHex}},      // MPLS_LABEL_2 / mplsLabelStackSection2
	72: {{"mpls_label_3", decodeHex}},      // MPLS_LABEL_3 / mplsLabelStackSection3
	73: {{"mpls_label_4", decodeHex}},      // MPLS_LABEL_4 / mplsLabelStackSection4
	74: {{"mpls_label_5", decodeHex}},      // MPLS_LABEL_5 / mplsLabelStackSection5
	75: {{"mpls_label_6", decodeHex}},      // MPLS_LABEL_6 / mplsLabelStackSection6
	76: {{"mpls_label_7", decodeHex}},      // MPLS_LABEL_7 / mplsLabelStackSection7
	77: {{"mpls_label_8", decodeHex}},      // MPLS_LABEL_8 / mplsLabelStackSection8
	78: {{"mpls_label_9", decodeHex}},      // MPLS_LABEL_9 / mplsLabelStackSection9
	79: {{"mpls_label_10", decodeHex}},     // MPLS_LABEL_10 / mplsLabelStackSection10
	80: {{"in_dst_mac", decodeHex}},        // IN_DST_MAC / destinationMacAddress
	81: {{"out_src_mac", decodeHex}},       // OUT_SRC_MAC / postSourceMacAddress
	82: {{"interface", decodeString}},      // IF_NAME / interfaceName
	83: {{"interface_desc", decodeString}}, // IF_DESC / interfaceDescription
	84: {{"sampler_name", decodeString}},   // SAMPLER_NAME / samplerName
	85: {{"in_total_bytes", decodeUint}},   // IN_PERMANENT_BYTES / octetTotalCount
	86: {{"in_total_packets", decodeUint}}, // IN_PERMANENT_PKTS / packetTotalCount
	// 87: vendor proprietary
	88: {{"fragment_offset", decodeUint}}, // FRAGMENT_OFFSET / fragmentOffset
	89: {
		{"fwd_status", decodeFwdStatus}, // FORWARDING STATUS / forwardingStatus
		{"fwd_reason", decodeFwdReason},
	},
	90: {{"mpls_vpn_rd", decodeHex}},        // MPLS PAL RD / mplsVpnRouteDistinguisher
	91: {{"mpls_prefix_len", decodeUint}},   // MPLS PREFIX LEN / mplsTopLabelPrefixLength
	92: {{"src_traffic_index", decodeUint}}, // SRC TRAFFIC INDEX / srcTrafficIndex
	93: {{"dst_traffic_index", decodeUint}}, // DST TRAFFIC INDEX / dstTrafficIndex
	94: {{"app_desc", decodeString}},        // APPLICATION DESCRIPTION / applicationDescription
	95: {{"app_id", decodeHex}},             // APPLICATION TAG / applicationId
	96: {{"app_name", decodeString}},        // APPLICATION NAME / applicationName
	// 97: undefined
	98: {{"out_dscp", decodeUint}},           // postipDiffServCodePoint / postIpDiffServCodePoint
	99: {{"replication_factor", decodeUint}}, // replication factor / multicastReplicationFactor
	// 100: deprecated / className
	101: {{"classification_engine_id", decodeUint}}, // undefined / classificationEngineId
	102: {{"l2_packet_section_offset", decodeUint}}, // layer2packetSectionOffset
	103: {{"l2_packet_section_size", decodeUint}},   // layer2packetSectionSize
	104: {{"l2_packet_section_data", decodeHex}},    // layer2packetSectionData
	// 105 - 127:  reserved

	// Common between Netflow v9 ASA extension
	// https://www.cisco.com/c/en/us/td/docs/security/asa/special/netflow/asa_netflow.html
	// and IPFIX.
	148: {{"flow_id", decodeUint}},         // flowId
	152: {{"flow_start_ms", decodeUint}},   // NF_F_FLOW_CREATE_TIME_MSEC / flowStartMilliseconds
	153: {{"flow_end_ms", decodeUint}},     // NF_F_FLOW_END_TIME_MSEC / flowEndMilliseconds
	176: {{"icmp_type", decodeUint}},       // NF_F_ICMP_TYPE / icmpTypeIPv4
	177: {{"icmp_code", decodeUint}},       // NF_F_ICMP_CODE / icmpCodeIPv4
	178: {{"icmp_type", decodeUint}},       // NF_F_ICMP_TYPE_IPV6 / icmpTypeIPv6
	179: {{"icmp_code", decodeUint}},       // NF_F_ICMP_CODE_IPV6 / icmpCodeIPv6
	225: {{"xlat_src", decodeIP}},          // NF_F_XLATE_SRC_ADDR_IPV4 / postNATSourceIPv4Address
	226: {{"xlat_dst", decodeIP}},          // NF_F_XLATE_DST_ADDR_IPV4 / postNATDestinationIPv4Address
	227: {{"xlat_src_port", decodeUint}},   // NF_F_XLATE_SRC_PORT / postNAPTSourceTransportPort
	228: {{"xlat_dst_port", decodeUint}},   // NF_F_XLATE_DST_PORT / postNAPTDestinationTransportPort
	231: {{"initiator_bytes", decodeUint}}, // NF_F_FWD_FLOW_DELTA_BYTES / initiatorOctets
	232: {{"responder_bytes", decodeUint}}, // NF_F_REV_FLOW_DELTA_BYTES / responderOctets
	233: {{"fw_event", decodeFWEvent}},     // NF_F_FW_EVENT / firewallEvent
	281: {{"xlat_src", decodeIP}},          // NF_F_XLATE_SRC_ADDR_IPV6 / postNATSourceIPv6Address
	282: {{"xlat_dst", decodeIP}},          // NF_F_XLATE_DST_ADDR_IPV6 / postNATDestinationIPv6Address
}

// Default field mappings specific to Netflow version 9
// From documentation at https://www.cisco.com/c/en/us/td/docs/security/asa/special/netflow/asa_netflow.html
var fieldMappingsNetflowV9 = map[uint16][]fieldMapping{
	323:   {{"event_time_ms", decodeUint}}, // NF_F_EVENT_TIME_MSEC
	324:   {{"event_time_us", decodeUint}}, // NF_F_EVENT_TIME_USEC
	325:   {{"event_time_ns", decodeUint}}, // NF_F_EVENT_TIME_NSEC
	33000: {{"in_acl_id", decodeHex}},      // NF_F_INGRESS_ACL_ID
	33001: {{"out_acl_id", decodeHex}},     // NF_F_EGRESS_ACL_ID
	33002: {{"fw_event_ext", decodeHex}},   // NF_F_FW_EXT_EVENT
	40000: {{"username", decodeString}},    // NF_F_USERNAME
}

// Default field mappings specific to Netflow version 9
// From documentation at
// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-information-elements
var fieldMappingsIPFIX = map[uint16][]fieldMapping{
	128: {{"bgp_next_as", decodeUint}},              // bgpNextAdjacentAsNumber
	129: {{"bgp_prev_as", decodeUint}},              // bgpPrevAdjacentAsNumber
	130: {{"exporter", decodeIP}},                   // exporterIPv4Address
	131: {{"exporter", decodeIP}},                   // exporterIPv6Address
	132: {{"dropped_bytes", decodeUint}},            // droppedOctetDeltaCount
	133: {{"dropped_packets", decodeUint}},          // droppedPacketDeltaCount
	134: {{"dropped_bytes_total", decodeUint}},      // droppedOctetTotalCount
	135: {{"dropped_packets_total", decodeUint}},    // droppedPacketTotalCount
	136: {{"flow_end_reason", decodeFlowEndReason}}, // flowEndReason
	137: {{"common_properties_id", decodeUint}},     // commonPropertiesId
	138: {{"observation_point_id", decodeUint}},     // observationPointId
	139: {
		{"icmp_type", func(b []byte) interface{} { return b[0] }}, // icmpTypeCodeIPv6
		{"icmp_code", func(b []byte) interface{} { return b[1] }},
	},
	140: {{"mpls_top_label_ip", decodeIP}}, // mplsTopLabelIPv6Address
	141: {{"linecard_id", decodeUint}},     // lineCardId
	142: {{"port_id", decodeUint}},         // portId
	143: {{"metering_pid", decodeUint}},    // meteringProcessId
	144: {{"exporting_pid", decodeUint}},   // exportingProcessId
	145: {{"template_id", decodeUint}},     // templateId
	146: {{"wlan_channel", decodeUint}},    // wlanChannelId
	147: {{"wlan_ssid", decodeString}},     // wlanSSID
	// 148: common
	149: {{"observation_domain_id", decodeUint}}, // observationDomainId
	150: {{"flow_start", decodeUint}},            // flowStartSeconds
	// 151 - 152: common
	153: {{"flow_end_ms", decodeUint}},              // flowEndMilliseconds
	154: {{"flow_start_us", decodeUint}},            // flowStartMicroseconds
	155: {{"flow_end_us", decodeUint}},              // flowEndMicroseconds
	156: {{"flow_start_ns", decodeUint}},            // flowStartNanoseconds
	157: {{"flow_end_ns", decodeUint}},              // flowEndNanoseconds
	158: {{"flow_start_delta_us", decodeUint}},      // flowStartDeltaMicroseconds
	159: {{"flow_end_delta_us", decodeUint}},        // flowEndDeltaMicroseconds
	160: {{"system_init_ms", decodeUint}},           // systemInitTimeMilliseconds
	161: {{"flow_duration_ms", decodeUint}},         // flowDurationMilliseconds
	162: {{"flow_duration_us", decodeUint}},         // flowDurationMicroseconds
	163: {{"flow_count_total", decodeUint}},         // observedFlowTotalCount
	164: {{"ignored_packet_total", decodeUint}},     // ignoredPacketTotalCount
	165: {{"ignored_bytes_total", decodeUint}},      // ignoredOctetTotalCount
	166: {{"notsent_flow_count_total", decodeUint}}, // notSentFlowTotalCount
	167: {{"notsent_packet_total", decodeUint}},     // notSentPacketTotalCount
	168: {{"notsent_bytes_total", decodeUint}},      // notSentOctetTotalCount
	169: {{"ipv6_dst_prefix", decodeIP}},            // destinationIPv6Prefix
	170: {{"ipv6_src_prefix", decodeIP}},            // sourceIPv6Prefix
	171: {{"out_bytes_total", decodeUint}},          // postOctetTotalCount
	172: {{"out_packets_total", decodeUint}},        // postPacketTotalCount
	173: {{"flow_key_indicator", decodeHex}},        // flowKeyIndicator
	174: {{"out_mcast_packets_total", decodeUint}},  // postMCastPacketTotalCount
	175: {{"out_mcast_bytes_total", decodeUint}},    // postMCastOctetTotalCount
	// 176 - 179: common
	180: {{"udp_src_port", decodeUint}},             // udpSourcePort
	181: {{"udp_dst_port", decodeUint}},             // udpDestinationPort
	182: {{"tcp_src_port", decodeUint}},             // tcpSourcePort
	183: {{"tcp_dst_port", decodeUint}},             // tcpDestinationPort
	184: {{"tcp_seq_number", decodeUint}},           // tcpSequenceNumber
	185: {{"tcp_ack_number", decodeUint}},           // tcpAcknowledgementNumber
	186: {{"tcp_window_size", decodeUint}},          // tcpWindowSize
	187: {{"tcp_urgent_ptr", decodeUint}},           // tcpUrgentPointer
	188: {{"tcp_header_len", decodeUint}},           // tcpHeaderLength
	189: {{"ip_header_len", decodeUint}},            // ipHeaderLength
	190: {{"ipv4_total_len", decodeUint}},           // totalLengthIPv4
	191: {{"ipv6_payload_len", decodeUint}},         // payloadLengthIPv6
	192: {{"ttl", decodeUint}},                      // ipTTL
	193: {{"ipv6_next_header", decodeUint}},         // nextHeaderIPv6
	194: {{"mpls_payload_len", decodeUint}},         // mplsPayloadLength
	195: {{"dscp", decodeUint}},                     // ipDiffServCodePoint
	196: {{"precedence", decodeUint}},               // ipPrecedence
	197: {{"fragement_flags", decodeFragmentFlags}}, // fragmentFlags
	198: {{"bytes_sqr_sum", decodeUint}},            // octetDeltaSumOfSquares
	199: {{"bytes_sqr_sum_total", decodeUint}},      // octetTotalSumOfSquares
	200: {{"mpls_top_label_ttl", decodeUint}},       // mplsTopLabelTTL
	201: {{"mpls_stack_len", decodeUint}},           // mplsLabelStackLength
	202: {{"mpls_stack_depth", decodeUint}},         // mplsLabelStackDepth
	203: {{"mpls_top_label_exp", decodeUint}},       // mplsTopLabelExp
	204: {{"ip_payload_len", decodeUint}},           // ipPayloadLength
	205: {{"udp_msg_len", decodeUint}},              // udpMessageLength
	206: {{"mcast", decodeUint}},                    // isMulticast
	207: {{"ipv4_inet_header_len", decodeUint}},     // ipv4IHL
	208: {{"ipv4_options", decodeIPv4Options}},      // ipv4Options
	209: {{"tcp_options", decodeHex}},               // tcpOptions
	210: {{"padding", decodeHex}},                   // paddingOctets
	211: {{"collector", decodeIP}},                  // collectorIPv4Address
	212: {{"collector", decodeIP}},                  // collectorIPv6Address
	213: {{"export_interface", decodeUint}},         // exportInterface
	214: {{"export_proto_version", decodeUint}},     //exportProtocolVersion
	215: {{"export_transport_proto", decodeUint}},   //exportTransportProtocol
	216: {{"collector_transport_port", decodeUint}}, //collectorTransportPort
	217: {{"exporter_transport_port", decodeUint}},  //exporterTransportPort
	218: {{"tcp_syn_total", decodeUint}},            // tcpSynTotalCount
	219: {{"tcp_fin_total", decodeUint}},            // tcpFinTotalCount
	220: {{"tcp_rst_total", decodeUint}},            // tcpRstTotalCount
	221: {{"tcp_psh_total", decodeUint}},            // tcpPshTotalCount
	222: {{"tcp_ack_total", decodeUint}},            // tcpAckTotalCount
	223: {{"tcp_urg_total", decodeUint}},            // tcpUrgTotalCount
	224: {{"ip_total_len", decodeUint}},             // ipTotalLength
	// 225 - 228: common
	229: {{"nat_origin_addr_realm", decodeUint}}, // natOriginatingAddressRealm
	230: {{"nat_event", decodeUint}},             // natEvent
	// 231 - 233: common
	234: {{"in_vrf_id", decodeUint}},                      // ingressVRFID
	235: {{"out_vrf_id", decodeUint}},                     // egressVRFID
	236: {{"vrf_name", decodeString}},                     // VRFname
	237: {{"out_mpls_top_label_exp", decodeUint}},         // postMplsTopLabelExp
	238: {{"tcp_window_scale", decodeUint}},               // tcpWindowScale
	239: {{"biflow_direction", decodeBiflowDirection}},    // biflowDirection
	240: {{"eth_header_len", decodeUint}},                 // ethernetHeaderLength
	241: {{"eth_payload_len", decodeUint}},                // ethernetPayloadLength
	242: {{"eth_total_len", decodeUint}},                  // ethernetTotalLength
	243: {{"vlan_id", decodeUint}},                        // dot1qVlanId
	244: {{"vlan_priority", decodeUint}},                  // dot1qPriority
	245: {{"vlan_customer_id", decodeUint}},               // dot1qCustomerVlanId
	246: {{"vlan_customer_priority", decodeUint}},         // dot1qCustomerPriority
	247: {{"metro_evc_id", decodeString}},                 // metroEvcId
	248: {{"metro_evc_type", decodeUint}},                 // metroEvcType
	249: {{"pseudo_wire_id", decodeUint}},                 // pseudoWireId
	250: {{"pseudo_wire_type", decodeHex}},                // pseudoWireType
	251: {{"pseudo_wire_ctrl_word", decodeHex}},           // pseudoWireControlWord
	252: {{"in_phy_interface", decodeUint}},               // ingressPhysicalInterface
	253: {{"out_phy_interface", decodeUint}},              // egressPhysicalInterface
	254: {{"out_vlan_id", decodeUint}},                    // postDot1qVlanId
	255: {{"out_vlan_customer_id", decodeUint}},           // postDot1qCustomerVlanId
	256: {{"eth_type", decodeHex}},                        // ethernetType
	257: {{"out_precedence", decodeUint}},                 // postIpPrecedence
	258: {{"collection_time_ms", decodeUint}},             // collectionTimeMilliseconds
	259: {{"export_sctp_stream_id", decodeUint}},          // exportSctpStreamId
	260: {{"max_export_time", decodeUint}},                // maxExportSeconds
	261: {{"max_flow_end_time", decodeUint}},              // maxFlowEndSeconds
	262: {{"msg_md5", decodeHex}},                         // messageMD5Checksum
	263: {{"msg_scope", decodeUint}},                      // messageScope
	264: {{"min_export_time", decodeUint}},                // minExportSeconds
	265: {{"min_flow_start_time", decodeUint}},            // minFlowStartSeconds
	266: {{"opaque_bytes", decodeUint}},                   // opaqueOctets
	267: { /* MUST BE IGNORED according to standard */ },  // sessionScope
	268: {{"max_flow_end_time_us", decodeUint}},           // maxFlowEndMicroseconds
	269: {{"max_flow_end_time_ms", decodeUint}},           // maxFlowEndMilliseconds
	270: {{"max_flow_end_time_ns", decodeUint}},           // maxFlowEndNanoseconds
	271: {{"min_flow_start_time_us", decodeUint}},         // minFlowStartMicroseconds
	272: {{"min_flow_start_time_ms", decodeUint}},         // minFlowStartMilliseconds
	273: {{"min_flow_start_time_ns", decodeUint}},         // minFlowStartNanoseconds
	274: {{"collector_cert", decodeString}},               // collectorCertificate
	275: {{"exporter_cert", decodeString}},                // exporterCertificate
	276: {{"data_records_reliability", decodeBool}},       // dataRecordsReliability
	277: {{"observation_point_type", decodeOpsPointType}}, // observationPointType
	278: {{"connection_new_count", decodeUint}},           // newConnectionDeltaCount
	279: {{"connection_duration_sum", decodeUint}},        // connectionSumDurationSeconds
	280: {{"connection_transaction_id", decodeUint}},      // connectionTransactionId
	// 281 - 282: common
	283: {{"nat_pool_id", decodeUint}},     // natPoolId
	284: {{"nat_pool_name", decodeString}}, // natPoolName
	285: {
		{"anon_stability_class", decodeAnonStabilityClass}, // anonymizationFlags
		{"anon_flags", decodeAnonFlags},
	},
	286: {{"anon_technique", decodeAnonTechnique}}, // anonymizationTechnique
	287: {{"information_element", decodeUint}},     // informationElementIndex
	288: {{"p2p", decodeTechnology}},               // p2pTechnology
	289: {{"tunnel", decodeTechnology}},            // tunnelTechnology
	290: {{"encryption", decodeTechnology}},        // encryptedTechnology
	291: { /* IGNORED for parse-ability */ },       // basicList
	292: { /* IGNORED for parse-ability */ },       // subTemplateList
	293: { /* IGNORED for parse-ability */ },       // subTemplateMultiList
	294: {{"bgp_validity_state", decodeUint}},      // bgpValidityState
	295: {{"ipsec_spi", decodeUint}},               // IPSecSPI
	296: {{"gre_key", decodeUint}},                 // greKey
	297: {{"nat_type", decodeIPNatType}},           // natType
	298: {{"initiator_packets", decodeUint}},       // initiatorPackets
	299: {{"responder_packets", decodeUint}},       // responderPackets
	//	30:  {{"", decodeUint}},                        //

	//	138: {{"observation_point_id", decodeUint}},     // observationPointId
	/*
	   :                      {{"", decodeUint}},              // 294
	   :                              {{"", decodeUint}},              // 295
	   :                                {{"", decodeUint}},              // 296
	   :                               {{"", decodeUint}},              // 297
	   :                      {{"", decodeUint}},              // 298
	   :                      {{"", decodeUint}},              // 299
	   observationDomainName:                 {{"", decodeUint}},              // 300
	   selectionSequenceId:                   {{"", decodeUint}},              // 301
	   selectorId:                            {{"", decodeUint}},              // 302
	   informationElementId:                  {{"", decodeUint}},              // 303
	   selectorAlgorithm:                     {{"", decodeUint}},              // 304
	   samplingPacketInterval:                {{"", decodeUint}},              // 305
	   samplingPacketSpace:                   {{"", decodeUint}},              // 306
	   samplingTimeInterval:                  {{"", decodeUint}},              // 307
	   samplingTimeSpace:                     {{"", decodeUint}},              // 308
	   samplingSize:                          {{"", decodeUint}},              // 309
	   samplingPopulation:                    {{"", decodeUint}},              // 310
	   samplingProbability:                   {{"", decodeUint}},              // 311
	   dataLinkFrameSize:                     {{"", decodeUint}},              // 312
	   ipHeaderPacketSection:                 {{"", decodeUint}},              // 313
	   ipPayloadPacketSection:                {{"", decodeUint}},              // 314
	   dataLinkFrameSection:                  {{"", decodeUint}},              // 315
	   mplsLabelStackSection:                 {{"", decodeUint}},              // 316
	   mplsPayloadPacketSection:              {{"", decodeUint}},              // 317
	   selectorIdTotalPktsObserved:           {{"", decodeUint}},              // 318
	   selectorIdTotalPktsSelected:           {{"", decodeUint}},              // 319
	   absoluteError:                         {{"", decodeUint}},              // 320
	   relativeError:                         {{"", decodeUint}},              // 321
	   observationTimeSeconds:                {{"", decodeUint}},              // 322
	   observationTimeMilliseconds:           {{"", decodeUint}},              // 323
	   observationTimeMicroseconds:           {{"", decodeUint}},              // 324
	   observationTimeNanoseconds:            {{"", decodeUint}},              // 325
	   digestHashValue:                       {{"", decodeUint}},              // 326
	   hashIPPayloadOffset:                   {{"", decodeUint}},              // 327
	   hashIPPayloadSize:                     {{"", decodeUint}},              // 328
	   hashOutputRangeMin:                    {{"", decodeUint}},              // 329
	   hashOutputRangeMax:                    {{"", decodeUint}},              // 330
	   hashSelectedRangeMin:                  {{"", decodeUint}},              // 331
	   hashSelectedRangeMax:                  {{"", decodeUint}},              // 332
	   hashDigestOutput:                      {{"", decodeUint}},              // 333
	   hashInitialiserValue:                  {{"", decodeUint}},              // 334
	   selectorName:                          {{"", decodeUint}},              // 335
	   upperCILimit:                          {{"", decodeUint}},              // 336
	   lowerCILimit:                          {{"", decodeUint}},              // 337
	   confidenceLevel:                       {{"", decodeUint}},              // 338
	   informationElementDataType:            {{"", decodeUint}},              // 339
	   informationElementDescription:         {{"", decodeUint}},              // 340
	   informationElementName:                {{"", decodeUint}},              // 341
	   informationElementRangeBegin:          {{"", decodeUint}},              // 342
	   informationElementRangeEnd:            {{"", decodeUint}},              // 343
	   informationElementSemantics:           {{"", decodeUint}},              // 344
	   informationElementUnits:               {{"", decodeUint}},              // 345
	   privateEnterpriseNumber:               {{"", decodeUint}},              // 346
	   virtualStationInterfaceId:             {{"", decodeUint}},              // 347
	   virtualStationInterfaceName:           {{"", decodeUint}},              // 348
	   virtualStationUUID:                    {{"", decodeUint}},              // 349
	   virtualStationName:                    {{"", decodeUint}},              // 350
	   layer2SegmentId:                       {{"", decodeUint}},              // 351
	   layer2OctetDeltaCount:                 {{"", decodeUint}},              // 352
	   layer2OctetTotalCount:                 {{"", decodeUint}},              // 353
	   ingressUnicastPacketTotalCount:        {{"", decodeUint}},              // 354
	   ingressMulticastPacketTotalCount:      {{"", decodeUint}},              // 355
	   ingressBroadcastPacketTotalCount:      {{"", decodeUint}},              // 356
	   egressUnicastPacketTotalCount:         {{"", decodeUint}},              // 357
	   egressBroadcastPacketTotalCount:       {{"", decodeUint}},              // 358
	   monitoringIntervalStartMilliSeconds:   {{"", decodeUint}},              // 359
	   monitoringIntervalEndMilliSeconds:     {{"", decodeUint}},              // 360
	   portRangeStart:                        {{"", decodeUint}},              // 361
	   portRangeEnd:                          {{"", decodeUint}},              // 362
	   portRangeStepSize:                     {{"", decodeUint}},              // 363
	   portRangeNumPorts:                     {{"", decodeUint}},              // 364
	   staMacAddress:                         {{"", decodeUint}},              // 365
	   staIPv4Address:                        {{"", decodeUint}},              // 366
	   wtpMacAddress:                         {{"", decodeUint}},              // 367
	   ingressInterfaceType:                  {{"", decodeUint}},              // 368
	   egressInterfaceType:                   {{"", decodeUint}},              // 369
	   rtpSequenceNumber:                     {{"", decodeUint}},              // 370
	   userName:                              {{"", decodeUint}},              // 371
	   applicationCategoryName:               {{"", decodeUint}},              // 372
	   applicationSubCategoryName:            {{"", decodeUint}},              // 373
	   applicationGroupName:                  {{"", decodeUint}},              // 374
	   originalFlowsPresent:                  {{"", decodeUint}},              // 375
	   originalFlowsInitiated:                {{"", decodeUint}},              // 376
	   originalFlowsCompleted:                {{"", decodeUint}},              // 377
	   distinctCountOfSourceIPAddress:        {{"", decodeUint}},              // 378
	   distinctCountOfDestinationIPAddress:   {{"", decodeUint}},              // 379
	   distinctCountOfSourceIPv4Address:      {{"", decodeUint}},              // 380
	   distinctCountOfDestinationIPv4Address: {{"", decodeUint}},              // 381
	   distinctCountOfSourceIPv6Address:      {{"", decodeUint}},              // 382
	   distinctCountOfDestinationIPv6Address: {{"", decodeUint}},              // 383
	   valueDistributionMethod:               {{"", decodeUint}},              // 384
	   rfc3550JitterMilliseconds:             {{"", decodeUint}},              // 385
	   rfc3550JitterMicroseconds:             {{"", decodeUint}},              // 386
	   rfc3550JitterNanoseconds:              {{"", decodeUint}},              // 387
	   dot1qDEI:                              {{"", decodeUint}},              // 388
	   dot1qCustomerDEI:                      {{"", decodeUint}},              // 389
	   flowSelectorAlgorithm:                 {{"", decodeUint}},              // 390
	   flowSelectedOctetDeltaCount:           {{"", decodeUint}},              // 391
	   flowSelectedPacketDeltaCount:          {{"", decodeUint}},              // 392
	   flowSelectedFlowDeltaCount:            {{"", decodeUint}},              // 393
	   selectorIDTotalFlowsObserved:          {{"", decodeUint}},              // 394
	   selectorIDTotalFlowsSelected:          {{"", decodeUint}},              // 395
	   samplingFlowInterval:                  {{"", decodeUint}},              // 396
	   samplingFlowSpacing:                   {{"", decodeUint}},              // 397
	   flowSamplingTimeInterval:              {{"", decodeUint}},              // 398
	   flowSamplingTimeSpacing:               {{"", decodeUint}},              // 399
	   hashFlowDomain:                        {{"", decodeUint}},              // 400
	   transportOctetDeltaCount:              {{"", decodeUint}},              // 401
	   transportPacketDeltaCount:             {{"", decodeUint}},              // 402
	   originalExporterIPv4Address:           {{"", decodeUint}},              // 403
	   originalExporterIPv6Address:           {{"", decodeUint}},              // 404
	   originalObservationDomainId:           {{"", decodeUint}},              // 405
	   intermediateProcessId:                 {{"", decodeUint}},              // 406
	   ignoredDataRecordTotalCount:           {{"", decodeUint}},              // 407
	   dataLinkFrameType:                     {{"", decodeUint}},              // 408
	   sectionOffset:                         {{"", decodeUint}},              // 409
	   sectionExportedOctets:                 {{"", decodeUint}},              // 410
	   dot1qServiceInstanceTag:               {{"", decodeUint}},              // 411
	   dot1qServiceInstanceId:                {{"", decodeUint}},              // 412
	   dot1qServiceInstancePriority:          {{"", decodeUint}},              // 413
	   dot1qCustomerSourceMacAddress:         {{"", decodeUint}},              // 414
	   dot1qCustomerDestinationMacAddress:    {{"", decodeUint}},              // 415
	   postLayer2OctetDeltaCount:             {{"", decodeUint}},              // 417
	   postMCastLayer2OctetDeltaCount:        {{"", decodeUint}},              // 418
	   postLayer2OctetTotalCount:             {{"", decodeUint}},              // 420
	   postMCastLayer2OctetTotalCount:        {{"", decodeUint}},              // 421
	   minimumLayer2TotalLength:              {{"", decodeUint}},              // 422
	   maximumLayer2TotalLength:              {{"", decodeUint}},              // 423
	   droppedLayer2OctetDeltaCount:          {{"", decodeUint}},              // 424
	   droppedLayer2OctetTotalCount:          {{"", decodeUint}},              // 425
	   ignoredLayer2OctetTotalCount:          {{"", decodeUint}},              // 426
	   notSentLayer2OctetTotalCount:          {{"", decodeUint}},              // 427
	   layer2OctetDeltaSumOfSquares:          {{"", decodeUint}},              // 428
	   layer2OctetTotalSumOfSquares:          {{"", decodeUint}},              // 429
	   layer2FrameDeltaCount:                 {{"", decodeUint}},              // 430
	   layer2FrameTotalCount:                 {{"", decodeUint}},              // 431
	   pseudoWireDestinationIPv4Address:      {{"", decodeUint}},              // 432
	   ignoredLayer2FrameTotalCount:          {{"", decodeUint}},              // 433
	   mibObjectValueInteger:                 {{"", decodeUint}},              // 434
	   mibObjectValueOctetString:             {{"", decodeUint}},              // 435
	   mibObjectValueOID:                     {{"", decodeUint}},              // 436
	   mibObjectValueBits:                    {{"", decodeUint}},              // 437
	   mibObjectValueIPAddress:               {{"", decodeUint}},              // 438
	   mibObjectValueCounter:                 {{"", decodeUint}},              // 439
	   mibObjectValueGauge:                   {{"", decodeUint}},              // 440
	   mibObjectValueTimeTicks:               {{"", decodeUint}},              // 441
	   mibObjectValueUnsigned:                {{"", decodeUint}},              // 442
	   mibObjectValueTable:                   {{"", decodeUint}},              // 443
	   mibObjectValueRow:                     {{"", decodeUint}},              // 444
	   mibObjectIdentifier:                   {{"", decodeUint}},              // 445
	   mibSubIdentifier:                      {{"", decodeUint}},              // 446
	   mibIndexIndicator:                     {{"", decodeUint}},              // 447
	   mibCaptureTimeSemantics:               {{"", decodeUint}},              // 448
	   mibContextEngineID:                    {{"", decodeUint}},              // 449
	   mibContextName:                        {{"", decodeUint}},              // 450
	   mibObjectName:                         {{"", decodeUint}},              // 451
	   mibObjectDescription:                  {{"", decodeUint}},              // 452
	   mibObjectSyntax:                       {{"", decodeUint}},              // 453
	   mibModuleName:                         {{"", decodeUint}},              // 454
	   mobileIMSI:                            {{"", decodeUint}},              // 455
	   mobileMSISDN:                          {{"", decodeUint}},              // 456
	   httpStatusCode:                        {{"", decodeUint}},              // 457
	   sourceTransportPortsLimit:             {{"", decodeUint}},              // 458
	   httpRequestMethod:                     {{"", decodeUint}},              // 459
	   httpRequestHost:                       {{"", decodeUint}},              // 460
	   httpRequestTarget:                     {{"", decodeUint}},              // 461
	   httpMessageVersion:                    {{"", decodeUint}},              // 462
	   natInstanceID:                         {{"", decodeUint}},              // 463
	   internalAddressRealm:                  {{"", decodeUint}},              // 464
	   externalAddressRealm:                  {{"", decodeUint}},              // 465
	   natQuotaExceededEvent:                 {{"", decodeUint}},              // 466
	   natThresholdEvent:                     {{"", decodeUint}},              // 467
	*/
}

// Decoder structure
type netflowDecoder struct {
	//FieldMappings map[string]TODO
	Log telegraf.Logger

	templates     map[string]*netflow.BasicTemplateSystem
	mappingsV9    map[uint16]fieldMapping
	mappingsIPFIX map[uint16]fieldMapping

	sync.Mutex
}

func (d *netflowDecoder) Decode(srcIP net.IP, payload []byte) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric

	src := srcIP.String()

	// Prepare the templates used to decode the messages
	d.Lock()
	if _, ok := d.templates[src]; !ok {
		d.templates[src] = netflow.CreateTemplateSystem()
	}
	templates := d.templates[src]
	d.Unlock()

	// Decode the overall message
	buf := bytes.NewBuffer(payload)
	packet, err := netflow.DecodeMessage(buf, templates)
	if err != nil {
		return nil, err
	}

	// Extract metrics
	switch msg := packet.(type) {
	case netflow.NFv9Packet:
		for _, flowsets := range msg.FlowSets {
			switch fs := flowsets.(type) {
			case netflow.TemplateFlowSet:
			case netflow.NFv9OptionsTemplateFlowSet:
			case netflow.OptionsDataFlowSet:
			case netflow.DataFlowSet:
				for _, record := range fs.Records {
					tags := map[string]string{
						"source":  srcIP.String(),
						"version": "NetFlowV9",
					}
					fields := make(map[string]interface{})
					t := time.Now()
					for _, value := range record.Values {
						for _, field := range d.decodeValueV9(value) {
							fields[field.Key] = field.Value
						}
					}
					metrics = append(metrics, metric.New("netflow", tags, fields, t))
				}
			}
		}
	case netflow.IPFIXPacket:
		for _, flowsets := range msg.FlowSets {
			switch fs := flowsets.(type) {
			case netflow.TemplateFlowSet:
			case netflow.IPFIXOptionsTemplateFlowSet:
			case netflow.OptionsDataFlowSet:
			case netflow.DataFlowSet:
				for _, record := range fs.Records {
					tags := map[string]string{
						"source":  srcIP.String(),
						"version": "IPFIX",
					}
					fields := make(map[string]interface{})
					t := time.Now()
					for _, value := range record.Values {
						for _, field := range d.decodeValueIPFIX(value) {
							fields[field.Key] = field.Value
						}
					}
					metrics = append(metrics, metric.New("netflow", tags, fields, t))
				}
			}
		}
	}

	return metrics, nil
}

func (d *netflowDecoder) Init() error {
	if err := initL4ProtoMapping(); err != nil {
		return fmt.Errorf("initializing layer 4 protocol mapping failed: %w", err)
	}
	if err := initIPv4OptionMapping(); err != nil {
		return fmt.Errorf("initializing IPv4 options mapping failed: %w", err)
	}

	d.templates = make(map[string]*netflow.BasicTemplateSystem)
	d.mappingsV9 = make(map[uint16]fieldMapping)
	d.mappingsIPFIX = make(map[uint16]fieldMapping)

	return nil
}

func (d *netflowDecoder) decodeValueV9(field netflow.DataField) []telegraf.Field {
	raw := field.Value.([]byte)

	// Check the user-specified mapping
	if m, found := d.mappingsV9[field.Type]; found {
		return []telegraf.Field{{Key: m.name, Value: m.decoder(raw)}}
	}

	// Check the version specific default field mappings
	if mappings, found := fieldMappingsNetflowV9[field.Type]; found {
		var fields []telegraf.Field
		for _, m := range mappings {
			fields = append(fields, telegraf.Field{
				Key:   m.name,
				Value: m.decoder(raw),
			})
		}
		return fields
	}

	// Check the common default field mappings
	if mappings, found := fieldMappingsNetflowCommon[field.Type]; found {
		var fields []telegraf.Field
		for _, m := range mappings {
			fields = append(fields, telegraf.Field{
				Key:   m.name,
				Value: m.decoder(raw),
			})
		}
		return fields
	}

	// Return the raw data if no mapping was found
	d.Log.Debugf("unknown data field %v", field)
	name := "type_" + strconv.FormatUint(uint64(field.Type), 10)
	return []telegraf.Field{{Key: name, Value: decodeHex(raw)}}
}

func (d *netflowDecoder) decodeValueIPFIX(field netflow.DataField) []telegraf.Field {
	raw := field.Value.([]byte)

	// Checking for reverse elements according to RFC5103
	var prefix string
	elementId := field.Type
	if field.Type&0x4000 != 0 {
		prefix = "rev_"
		elementId = field.Type & (0x4000 ^ 0xffff)
	}

	// Check the user-specified mapping
	if m, found := d.mappingsIPFIX[elementId]; found {
		return []telegraf.Field{{Key: prefix + m.name, Value: m.decoder(raw)}}
	}

	// Check the version specific default field mappings
	if mappings, found := fieldMappingsIPFIX[elementId]; found {
		var fields []telegraf.Field
		for _, m := range mappings {
			fields = append(fields, telegraf.Field{
				Key:   prefix + m.name,
				Value: m.decoder(raw),
			})
		}
		return fields
	}

	// Check the common default field mappings
	if mappings, found := fieldMappingsNetflowCommon[elementId]; found {
		var fields []telegraf.Field
		for _, m := range mappings {
			fields = append(fields, telegraf.Field{
				Key:   prefix + m.name,
				Value: m.decoder(raw),
			})
		}
		return fields
	}

	// Return the raw data if no mapping was found
	d.Log.Debugf("unknown data field %v", field)
	name := "type_" + strconv.FormatUint(uint64(field.Type), 10)
	return []telegraf.Field{{Key: name, Value: decodeHex(raw)}}
}
