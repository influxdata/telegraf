package serializers

import (
	"github.com/influxdata/telegraf"

	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
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
}

// Config is a struct that covers the data types needed for all serializer types,
// and can be used to instantiate _any_ of the serializers.
type Config struct {
	// Dataformat can be one of: influx, graphite
	DataFormat string

	// Prefix to add to all measurements, only supports Graphite
	Prefix string

	// Template for converting telegraf metrics into Graphite
	// only supports Graphite
	Template string
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config, timestamp_units internal.Duration) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "influx":
		serializer, err = NewInfluxSerializer()
	case "graphite":
		serializer, err = NewGraphiteSerializer(config.Prefix, config.Template)
	case "json":
		serializer, err = NewJsonSerializer(timestamp_units)
	}
	return serializer, err
}

func NewJsonSerializer(timestamp_units internal.Duration) (Serializer, error) {
	if timestamp_units.Duration.Seconds() <= 0 {
		timestamp_units.Duration, _ = time.ParseDuration("1s")
	}
	return &json.JsonSerializer{
		TimestampUnits: timestamp_units,
	}, nil
}

func NewInfluxSerializer() (Serializer, error) {
	return &influx.InfluxSerializer{}, nil
}

func NewGraphiteSerializer(prefix, template string) (Serializer, error) {
	return &graphite.GraphiteSerializer{
		Prefix:   prefix,
		Template: template,
	}, nil
}
