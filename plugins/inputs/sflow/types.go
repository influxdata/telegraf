package sflow

import (
	"net"
	"strconv"
)

const (
	IPProtocolTCP uint8 = 6
	IPProtocolUDP uint8 = 17
)

var ETypeMap = map[uint16]string{
	0x0800: "IPv4",
	0x86DD: "IPv6",
}

type ContainsMetricData interface {
	GetTags() map[string]string
	GetFields() map[string]interface{}
}

// V5Format answers and decoder.Directive capable of decoding sFlow v5 packets in accordance
// with SFlow v5 specification at https://sflow.org/sflow_version_5.txt
type V5Format struct {
	Version        uint32
	AgentAddress   net.IPAddr
	SubAgentID     uint32
	SequenceNumber uint32
	Uptime         uint32
	Samples        []Sample
}

type SampleType uint32

const (
	SampleTypeFlowSample         SampleType = 1 // sflow_version_5.txt line: 1614
	SampleTypeFlowSampleExpanded SampleType = 3 // sflow_version_5.txt line: 1698
)

type SampleData interface{}

type Sample struct {
	SampleType SampleType
	SampleData SampleDataFlowSampleExpanded
}

type SampleDataFlowSampleExpanded struct {
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
	FlowRecords     []FlowRecord
}

type FlowFormatType uint32

const (
	FlowFormatTypeRawPacketHeader FlowFormatType = 1 // sflow_version_5.txt line: 1938
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
	HeaderProtocolTypeEthernetISO88023: "ETHERNET-ISO88023", // sflow_version_5.txt line: 1920
}

type Header ContainsMetricData

type RawPacketHeaderFlowData struct {
	HeaderProtocol HeaderProtocolType
	FrameLength    uint32
	Bytes          uint32
	StrippedOctets uint32
	HeaderLength   uint32
	Header         Header
}

func (h RawPacketHeaderFlowData) GetTags() map[string]string {
	var t map[string]string
	if h.Header != nil {
		t = h.Header.GetTags()
	} else {
		t = map[string]string{}
	}
	t["header_protocol"] = HeaderProtocolMap[h.HeaderProtocol]
	return t
}
func (h RawPacketHeaderFlowData) GetFields() map[string]interface{} {
	var f map[string]interface{}
	if h.Header != nil {
		f = h.Header.GetFields()
	} else {
		f = map[string]interface{}{}
	}
	f["bytes"] = h.Bytes
	f["frame_length"] = h.FrameLength
	f["header_length"] = h.HeaderLength
	return f
}

type IPHeader ContainsMetricData

type EthHeader struct {
	DestinationMAC        [6]byte
	SourceMAC             [6]byte
	TagProtocolIdentifier uint16
	TagControlInformation uint16
	EtherTypeCode         uint16
	EtherType             string
	IPHeader              IPHeader
}

func (h EthHeader) GetTags() map[string]string {
	var t map[string]string
	if h.IPHeader != nil {
		t = h.IPHeader.GetTags()
	} else {
		t = map[string]string{}
	}
	t["src_mac"] = net.HardwareAddr(h.SourceMAC[:]).String()
	t["dst_mac"] = net.HardwareAddr(h.DestinationMAC[:]).String()
	t["ether_type"] = h.EtherType
	return t
}
func (h EthHeader) GetFields() map[string]interface{} {
	if h.IPHeader != nil {
		return h.IPHeader.GetFields()
	}
	return map[string]interface{}{}
}

type ProtocolHeader ContainsMetricData

// https://en.wikipedia.org/wiki/IPv4#Header
type IPV4Header struct {
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
	ProtocolHeader       ProtocolHeader
}

func (h IPV4Header) GetTags() map[string]string {
	var t map[string]string
	if h.ProtocolHeader != nil {
		t = h.ProtocolHeader.GetTags()
	} else {
		t = map[string]string{}
	}
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
	f["ip_dscp"] = strconv.FormatUint(uint64(h.DSCP), 10)
	f["ip_ecn"] = strconv.FormatUint(uint64(h.ECN), 10)
	f["ip_flags"] = h.Flags
	f["ip_fragment_offset"] = h.FragmentOffset
	f["ip_total_length"] = h.TotalLength
	f["ip_ttl"] = h.TTL
	return f
}

// https://en.wikipedia.org/wiki/IPv6_packet
type IPV6Header struct {
	DSCP            uint8
	ECN             uint8
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
	f["ip_dscp"] = strconv.FormatUint(uint64(h.DSCP), 10)
	f["ip_ecn"] = strconv.FormatUint(uint64(h.ECN), 10)
	f["payload_length"] = h.PayloadLength
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
