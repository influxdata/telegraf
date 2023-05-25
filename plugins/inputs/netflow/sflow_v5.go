package netflow

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/netsampler/goflow2/decoders/sflow"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Decoder structure
type sflowv5Decoder struct {
	Log telegraf.Logger

	warnedCounterRaw map[uint32]bool
	warnedFlowRaw    map[int64]bool
}

func (d *sflowv5Decoder) Init() error {
	if err := initL4ProtoMapping(); err != nil {
		return fmt.Errorf("initializing layer 4 protocol mapping failed: %w", err)
	}
	d.warnedCounterRaw = make(map[uint32]bool)
	d.warnedFlowRaw = make(map[int64]bool)

	return nil
}

func (d *sflowv5Decoder) Decode(srcIP net.IP, payload []byte) ([]telegraf.Metric, error) {
	src := srcIP.String()

	// Decode the message
	buf := bytes.NewBuffer(payload)
	packet, err := sflow.DecodeMessage(buf)
	if err != nil {
		return nil, err
	}

	// Extract metrics
	msg, ok := packet.(sflow.Packet)
	if !ok {
		return nil, fmt.Errorf("unexpected message type %T", packet)
	}

	t := time.Unix(0, int64(msg.Uptime)*int64(time.Millisecond))
	metrics := make([]telegraf.Metric, 0, len(msg.Samples))
	for _, s := range msg.Samples {
		tags := map[string]string{
			"source":  src,
			"version": "sFlowV5",
		}

		switch sample := s.(type) {
		case sflow.FlowSample:
			fields := map[string]interface{}{
				"ip_version":        decodeSflowIPVersion(msg.IPVersion),
				"sys_uptime":        msg.Uptime,
				"agent_ip":          decodeIP(msg.AgentIP),
				"agent_subid":       msg.SubAgentId,
				"seq_number":        sample.Header.SampleSequenceNumber,
				"sampling_interval": sample.SamplingRate,
				"in_total_packets":  sample.SamplePool,
				"sampling_drops":    sample.Drops,
				"in_snmp":           sample.Input,
			}
			if sample.Output>>31 == 0 {
				fields["out_snmp"] = sample.Output & 0x7fffffff
			}
			// Decode the source information
			if name := decodeSflowSourceInterface(sample.Header.SourceIdType); name != "" {
				fields[name] = sample.Header.SourceIdValue
			}
			// Decode the sampling direction
			if sample.Header.SourceIdValue == sample.Input {
				fields["direction"] = "ingress"
			} else {
				fields["direction"] = "egress"
			}
			recordFields, err := d.decodeFlowRecords(sample.Records)
			if err != nil {
				return nil, err
			}
			for k, v := range recordFields {
				fields[k] = v
			}
			metrics = append(metrics, metric.New("netflow", tags, fields, t))
		case sflow.ExpandedFlowSample:
			fields := map[string]interface{}{
				"ip_version":        decodeSflowIPVersion(msg.IPVersion),
				"sys_uptime":        msg.Uptime,
				"agent_ip":          decodeIP(msg.AgentIP),
				"agent_subid":       msg.SubAgentId,
				"seq_number":        sample.Header.SampleSequenceNumber,
				"sampling_interval": sample.SamplingRate,
				"in_total_packets":  sample.SamplePool,
				"sampling_drops":    sample.Drops,
				"in_snmp":           sample.InputIfValue,
			}
			if sample.OutputIfFormat == 0 {
				fields["out_snmp"] = sample.OutputIfValue
			}
			// Decode the source information
			if name := decodeSflowSourceInterface(sample.Header.SourceIdType); name != "" {
				fields[name] = sample.Header.SourceIdValue
			}
			// Decode the sampling direction
			if sample.Header.SourceIdValue == sample.InputIfValue {
				fields["direction"] = "ingress"
			} else {
				fields["direction"] = "egress"
			}
			recordFields, err := d.decodeFlowRecords(sample.Records)
			if err != nil {
				return nil, err
			}
			for k, v := range recordFields {
				fields[k] = v
			}
			metrics = append(metrics, metric.New("netflow", tags, fields, t))
		case sflow.CounterSample:
			fields := map[string]interface{}{
				"ip_version":  decodeSflowIPVersion(msg.IPVersion),
				"sys_uptime":  msg.Uptime,
				"agent_ip":    decodeIP(msg.AgentIP),
				"agent_subid": msg.SubAgentId,
				"seq_number":  sample.Header.SampleSequenceNumber,
			}
			// Decode the source information
			if name := decodeSflowSourceInterface(sample.Header.SourceIdType); name != "" {
				fields[name] = sample.Header.SourceIdValue
			}
			recordFields, err := d.decodeCounterRecords(sample.Records)
			if err != nil {
				return nil, err
			}
			for k, v := range recordFields {
				fields[k] = v
			}
			metrics = append(metrics, metric.New("netflow", tags, fields, t))
		default:
			return nil, fmt.Errorf("unknown record type %T", s)
		}
	}

	return metrics, nil
}

