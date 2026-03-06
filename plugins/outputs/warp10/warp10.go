//go:generate ../../../tools/readme_config_includer/generator
package warp10

import (
	"bytes"
	_ "embed"
	"errors"
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
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultClientTimeout = 15 * time.Second
)

// Warp10 output plugin
type Warp10 struct {
	Prefix             string          `toml:"prefix"`
	WarpURL            string          `toml:"warp_url"`
	Token              config.Secret   `toml:"token"`
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

func (*Warp10) SampleConfig() string {
	return sampleConfig
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
	return strings.Join(collectString, "")
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
		return fmt.Errorf("unable to create new request %q: %w", addr, err)
	}

	req.Header.Set("Content-Type", "text/plain")
	token, err := w.Token.Get()
	if err != nil {
		return fmt.Errorf("getting token failed: %w", err)
	}
	req.Header.Set("X-Warp10-Token", token.String())
	token.Destroy()

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//nolint:errcheck // err can be ignored since it is just for logging
		body, _ := io.ReadAll(resp.Body)
		werr := HandleError(string(body), w.MaxStringErrorSize)
		fullErr := fmt.Errorf("%s: %w", w.WarpURL, werr)

		if !werr.Retryable {
			w.Log.Errorf("Non-retryable error, metrics will be dropped: %s", werr.Error())
			// Return PartialWriteError with all metrics rejected
			indices := make([]int, len(metrics))
			for i := range metrics {
				indices[i] = i
			}
			return &internal.PartialWriteError{
				Err:           fullErr,
				MetricsReject: indices,
			}
		}

		// Retryable error - return plain error, metrics stay in buffer for retry
		return fullErr
	}

	return nil
}

func buildTags(tags []*telegraf.Tag) []string {
	tagsString := make([]string, 0, len(tags)+1)
	for _, tag := range tags {
		key := url.QueryEscape(tag.Key)
		value := url.QueryEscape(tag.Value)
		tagsString = append(tagsString, fmt.Sprintf("%s=%s", key, value))
	}
	tagsString = append(tagsString, "source=telegraf")
	sort.Strings(tagsString)
	return tagsString
}

func buildValue(v interface{}) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = intToString(p)
	case string:
		retv = fmt.Sprintf("'%s'", strings.ReplaceAll(p, "'", "\\'"))
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

/*
Warp10 supports Infinity/-Infinity/NaN
<'
// class{label=value} 42.0
0// class-1{label=value}{attribute=value} 42
=1// Infinity
'>
PARSE

<'
// class{label=value} 42.0
0// class-1{label=value}{attribute=value} 42
=1// -Infinity
'>
PARSE

<'
// class{label=value} 42.0
0// class-1{label=value}{attribute=value} 42
=1// NaN
'>
PARSE
*/
func floatToString(inputNum float64) string {
	switch {
	case math.IsNaN(inputNum):
		return "NaN"
	case math.IsInf(inputNum, -1):
		return "-Infinity"
	case math.IsInf(inputNum, 1):
		return "Infinity"
	}
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}

// Close close
func (*Warp10) Close() error {
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

// HandleError read http error body and return a corresponding HTTPError with retry information.
// Note: Warp10 doesn't follow REST conventions - it returns 200/500 status codes, so error type
// is determined by parsing the response body rather than the HTTP status code.
func HandleError(body string, maxStringSize int) *internal.HTTPError {
	if body == "" {
		return &internal.HTTPError{Err: errors.New("empty return"), Retryable: true}
	}

	// Non-retryable authentication/authorization errors
	if strings.Contains(body, "Invalid token") {
		return &internal.HTTPError{Err: errors.New("invalid token"), Retryable: false}
	}

	if strings.Contains(body, "Write token missing") {
		return &internal.HTTPError{Err: errors.New("write token missing"), Retryable: false}
	}

	if strings.Contains(body, "Token Expired") {
		return &internal.HTTPError{Err: errors.New("token expired"), Retryable: false}
	}

	if strings.Contains(body, "Token revoked") {
		return &internal.HTTPError{Err: errors.New("token revoked"), Retryable: false}
	}

	if strings.Contains(body, "Application suspended or closed") {
		return &internal.HTTPError{Err: errors.New("application suspended or closed"), Retryable: false}
	}

	// Retryable errors (rate limits, temporary issues)
	if strings.Contains(body, "exceed your Monthly Active Data Streams limit") || strings.Contains(body, "exceed the Monthly Active Data Streams limit") {
		return &internal.HTTPError{Err: errors.New("exceeded Monthly Active Data Streams limit"), Retryable: true}
	}

	if strings.Contains(body, "Daily Data Points limit being already exceeded") {
		return &internal.HTTPError{Err: errors.New("exceeded Daily Data Points limit"), Retryable: true}
	}

	if strings.Contains(body, "broken pipe") {
		return &internal.HTTPError{Err: errors.New("broken pipe"), Retryable: true}
	}

	// Unknown errors - retryable by default
	msg := body
	if len(body) >= maxStringSize {
		msg = body[0:maxStringSize]
	}
	return &internal.HTTPError{Err: errors.New(msg), Retryable: true}
}
