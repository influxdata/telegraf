package newrelic

// newrelic.go
import (
	"context"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
)

// NewRelic nr structure
type NewRelic struct {
	harvestor    *telemetry.Harvester
	InsightsKey  string `toml:"insights_key"`
	MetricPrefix string `toml:"metric_prefix"`
}

// Description returns a one-sentence description on the Output
func (nr *NewRelic) Description() string {
	return "Send metrics to New Relic metrics endpoint"
}

// SampleConfig : return  default configuration of the Output
func (nr *NewRelic) SampleConfig() string {
	return `
	## New Relic Insights API key (required)
	insights_key = "insights api key"
	#metric_prefix if defined, prefix's metrics name for easy identification (optional)
	# metric_prefix = "Telegraf_"
`
}

// Connect to the Output
func (nr *NewRelic) Connect() error {
	if nr.InsightsKey == "" {
		return fmt.Errorf("InsightKey is a required for newrelic")
	}
	var err error
	nr.harvestor, err = telemetry.NewHarvester(telemetry.ConfigAPIKey(nr.InsightsKey))
	if err != nil {
		return fmt.Errorf("unable to connect to newrelic %v", err)
	}
	return nil
}

// Close any connections to the Output
func (nr *NewRelic) Close() error {
	nr.harvestor = nil
	return nil
}

// Write takes in group of points to be written to the Output
func (nr *NewRelic) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		// create tag map
		tags := make(map[string]interface{})
		for k, v := range metric.Tags() {
			tags[k] = v
		}
		for k, v := range metric.Fields() {
			var mvalue float64
			var mname string
			if nr.MetricPrefix != "" {
				mname = nr.MetricPrefix + "." + metric.Name() + "." + k
			} else {
				mname = metric.Name() + "." + k
			}
			switch n := v.(type) {
			case int64:
				mvalue = float64(n)
			case uint64:
				mvalue = float64(n)
			case float64:
				mvalue = float64(n)
			case bool:
				mvalue = float64(0)
				if n {
					mvalue = float64(1)
				}
			default:
				return fmt.Errorf("Undefined field type: %T", v)
			}

			nr.harvestor.RecordMetric(telemetry.Gauge{
				Timestamp:  metric.Time(),
				Value:      mvalue,
				Name:       mname,
				Attributes: tags})
		}
	}
	// By default, the Harvester sends metrics and spans to the New Relic
	// backend every 5 seconds.  You can force data to be sent at any time
	// using HarvestNow.
	nr.harvestor.HarvestNow(context.Background())
	return nil
}
func init() {
	outputs.Add("newrelic", func() telegraf.Output {
		return &NewRelic{}
	})
}