func (d *sflowv5Decoder) decodeFlowRecords(records []sflow.FlowRecord) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	for _, r := range records {
		if r.Data == nil {
			continue
		}
		switch record := r.Data.(type) {
		case sflow.SampledHeader:
			fields["l2_protocol"] = decodeSflowHeaderProtocol(record.Protocol)
			fields["l2_bytes"] = record.FrameLength
			pktfields, err := d.decodeRawHeaderSample(&record)
			if err != nil {
				return nil, err
			}
			for k, v := range pktfields {
				fields[k] = v
			}
		case sflow.SampledEthernet:
			fields["eth_total_len"] = record.Length
			fields["in_src_mac"] = decodeMAC(record.SrcMac)
			fields["out_dst_mac"] = decodeMAC(record.DstMac)
			fields["datalink_frame_type"] = layers.EthernetType(record.EthType & 0x0000ffff).String()
		case sflow.SampledIPv4:
			fields["ipv4_total_len"] = record.Base.Length
			fields["protocol"] = mapL4Proto(uint8(record.Base.Protocol & 0x000000ff))
			fields["src"] = decodeIP(record.Base.SrcIP)
			fields["dst"] = decodeIP(record.Base.DstIP)
			fields["src_port"] = record.Base.SrcPort
			fields["dst_port"] = record.Base.DstPort
			fields["src_tos"] = record.Tos
			fields["tcp_flags"] = decodeTCPFlags([]byte{byte(record.Base.TcpFlags & 0x000000ff)})
		case sflow.SampledIPv6:
			fields["ipv6_total_len"] = record.Base.Length
			fields["protocol"] = mapL4Proto(uint8(record.Base.Protocol & 0x000000ff))
			fields["src"] = decodeIP(record.Base.SrcIP)
			fields["dst"] = decodeIP(record.Base.DstIP)
			fields["src_port"] = record.Base.SrcPort
			fields["dst_port"] = record.Base.DstPort
			fields["tcp_flags"] = decodeTCPFlags([]byte{byte(record.Base.TcpFlags & 0x000000ff)})
		case sflow.ExtendedSwitch:
			fields["vlan_src"] = record.SrcVlan
			fields["vlan_src_priority"] = record.SrcPriority
			fields["vlan_dst"] = record.DstVlan
			fields["vlan_dst_priority"] = record.DstPriority
		case sflow.ExtendedRouter:
			fields["next_hop"] = decodeIP(record.NextHop)
			fields["src_mask"] = record.SrcMaskLen
			fields["dst_mask"] = record.DstMaskLen
		case sflow.ExtendedGateway:
			fields["next_hop"] = decodeIP(record.NextHop)
			fields["bgp_src_as"] = record.SrcAS
			fields["bgp_dst_as"] = record.ASDestinations
			fields["bgp_next_hop"] = decodeIP(record.NextHop)
			fields["bgp_prev_as"] = record.SrcPeerAS
			if len(record.ASPath) > 0 {
				fields["bgp_next_as"] = record.ASPath[0]
			}
		default:
			return nil, fmt.Errorf("unhandled flow record type %T", r.Data)
		}
	}
	return fields, nil
}

