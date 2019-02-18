package serializers

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/serializers/carbon2"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/plugins/serializers/nowmetric"
	"github.com/influxdata/telegraf/plugins/serializers/splunkmetric"
)

// SerializerOutput is an interface for output plugins that are able to
// serialize telegraf metrics into arbitrary data formats.
type SerializerOutput interface {
	// SetSerializer sets the serializer function for the interface.
	SetSerializer(serializer Serializer)
}

// Serializer is an interface defining functions that a serializer plugin must
// satisfy.
type Serializer interface {
	// Serialize takes a single telegraf metric and turns it into a byte buffer.
	// separate metrics should be separated by a newline, and there should be
	// a newline at the end of the buffer.
	Serialize(metric telegraf.Metric) ([]byte, error)

	// SerializeBatch takes an array of telegraf metric and serializes it into
	// a byte buffer.  This method is not required to be suitable for use with
	// line oriented framing.
	SerializeBatch(metrics []telegraf.Metric) ([]byte, error)
}

// Config is a struct that covers the data types needed for all serializer types,
// and can be used to instantiate _any_ of the serializers.
type Config struct {
	// Dataformat can be one of: influx, graphite, or json
	DataFormat string

	// Support tags in graphite protocol
	GraphiteTagSupport bool

	// Maximum line length in bytes; influx format only
	InfluxMaxLineBytes int

	// Sort field keys, set to true only when debugging as it less performant
	// than unsorted fields; influx format only
	InfluxSortFields bool

	// Support unsigned integer output; influx format only
	InfluxUintSupport bool

	// Prefix to add to all measurements, only supports Graphite
	Prefix string

	// Template for converting telegraf metrics into Graphite
	// only supports Graphite
	Template string

	// Timestamp units to use for JSON formatted output
	TimestampUnits time.Duration

	// Include HEC routing fields for splunkmetric output
	HecRouting bool
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "influx":
		serializer, err = NewInfluxSerializerConfig(config)
	case "graphite":
		serializer, err = NewGraphiteSerializer(config.Prefix, config.Template, config.GraphiteTagSupport)
	case "json":
		serializer, err = NewJsonSerializer(config.TimestampUnits)
	case "splunkmetric":
		serializer, err = NewSplunkmetricSerializer(config.HecRouting)
	case "nowmetric":
		serializer, err = NewNowSerializer()
	case "carbon2":
		serializer, err = NewCarbon2Serializer()
	default:
		err = fmt.Errorf("Invalid data format: %s", config.DataFormat)
	}
	return serializer, err
}

func NewJsonSerializer(timestampUnits time.Duration) (Serializer, error) {
	return json.NewSerializer(timestampUnits)
}

func NewCarbon2Serializer() (Serializer, error) {
	return carbon2.NewSerializer()
}

func NewSplunkmetricSerializer(splunkmetric_hec_routing bool) (Serializer, error) {
	return splunkmetric.NewSerializer(splunkmetric_hec_routing)
}

func NewNowSerializer() (Serializer, error) {
	return nowmetric.NewSerializer()
}

func NewInfluxSerializerConfig(config *Config) (Serializer, error) {
	var sort influx.FieldSortOrder
	if config.InfluxSortFields {
		sort = influx.SortFields
	}

	var typeSupport influx.FieldTypeSupport
	if config.InfluxUintSupport {
		typeSupport = typeSupport + influx.UintSupport
	}

	s := influx.NewSerializer()
	s.SetMaxLineBytes(config.InfluxMaxLineBytes)
	s.SetFieldSortOrder(sort)
	s.SetFieldTypeSupport(typeSupport)
	return s, nil
}

func NewInfluxSerializer() (Serializer, error) {
	return influx.NewSerializer(), nil
}

func NewGraphiteSerializer(prefix, template string, tag_support bool) (Serializer, error) {
	return &graphite.GraphiteSerializer{
		Prefix:     prefix,
		Template:   template,
		TagSupport: tag_support,
	}, nil
}
