//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package http

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_signer "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	common_aws "github.com/influxdata/telegraf/plugins/common/aws"
	common_gcp "github.com/influxdata/telegraf/plugins/common/gcp"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultContentType = "text/plain; charset=utf-8"
	defaultMethod      = http.MethodPost
	defaultOnError     = "keep"
	tagStatusCode      = "status_code"
	tagResult          = "result"
	fieldHTTPError     = "http_error"
)

type HTTP struct {
	URL             string `toml:"url"`
	Method          string `toml:"method"`
	DropOriginal    bool   `toml:"drop_original"`
	Merge           string `toml:"merge"`
	OnError         string `toml:"on_error"`
	ContentEncoding string `toml:"content_encoding"`

	Username config.Secret             `toml:"username"`
	Password config.Secret             `toml:"password"`
	Headers  map[string]*config.Secret `toml:"headers"`

	SuccessStatusCodes []int `toml:"success_status_codes"`

	AwsService      string `toml:"aws_service"`
	CredentialsFile string `toml:"google_application_credentials"`

	common_http.HTTPClientConfig
	common_aws.CredentialConfig

	Log telegraf.Logger `toml:"-"`

	client      *http.Client
	serializer  telegraf.Serializer
	parserFunc  telegraf.ParserFunc
	awsCfg      *aws.Config
	oauth2Token *oauth2.Token
}

func (*HTTP) SampleConfig() string {
	return sampleConfig
}

func (h *HTTP) SetSerializer(serializer telegraf.Serializer) {
	h.serializer = serializer
}

func (h *HTTP) SetParserFunc(fn telegraf.ParserFunc) {
	h.parserFunc = fn
}

func (h *HTTP) Init() error {
	if h.URL == "" {
		return errors.New("url is required")
	}
	if h.serializer == nil {
		return errors.New("serializer not configured")
	}
	if h.parserFunc == nil {
		return errors.New("parser not configured")
	}

	switch h.Merge {
	case "":
		h.Merge = "none"
	case "none", "override", "parent":
	case "override-with-timestamp", "parent-with-timestamp":
	default:
		return fmt.Errorf("unrecognized merge value: %s", h.Merge)
	}

	switch h.OnError {
	case "":
		h.OnError = defaultOnError
	case "keep", "drop":
	default:
		return fmt.Errorf("invalid on_error %q", h.OnError)
	}

	if h.AwsService != "" {
		cfg, err := h.CredentialConfig.Credentials()
		if err == nil {
			h.awsCfg = &cfg
		}
	}

	if h.Method == "" {
		h.Method = defaultMethod
	}
	h.Method = strings.ToUpper(h.Method)
	if h.Method == http.MethodGet {
		return fmt.Errorf("invalid method [%s] %s: GET is not supported", h.URL, h.Method)
	}
	if h.Method != http.MethodPost && h.Method != http.MethodPut && h.Method != http.MethodPatch {
		return fmt.Errorf("invalid method [%s] %s", h.URL, h.Method)
	}

	if len(h.SuccessStatusCodes) == 0 {
		h.SuccessStatusCodes = []int{http.StatusOK}
	}

	ctx := context.Background()
	client, err := h.HTTPClientConfig.CreateClient(ctx, h.Log)
	if err != nil {
		return err
	}
	h.client = client

	return nil
}

type httpResult struct {
	statusCode int
	body       []byte
}

type requestError struct {
	result string
	resp   *httpResult
	err    error
}

func (e *requestError) Error() string {
	return e.err.Error()
}

func (e *requestError) Unwrap() error {
	return e.err
}

func requestErr(result string, resp *httpResult, err error) error {
	return &requestError{
		result: result,
		resp:   resp,
		err:    err,
	}
}

func (h *HTTP) Apply(in ...telegraf.Metric) []telegraf.Metric {
	results := make([]telegraf.Metric, 0, len(in))
	for _, metric := range in {
		out, err := h.processMetric(metric)
		if err != nil {
			h.Log.Errorf("%v", err)
			if h.OnError == "drop" {
				metric.Drop()
				continue
			}
			annotateErrorMetric(metric, err)
			results = append(results, metric)
			continue
		}
		results = append(results, out...)
	}
	return results
}

