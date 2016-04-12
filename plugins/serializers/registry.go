package serializers

import (
	"github.com/influxdata/telegraf"

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
	// Serialize takes a single telegraf metric and turns it into a string.
	Serialize(metric telegraf.Metric) ([]string, error)
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
func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "influx":
		serializer, err = NewInfluxSerializer()
	case "graphite":
		serializer, err = NewGraphiteSerializer(config.Prefix, config.Template)
	case "json":
		serializer, err = NewJsonSerializer()
	}
	return serializer, err
}

func NewJsonSerializer() (Serializer, error) {
	return &json.JsonSerializer{}, nil
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