func (d *sflowv5Decoder) decodeRawHeaderSample(record *sflow.SampledHeader) (map[string]interface{}, error) {
	var packet gopacket.Packet
	switch record.Protocol {
	case 1: // ETHERNET-ISO8023
		packet = gopacket.NewPacket(record.HeaderData, layers.LayerTypeEthernet, gopacket.Default)
	case 2: // ISO88024-TOKENBUS
		fallthrough
	case 3: // ISO88025-TOKENRING
		fallthrough
	case 4: // FDDI
		fallthrough
	case 5: // FRAME-RELAY
		fallthrough
	case 6: // X25
		fallthrough
	case 7: // PPP
		fallthrough
	case 8: // SMDS
		fallthrough
	case 9: // AAL5
		fallthrough
	case 10: // AAL5-IP
		fallthrough
	case 11: // IPv4
		fallthrough
	case 12: // IPv6
		fallthrough
	case 13: // MPLS
		fallthrough
	default:
		return nil, fmt.Errorf("unhandled protocol %d", record.Protocol)
	}

	fields := make(map[string]interface{})
	for _, pkt := range packet.Layers() {
		switch l := pkt.(type) {
		case *layers.Ethernet:
			fields["in_src_mac"] = l.SrcMAC
			fields["out_dst_mac"] = l.DstMAC
			fields["datalink_frame_type"] = l.EthernetType.String()
			if l.Length > 0 {
				fields["eth_header_len"] = l.Length
			}
		case *layers.Dot1Q:
			fields["vlan_id"] = l.VLANIdentifier
			fields["vlan_priority"] = l.Priority
			fields["vlan_drop_eligible"] = l.DropEligible
		case *layers.IPv4:
			fields["ip_version"] = l.Version
			fields["ipv4_inet_header_len"] = l.IHL
			fields["src_tos"] = l.TOS
			fields["ipv4_total_len"] = l.Length
			fields["ipv4_id"] = l.Id // ?
			fields["ttl"] = l.TTL
			fields["protocol"] = mapL4Proto(uint8(l.Protocol))
			fields["src"] = l.SrcIP.String()
			fields["dst"] = l.DstIP.String()

			flags := []byte("........")
			switch {
			case l.Flags&layers.IPv4EvilBit > 0:
				flags[7] = byte('E')
			case l.Flags&layers.IPv4DontFragment > 0:
				flags[6] = byte('D')
			case l.Flags&layers.IPv4MoreFragments > 0:
				flags[5] = byte('M')
			}
			fields["fragment_flags"] = string(flags)
			fields["fragment_offset"] = l.FragOffset
			fields["ip_total_len"] = l.Length
		case *layers.IPv6:
			fields["ip_version"] = l.Version
			fields["ipv6_total_len"] = l.Length
			fields["ttl"] = l.HopLimit
			fields["protocol"] = mapL4Proto(uint8(l.NextHeader))
			fields["src"] = l.SrcIP.String()
			fields["dst"] = l.DstIP.String()
			fields["ip_total_len"] = l.Length
		case *layers.TCP:
			fields["src_port"] = l.SrcPort
			fields["dst_port"] = l.DstPort
			fields["tcp_seq_number"] = l.Seq
			fields["tcp_ack_number"] = l.Ack
			fields["tcp_window_size"] = l.Window
			fields["tcp_urgent_ptr"] = l.Urgent
			flags := []byte("........")
			switch {
			case l.FIN:
				flags[7] = byte('F')
			case l.SYN:
				flags[6] = byte('S')
			case l.RST:
				flags[5] = byte('R')
			case l.PSH:
				flags[4] = byte('P')
			case l.ACK:
				flags[3] = byte('A')
			case l.URG:
				flags[2] = byte('U')
			case l.ECE:
				flags[1] = byte('E')
			case l.CWR:
				flags[0] = byte('C')
			}
			fields["tcp_flags"] = string(flags)
		case *layers.UDP:
			fields["src_port"] = l.SrcPort
			fields["dst_port"] = l.DstPort
			fields["ip_total_len"] = l.Length
		case *gopacket.Payload:
			// Ignore the payload
		default:
			ltype := int64(pkt.LayerType())
			if !d.warnedFlowRaw[ltype] {
				contents := hex.EncodeToString(pkt.LayerContents())
				payload := hex.EncodeToString(pkt.LayerPayload())
				d.Log.Warnf("Unknown flow raw flow message %s (%d):", pkt.LayerType().String(), pkt.LayerType())
				d.Log.Warnf("  contents: %s", contents)
				d.Log.Warnf("  payload:  %s", payload)

				d.Log.Warn("This message is only printed once.")
			}
			d.warnedFlowRaw[ltype] = true
		}
	}
	return fields, nil
}

