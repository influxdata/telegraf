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
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultClientTimeout = 15 * time.Second
)

// tokenAuthError indicates a non-retryable authentication error from Warp 10.
type tokenAuthError struct {
	msg string
}

func (e *tokenAuthError) Error() string {
	return e.msg
}

// Warp10 output plugin
type Warp10 struct {
	Prefix             string          `toml:"prefix"`
	WarpURL            string          `toml:"warp_url"`
	Token              config.Secret   `toml:"token"`
	Timeout            config.Duration `toml:"timeout"`
	PrintErrorBody     bool            `toml:"print_error_body"`
	MaxStringErrorSize int             `toml:"max_string_error_size"`
	AuthErrorRetries   uint            `toml:"auth_error_retries"`
	client             *http.Client
	failedToken        string
	authRetriesLeft    uint
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

	token, err := w.Token.Get()
	if err != nil {
		return fmt.Errorf("getting token failed: %w", err)
	}
	currentToken := token.String()
	token.Destroy()

	// Check if we are in a failure state
	if w.failedToken != "" {
		if w.failedToken != currentToken {
			// Token changed (e.g. secret-store refresh), reset failure state
			w.failedToken = ""
			w.authRetriesLeft = 0
		} else if w.authRetriesLeft > 0 {
			// Still waiting — decrement and drop metrics
			w.authRetriesLeft--
			w.Log.Warnf("Dropping metrics: auth error retry cooldown (%d flushes left)", w.authRetriesLeft)
			return nil
		}
		// authRetriesLeft == 0: retry this flush
	}

	addr := w.WarpURL + "/api/v0/update"
	req, err := http.NewRequest("POST", addr, bytes.NewBufferString(payload))
	if err != nil {
		return fmt.Errorf("unable to create new request %q: %w", addr, err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Warp10-Token", currentToken)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		respErr := handleError(string(body), w.MaxStringErrorSize)

		var authErr *tokenAuthError
		if errors.As(respErr, &authErr) {
			w.failedToken = currentToken
			w.authRetriesLeft = w.AuthErrorRetries
		}

		if w.PrintErrorBody {
			return fmt.Errorf("%s: %w", w.WarpURL, respErr)
		}

		if len(resp.Status) < w.MaxStringErrorSize {
			return fmt.Errorf("%s: %s", w.WarpURL, resp.Status)
		}
		return fmt.Errorf("%s: %s", w.WarpURL, resp.Status[0:w.MaxStringErrorSize])
	}

	// Success — clear any failure state
	if w.failedToken != "" {
		w.failedToken = ""
		w.authRetriesLeft = 0
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

// handleError parses the HTTP response body and returns the corresponding error.
// For authentication errors it returns a *tokenAuthError (non-retryable).
// For all other errors it returns a plain error (retryable).
func handleError(body string, maxStringSize int) error {
	if body == "" {
		return errors.New("empty return")
	}

	switch {
	case strings.Contains(body, "Invalid token"):
		return &tokenAuthError{msg: "Invalid token"}
	case strings.Contains(body, "Write token missing"):
		return &tokenAuthError{msg: "Write token missing"}
	case strings.Contains(body, "Token Expired"):
		return &tokenAuthError{msg: "Token Expired"}
	case strings.Contains(body, "Token revoked"):
		return &tokenAuthError{msg: "Token revoked"}
	case strings.Contains(body, "Application suspended or closed"):
		return &tokenAuthError{msg: "Application suspended or closed"}
	}

	// Retryable errors (quota limits, transient failures)
	switch {
	case strings.Contains(body, "exceed your Monthly Active Data Streams limit"),
		strings.Contains(body, "exceed the Monthly Active Data Streams limit"):
		return errors.New("Exceeded Monthly Active Data Streams limit")
	case strings.Contains(body, "Daily Data Points limit being already exceeded"):
		return errors.New("Exceeded Daily Data Points limit")
	case strings.Contains(body, "broken pipe"):
		return errors.New("broken pipe")
	}

	if len(body) < maxStringSize {
		return errors.New(body)
	}
	return errors.New(body[0:maxStringSize])
}
