package signalfx

import (
	"fmt"
	"log"
	"regexp"

	"golang.org/x/net/context"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"
)

type SignalFx struct {
	AuthToken string `toml:"auth_token"`
	UserAgent string `toml:"user_agent"`
	Endpoint  string `toml:"endpoint"`

	sink *sfxclient.HTTPDatapointSink
}

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

var sampleConfig = `
  ## Your organization's SignalFx API access token.
  auth_token = "SuperSecretToken"

  ## Optional HTTP User Agent value; Overrides the default.
  # user_agent = "Telegraf collector"

  ## Optional SignalFX API endpoint value; Overrides the default.
  # endpoint = "https://ingest.signalfx.com/v2/datapoint"
`

func (s *SignalFx) Description() string {
	return "Send Telegraf metrics to SignalFx"
}

func (s *SignalFx) SampleConfig() string {
	return sampleConfig
}

func (s *SignalFx) Connect() error {
	s.sink = sfxclient.NewHTTPDatapointSink()
	s.sink.AuthToken = s.AuthToken
	if len(s.UserAgent) > 0 {
		s.sink.UserAgent = s.UserAgent
	}
	if len(s.Endpoint) > 0 {
		s.sink.Endpoint = s.Endpoint
	}

	return nil
}

func (s *SignalFx) Close() error {
	return nil
}

func (s *SignalFx) Write(metrics []telegraf.Metric) error {
	var datapoints []*datapoint.Datapoint
	for _, metric := range metrics {
		// Sanitize metric name.
		metricName := metric.Name()
		metricName = invalidNameCharRE.ReplaceAllString(metricName, "_")

		// Get a type if it's available, defaulting to Gauge.
		var sfMetricType datapoint.MetricType
		switch metric.Type() {
		case telegraf.Counter:
			sfMetricType = datapoint.Counter
		case telegraf.Gauge:
			sfMetricType = datapoint.Gauge
		default:
			sfMetricType = datapoint.Gauge
		}

		// One SignalFx metric per field.
		for fieldName, fieldValue := range metric.Fields() {
			var sfValue datapoint.Value
			switch fieldValue.(type) {
			case float64:
				sfValue = datapoint.NewFloatValue(fieldValue.(float64))
			case int64:
				sfValue = datapoint.NewIntValue(fieldValue.(int64))
			default:
				log.Printf("Unhandled type %T for field %s\n", fieldValue, fieldName)
				continue
			}

			// Sanitize field name.
			fieldName = invalidNameCharRE.ReplaceAllString(fieldName, "_")

			var sfMetricName string
			if fieldName == "value" {
				sfMetricName = metricName
			} else {
				sfMetricName = fmt.Sprintf("%s.%s", metricName, fieldName)
			}

			datapoint := datapoint.New(sfMetricName, metric.Tags(), sfValue, sfMetricType, metric.Time())
			datapoints = append(datapoints, datapoint)
		}
	}

	ctx := context.Background()
	err := s.sink.AddDatapoints(ctx, datapoints)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	outputs.Add("signalfx", func() telegraf.Output { return &SignalFx{} })
}
