package types

import (
	"net"
	"strconv"
)

const (
	AddressTypeIPv4 uint32 = 1 // line: 1383
	AddressTypeIPv6 uint32 = 2 // line: 1384

	IPProtocolTCP uint8 = 6
	IPProtocolUDP uint8 = 17

	metricName = "sflow"
)

var ETypeMap = map[uint16]string{
	0x0800: "IPv4",
	0x86DD: "IPv6",
}

var IPvMap = map[uint32]string{
	1: "IPV4", // line: 1383
	2: "IPV6", // line: 1384
}

type ContainsMetricData interface {
	GetTags() map[string]string
	GetFields() map[string]interface{}
}

// V5Format answers and decoder.Directive capable of decoding sFlow v5 packets in accordance
// with SFlow v5 specification at https://sflow.org/sflow_version_5.txt
type V5Format struct {
	Version        uint32
	AgentAddress   net.IPAddr `tag:"agent_address"`
	SubAgentID     uint32
	SequenceNumber uint32
	Uptime         uint32
	Samples        []Sample
}

type SampleType uint32

const (
	SampleTypeFlowSample         SampleType = 1 // line: 1614
	SampleTypeFlowSampleExpanded SampleType = 3 // line: 1698
)

type SampleData interface{}

type Sample struct {
	SampleType SampleType
	SampleData SampleDataFlowSampleExpanded
}

type SampleDataFlowSampleExpanded struct {
	SequenceNumber  uint32
	SourceIDType    uint32 `tag:"source_id_type"`
	SourceIDIndex   uint32 `tag:"source_id_index"`
	SamplingRate    uint32 `tag:"sampling_rate"`
	SamplePool      uint32
	Drops           uint32
	SampleDirection string // ingress/egress
	InputIfFormat   uint32
	InputIfIndex    uint32
	OutputIfFormat  uint32
	OutputIfIndex   uint32
	FlowRecords     []FlowRecord
}

type FlowFormatType uint32

const (
	FlowFormatTypeRawPacketHeader FlowFormatType = 1 // line: 1938
)

type FlowData ContainsMetricData

type FlowRecord struct {
	FlowFormat FlowFormatType
	FlowData   FlowData
}

type HeaderProtocolType uint32

const (
	HeaderProtocolTypeEthernetISO88023  HeaderProtocolType = 1
	HeaderProtocolTypeISO88024TokenBus  HeaderProtocolType = 2
	HeaderProtocolTypeISO88025TokenRing HeaderProtocolType = 3
	HeaderProtocolTypeFDDI              HeaderProtocolType = 4
	HeaderProtocolTypeFrameRelay        HeaderProtocolType = 5
	HeaderProtocolTypeX25               HeaderProtocolType = 6
	HeaderProtocolTypePPP               HeaderProtocolType = 7
	HeaderProtocolTypeSMDS              HeaderProtocolType = 8
	HeaderProtocolTypeAAL5              HeaderProtocolType = 9
	HeaderProtocolTypeAAL5IP            HeaderProtocolType = 10 /* e.g. Cisco AAL5 mux */
	HeaderProtocolTypeIPv4              HeaderProtocolType = 11
	HeaderProtocolTypeIPv6              HeaderProtocolType = 12
	HeaderProtocolTypeMPLS              HeaderProtocolType = 13
	HeaderProtocolTypePOS               HeaderProtocolType = 14 /* RFC 1662, 2615 */
)

var HeaderProtocolMap = map[HeaderProtocolType]string{
	HeaderProtocolTypeEthernetISO88023: "ETHERNET-ISO88023", // line: 1920
}

type Header ContainsMetricData

type RawPacketHeaderFlowData struct {
	HeaderProtocol HeaderProtocolType `tag:"header_protocol"`
	FrameLength    uint32             `field:"frame_length"`
	Bytes          uint32             `field:"bytes"`
	StrippedOctets uint32
	HeaderLength   uint32 `field:"header_length"`
	Header         Header // consider making this an interface if we're adding more types
}

func (h RawPacketHeaderFlowData) GetTags() map[string]string {
	t := h.Header.GetTags()
	t["header_protocol"] = HeaderProtocolMap[h.HeaderProtocol]
	return t
}
func (h RawPacketHeaderFlowData) GetFields() map[string]interface{} {
	f := h.Header.GetFields()
	f["bytes"] = h.Bytes
	f["frame_length"] = h.FrameLength
	f["header_length"] = h.HeaderLength
	return f
}

type IPHeader ContainsMetricData

type EthHeader struct {
	DestinationMAC        [6]byte `tag:"dst_mac"`
	SourceMAC             [6]byte `tag:"src_mac"`
	TagProtocolIdentifier uint16
	TagControlInformation uint16
	EtherTypeCode         uint16
	EtherType             string `tag:"ether_type"`
	IPHeader              IPHeader
}

