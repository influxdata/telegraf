package telegraf

import (
	"time"

	// TODO remove
	"github.com/influxdata/influxdb/client/v2"
)

// ValueType is an enumeration of metric types that represent a simple value.
type ValueType int

// Possible values for the ValueType enum.
const (
	_ ValueType = iota
	Counter
	Gauge
	Untyped
)

type Metric interface {
	Serialize() []byte
	String() string // convenience function for string(Serialize())
	Copy() Metric
	// Split will attempt to return multiple metrics with the same timestamp
	// whose string representations are no longer than maxSize.
	// Metrics with a single field may exceed the requested size.
	Split(maxSize int) []Metric

	// Tag functions
	HasTag(key string) bool
	AddTag(key, value string)
	RemoveTag(key string)

	// Field functions
	HasField(key string) bool
	AddField(key string, value interface{})
	RemoveField(key string) error

	// Name functions
	SetName(name string)
	SetPrefix(prefix string)
	SetSuffix(suffix string)

	// Getting data structure functions
	Name() string
	Tags() map[string]string
	Fields() map[string]interface{}
	Time() time.Time
	UnixNano() int64
	Type() ValueType
	Len() int // returns the length of the serialized metric, including newline
	HashID() uint64

	// aggregator things:
	SetAggregate(bool)
	IsAggregate() bool

	// Point returns a influxdb client.Point object
	// TODO remove this function
	Point() *client.Point
}
