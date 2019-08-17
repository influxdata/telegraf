package warp10

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultClientTimeout = 15 * time.Second
)

// Warp10 output plugin
type Warp10 struct {
	Prefix  string
	WarpURL string
	Token   string
	Timeout internal.Duration `toml:"timeout"`
	client  *http.Client
	tls.ClientConfig
}

var sampleConfig = `
  # prefix for metrics class Name
  prefix = "Prefix"
  ## POST HTTP(or HTTPS) ##
  # Url name of the Warp 10 server
  warp_url = "WarpUrl"
  # Token to access your app on warp 10
  token = "Token"
  # Warp 10 query timeout, by default 15s
  timeout = "15s"
  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
`

// MetricLine Warp 10 metrics
type MetricLine struct {
	Metric    string
	Timestamp int64
	Value     string
	Tags      string
}

func (w *Warp10) createClient() (*http.Client, error) {
	tlsCfg, err := w.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if w.Timeout.Duration == 0 {
		w.Timeout.Duration = defaultClientTimeout
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: w.Timeout.Duration,
	}

	return client, nil
}

// Connect to warp10
func (w *Warp10) Connect() error {
	client, err := w.createClient()
	if err != nil {
		return err
	}

	w.client = client
	return nil
}

// GenWarp10Payload compute Warp 10 metrics payload
func (w *Warp10) GenWarp10Payload(metrics []telegraf.Metric, now time.Time) string {
	collectString := make([]string, 0)
	for _, mm := range metrics {

		for _, field := range mm.FieldList() {

			metric := &MetricLine{
				Metric:    fmt.Sprintf("%s%s", w.Prefix, mm.Name()+"."+field.Key),
				Timestamp: now.UnixNano() / 1000,
			}

			metricValue, err := buildValue(field.Value)
			if err != nil {
				log.Printf("E! [outputs.warp10] Could not encode value: %v", err)
				continue
			}
			metric.Value = metricValue

			tagsSlice := buildTags(mm.TagList())
			metric.Tags = strings.Join(tagsSlice, ",")

			messageLine := fmt.Sprintf("%d// %s{%s} %s\n", metric.Timestamp, metric.Metric, metric.Tags, metric.Value)

			collectString = append(collectString, messageLine)
		}
	}
	return fmt.Sprint(strings.Join(collectString, ""))
}

// Write metrics to Warp10
func (w *Warp10) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}
	var now = time.Now()
	payload := w.GenWarp10Payload(metrics, now)

	req, err := http.NewRequest("POST", w.WarpURL+"/api/v0/update", bytes.NewBufferString(payload))
	req.Header.Set("X-Warp10-Token", w.Token)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(w.WarpURL + ": " + w.HandleError(string(body)))
	}

	return nil
}

func buildTags(tags []*telegraf.Tag) []string {

	tagsString := make([]string, len(tags)+1)
	indexSource := 0
	for index, tag := range tags {
		tagsString[index] = fmt.Sprintf("%s=%s", tag.Key, tag.Value)
		indexSource = index
	}
	indexSource++
	tagsString[indexSource] = fmt.Sprintf("source=telegraf")
	sort.Strings(tagsString)
	return tagsString
}

func buildValue(v interface{}) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = intToString(int64(p))
	case string:
		retv = fmt.Sprintf("'%s'", strings.Replace(p, "'", "\\'", -1))
	case bool:
		retv = boolToString(bool(p))
	case uint64:
		retv = uIntToString(uint64(p))
	case float64:
		retv = floatToString(float64(p))
	default:
		retv = "'" + strings.Replace(fmt.Sprintf("%s", p), "'", "\\'", -1) + "'"
	}
	return retv, nil
}

func intToString(inputNum int64) string {
	return strconv.FormatInt(inputNum, 10)
}

func boolToString(inputBool bool) string {
	return strconv.FormatBool(inputBool)
}

func uIntToString(inputNum uint64) string {
	return strconv.FormatUint(inputNum, 10)
}

func floatToString(inputNum float64) string {
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}

// SampleConfig get config
func (w *Warp10) SampleConfig() string {
	return sampleConfig
}

// Description get description
func (w *Warp10) Description() string {
	return "Configuration for Warp server to send metrics to"
}

// Close close
func (w *Warp10) Close() error {
	return nil
}

func init() {
	outputs.Add("warp10", func() telegraf.Output {
		return &Warp10{}
	})
}

// HandleError read http error body and return a corresponding error
func (w *Warp10) HandleError(body string) string {
	if body == "" {
		return "Empty return"
	}

	if strings.Contains(body, "Invalid token") {
		return fmt.Sprintf("Invalid token: %v", w.Token)
	}

	if strings.Contains(body, "Write token missing") {
		return "Write token missing"
	}

	if strings.Contains(body, "Token Expired") {
		return fmt.Sprintf("Token Expired: %v", w.Token)
	}

	if strings.Contains(body, "Token revoked") {
		return fmt.Sprintf("Token revoked: %v", w.Token)
	}

	if strings.Contains(body, "exceed your Monthly Active Data Streams limit") || strings.Contains(body, "exceed the Monthly Active Data Streams limit") {
		reg := regexp.MustCompile(`Monthly Active Data Streams limit(?: for application.*)? \((\d+)(.\d+)?(E-\d)?\)`)
		parts := reg.FindStringSubmatch(body)
		limit := "-1"
		if len(parts) > 1 {
			limit = parts[1]
		}
		return fmt.Sprintf("MADS exceeded: %v", limit)
	}

	if strings.Contains(body, "Daily Data Points limit being already exceeded") {
		reg := regexp.MustCompile(`Current maximum rate is \((\d+)(.\d+)?(E-\d)?\) datapoints/s`)
		parts := reg.FindStringSubmatch(body)
		limit := "-1"
		if len(parts) > 1 {
			limit = parts[1]
		}
		return fmt.Sprintf("DDP exceeded: %v", limit)
	}

	if strings.Contains(body, "Parse error at") {
		reg := regexp.MustCompile(`<pre>\s*Parse error at &apos;(.*)&apos;</pre>`)
		parts := reg.FindStringSubmatch(body)
		str := ""
		if len(parts) > 1 {
			str = parts[1]
		}
		return fmt.Sprintf("Parse error at: %v", str)
	}

	if strings.Contains(body, "Application suspended or closed") {
		return "Application suspended or closed"
	}

	if strings.Contains(body, "For input string") {
		reg := regexp.MustCompile(`<pre>\s*For input string: &quot;(.*)&quot;</pre>`)
		parts := reg.FindStringSubmatch(body)
		str := ""
		if len(parts) > 1 {
			str = parts[1]
		}
		return fmt.Sprintf("For input string: %v", str)
	}

	if strings.Contains(body, "broken pipe") {
		return "broken pipe"
	}

	maxStringSixe := 511
	if len(body) < maxStringSixe {
		return body
	}
	return body[0:maxStringSixe]
}
