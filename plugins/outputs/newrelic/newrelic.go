package newrelic

// newrelic.go
import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/cumulative"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// NewRelic nr structure
type NewRelic struct {
	InsightsKey  string          `toml:"insights_key"`
	MetricPrefix string          `toml:"metric_prefix"`
	Timeout      config.Duration `toml:"timeout"`
	HTTPProxy    string          `toml:"http_proxy"`
	MetricURL    string          `toml:"metric_url"`

	harvestor   *telemetry.Harvester
	dc          *cumulative.DeltaCalculator
	savedErrors map[int]interface{}
	errorCount  int
	client      http.Client
}

// Description returns a one-sentence description on the Output
func (nr *NewRelic) Description() string {
	return "Send metrics to New Relic metrics endpoint"
}

// SampleConfig : return  default configuration of the Output
func (nr *NewRelic) SampleConfig() string {
	return `
  ## The 'insights_key' parameter requires a NR license key.
  ## New Relic recommends you create one
  ## with a convenient name such as TELEGRAF_INSERT_KEY.
  ## reference: https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/#ingest-license-key
  # insights_key = "New Relic License Key Here"

  ## Prefix to add to add to metric name for easy identification.
  ## This is very useful if your metric names are ambiguous.
  # metric_prefix = ""

  ## Timeout for writes to the New Relic API.
  # timeout = "15s"

  ## HTTP Proxy override. If unset use values from the standard
  ## proxy environment variables to determine proxy, if any.
  # http_proxy = "http://corporate.proxy:3128"

  ## Metric URL override to enable geographic location endpoints.
  # If not set use values from the standard 
  # metric_url = "https://metric-api.newrelic.com/metric/v1"
`
}

// Connect to the Output
func (nr *NewRelic) Connect() error {
	if nr.InsightsKey == "" {
		return fmt.Errorf("InsightKey is a required for newrelic")
	}
	err := nr.initClient()
	if err != nil {
		return err
	}

	nr.harvestor, err = telemetry.NewHarvester(telemetry.ConfigAPIKey(nr.InsightsKey),
		telemetry.ConfigHarvestPeriod(0),
		func(cfg *telemetry.Config) {
			cfg.Product = "NewRelic-Telegraf-Plugin"
			cfg.ProductVersion = "1.0"
			cfg.HarvestTimeout = time.Duration(nr.Timeout)
			cfg.Client = &nr.client
			cfg.ErrorLogger = func(e map[string]interface{}) {
				var errorString string
				for k, v := range e {
					errorString += fmt.Sprintf("%s = %s ", k, v)
				}
				nr.errorCount++
				nr.savedErrors[nr.errorCount] = errorString
			}
			if nr.MetricURL != "" {
				cfg.MetricsURLOverride = nr.MetricURL
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
	nr.client.CloseIdleConnections()
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
				mvalue = n
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
				return fmt.Errorf("undefined field type: %T", field.Value)
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
			Timeout: config.Duration(time.Second * 15),
		}
	})
}

func (nr *NewRelic) initClient() error {
	if nr.HTTPProxy == "" {
		nr.client = http.Client{}
		return nil
	}

	proxyURL, err := url.Parse(nr.HTTPProxy)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	nr.client = http.Client{
		Transport: transport,
	}
	return nil
}
