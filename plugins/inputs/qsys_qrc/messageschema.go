package qsys_qrc

type JSONRPC struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int32       `json:"id"`
}

type EngineStatusReplyData struct {
	Platform    string
	State       string
	DesignName  string
	DesignCode  string
	IsRedundant bool
	IsEmulator  bool
	Status      struct {
		Code   int32
		String string
	}
}

type EngineStatusReply struct {
	Version string                `json:"jsonrpc"`
	ID      int32                 `json:"id"`
	Result  EngineStatusReplyData `json:"result"`
}
