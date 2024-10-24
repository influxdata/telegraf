package sflow

import (
	"net"
	"strconv"
)

const (
	ipProtocolTCP uint8 = 6
	ipProtocolUDP uint8 = 17
)

var eTypeMap = map[uint16]string{
	0x0800: "IPv4",
	0x86DD: "IPv6",
}

type containsMetricData interface {
	getTags() map[string]string
	getFields() map[string]interface{}
}

// v5Format answers and decoder.Directive capable of decoding sFlow v5 packets in accordance
// with SFlow v5 specification at https://sflow.org/sflow_version_5.txt
type v5Format struct {
	Version        uint32
	AgentAddress   net.IPAddr
	SubAgentID     uint32
	SequenceNumber uint32
	Uptime         uint32
	Samples        []sample
}

type sampleType uint32

const (
	sampleTypeFlowSample         sampleType = 1 // sflow_version_5.txt line: 1614
	sampleTypeFlowSampleExpanded sampleType = 3 // sflow_version_5.txt line: 1698
)

type sample struct {
	SampleType sampleType
	SampleData sampleDataFlowSampleExpanded
}

type sampleDataFlowSampleExpanded struct {
	SequenceNumber  uint32
	SourceIDType    uint32
	SourceIDIndex   uint32
	SamplingRate    uint32
	SamplePool      uint32
	Drops           uint32
	SampleDirection string // ingress/egress
	InputIfFormat   uint32
	InputIfIndex    uint32
	OutputIfFormat  uint32
	OutputIfIndex   uint32
	FlowRecords     []flowRecord
}

type flowFormatType uint32

const (
	flowFormatTypeRawPacketHeader flowFormatType = 1 // sflow_version_5.txt line: 1938
)

type flowData containsMetricData

type flowRecord struct {
	FlowFormat flowFormatType
	FlowData   flowData
}

type headerProtocolType uint32

const (
	headerProtocolTypeEthernetISO88023  headerProtocolType = 1
	headerProtocolTypeISO88024TokenBus  headerProtocolType = 2
	headerProtocolTypeISO88025TokenRing headerProtocolType = 3
	headerProtocolTypeFDDI              headerProtocolType = 4
	headerProtocolTypeFrameRelay        headerProtocolType = 5
	headerProtocolTypeX25               headerProtocolType = 6
	headerProtocolTypePPP               headerProtocolType = 7
	headerProtocolTypeSMDS              headerProtocolType = 8
	headerProtocolTypeAAL5              headerProtocolType = 9
	headerProtocolTypeAAL5IP            headerProtocolType = 10 /* e.g. Cisco AAL5 mux */
	headerProtocolTypeIPv4              headerProtocolType = 11
	headerProtocolTypeIPv6              headerProtocolType = 12
	headerProtocolTypeMPLS              headerProtocolType = 13
	headerProtocolTypePOS               headerProtocolType = 14 /* RFC 1662, 2615 */
)

var headerProtocolMap = map[headerProtocolType]string{
	headerProtocolTypeEthernetISO88023: "ETHERNET-ISO88023", // sflow_version_5.txt line: 1920
}

type header containsMetricData

type rawPacketHeaderFlowData struct {
	HeaderProtocol headerProtocolType
	FrameLength    uint32
	Bytes          uint32
	StrippedOctets uint32
	HeaderLength   uint32
	Header         header
}

func (h rawPacketHeaderFlowData) getTags() map[string]string {
	var t map[string]string
	if h.Header != nil {
		t = h.Header.getTags()
	} else {
		t = make(map[string]string, 1)
	}
	t["header_protocol"] = headerProtocolMap[h.HeaderProtocol]
	return t
}
func (h rawPacketHeaderFlowData) getFields() map[string]interface{} {
	var f map[string]interface{}
	if h.Header != nil {
		f = h.Header.getFields()
	} else {
		f = make(map[string]interface{}, 3)
	}
	f["bytes"] = h.Bytes
	f["frame_length"] = h.FrameLength
	f["header_length"] = h.HeaderLength
	return f
}

type ipHeader containsMetricData

type ethHeader struct {
	DestinationMAC        [6]byte
	SourceMAC             [6]byte
	TagProtocolIdentifier uint16
	TagControlInformation uint16
	EtherTypeCode         uint16
	EtherType             string
	IPHeader              ipHeader
}

func (h ethHeader) getTags() map[string]string {
	var t map[string]string
	if h.IPHeader != nil {
		t = h.IPHeader.getTags()
	} else {
		t = make(map[string]string, 3)
	}
	t["src_mac"] = net.HardwareAddr(h.SourceMAC[:]).String()
	t["dst_mac"] = net.HardwareAddr(h.DestinationMAC[:]).String()
	t["ether_type"] = h.EtherType
	return t
}
func (h ethHeader) getFields() map[string]interface{} {
	if h.IPHeader != nil {
		return h.IPHeader.getFields()
	}
	return make(map[string]interface{})
}