func (d *sflowv5Decoder) decodeCounterRecords(records []sflow.CounterRecord) (map[string]interface{}, error) {
	for _, r := range records {
		if r.Data == nil {
			continue
		}
		switch record := r.Data.(type) {
		case sflow.IfCounters:
			fields := map[string]interface{}{
				"interface":                   record.IfIndex,
				"interface_type":              record.IfType,
				"speed":                       record.IfSpeed,
				"in_bytes":                    record.IfInOctets,
				"in_unicast_packets_total":    record.IfInUcastPkts,
				"in_mcast_packets_total":      record.IfInMulticastPkts,
				"in_broadcast_packets_total":  record.IfInBroadcastPkts,
				"in_dropped_packets":          record.IfInDiscards,
				"in_errors":                   record.IfInErrors,
				"in_unknown_protocol":         record.IfInUnknownProtos,
				"out_bytes":                   record.IfOutOctets,
				"out_unicast_packets_total":   record.IfOutUcastPkts,
				"out_mcast_packets_total":     record.IfOutMulticastPkts,
				"out_broadcast_packets_total": record.IfOutBroadcastPkts,
				"out_dropped_packets":         record.IfOutDiscards,
				"out_errors":                  record.IfOutErrors,
				"promiscuous":                 record.IfPromiscuousMode,
			}
			if record.IfStatus == 0 {
				fields["status"] = "down"
			} else {
				fields["status"] = "up"
			}
			return fields, nil
		case sflow.EthernetCounters:
			fields := map[string]interface{}{
				"type":                    "IEEE 802.3",
				"collision_frames_single": record.Dot3StatsSingleCollisionFrames,
				"collision_frames_multi":  record.Dot3StatsMultipleCollisionFrames,
				"collisions_late":         record.Dot3StatsLateCollisions,
				"collisions_excessive":    record.Dot3StatsExcessiveCollisions,
				"deferred":                record.Dot3StatsDeferredTransmissions,
				"errors_alignment":        record.Dot3StatsAlignmentErrors,
				"errors_fcs":              record.Dot3StatsFCSErrors,
				"errors_sqetest":          record.Dot3StatsSQETestErrors,
				"errors_internal_mac_tx":  record.Dot3StatsInternalMacTransmitErrors,
				"errors_internal_mac_rx":  record.Dot3StatsInternalMacReceiveErrors,
				"errors_carrier_sense":    record.Dot3StatsCarrierSenseErrors,
				"errors_frame_too_long":   record.Dot3StatsFrameTooLongs,
				"errors_symbols":          record.Dot3StatsSymbolErrors,
			}
			return fields, nil
		case *sflow.FlowRecordRaw:
			switch r.Header.DataFormat {
			case 1005:
				// Openflow port-name
				if len(record.Data) < 4 {
					return nil, fmt.Errorf("invalid data for raw counter %+v", r)
				}
				fields := map[string]interface{}{
					"port_name": string(record.Data[4:]),
				}
				return fields, nil
			default:
				if !d.warnedCounterRaw[r.Header.DataFormat] {
					data := hex.EncodeToString(record.Data)
					d.Log.Warnf("Unknown counter raw flow message %d: %s", r.Header.DataFormat, data)
					d.Log.Warn("This message is only printed once.")
				}
				d.warnedCounterRaw[r.Header.DataFormat] = true
			}
		default:
			return nil, fmt.Errorf("unhandled counter record type %T", r.Data)
		}
	}
	return nil, nil
}
