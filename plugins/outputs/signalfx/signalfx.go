package signalfx

import (
	"log"

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
		// One SignalFx metric per field.
		for fieldName, fieldValue := range metric.Fields() {
			var value datapoint.Value
			switch fieldValue.(type) {
			case float64:
				value = datapoint.NewFloatValue(fieldValue.(float64))
			case int64:
				value = datapoint.NewIntValue(fieldValue.(int64))
			default:
				log.Printf("Unhandled type %T for field %s\n", fieldValue, fieldName)
				continue
			}

			metricName := metric.Name() + "." + fieldName
			datapoint := datapoint.New(metricName, metric.Tags(), value, datapoint.Gauge, metric.Time())
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
