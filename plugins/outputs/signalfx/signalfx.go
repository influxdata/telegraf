package http

import (
	"context"
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"
)

var sampleConfig = `
  api_token = xxx
`

// SignalFX knows how to send metrics to the signalfx cloud
type SignalFX struct {
	APIToken string `toml:"api_token"`

	sink *sfxclient.HTTPSink
}

// Connect initializes the connection to the signalfx cloud
func (s *SignalFX) Connect() error {
	s.sink = sfxclient.NewHTTPSink()
	s.sink.AuthToken = s.APIToken

	return nil
}

// Close - NOOP since the `sink` does not expose a way to close the connection
func (s *SignalFX) Close() error {
	return nil
}

// Description returns a string describing this plugins functionality
func (s *SignalFX) Description() string {
	return "A plugin that can transmit metrics to SignalFX"
}

// SampleConfig returns the sample configuration for this plugin
func (s *SignalFX) SampleConfig() string {
	return sampleConfig
}

// Write sends the metrics to the signalfx server
func (s *SignalFX) Write(metrics []telegraf.Metric) error {
	d := []*datapoint.Datapoint{}
	for _, m := range metrics {
		d = append(d, telegrafMetricToSignalFXDatapoints(m)...)
	}
	if len(d) <= 0 {
		return nil
	}
	err := s.sink.AddDatapoints(context.TODO(), d)
	if err != nil {
		log.Println("E! Failed to add datapoints to signalfx sink:", err)
		return err
	}
	return nil
}

func telegrafMetricToSignalFXDatapoints(m telegraf.Metric) []*datapoint.Datapoint {
	fields := m.FieldList()
	datapoints := make([]*datapoint.Datapoint, 0, len(fields))

	sfxType := telegrafTypeToSignalFXType(m.Type())
	sfxTime := m.Time()

	for _, field := range fields {
		sfxName := m.Name()
		if field.Key != "value" {
			sfxName = fmt.Sprintf("%s.%s", sfxName, field.Key)
		}

		sfxValue, err := datapoint.CastMetricValue(field.Value)
		if err != nil {
			// Intentionally emitting a debug level msg (not error) since
			// plugins can often be naughty and emit fields that are strings
			// (i.e. not numeric/chartable). E.g. system.uptime_format
			log.Println("D! Failed to cast value for", field.Key)
			continue
		}
		d := datapoint.New(sfxName, m.Tags(), sfxValue, sfxType, sfxTime)
		datapoints = append(datapoints, d)
	}
	return datapoints
}

func telegrafTypeToSignalFXType(t telegraf.ValueType) datapoint.MetricType {
	switch t {
	case telegraf.Counter:
		return datapoint.Counter
	case telegraf.Gauge:
		return datapoint.Gauge
	default:
		// All other telegraf types are sent as gauges (since they do not map
		// on to any of the SignalFX types)
		return datapoint.Gauge
	}
}

func init() {
	outputs.Add("signalfx", func() telegraf.Output {
		return &SignalFX{}
	})
}
