package serializers

import (
	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
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
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "influx":
		serializer, err = NewInfluxSerializer()
	case "graphite":
		serializer, err = NewGraphiteSerializer(config.Prefix)
	}
	return serializer, err
}

func NewInfluxSerializer() (Serializer, error) {
	return &influx.InfluxSerializer{}, nil
}

func NewGraphiteSerializer(prefix string) (Serializer, error) {
	return &graphite.GraphiteSerializer{
		Prefix: prefix,
	}, nil
}
