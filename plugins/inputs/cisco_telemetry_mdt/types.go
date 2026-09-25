package cisco_telemetry_mdt

type nxPayloadXfromStructure struct {
	Name string `json:"Name"`
	Prop []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"prop"`
}

// TCP Dialout telemetry framing header
type header struct {
	MsgType       uint16
	MsgEncap      uint16
	MsgHdrVersion uint16
	MsgFlags      uint16
	MsgLen        uint32
}
