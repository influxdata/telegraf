// Package circonus contains the output plugin used to output metric data to
// the Circonus platform.
package circonus

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	itls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// Circonus values are used to output data to the Circonus platform.
type Circonus struct {
	itls.ClientConfig
	Timeout               internal.Duration `toml:"timeout"`
	Broker                string            `toml:"broker"`
	ExcludeBrokers        []string          `toml:"exclude_brokers"`
	Checks                map[string]string `toml:"checks"`
	client                *http.Client
	apicli                *http.Client
	api                   apiClient
	APIURL                string `toml:"api_url"`
	APIToken              string `toml:"api_token"`
	APIApp                string `toml:"api_app"`
	APITLSCA              string `toml:"api_tls_ca"`
	APIInsecureSkipVerify bool   `toml:"api_insecure_skip_verify"`
}

// Init performs initialization of a Circonus client.
func (c *Circonus) Init() error {
	return c.getAPIClient()
}

var sampleConfig = `
  ## Connection timeout:
  # timeout = "5s"

  ## Checks is a map of regexp patterns and submission URL's of Circonus
  ## HTTPTrap checks to which metrics with names mattching the patterns will
  ## be sent:
  # checks = { ".*" = "https://broker1.example.net:43191/module/httptrap/11223344-5566-7788-9900-aabbccddeeff/example" }
  
  ## If the CID of a broker is provided:
  # broker = "/broker/1"
  ## or automatic broker lookup can be used if broker is set to "auto":
  # broker = "auto"
  ## brokers can be excluded by adding their CID to the exclude list:
  # exclude_brokers = [ "/broker/2" ]
  ## then a check can be automatically created for metrics collected with
  ## this Telegraf plugin by entering "auto" for the submission URL:
  checks = { ".*" = "auto" }
  
  ## Optional Broker TLS Configuration, note any brokers used by this plugin
  ## must share the same CA and certificate files, if this info is not provided,
  ## the broker CA data will be retrived using the API:
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification:
  # insecure_skip_verify = false

  ## Circonus API Configuration, this is required for automatic check creation
  ## and automatic check lookup, and retrieving broker CA information:
  # api_url = "https://api.circonus.com/"
  # api_token = "11223344-5566-7788-9900-aabbccddeeff"
  # api_app = "telegraf"
  ## Optional API TLS Configuration: 
  # api_tls_ca = "/etc/telegraf/api_ca.pem"
  ## Use TLS but skip chain & host verification:
  # api_insecure_skip_verify = false
`

var description = "Configuration for Circonus output plugin."

// Metrics values are maps of metric data keyed by metric name.
type Metrics map[string]Metric

// Metric values describe a single metric data value.
type Metric struct {
	Name  string      `json:"-"`
	Type  string      `json:"_type"`
	Value interface{} `json:"_value"`
}

// Conenct creates the client connection to the Circonus broker.
func (c *Circonus) Connect() error {
	tlsCfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}

	if tlsCfg.RootCAs == nil && !tlsCfg.InsecureSkipVerify {
		data, err := c.api.Get("/pki/ca.crt")
		if err != nil {
			return fmt.Errorf("unable to fetch ca certificate: %w", err)
		}

		type cacert struct {
			Contents string `json:"contents"`
		}

		var cadata cacert

		if err := json.Unmarshal(data, &cadata); err != nil {
			return fmt.Errorf("error unmarshalling certificate data: %w", err)
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(cadata.Contents)) {
			return fmt.Errorf("unable to add Circonus broker CA certificate to certtificate pool")
		}

		tlsCfg.RootCAs = cp
	}

	c.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: c.Timeout.Duration,
	}

	c.apicli = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: c.Timeout.Duration,
	}

	if c.APITLSCA != "" {
		cert, err := ioutil.ReadFile(c.APITLSCA)
		if err != nil {
			return fmt.Errorf("unable to configure Circonus API client: %w", err)
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(cert) {
			return fmt.Errorf("unable to add Circonus API CA certificate to certtificate pool")
		}

		c.apicli.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: cp},
			Proxy:           http.ProxyFromEnvironment,
		}
	}

	return nil
}

