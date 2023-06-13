package telegraf

import (
	"strings"
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

type Namer struct {
	Name   string
	Prefix string
	Suffix string
}

func (n *Namer) Value() string {
	return n.Prefix + n.Name + n.Suffix
}

func (n *Namer) SetName(name string) {
	n.Name = name
}

// SetPrefix on name, if the prefix does not start with a `+` sign,
// it directly overwrites the prefix setting of the metric, otherwise add it.
func (n *Namer) SetPrefix(prefix string) {
	if strings.HasPrefix(prefix, "+") {
		n.Prefix = prefix[1:] + n.Prefix
	} else {
		n.Prefix = prefix
	}
}

// SetSuffix on name, if the suffix does not start with a `+` sign,
// it directly overwrites the suffix setting of the metric, otherwise add it.
func (n *Namer) SetSuffix(suffix string) {
	if strings.HasPrefix(suffix, "+") {
		n.Suffix += suffix[1:]
	} else {
		n.Suffix = suffix
	}
}

func (n *Namer) Copy() *Namer {
	return &Namer{
		Name:   n.Name,
		Prefix: n.Prefix,
		Suffix: n.Suffix,
	}
}

// Tag represents a single tag key and value.
type Tag struct {
	Key   string
	Value string
}

// Field represents a single field key and value.
type Field struct {
	Key   string
	Value interface{}
}

// Metric is the type of data that is processed by Telegraf.  Input plugins,
// and to a lesser degree, Processor and Aggregator plugins create new Metrics
// and Output plugins write them.
//
//nolint:interfacebloat // conditionally allow to contain more methods
type Metric interface {
	// Name is the primary identifier for the Metric and corresponds to the
	// measurement in the InfluxDB data model.
	// The final name should be m.Name().Value().
	//
	// This method is deprecated, use Namer().Value() instead.
	Name() string

	// Namer is used to fetch the primary identifier for the Metric
	// and corresponds to the measurement in the InfluxDB data model.
	// The final name should be m.Namer().Value().
	Namer() *Namer

	// Tags returns the tags as a map.  This method is deprecated, use TagList instead.
	Tags() map[string]string

	// TagList returns the tags as a slice ordered by the tag key in lexical
	// bytewise ascending order.  The returned value should not be modified,
	// use the AddTag or RemoveTag methods instead.
	TagList() []*Tag

	// Fields returns the fields as a map.  This method is deprecated, use FieldList instead.
	Fields() map[string]interface{}

	// FieldList returns the fields as a slice in an undefined order.  The
	// returned value should not be modified, use the AddField or RemoveField
	// methods instead.
	FieldList() []*Field

	// Time returns the timestamp of the metric.
	Time() time.Time

	// Type returns a general type for the entire metric that describes how you
	// might interpret, aggregate the values. Used by prometheus and statsd.
	Type() ValueType

	// SetName sets the metric name
	// equivalent to m.Namer().SetName(nameOverride).
	//
	// This method is deprecated, use Namer().SetName instead.
	SetName(name string)

	// SetPrefix sets a string to the front of the metric name.  It is
	// equivalent to m.Namer().SetPrefix(prefix). It is different with
	// AddPrefix sets a string to the front of the prefixed metric name,
	// which equivalent to m.Namer().SetPrefix("+" + prefix).
	//
	// This method is deprecated, use Namer().SetPrefix instead.
	SetPrefix(prefix string)

	// SetSuffix sets a string to the back of the metric name.  It is
	// equivalent to m.Namer().SetSuffix(suffix). It is different with
	// AddSuffix sets a string to the back of the suffixed metric name,
	// which equivalent to m.Namer().SetSuffix("+" + suffix).
	//
	// This method is deprecated, use Namer().SetSuffix instead.
	SetSuffix(suffix string)

	// GetTag returns the value of a tag and a boolean to indicate if it was set.
	GetTag(key string) (string, bool)

	// HasTag returns true if the tag is set on the Metric.
	HasTag(key string) bool

	// AddTag sets the tag on the Metric.  If the Metric already has the tag
	// set then the current value is replaced.
	AddTag(key, value string)

	// RemoveTag removes the tag if it is set.
	RemoveTag(key string)

	// GetField returns the value of a field and a boolean to indicate if it was set.
	GetField(key string) (interface{}, bool)

	// HasField returns true if the field is set on the Metric.
	HasField(key string) bool

	// AddField sets the field on the Metric.  If the Metric already has the field
	// set then the current value is replaced.
	AddField(key string, value interface{})

	// RemoveField removes the tag if it is set.
	RemoveField(key string)

	// SetTime sets the timestamp of the Metric.
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

	// Drop marks the metric as processed successfully without being written
	// to any output.
	Drop()
}

// TemplateMetric is an interface to use in templates (e.g text/template)
// to generate complex strings from metric properties
// e.g. '{{.Neasurement}}-{{.Tag "foo"}}-{{.Field "bar"}}'
type TemplateMetric interface {
	Name() string
	Tag(key string) string
	Field(key string) interface{}
}
