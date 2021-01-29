package newrelic

// newrelic.go
import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/newrelic/newrelic-telemetry-sdk-go/cumulative"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
)

// NewRelic nr structure
type NewRelic struct {
	InsightsKey  string            `toml:"insights_key"`
	MetricPrefix string            `toml:"metric_prefix"`
	Timeout      internal.Duration `toml:"timeout"`

	harvestor   *telemetry.Harvester
	dc          *cumulative.DeltaCalculator
	savedErrors map[int]interface{}
	errorCount  int
	Client      http.Client `toml:"-" json:"-"`
}

// Description returns a one-sentence description on the Output
func (nr *NewRelic) Description() string {
	return "Send metrics to New Relic metrics endpoint"
}

// SampleConfig : return  default configuration of the Output
func (nr *NewRelic) SampleConfig() string {
	return `
  ## New Relic Insights API key
  insights_key = "insights api key"

  ## Prefix to add to add to metric name for easy identification.
  # metric_prefix = ""

  ## Timeout for writes to the New Relic API.
  # timeout = "15s"
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
			cfg.Product = "NewRelic-Telegraf-Plugin"
			cfg.ProductVersion = "1.0"
			cfg.HarvestTimeout = nr.Timeout.Duration
			cfg.Client = &nr.Client
			cfg.ErrorLogger = func(e map[string]interface{}) {
				var errorString string
				for k, v := range e {
					errorString += fmt.Sprintf("%s = %s ", k, v)
				}
				nr.errorCount++
				nr.savedErrors[nr.errorCount] = errorString
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
	nr.errorCount = 0
	nr.Client.CloseIdleConnections()
	return nil
}

// Write takes in group of points to be written to the Output
func (nr *NewRelic) Write(metrics []telegraf.Metric) error {
	nr.errorCount = 0
	nr.savedErrors = make(map[int]interface{})

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
				continue
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
	if nr.errorCount != 0 {
		return fmt.Errorf("unable to harvest metrics  %s ", nr.savedErrors[nr.errorCount])
	}
	return nil
}

func init() {
	outputs.Add("newrelic", func() telegraf.Output {
		return &NewRelic{
			Timeout: internal.Duration{Duration: time.Second * 15},
			Client:  http.Client{},
		}
	})
}