// Write is used to write metric data to Circonus checks.
func (c *Circonus) Write(metrics []telegraf.Metric) error {
	mets := map[string]Metrics{}

	for _, m := range metrics {
		fieldMetrics := c.buildMetrics(m)

		for _, met := range fieldMetrics {
			for pat, host := range c.Checks {
				re, err := regexp.Compile(pat)
				if err != nil {
					return fmt.Errorf("unable to compile metric regexp: %s: %s",
						pat, err.Error())
				}

				if re.MatchString(met.Name) {
					if mets[host] == nil {
						mets[host] = Metrics{}
					}

					mets[host][met.Name] = met

					break
				}
			}
		}
	}

	autoHost := ""
	cli := c.client

	for host, ms := range mets {
		b, err := json.Marshal(ms)
		if err != nil {
			return fmt.Errorf("unable to marshal metric data, %s\n",
				err.Error())
		}

		if strings.ToLower(host) == "auto" {
			if autoHost == "" {
				autoHost, err := c.getSubmissionURL()
				if err != nil {
					return err
				}

				host = autoHost
				if strings.HasPrefix(host, "https") &&
					strings.HasPrefix(host, c.APIURL) {
					cli = c.apicli
				}
			}
		}

		req, err := http.NewRequest("PUT", host, bytes.NewBuffer(b))
		if err != nil {
			return fmt.Errorf("unable to create http.Request, %s\n",
				err.Error())
		}

		req.Header.Add("Content-Type", "application/json")
		resp, err := cli.Do(req)
		if err != nil {
			return fmt.Errorf("error sending metric data, %s\n", err.Error())
		}

		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 209 {
			return fmt.Errorf("received error status code, %d\n",
				resp.StatusCode)
		}
	}

	return nil
}

// buildNumerics constructs numeric metrics from a telegraf metric.
func (c *Circonus) buildNumerics(m telegraf.Metric) []Metric {
	metrics := []Metric{}
	fields := m.FieldList()

	for _, field := range fields {
		mn := m.Name() + "." + field.Key
		mt := "L"
		fv := field.Value
		met := Metric{
			Type:  mt,
			Value: fv,
		}

		if strings.HasSuffix(mn, "__value") {
			mn = mn[:len(mn)-7]
		}

		tags := m.TagList()
		if len(tags) > 0 {
			mn += "|ST["

			for i, t := range tags {
				if i > 0 {
					mn += ","
				}

				mn += t.Key + ":" + t.Value
			}

			mn += "]"
		}

		met.Name = mn
		metrics = append(metrics, met)
	}

	return metrics
}

// buildTexts constructs text metrics from a telegraf metric.
func (c *Circonus) buildTexts(m telegraf.Metric) []Metric {
	metrics := []Metric{}
	fields := m.FieldList()

	for _, field := range fields {
		mn := m.Name() + "." + field.Key
		mt := "s"
		fv := field.Value
		met := Metric{
			Type:  mt,
			Value: fv,
		}

		if strings.HasSuffix(mn, "__value") {
			mn = mn[:len(mn)-7]
		}

		tags := m.TagList()
		if len(tags) > 0 {
			mn += "|ST["

			for i, t := range tags {
				if i > 0 {
					mn += ","
				}

				mn += t.Key + ":" + t.Value
			}

			mn += "]"
		}

		met.Name = mn
		metrics = append(metrics, met)
	}

	return metrics
}

// buildHistogram constructs histogram metrics from a telegraf metric.
func (c *Circonus) buildHistogram(m telegraf.Metric) []Metric {
	metrics := []Metric{}
	fields := m.FieldList()

	mn := m.Name()
	mt := "n"
	hv := []string{}

	for _, f := range fields {
		v, err := strconv.ParseFloat(f.Key, 64)
		if err != nil {
			continue
		}

		hv = append(hv, fmt.Sprintf("H[%.1e]=%v", v, f.Value))
	}

	fv := hv

	met := Metric{
		Type:  mt,
		Value: fv,
	}

	if strings.HasSuffix(mn, "__value") {
		mn = mn[:len(mn)-7]
	}

	tags := m.TagList()
	if len(tags) > 0 {
		mn += "|ST["

		for i, t := range tags {
			if i > 0 {
				mn += ","
			}

			mn += t.Key + ":" + t.Value
		}

		mn += "]"
	}

	met.Name = mn
	metrics = append(metrics, met)

	return metrics
}

// buildMetrics constructs Circonus metrics from a telegraf metric.
func (c *Circonus) buildMetrics(m telegraf.Metric) []Metric {
	switch m.Type() {
	case telegraf.Counter, telegraf.Gauge, telegraf.Summary:
		return c.buildNumerics(m)
	case telegraf.Untyped:
		fields := m.FieldList()
		if s, ok := fields[0].Value.(string); ok {
			if strings.Contains(s, "H[") && strings.Contains(s, "]=") {
				return c.buildHistogram(m)
			} else {
				return c.buildTexts(m)
			}
		} else {
			return c.buildNumerics(m)
		}
	case telegraf.Histogram:
		return c.buildHistogram(m)
	default:
		return []Metric{}
	}
}

// SampleConfig returns the sample Circonus plugin configuration.
func (c *Circonus) SampleConfig() string {
	return sampleConfig
}

// Description returns a description of the Circonus plugin configuration.
func (c *Circonus) Description() string {
	return description
}

// Close will close the Circonus client connection.
func (c *Circonus) Close() error {
	return nil
}

func init() {
	outputs.Add("circonus", func() telegraf.Output {
		return &Circonus{}
	})
}
