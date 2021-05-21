package dynatrace

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"

	dtMetric "github.com/dynatrace-oss/dynatrace-metric-utils-go/metric"
	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/apiconstants"
	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/dimensions"
)

// Dynatrace Configuration for the Dynatrace output plugin
type Dynatrace struct {
	URL               string          `toml:"url"`
	APIToken          string          `toml:"api_token"`
	Prefix            string          `toml:"prefix"`
	Log               telegraf.Logger `toml:"-"`
	Timeout           config.Duration `toml:"timeout"`
	AddCounterMetrics []string        `toml:"additional_counters"`

	tls.ClientConfig

	client *http.Client

	loggedMetrics map[string]bool // New empty set
}

const sampleConfig = `
  ## For usage with the Dynatrace OneAgent you can omit any configuration,
  ## the only requirement is that the OneAgent is running on the same host.
  ## Only setup environment url and token if you want to monitor a Host without the OneAgent present.
  ##
  ## Your Dynatrace environment URL.
  ## For Dynatrace OneAgent you can leave this empty or set it to "http://127.0.0.1:14499/metrics/ingest" (default)
  ## For Dynatrace SaaS environments the URL scheme is "https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest"
  ## For Dynatrace Managed environments the URL scheme is "https://{your-domain}/e/{your-environment-id}/api/v2/metrics/ingest"
  url = ""

  ## Your Dynatrace API token. 
  ## Create an API token within your Dynatrace environment, by navigating to Settings > Integration > Dynatrace API
  ## The API token needs data ingest scope permission. When using OneAgent, no API token is required.
  api_token = "" 

  ## Optional prefix for metric names (e.g.: "telegraf")
  prefix = "telegraf"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Optional flag for ignoring tls certificate check
  # insecure_skip_verify = false


  ## Connection timeout, defaults to "5s" if not set.
  timeout = "5s"

  ## If you want to convert values represented as gauges to counters, add the metric names here
  additional_counters = [ ]
`

// Connect Connects the Dynatrace output plugin to the Telegraf stream
func (d *Dynatrace) Connect() error {
	return nil
}

// Close Closes the Dynatrace output plugin
func (d *Dynatrace) Close() error {
	d.client = nil
	return nil
}

// SampleConfig Returns a sample configuration for the Dynatrace output plugin
func (d *Dynatrace) SampleConfig() string {
	return sampleConfig
}

// Description returns the description for the Dynatrace output plugin
func (d *Dynatrace) Description() string {
	return "Send telegraf metrics to a Dynatrace environment"
}

func (d *Dynatrace) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	lines := []string{}

	for _, tm := range metrics {
		dims := []dimensions.Dimension{}
		for _, tag := range tm.TagList() {
			// Ignore special tags for histogram and summary types.
			switch tm.Type() {
			case telegraf.Histogram:
				if tag.Key == "le" || tag.Key == "gt" {
					continue
				}
			case telegraf.Summary:
				if tag.Key == "quantile" {
					continue
				}
			}
			dims = append(dims, dimensions.NewDimension(tag.Key, tag.Value))
		}

		metricType := tm.Type()
		for _, field := range tm.FieldList() {
			metricName := tm.Name() + "." + field.Key
			for _, i := range d.AddCounterMetrics {
				if metricName == i {
					metricType = telegraf.Counter
				}
			}

			typeOpt := getTypeOption(metricType, field)

			if typeOpt == nil {
				// Unsupported type. Log only once per unsupported metric name
				if !d.loggedMetrics[metricName] {
					d.Log.Warnf("Unsupported type for %s", metricName)
					d.loggedMetrics[metricName] = true
				}
				continue
			}

			name := tm.Name() + "." + field.Key
			dm, err := dtMetric.NewMetric(
				name,
				dtMetric.WithPrefix(d.Prefix),
				dtMetric.WithDimensions(
					dimensions.MergeLists(
						// dimensions.NewNormalizedDimensionList(e.opts.DefaultDimensions...),
						dimensions.NewNormalizedDimensionList(dims...),
					),
				),
				typeOpt,
			)

			if err != nil {
				d.Log.Warn(fmt.Sprintf("failed to normalize metric: %s - %s", name, err.Error()))
				continue
			}

			line, err := dm.Serialize()

			if err != nil {
				d.Log.Warn(fmt.Sprintf("failed to serialize metric: %s - %s", name, err.Error()))
				continue
			}

			lines = append(lines, line)
		}
	}

	limit := apiconstants.GetPayloadLinesLimit()
	for i := 0; i < len(lines); i += limit {
		batch := lines[i:min(i+limit, len(lines))]

		output := strings.Join(batch, "\n")
		if output != "" {
			if err := d.send(output); err != nil {
				return fmt.Errorf("error processing data:, %s", err.Error())
			}
		}
	}

	return nil
}

func (d *Dynatrace) send(msg string) error {
	var err error
	req, err := http.NewRequest("POST", d.URL, bytes.NewBufferString(msg))
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		return fmt.Errorf("error while creating HTTP request:, %s", err.Error())
	}
	req.Header.Add("Content-Type", "text/plain; charset=UTF-8")

	if len(d.APIToken) != 0 {
		req.Header.Add("Authorization", "Api-Token "+d.APIToken)
	}
	// add user-agent header to identify metric source
	req.Header.Add("User-Agent", "telegraf")

	resp, err := d.client.Do(req)
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		return fmt.Errorf("error while sending HTTP request:, %s", err.Error())
	}
	defer resp.Body.Close()

	// print metric line results as info log
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusBadRequest {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			d.Log.Errorf("Dynatrace error reading response")
		}
		bodyString := string(bodyBytes)
		d.Log.Debugf("Dynatrace returned: %s", bodyString)
	} else {
		return fmt.Errorf("request failed with response code:, %d", resp.StatusCode)
	}
	return nil
}

func (d *Dynatrace) Init() error {
	if len(d.URL) == 0 {
		d.Log.Infof("Dynatrace URL is empty, defaulting to OneAgent metrics interface")
		d.URL = apiconstants.GetDefaultOneAgentEndpoint()
	}
	if d.URL != apiconstants.GetDefaultOneAgentEndpoint() && len(d.APIToken) == 0 {
		d.Log.Errorf("Dynatrace api_token is a required field for Dynatrace output")
		return fmt.Errorf("api_token is a required field for Dynatrace output")
	}

	tlsCfg, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(d.Timeout),
	}
	return nil
}

func init() {
	outputs.Add("dynatrace", func() telegraf.Output {
		return &Dynatrace{
			Timeout: config.Duration(time.Second * 5),
		}
	})
}

func getTypeOption(metricType telegraf.ValueType, field *telegraf.Field) dtMetric.MetricOption {
	if metricType == telegraf.Counter {
		switch v := field.Value.(type) {
		case float64:
			return dtMetric.WithFloatCounterValueTotal(v)
		case uint64:
			return dtMetric.WithIntCounterValueTotal(int64(v))
		case int64:
			return dtMetric.WithIntCounterValueTotal(v)
		default:
			return nil
		}
	}

	switch v := field.Value.(type) {
	case float64:
		return dtMetric.WithFloatGaugeValue(v)
	case uint64:
		return dtMetric.WithIntGaugeValue(int64(v))
	case int64:
		return dtMetric.WithIntGaugeValue(32)
	case bool:
		if v {
			return dtMetric.WithIntGaugeValue(1)
		}
		return dtMetric.WithIntGaugeValue(0)
	}

	return nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