type protocolHeader containsMetricData

// https://en.wikipedia.org/wiki/IPv4#Header
type ipV4Header struct {
	Version              uint8 // 4 bit
	InternetHeaderLength uint8 // 4 bit
	DSCP                 uint8
	ECN                  uint8
	TotalLength          uint16
	Identification       uint16
	Flags                uint8
	FragmentOffset       uint16
	TTL                  uint8
	Protocol             uint8 // https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
	HeaderChecksum       uint16
	SourceIP             [4]byte
	DestIP               [4]byte
	ProtocolHeader       protocolHeader
}

func (h ipV4Header) getTags() map[string]string {
	var t map[string]string
	if h.ProtocolHeader != nil {
		t = h.ProtocolHeader.getTags()
	} else {
		t = make(map[string]string, 2)
	}
	t["src_ip"] = net.IP(h.SourceIP[:]).String()
	t["dst_ip"] = net.IP(h.DestIP[:]).String()
	return t
}
func (h ipV4Header) getFields() map[string]interface{} {
	var f map[string]interface{}
	if h.ProtocolHeader != nil {
		f = h.ProtocolHeader.getFields()
	} else {
		f = make(map[string]interface{}, 6)
	}
	f["ip_dscp"] = strconv.FormatUint(uint64(h.DSCP), 10)
	f["ip_ecn"] = strconv.FormatUint(uint64(h.ECN), 10)
	f["ip_flags"] = h.Flags
	f["ip_fragment_offset"] = h.FragmentOffset
	f["ip_total_length"] = h.TotalLength
	f["ip_ttl"] = h.TTL
	return f
}

// https://en.wikipedia.org/wiki/IPv6_packet
type ipV6Header struct {
	DSCP            uint8
	ECN             uint8
	PayloadLength   uint16
	NextHeaderProto uint8 // tcp/udp?
	HopLimit        uint8
	SourceIP        [16]byte
	DestIP          [16]byte
	ProtocolHeader  protocolHeader
}

func (h ipV6Header) getTags() map[string]string {
	var t map[string]string
	if h.ProtocolHeader != nil {
		t = h.ProtocolHeader.getTags()
	} else {
		t = make(map[string]string, 2)
	}
	t["src_ip"] = net.IP(h.SourceIP[:]).String()
	t["dst_ip"] = net.IP(h.DestIP[:]).String()
	return t
}
func (h ipV6Header) getFields() map[string]interface{} {
	var f map[string]interface{}
	if h.ProtocolHeader != nil {
		f = h.ProtocolHeader.getFields()
	} else {
		f = make(map[string]interface{}, 3)
	}
	f["ip_dscp"] = strconv.FormatUint(uint64(h.DSCP), 10)
	f["ip_ecn"] = strconv.FormatUint(uint64(h.ECN), 10)
	f["payload_length"] = h.PayloadLength
	return f
}

// https://en.wikipedia.org/wiki/Transmission_Control_Protocol
type tcpHeader struct {
	SourcePort       uint16
	DestinationPort  uint16
	Sequence         uint32
	AckNumber        uint32
	TCPHeaderLength  uint8
	Flags            uint16
	TCPWindowSize    uint16
	Checksum         uint16
	TCPUrgentPointer uint16
}

func (h tcpHeader) getTags() map[string]string {
	t := map[string]string{
		"dst_port": strconv.FormatUint(uint64(h.DestinationPort), 10),
		"src_port": strconv.FormatUint(uint64(h.SourcePort), 10),
	}
	return t
}
func (h tcpHeader) getFields() map[string]interface{} {
	return map[string]interface{}{
		"tcp_header_length":  h.TCPHeaderLength,
		"tcp_urgent_pointer": h.TCPUrgentPointer,
		"tcp_window_size":    h.TCPWindowSize,
	}
}

type udpHeader struct {
	SourcePort      uint16
	DestinationPort uint16
	UDPLength       uint16
	Checksum        uint16
}

func (h udpHeader) getTags() map[string]string {
	t := map[string]string{
		"dst_port": strconv.FormatUint(uint64(h.DestinationPort), 10),
		"src_port": strconv.FormatUint(uint64(h.SourcePort), 10),
	}
	return t
}
func (h udpHeader) getFields() map[string]interface{} {
	return map[string]interface{}{
		"udp_length": h.UDPLength,
	}
}
