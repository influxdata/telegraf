package newrelic

// newrelic.go
import (
	"context"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/newrelic/newrelic-telemetry-sdk-go/cumulative"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
)

// NewRelic nr structure
type NewRelic struct {
	harvestor    *telemetry.Harvester
	dc           *cumulative.DeltaCalculator
	InsightsKey  string `toml:"insights_key"`
	MetricPrefix string `toml:"metric_prefix"`
	savedErrors  []map[string]interface{}
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
	# metric_prefix = ""
`
}

// Connect to the Output
func (nr *NewRelic) Connect() error {
	if nr.InsightsKey == "" {
		return fmt.Errorf("InsightKey is a required for newrelic")
	}
	var err error
	nr.harvestor, err = telemetry.NewHarvester(telemetry.ConfigAPIKey(nr.InsightsKey),
		telemetry.ConfigHarvestPeriod(0),
		func(cfg *telemetry.Config) {
			cfg.Product = "NewRelic-Telgraf-Plugin"
			cfg.ProductVersion = "1.0"
			cfg.ErrorLogger = func(e map[string]interface{}) {
				nr.savedErrors = append(nr.savedErrors, e)
			}
		})
	if err != nil {
		return fmt.Errorf("unable to connect to newrelic %v", err)
	}

	nr.dc = cumulative.NewDeltaCalculator()
	return nil
}

// Close any connections to the Output
func (nr *NewRelic) Close() error {
	nr.harvestor = nil
	nr.dc = nil
	nr.savedErrors = nil
	return nil
}

// Write takes in group of points to be written to the Output
func (nr *NewRelic) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		// create tag map
		tags := make(map[string]interface{})
		for _, tag := range metric.TagList() {
			tags[tag.Key] = tag.Value
		}
		for _, field := range metric.FieldList() {
			var mvalue float64
			var mname string
			if nr.MetricPrefix != "" {
				mname = nr.MetricPrefix + "." + metric.Name() + "." + field.Key
			} else {
				mname = metric.Name() + "." + field.Key
			}
			switch n := field.Value.(type) {
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
			case string:
				// Do not log everytime we encounter string
				// we just skip
			default:
				return fmt.Errorf("Undefined field type: %T", field.Value)
			}

			switch metric.Type() {
			case telegraf.Counter:
				if counter, ok := nr.dc.CountMetric(mname, tags, mvalue, metric.Time()); ok {
					nr.harvestor.RecordMetric(counter)
				}
			default:
				nr.harvestor.RecordMetric(telemetry.Gauge{
					Timestamp:  metric.Time(),
					Value:      mvalue,
					Name:       mname,
					Attributes: tags})
			}
		}
	}
	// By default, the Harvester sends metrics and spans to the New Relic
	// backend every 5 seconds.  You can force data to be sent at any time
	// using HarvestNow.
	nr.harvestor.HarvestNow(context.Background())

	//Check if we encountered errors
	if len(nr.savedErrors) != 0 {
		// we have errors, build error string
		er := fmt.Sprintf("%#v", nr.savedErrors)
		nr.savedErrors = nil
		return fmt.Errorf("unable to harvest metrics  %s", er)
	}
	return nil
}

func init() {
	outputs.Add("newrelic", func() telegraf.Output {
		return &NewRelic{}
	})
}
