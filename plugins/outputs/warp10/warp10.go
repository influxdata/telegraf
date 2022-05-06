package warp10

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultClientTimeout = 15 * time.Second
)

// Warp10 output plugin
type Warp10 struct {
	Prefix             string          `toml:"prefix"`
	WarpURL            string          `toml:"warp_url"`
	Token              string          `toml:"token"`
	Timeout            config.Duration `toml:"timeout"`
	PrintErrorBody     bool            `toml:"print_error_body"`
	MaxStringErrorSize int             `toml:"max_string_error_size"`
	client             *http.Client
	tls.ClientConfig
	Log telegraf.Logger `toml:"-"`
}

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

	if w.Timeout == 0 {
		w.Timeout = config.Duration(defaultClientTimeout)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(w.Timeout),
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
func (w *Warp10) GenWarp10Payload(metrics []telegraf.Metric) string {
	collectString := make([]string, 0)
	for _, mm := range metrics {
		for _, field := range mm.FieldList() {
			metric := &MetricLine{
				Metric:    fmt.Sprintf("%s%s", w.Prefix, mm.Name()+"."+field.Key),
				Timestamp: mm.Time().UnixNano() / 1000,
			}

			metricValue, err := buildValue(field.Value)
			if err != nil {
				w.Log.Errorf("Could not encode value: %v", err)
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
	payload := w.GenWarp10Payload(metrics)
	if payload == "" {
		return nil
	}

	addr := w.WarpURL + "/api/v0/update"
	req, err := http.NewRequest("POST", addr, bytes.NewBufferString(payload))
	if err != nil {
		return fmt.Errorf("unable to create new request '%s': %s", addr, err)
	}

	req.Header.Set("X-Warp10-Token", w.Token)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if w.PrintErrorBody {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf(w.WarpURL + ": " + w.HandleError(string(body), w.MaxStringErrorSize))
		}

		if len(resp.Status) < w.MaxStringErrorSize {
			return fmt.Errorf(w.WarpURL + ": " + resp.Status)
		}

		return fmt.Errorf(w.WarpURL + ": " + resp.Status[0:w.MaxStringErrorSize])
	}

	return nil
}

func buildTags(tags []*telegraf.Tag) []string {
	tagsString := make([]string, len(tags)+1)
	indexSource := 0
	for index, tag := range tags {
		key := url.QueryEscape(tag.Key)
		value := url.QueryEscape(tag.Value)
		tagsString[index] = fmt.Sprintf("%s=%s", key, value)
		indexSource = index
	}
	indexSource++
	tagsString[indexSource] = "source=telegraf"
	sort.Strings(tagsString)
	return tagsString
}

func buildValue(v interface{}) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = intToString(p)
	case string:
		retv = fmt.Sprintf("'%s'", strings.Replace(p, "'", "\\'", -1))
	case bool:
		retv = boolToString(p)
	case uint64:
		if p <= uint64(math.MaxInt64) {
			retv = strconv.FormatInt(int64(p), 10)
		} else {
			retv = strconv.FormatInt(math.MaxInt64, 10)
		}
	case float64:
		retv = floatToString(p)
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
	return retv, nil
}

func intToString(inputNum int64) string {
	return strconv.FormatInt(inputNum, 10)
}

func boolToString(inputBool bool) string {
	return strconv.FormatBool(inputBool)
}

func floatToString(inputNum float64) string {
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}

// Close close
func (w *Warp10) Close() error {
	return nil
}

// Init Warp10 struct
func (w *Warp10) Init() error {
	if w.MaxStringErrorSize <= 0 {
		w.MaxStringErrorSize = 511
	}
	return nil
}

func init() {
	outputs.Add("warp10", func() telegraf.Output {
		return &Warp10{}
	})
}

// HandleError read http error body and return a corresponding error
func (w *Warp10) HandleError(body string, maxStringSize int) string {
	if body == "" {
		return "Empty return"
	}

	if strings.Contains(body, "Invalid token") {
		return "Invalid token"
	}

	if strings.Contains(body, "Write token missing") {
		return "Write token missing"
	}

	if strings.Contains(body, "Token Expired") {
		return "Token Expired"
	}

	if strings.Contains(body, "Token revoked") {
		return "Token revoked"
	}

	if strings.Contains(body, "exceed your Monthly Active Data Streams limit") || strings.Contains(body, "exceed the Monthly Active Data Streams limit") {
		return "Exceeded Monthly Active Data Streams limit"
	}

	if strings.Contains(body, "Daily Data Points limit being already exceeded") {
		return "Exceeded Daily Data Points limit"
	}

	if strings.Contains(body, "Application suspended or closed") {
		return "Application suspended or closed"
	}

	if strings.Contains(body, "broken pipe") {
		return "broken pipe"
	}

	if len(body) < maxStringSize {
		return body
	}
	return body[0:maxStringSize]
}
