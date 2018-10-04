package netflow

type V5Header struct {
	Version          uint16
	Count            uint16
	SysUptime        uint32
	UnixSeconds      uint32
	UnixNanoSeconds  uint32
	FlowSequence     uint32
	EngineType       uint8
	EngineID         uint8
	SamplingInterval uint16
}

type V5FlowRecord struct {
	SrcAddr  uint32
	DstAddr  uint32
	Nexthop  uint32
	Input    uint16
	Output   uint16
	Packets  uint32
	Bytes    uint32
	First    uint32
	Last     uint32
	SrcPort  uint16
	DstPort  uint16
	Padding1 uint8
	TCPFlags uint8
	Protocol uint8
	ToS      uint8
	SrcAS    uint16
	DstAS    uint16
	SrcMask  uint8
	DstMask  uint8
	Padding2 uint16
}
