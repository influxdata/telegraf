package dynatrace

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	oneAgentMetricsUrl = "http://127.0.0.1:14499/metrics/ingest"
)

var (
	reNameAllowedCharList = regexp.MustCompile("[^A-Za-z0-9.]+")
	maxDimKeyLen          = 100
	maxMetricKeyLen       = 250
)

// Dynatrace Configuration for the Dynatrace output plugin
type Dynatrace struct {
	EnvironmentURL       string          `toml:"environment_url"`
	EnvironmentAPIToken  string          `toml:"environmentApiToken"`
	SkipCertificateCheck bool            `toml:"skipCertificateCheck"`
	Prefix               string          `toml:"prefix"`
	Log                  telegraf.Logger `toml:"log"`

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
  environmentURL = ""

  ## Your Dynatrace API token. 
  ## Create an API token within your Dynatrace environment, by navigating to Settings > Integration > Dynatrace API
  ## The API token needs data ingest scope permission. When using OneAgent, no API token is required.
  environmentApiToken = "" 

  ## Optional prefix for metric names (e.g.: "telegraf.")
  prefix = "telegraf."
  
  ## Optional flag for ignoring tls certificate check
  skipCertificateCheck = false
`

// Connect Connects the Dynatrace output plugin to the Telegraf stream
func (d *Dynatrace) Connect() error {
	if len(d.EnvironmentURL) == 0 {
		d.Log.Infof("Dynatrace environmentURL is empty, defaulting to OneAgent metrics interface")
		d.EnvironmentURL = oneAgentMetricsUrl
	}
	if d.EnvironmentURL != oneAgentMetricsUrl && len(d.EnvironmentAPIToken) == 0 {
		d.Log.Errorf("Dynatrace environmentApiToken is a required field for Dynatrace output")
		return fmt.Errorf("environmentApiToken is a required field for Dynatrace output")
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: d.SkipCertificateCheck},
		},
		Timeout: 5 * time.Second,
	}
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

	// Strip Digits if they are at the beginning of the string
	normalizedString := ""
	firstChars := true

	for _, char := range s {
		if firstChars && (unicode.IsDigit(char) || char == '_') {
			continue
		} else {
			firstChars = false
		}
		normalizedString += string(char)
	}

	for strings.HasPrefix(normalizedString, "_") {
		normalizedString = normalizedString[1:]
	}

	if len(normalizedString) > max {
		normalizedString = normalizedString[:max]
	}

	for strings.HasSuffix(normalizedString, "_") {
		normalizedString = normalizedString[:len(normalizedString)-1]
	}

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
	var tagb bytes.Buffer
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		// first write the tags into a buffer
		tagb.Reset()
		if len(metric.Tags()) > 0 {
			for tk, tv := range metric.Tags() {
				tagKey, err := d.normalize(tk, maxDimKeyLen)
				if err != nil {
					continue
				}
				fmt.Fprintf(&tagb, ",%s=%s", strings.ToLower(tagKey), d.escape(tv))

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
					d.Log.Infof("Dynatrace type not supported! %s", v)
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
				fmt.Fprintf(&buf, "%s", metricID)
				// add the tag string
				if len(tagb.String()) > 0 {
					fmt.Fprintf(&buf, "%s", tagb.String())
				}

				// write measured value
				fmt.Fprintf(&buf, " %v\n", value)
			}
		}
	}

	return d.send(buf.Bytes())
}

func (d *Dynatrace) send(msg []byte) error {
	var err error
	req, err := http.NewRequest("POST", d.EnvironmentURL, bytes.NewBuffer(msg))
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		return fmt.Errorf("Dynatrace error while creating HTTP request:, %s", err.Error())
	}
	req.Header.Add("Content-Type", "text/plain; charset=UTF-8")

	if len(d.EnvironmentAPIToken) != 0 {
		req.Header.Add("Authorization", "Api-Token "+d.EnvironmentAPIToken)
	}
	// add user-agent header to identify metric source
	req.Header.Add("User-Agent", "telegraf")

	resp, err := d.client.Do(req)
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		fmt.Println(req)
		return fmt.Errorf("Dynatrace error while sending HTTP request:, %s", err.Error())
	}
	defer resp.Body.Close()

	// print metric line results as info log
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			d.Log.Errorf("Dynatrace error reading response")
		}
		bodyString := string(bodyBytes)
		d.Log.Infof("Dynatrace returned: %s", bodyString)
	} else {
		return fmt.Errorf("Dynatrace request failed with response code:, %d", resp.StatusCode)
	}

	return nil
}

func init() {
	outputs.Add("dynatrace", func() telegraf.Output {
		return &Dynatrace{}
	})
}
