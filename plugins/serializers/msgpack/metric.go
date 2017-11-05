package msgpack

import "time"

//go:generate msgp

// Metric is structure to define MessagePack message format
// will be used by msgp code generator
type Metric struct {
	Name   string                 `msg:"name"`
	Time   time.Time              `msg:"time"`
	Tags   map[string]string      `msg:"tags"`
	Fields map[string]interface{} `msg:"fields"`
}
