package dynatrace

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	oneAgentMetricsUrl   = "http://127.0.0.1:14499/metrics/ingest"
	dtIngestApiLineLimit = 1000
)

var (
	reNameAllowedCharList = regexp.MustCompile("[^A-Za-z0-9.-]+")
	maxDimKeyLen          = 100
	maxMetricKeyLen       = 250
)

// Dynatrace Configuration for the Dynatrace output plugin
type Dynatrace struct {
	URL               string            `toml:"url"`
	APIToken          string            `toml:"api_token"`
	Prefix            string            `toml:"prefix"`
	Log               telegraf.Logger   `toml:"-"`
	Timeout           internal.Duration `toml:"timeout"`
	AddCounterMetrics []string          `toml:"additional_counters"`
	State             map[string]string
	SendCounter       int

	tls.ClientConfig

	client *http.Client
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

  ## Optional prefix for metric names (e.g.: "telegraf.")
  prefix = "telegraf."

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

// Normalizes a metric keys or metric dimension identifiers
// according to Dynatrace format.
func (d *Dynatrace) normalize(s string, max int) (string, error) {
	s = reNameAllowedCharList.ReplaceAllString(s, "_")

	// Strip Digits and underscores if they are at the beginning of the string
	normalizedString := strings.TrimLeft(s, "_0123456789")

	for strings.HasPrefix(normalizedString, "_") {
		normalizedString = normalizedString[1:]
	}

	if len(normalizedString) > max {
		normalizedString = normalizedString[:max]
	}

	for strings.HasSuffix(normalizedString, "_") {
		normalizedString = normalizedString[:len(normalizedString)-1]
	}

	normalizedString = strings.ReplaceAll(normalizedString, "..", "_")

	if len(normalizedString) == 0 {
		return "", fmt.Errorf("error normalizing the string: %s", s)
	}
	return normalizedString, nil
}

func (d *Dynatrace) escape(v string) string {
	return strconv.Quote(v)
}

func (d *Dynatrace) Write(metrics []telegraf.Metric) error {
	var buf bytes.Buffer
	metricCounter := 1
	var tagb bytes.Buffer
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		// first write the tags into a buffer
		tagb.Reset()
		if len(metric.Tags()) > 0 {
			keys := make([]string, 0, len(metric.Tags()))
			for k := range metric.Tags() {
				keys = append(keys, k)
			}
			// sort tag keys to expect the same order in ech run
			sort.Strings(keys)

			for _, k := range keys {
				tagKey, err := d.normalize(k, maxDimKeyLen)
				if err != nil {
					continue
				}
				if len(metric.Tags()[k]) > 0 {
					fmt.Fprintf(&tagb, ",%s=%s", strings.ToLower(tagKey), d.escape(metric.Tags()[k]))
				}
			}
		}
		if len(metric.Fields()) > 0 {
			for k, v := range metric.Fields() {
				var value string
				switch v := v.(type) {
				case string:
					continue
				case float64:
					if !math.IsNaN(v) && !math.IsInf(v, 0) {
						value = fmt.Sprintf("%f", v)
					} else {
						continue
					}
				case uint64:
					value = strconv.FormatUint(v, 10)
				case int64:
					value = strconv.FormatInt(v, 10)
				case bool:
					if v {
						value = "1"
					} else {
						value = "0"
					}
				default:
					d.Log.Debugf("Dynatrace type not supported! %s", v)
					continue
				}

				// metric name
				metricKey, err := d.normalize(k, maxMetricKeyLen)
				if err != nil {
					continue
				}

				metricID, err := d.normalize(d.Prefix+metric.Name()+"."+metricKey, maxMetricKeyLen)
				// write metric name combined with its field
				if err != nil {
					continue
				}
				// write metric id,tags and value

				metricType := metric.Type()
				for _, i := range d.AddCounterMetrics {
					if metric.Name()+"."+metricKey == i {
						metricType = telegraf.Counter
					}
				}

				switch metricType {
				case telegraf.Counter:
					var delta float64

					// Check if LastValue exists
					if lastvalue, ok := d.State[metricID+tagb.String()]; ok {
						// Convert Strings to Floats
						floatLastValue, err := strconv.ParseFloat(lastvalue, 32)
						if err != nil {
							d.Log.Debugf("Could not parse last value: %s", lastvalue)
						}
						floatCurrentValue, err := strconv.ParseFloat(value, 32)
						if err != nil {
							d.Log.Debugf("Could not parse current value: %s", value)
						}
						if floatCurrentValue >= floatLastValue {
							delta = floatCurrentValue - floatLastValue
							fmt.Fprintf(&buf, "%s%s count,delta=%f\n", metricID, tagb.String(), delta)
						}
					}
					d.State[metricID+tagb.String()] = value

				default:
					fmt.Fprintf(&buf, "%s%s %v\n", metricID, tagb.String(), value)
				}

				if metricCounter%dtIngestApiLineLimit == 0 {
					err = d.send(buf.Bytes())
					if err != nil {
						return err
					}
					buf.Reset()
				}
				metricCounter++
			}
		}
	}
	d.SendCounter++
	// in typical interval of 10s, we will clean the counter state once in 24h which is 8640 iterations

	if d.SendCounter%8640 == 0 {
		d.State = make(map[string]string)
	}
	return d.send(buf.Bytes())
}

func (d *Dynatrace) send(msg []byte) error {
	var err error
	req, err := http.NewRequest("POST", d.URL, bytes.NewBuffer(msg))
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
	d.State = make(map[string]string)
	if len(d.URL) == 0 {
		d.Log.Infof("Dynatrace URL is empty, defaulting to OneAgent metrics interface")
		d.URL = oneAgentMetricsUrl
	}
	if d.URL != oneAgentMetricsUrl && len(d.APIToken) == 0 {
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
		Timeout: d.Timeout.Duration,
	}
	return nil
}

func init() {
	outputs.Add("dynatrace", func() telegraf.Output {
		return &Dynatrace{
			Timeout:     internal.Duration{Duration: time.Second * 5},
			SendCounter: 0,
		}
	})
}
