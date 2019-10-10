package qsys_qrc

import "encoding/json"

// JSONRPC is the most generic form of a message
type JSONRPC struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int32       `json:"id"`
}

// EngineStatusReplyData is the response from the core for a StatusGet message
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

// EngineStatusReply wraps EngineStatusReplyData with JSONRPC fields
type EngineStatusReply struct {
	Version string                `json:"jsonrpc"`
	ID      int32                 `json:"id"`
	Result  EngineStatusReplyData `json:"result"`
}

// When querying named controls the core sends back an array of these objects
type ControlValue struct {
	Name string `json:"Name"`
	Value json.Token `json:"Value"`
	String string `json:"String"`
}

// ControlGetReply is the message the core responds with to a Control.Get query
type ControlGetReply struct {
	Version string `json:"jsonrpc"`
	ID int32 `json:"id"`
	Result []ControlValue `json:"result"`
	Error struct {
		Code int32 `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}