func (h EthHeader) GetTags() map[string]string {
	t := h.IPHeader.GetTags()
	t["src_mac"] = net.HardwareAddr(h.SourceMAC[:]).String()
	t["dst_mac"] = net.HardwareAddr(h.DestinationMAC[:]).String()
	t["ether_type"] = h.EtherType
	return t
}
func (h EthHeader) GetFields() map[string]interface{} {
	return h.IPHeader.GetFields()
}

type ProtocolHeader ContainsMetricData

// https://en.wikipedia.org/wiki/IPv4#Header
type IPV4Header struct {
	Version              uint8  // 4 bit
	InternetHeaderLength uint8  // 4 bit
	DSCP                 uint8  `tag:"ip_dscp"` // Differentiated Services Code Point: 6 bit
	ECN                  uint8  `tag:"ip_ecn"`  // Explicit Congestion Notification: 2 bit
	TotalLength          uint16 `field:"ip_total_length"`
	Identification       uint16
	Flags                uint8
	FragmentOffset       uint16
	TTL                  uint8
	Protocol             uint8 // https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
	HeaderChecksum       uint16
	SourceIP             [4]byte
	DestIP               [4]byte
	ProtocolHeader       ProtocolHeader
}

func (h IPV4Header) GetTags() map[string]string {
	var t map[string]string
	if h.ProtocolHeader != nil {
		t = h.ProtocolHeader.GetTags()
	} else {
		t = map[string]string{}
	}
	t["ip_dscp"] = strconv.FormatUint(uint64(h.DSCP), 10)
	t["ip_ecn"] = strconv.FormatUint(uint64(h.ECN), 10)
	t["src_ip"] = net.IP(h.SourceIP[:]).String()
	t["dst_ip"] = net.IP(h.DestIP[:]).String()
	return t
}
func (h IPV4Header) GetFields() map[string]interface{} {
	var f map[string]interface{}
	if h.ProtocolHeader != nil {
		f = h.ProtocolHeader.GetFields()
	} else {
		f = map[string]interface{}{}
	}
	f["ip_flags"] = h.Flags
	f["ip_fragment_offset"] = h.FragmentOffset
	f["ip_total_length"] = h.TotalLength
	f["ip_ttl"] = h.TTL
	return f
}

// https://en.wikipedia.org/wiki/IPv6_packet
type IPV6Header struct {
	DSCP            uint8 `tag:"ip_dscp"`
	ECN             uint8 `tag:"ip_ecn"`
	PayloadLength   uint16
	NextHeaderProto uint8 // tcp/udp?
	HopLimit        uint8
	SourceIP        [16]byte
	DestIP          [16]byte
	ProtocolHeader  ProtocolHeader
}

func (h IPV6Header) GetTags() map[string]string {
	var t map[string]string
	if h.ProtocolHeader != nil {
		t = h.ProtocolHeader.GetTags()
	} else {
		t = map[string]string{}
	}
	t["ip_dscp"] = strconv.FormatUint(uint64(h.DSCP), 10)
	t["ip_ecn"] = strconv.FormatUint(uint64(h.ECN), 10)
	t["src_ip"] = net.IP(h.SourceIP[:]).String()
	t["dst_ip"] = net.IP(h.DestIP[:]).String()
	return t
}
func (h IPV6Header) GetFields() map[string]interface{} {
	var f map[string]interface{}
	if h.ProtocolHeader != nil {
		f = h.ProtocolHeader.GetFields()
	} else {
		f = map[string]interface{}{}
	}
	// f["ip_total_length"] = h.PayloadLength // is this the same?
	return f
}

// https://en.wikipedia.org/wiki/Transmission_Control_Protocol
type TCPHeader struct {
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

func (h TCPHeader) GetTags() map[string]string {
	t := map[string]string{
		"dst_port": strconv.FormatUint(uint64(h.DestinationPort), 10),
		"src_port": strconv.FormatUint(uint64(h.SourcePort), 10),
	}
	return t
}
func (h TCPHeader) GetFields() map[string]interface{} {
	return map[string]interface{}{
		"tcp_header_length":  h.TCPHeaderLength,
		"tcp_urgent_pointer": h.TCPUrgentPointer,
		"tcp_window_size":    h.TCPWindowSize,
	}
}

type UDPHeader struct {
	SourcePort      uint16
	DestinationPort uint16
	UDPLength       uint16
	Checksum        uint16
}

func (h UDPHeader) GetTags() map[string]string {
	t := map[string]string{
		"dst_port": strconv.FormatUint(uint64(h.DestinationPort), 10),
		"src_port": strconv.FormatUint(uint64(h.SourcePort), 10),
	}
	return t
}
func (h UDPHeader) GetFields() map[string]interface{} {
	return map[string]interface{}{
		"udp_length": h.UDPLength,
	}
}