func (h *HTTP) processMetric(metric telegraf.Metric) ([]telegraf.Metric, error) {
	reqBody, err := h.serializer.Serialize(metric)
	if err != nil {
		return nil, requestErr("connection_failed", nil, fmt.Errorf("serializing metric failed: %w", err))
	}

	resp, err := h.doRequest(reqBody)
	if err != nil {
		return nil, err
	}

	parser, err := h.parserFunc()
	if err != nil {
		return nil, requestErr("connection_failed", resp, fmt.Errorf("instantiating parser failed: %w", err))
	}
	h.configureParser(parser)

	responseMetrics, err := parser.Parse(resp.body)
	if err != nil {
		return nil, requestErr("body_read_error", resp, fmt.Errorf("parsing response failed: %w", err))
	}
	if len(responseMetrics) == 0 {
		return nil, requestErr("body_read_error", resp, errors.New("parser returned no metrics"))
	}

	for _, m := range responseMetrics {
		if m.Name() == "" || m.Name() == "http" {
			m.SetName(metric.Name())
		}
	}

	var newMetrics []telegraf.Metric
	if !h.DropOriginal {
		newMetrics = append(newMetrics, metric)
	} else {
		metric.Drop()
	}
	newMetrics = append(newMetrics, responseMetrics...)

	var results []telegraf.Metric
	switch h.Merge {
	case "override":
		results = []telegraf.Metric{mergeAll(newMetrics[0], newMetrics[1:], false)}
	case "override-with-timestamp":
		results = []telegraf.Metric{mergeAll(newMetrics[0], newMetrics[1:], true)}
	case "parent":
		results = mergeIndividual(metric, responseMetrics, false)
		if !h.DropOriginal {
			results = append([]telegraf.Metric{metric}, results...)
		}
	case "parent-with-timestamp":
		results = mergeIndividual(metric, responseMetrics, true)
		if !h.DropOriginal {
			results = append([]telegraf.Metric{metric}, results...)
		}
	default:
		results = newMetrics
	}

	addSuccessMetadata(results, resp)
	return results, nil
}

func (h *HTTP) configureParser(parser telegraf.Parser) {
	if h.Merge != "override-with-timestamp" && h.Merge != "parent-with-timestamp" {
		return
	}

	unwrapped := parser
	if rp, ok := parser.(*models.RunningParser); ok {
		unwrapped = rp.Parser
	}
	if ptfp, ok := unwrapped.(telegraf.ParserTimeFuncPlugin); ok {
		ptfp.SetTimeFunc(func() time.Time { return time.Time{} })
	} else {
		h.Log.Warnf("Parser will always create a timestamp in merge-mode %q!", h.Merge)
	}
}

