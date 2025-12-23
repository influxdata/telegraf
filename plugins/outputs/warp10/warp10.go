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
	// maxAuthFailures is the maximum number of consecutive authentication failures
	// with the same token before dropping metrics. Prevents unbounded buffer growth
	// when using static tokens that will never change.
	maxAuthFailures = 3
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

	// lastFailedToken stores the token value that caused an authentication error.
	// Used to detect when the token has been refreshed by a secret-store.
	lastFailedToken  string
	authFailureCount int
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
	// Count total fields for preallocation
	totalFields := 0
	for _, mm := range metrics {
		totalFields += len(mm.FieldList())
	}

	collectString := make([]string, 0, totalFields)
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

// checkAuthFailureState checks if we're in an authentication failure state and determines
// whether to proceed with the write, skip it, or drop the metrics.
// Returns:
//   - proceed: true if the write should proceed, false otherwise
//   - err: non-nil if metrics should be skipped (plain error) or dropped (PartialWriteError)
func (w *Warp10) checkAuthFailureState(currentToken string, metricCount int) (proceed bool, err error) {
	if w.lastFailedToken == "" {
		return true, nil
	}

	if currentToken != w.lastFailedToken {
		// Token changed - clear failure state and proceed
		w.Log.Infof("Token changed, resuming writes after previous authentication failure")
		w.clearAuthFailureState()
		return true, nil
	}

	// Same token as the one that failed
	w.authFailureCount++
	if w.authFailureCount >= maxAuthFailures {
		// Max retries reached - drop metrics
		w.Log.Errorf("Authentication failure persists after %d attempts with same token, dropping metrics", w.authFailureCount)
		w.clearAuthFailureState()
		indices := make([]int, metricCount)
		for i := range indices {
			indices[i] = i
		}
		return false, &internal.PartialWriteError{
			Err:           errors.New("authentication failure: max retries exceeded"),
			MetricsReject: indices,
		}
	}

	// Same token, not yet at max - skip write to wait for token refresh
	w.Log.Debugf("Skipping write, waiting for token refresh (attempt %d/%d)", w.authFailureCount, maxAuthFailures)
	return false, fmt.Errorf("authentication failure pending token refresh (attempt %d/%d)", w.authFailureCount, maxAuthFailures)
}

// recordAuthFailure records an authentication failure with the given token.
func (w *Warp10) recordAuthFailure(token string) {
	w.lastFailedToken = token
	w.authFailureCount = 1
}

// clearAuthFailureState clears the authentication failure tracking state.
func (w *Warp10) clearAuthFailureState() {
	w.lastFailedToken = ""
	w.authFailureCount = 0
}

// Write metrics to Warp10
func (w *Warp10) Write(metrics []telegraf.Metric) error {
	payload := w.GenWarp10Payload(metrics)
	if payload == "" {
		return nil
	}

	// Get token once at the beginning for both auth failure check and request
	token, err := w.Token.Get()
	if err != nil {
		return fmt.Errorf("getting token failed: %w", err)
	}
	currentTokenValue := token.String()
	token.Destroy()

	// Check if we should proceed based on auth failure state
	proceed, err := w.checkAuthFailureState(currentTokenValue, len(metrics))
	if !proceed {
		return err
	}

	addr := w.WarpURL + "/api/v0/update"
	req, err := http.NewRequest("POST", addr, bytes.NewBufferString(payload))
	if err != nil {
		return fmt.Errorf("unable to create new request %q: %w", addr, err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Warp10-Token", currentTokenValue)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.Log.Debugf("Failed to read response body: %v", err)
		}
		werr := HandleError(string(body), w.MaxStringErrorSize)
		fullErr := fmt.Errorf("%s: %w", w.WarpURL, werr)

		if !werr.Retryable {
			// Authentication error - record failure and return error for retry
			w.recordAuthFailure(currentTokenValue)
			w.Log.Warnf("Authentication error, will retry if token changes: %s", werr.Error())
			// Return plain error (not PartialWriteError) to keep metrics in buffer
			return fullErr
		}

		// Retryable error - clear auth failure state, return plain error for normal retry
		w.clearAuthFailureState()
		return fullErr
	}

	// Success - clear any failure state
	w.clearAuthFailureState()
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
