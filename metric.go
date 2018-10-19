package telegraf

import (
	"time"
)

// ValueType is an enumeration of metric types that represent a simple value.
type ValueType int

// Possible values for the ValueType enum.
const (
	_ ValueType = iota
	Counter
	Gauge
	Untyped
	Summary
	Histogram
)

type Tag struct {
	Key   string
	Value string
}

type Field struct {
	Key   string
	Value interface{}
}

type Metric interface {
	// Getting data structure functions
	Name() string
	Tags() map[string]string
	TagList() []*Tag
	Fields() map[string]interface{}
	FieldList() []*Field
	Time() time.Time
	Type() ValueType

	// Name functions
	SetName(name string)
	AddPrefix(prefix string)
	AddSuffix(suffix string)

	// Tag functions
	GetTag(key string) (string, bool)
	HasTag(key string) bool
	AddTag(key, value string)
	RemoveTag(key string)

	// Field functions
	GetField(key string) (interface{}, bool)
	HasField(key string) bool
	AddField(key string, value interface{})
	RemoveField(key string)

	SetTime(t time.Time)

	// HashID returns an unique identifier for the series.
	HashID() uint64

	// Copy returns a deep copy of the Metric.
	Copy() Metric

	// Accept marks the metric as processed successfully and written to an
	// output.
	Accept()

	// Reject marks the metric as processed unsuccessfully.
	Reject()

	// Remove marks the metric as processed without being written to any
	// output.
	Remove()

	// Mark Metric as an aggregate
	SetAggregate(bool)
	IsAggregate() bool
}