func (h *HTTP) doRequest(reqBody []byte) (*httpResult, error) {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)

	if h.ContentEncoding == "gzip" {
		rc := internal.CompressWithGzip(reqBodyBuffer)
		defer rc.Close()
		reqBodyBuffer = rc
	}

	var payloadHash *string
	if h.awsCfg != nil {
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, reqBodyBuffer); err != nil {
			return nil, err
		}

		sum := sha256.Sum256(buf.Bytes())
		reqBodyBuffer = buf

		hash := hex.EncodeToString(sum[:])
		payloadHash = &hash
	}

	req, err := http.NewRequest(h.Method, h.URL, reqBodyBuffer)
	if err != nil {
		return nil, err
	}

	if h.awsCfg != nil {
		signer := aws_signer.NewSigner()
		ctx := context.Background()

		credentials, err := h.awsCfg.Credentials.Retrieve(ctx)
		if err != nil {
			return nil, err
		}

		if err := signer.SignHTTP(ctx, credentials, req, *payloadHash, h.AwsService, h.Region, time.Now().UTC()); err != nil {
			return nil, err
		}
	}

	if !h.Username.Empty() || !h.Password.Empty() {
		username, err := h.Username.Get()
		if err != nil {
			return nil, fmt.Errorf("getting username failed: %w", err)
		}
		password, err := h.Password.Get()
		if err != nil {
			username.Destroy()
			return nil, fmt.Errorf("getting password failed: %w", err)
		}
		req.SetBasicAuth(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	}

	if h.CredentialsFile != "" {
		token, err := h.getAccessToken(context.Background(), h.URL)
		if err != nil {
			return nil, err
		}
		token.SetAuthHeader(req)
	}

	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Set("Content-Type", defaultContentType)
	if h.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range h.Headers {
		secret, err := v.Get()
		if err != nil {
			return nil, err
		}

		headerVal := secret.String()
		if strings.EqualFold(k, "host") {
			req.Host = headerVal
		} else {
			req.Header.Set(k, headerVal)
		}

		secret.Destroy()
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, requestErr(classifyNetworkError(err), nil, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, requestErr("body_read_error", &httpResult{statusCode: resp.StatusCode},
			fmt.Errorf("reading body failed: %w", err))
	}

	result := &httpResult{
		statusCode: resp.StatusCode,
		body:       body,
	}

	responseHasSuccessCode := false
	for _, statusCode := range h.SuccessStatusCodes {
		if resp.StatusCode == statusCode {
			responseHasSuccessCode = true
			break
		}
	}
	if !responseHasSuccessCode {
		errorLine := firstLine(body)
		return result, requestErr("response_status_code_mismatch", result, fmt.Errorf(
			"received status code %d (%s), expected any value out of %v. body: %s",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			h.SuccessStatusCodes,
			errorLine,
		))
	}

	return result, nil
}

func firstLine(body []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func addSuccessMetadata(metrics []telegraf.Metric, resp *httpResult) {
	for _, m := range metrics {
		setResult(m, "success")
		if resp != nil && resp.statusCode > 0 {
			m.AddTag(tagStatusCode, strconv.Itoa(resp.statusCode))
		}
	}
}

func annotateErrorMetric(metric telegraf.Metric, err error) {
	result := "connection_failed"
	var resp *httpResult

	var re *requestError
	if errors.As(err, &re) {
		result = re.result
		resp = re.resp
	} else if networkResult := classifyNetworkError(err); networkResult != "" {
		result = networkResult
	}

	if resp != nil && resp.statusCode > 0 {
		metric.AddTag(tagStatusCode, strconv.Itoa(resp.statusCode))
	}
	metric.AddField(fieldHTTPError, err.Error())
	setResult(metric, result)
}

func setResult(metric telegraf.Metric, result string) {
	metric.AddTag(tagResult, result)
}

func classifyNetworkError(err error) string {
	var timeoutErr net.Error
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		return "timeout"
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var opErr *net.OpError
		if errors.As(urlErr, &opErr) {
			var dnsErr *net.DNSError
			if errors.As(opErr, &dnsErr) {
				return "dns_error"
			}
		}
	}

	return "connection_failed"
}

func mergeIndividual(base telegraf.Metric, metrics []telegraf.Metric, mergeTime bool) []telegraf.Metric {
	result := make([]telegraf.Metric, 0, len(metrics))
	for _, metric := range metrics {
		out := base.Copy()
		for _, field := range metric.FieldList() {
			out.AddField(field.Key, field.Value)
		}
		for _, tag := range metric.TagList() {
			out.AddTag(tag.Key, tag.Value)
		}
		out.SetName(metric.Name())
		if mergeTime && !metric.Time().IsZero() {
			out.SetTime(metric.Time())
		}
		result = append(result, out)
	}

	return result
}

func mergeAll(base telegraf.Metric, metrics []telegraf.Metric, mergeTime bool) telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			base.AddField(field.Key, field.Value)
		}
		for _, tag := range metric.TagList() {
			base.AddTag(tag.Key, tag.Value)
		}
		base.SetName(metric.Name())

		if mergeTime && !metric.Time().IsZero() {
			base.SetTime(metric.Time())
		}
	}
	return base
}

func (h *HTTP) getAccessToken(ctx context.Context, audience string) (*oauth2.Token, error) {
	if h.oauth2Token.Valid() {
		return h.oauth2Token, nil
	}

	credType, err := common_gcp.ParseCredentialType(h.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials file type: %w", err)
	}

	ts, err := idtoken.NewTokenSource(ctx, audience, idtoken.WithAuthCredentialsFile(idtoken.CredentialsType(credType), h.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("error creating oauth2 token source: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("error fetching oauth2 token: %w", err)
	}

	h.oauth2Token = token

	return token, nil
}

func init() {
	processors.Add("http", func() telegraf.Processor {
		return &HTTP{
			Method:  defaultMethod,
			OnError: defaultOnError,
		}
	})
}
