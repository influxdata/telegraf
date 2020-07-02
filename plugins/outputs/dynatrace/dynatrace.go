package dynatrace

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// Dynatrace Configuration for the Dynatrace output plugin
type Dynatrace struct {
	EnvironmentURL       string          `toml:"environmentURL"`
	EnvironmentAPIToken  string          `toml:"environmentApiToken"`
	SkipCertificateCheck bool            `toml:"skipCertificateCheck"`
	Log                  telegraf.Logger `toml:"log"`

	client *http.Client
}

var sampleConfig = `
  ## Your Dynatrace environment URL. 
  ## For Dynatrace SaaS environments the URL scheme is "https://{your-environment-id}.live.dynatrace.com"
  ## For Dynatrace Managed environments the URL scheme is "https://{your-domain}/e/{your-environment-id}"
  environmentURL = ""

  ## Your Dynatrace API token. 
  ## Create an API token within your Dynatrace environment, by navigating to Settings > Integration > Dynatrace API
  ## The API token needs data ingest scope permission.
  environmentApiToken = "" 
`

// Connect Connects the Dynatrace output plugin to the Telegraf stream
func (d *Dynatrace) Connect() error {
	if len(d.EnvironmentURL) == 0 {
		d.Log.Errorf("Dynatrace environmentURL is a required field for Dynatrace output")
		return fmt.Errorf("environmentURL is a required field for Dynatrace output")
	}
	if len(d.EnvironmentAPIToken) == 0 {
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
	var err error

	return err
}

// SampleConfig Returns a sample configuration for the Dynatrace output plugin
func (d *Dynatrace) SampleConfig() string {
	return sampleConfig
}

// Description returns the description for the Dynatrace output plugin
func (d *Dynatrace) Description() string {
	return "Send telegraf metrics to a Dynatrace environment"
}

func (d *Dynatrace) convertKey(v string) string {
	kEs := strings.ToLower(v)
	sEs := strings.Replace(kEs, " ", "_", -1)
	return sEs
}

func (d *Dynatrace) escape(v string) string {
	vEs := strings.Replace(v, "\\", "\\\\", -1)
	return "\"" + vEs + "\""
}

func (d *Dynatrace) Write(metrics []telegraf.Metric) error {
	var err error
	var buf bytes.Buffer
	var tagb bytes.Buffer
	if len(metrics) == 0 {
		return err
	}

	for _, metric := range metrics {
		// first write the tags into a buffer
		tagb.Reset()
		if len(metric.Tags()) > 0 {
			for tk, tv := range metric.Tags() {
				fmt.Fprintf(&tagb, ",%s=%s", d.convertKey(tk), d.escape(tv))
			}
		}
		if len(metric.Fields()) > 0 {
			for k, v := range metric.Fields() {
				var value string
				// first check if value type is supported
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

				// write metric name combined with its field
				fmt.Fprintf(&buf, "%s.%s", metric.Name(), k)
				// add the tag string
				if len(tagb.String()) > 0 {
					fmt.Fprintf(&buf, "%s", tagb.String())
				}

				// write measured value
				fmt.Fprintf(&buf, " %v\n", value)
			}
		}
	}
	//d.Log.Infof("%s", buf.String())
	// send it
	d.send(buf.Bytes())
	return err
}

func (d *Dynatrace) send(msg []byte) error {
	var err error
	req, err := http.NewRequest("POST", d.EnvironmentURL+"/api/v2/metrics/ingest", bytes.NewBuffer(msg))
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		return fmt.Errorf("Dynatrace error while creating HTTP request:, %s", err.Error())
	}
	req.Header.Add("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Add("Authorization", "Api-Token "+d.EnvironmentAPIToken)
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
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			d.Log.Errorf("Dynatrace error reading response")
		}
		bodyString := string(bodyBytes)
		d.Log.Infof("Dynatrace returned: %s", bodyString)
	} else {
		return fmt.Errorf("Dynatrace request failed with response code:, %d", resp.StatusCode)
	}

	return err
}

func init() {
	outputs.Add("dynatrace", func() telegraf.Output {
		return &Dynatrace{}
	})
}
