package http_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"go.starlark.net/starlark"
)

const (
	// defaultResponseBodyMaxSize is the default maximum response body size, in bytes.
	// if the response body is over this size, we will raise a body_read_error.
	defaultResponseBodyMaxSize = 32 * 1024 * 1024
)

// HTTPResponse struct
type HTTPResponse struct {
	Address         string   // deprecated in 1.12
	URLs            []string `toml:"urls"`
	HTTPProxy       string   `toml:"http_proxy"`
	Body            string
	Method          string
	ResponseTimeout config.Duration
	HTTPHeaderTags  map[string]string `toml:"http_header_tags"`
	Headers         map[string]string
	FollowRedirects bool
	// Absolute path to file with Bearer token
	BearerToken         string      `toml:"bearer_token"`
	ResponseBodyField   string      `toml:"response_body_field"`
	ResponseBodyMaxSize config.Size `toml:"response_body_max_size"`
	ResponseStringMatch string
	ResponseStatusCode  int
	Interface           string
	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	tls.ClientConfig

	initalBody    string            // Initial body saves the original body.
	initialHeader map[string]string // Initial header saves the original header.
	initialScript string            // Initial script saves the original script

	ResponseSetENV map[string]string `toml:"response_set_env"`
	ScriptSetENV   map[string]string `toml:"script_set_env"`
	Script         string            `toml:"script"`

	thread   *starlark.Thread
	builtins starlark.StringDict

	Log telegraf.Logger

	compiledStringMatch *regexp.Regexp
	client              httpClient
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (h *HTTPResponse) Init() error {
	h.thread = &starlark.Thread{
		Print: func(_ *starlark.Thread, msg string) { h.Log.Debug(msg) },
	}
	builtins := starlark.StringDict{}
	builtins["md5"] = starlark.NewBuiltin("md5", builtinMD5)
	builtins["sha256"] = starlark.NewBuiltin("sha256", builtinSHA256)
	builtins["now"] = starlark.NewBuiltin("now", builtinNow)
	builtins["rand"] = starlark.NewBuiltin("rand", builtinRand)
	h.builtins = builtins
	return nil
}

// Description returns the plugin Description
func (h *HTTPResponse) Description() string {
	return "HTTP/HTTPS request given an address a method and a timeout"
}

var sampleConfig = `
  ## Deprecated in 1.12, use 'urls'
  ## Server address (default http://localhost)
  # address = "http://localhost"

  ## List of urls to query.
  # urls = ["http://localhost"]

  ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
  # http_proxy = "http://localhost:8888"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## HTTP Request Method
  # method = "GET"

  ## Whether to follow redirects from the server (defaults to false)
  # follow_redirects = false

  ## Optional file with Bearer token
  ## file content is added as an Authorization header
  # bearer_token = "/path/to/file"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''

  ## Optional name of the field that will contain the body of the response.
  ## By default it is set to an empty String indicating that the body's content won't be added
  # response_body_field = ''

  ## Maximum allowed HTTP response body size in bytes.
  ## 0 means to use the default of 32MiB.
  ## If the response body size exceeds this limit a "body_read_error" will be raised
  # response_body_max_size = "32MiB"

  ## Optional substring or regex match in body of the response (case sensitive)
  # response_string_match = "\"service_status\": \"up\""
  # response_string_match = "ok"
  # response_string_match = "\".*_status\".?:.?\"up\""

  ## Expected response status code.
  ## The status code of the response is compared to this value. If they match, the field
  ## "response_status_code_match" will be 1, otherwise it will be 0. If the
  ## expected status code is 0, the check is disabled and the field won't be added.
  # response_status_code = 0

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional pre-request script using Starlark.
  # script = 'timestamp = now()'

  ## Optional environment variable exporting from pre-request script
  ## Assign variable name from script to environment variable name, then reference
  ## the env using ${TIMESTAMP} notation in body, header or script.
  [inputs.http_response.script_set_env]
  # TIMESTAMP = "timestamp"

  ## Optional environment variable exporting from response body
  ## Assign field name from response body to environment variable name, then reference
  ## the env using ${SESSIONID} notation in body, header or script.
  [inputs.http_response.response_set_env]
  SESSIONID = "sessionId"

  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"

  ## Optional setting to map response http headers into tags
  ## If the http header is not present on the request, no corresponding tag will be added
  ## If multiple instances of the http header are present, only the first value will be used
  # http_header_tags = {"HTTP_HEADER" = "TAG_NAME"}

  ## Interface to use when dialing an address
  # interface = "eth0"
`

// SampleConfig returns the plugin SampleConfig
func (h *HTTPResponse) SampleConfig() string {
	return sampleConfig
}

// ErrRedirectAttempted indicates that a redirect occurred
var ErrRedirectAttempted = errors.New("redirect")

// Set the proxy. A configured proxy overwrites the system wide proxy.
func getProxyFunc(httpProxy string) func(*http.Request) (*url.URL, error) {
	if httpProxy == "" {
		return http.ProxyFromEnvironment
	}
	proxyURL, err := url.Parse(httpProxy)
	if err != nil {
		return func(_ *http.Request) (*url.URL, error) {
			return nil, errors.New("bad proxy: " + err.Error())
		}
	}
	return func(r *http.Request) (*url.URL, error) {
		return proxyURL, nil
	}
}

// createHTTPClient creates an http client which will timeout at the specified
// timeout period and can follow redirects if specified
func (h *HTTPResponse) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{}

	if h.Interface != "" {
		dialer.LocalAddr, err = localAddress(h.Interface)
		if err != nil {
			return nil, err
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:             getProxyFunc(h.HTTPProxy),
			DialContext:       dialer.DialContext,
			DisableKeepAlives: true,
			TLSClientConfig:   tlsCfg,
		},
		Timeout: time.Duration(h.ResponseTimeout),
	}

	if !h.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client, nil
}

func localAddress(interfaceName string) (net.Addr, error) {
	i, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	addrs, err := i.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if naddr, ok := addr.(*net.IPNet); ok {
			// leaving port set to zero to let kernel pick
			return &net.TCPAddr{IP: naddr.IP}, nil
		}
	}

	return nil, fmt.Errorf("cannot create local address for interface %q", interfaceName)
}

func setResult(resultString string, fields map[string]interface{}, tags map[string]string) {
	resultCodes := map[string]int{
		"success":                       0,
		"response_string_mismatch":      1,
		"body_read_error":               2,
		"connection_failed":             3,
		"timeout":                       4,
		"dns_error":                     5,
		"response_status_code_mismatch": 6,
	}

	tags["result"] = resultString
	fields["result_type"] = resultString
	fields["result_code"] = resultCodes[resultString]
}

func setError(err error, fields map[string]interface{}, tags map[string]string) error {
	if timeoutError, ok := err.(net.Error); ok && timeoutError.Timeout() {
		setResult("timeout", fields, tags)
		return timeoutError
	}

	urlErr, isURLErr := err.(*url.Error)
	if !isURLErr {
		return nil
	}

	opErr, isNetErr := (urlErr.Err).(*net.OpError)
	if isNetErr {
		switch e := (opErr.Err).(type) {
		case *net.DNSError:
			setResult("dns_error", fields, tags)
			return e
		case *net.ParseError:
			// Parse error has to do with parsing of IP addresses, so we
			// group it with address errors
			setResult("address_error", fields, tags)
			return e
		}
	}

	return nil
}

func (h *HTTPResponse) findField(field, value string, m map[string]interface{}) string {
	if value != "" || m == nil {
		return value
	}
	for k, v := range m {
		if k == field {
			return fmt.Sprintf("%s", v)
		}
		vv, ok := v.(map[string]interface{})
		if ok {
			value := h.findField(field, value, vv)
			if value != "" {
				return value
			}
		}
	}
	return ""
}

// setENV replaces environment variable reference with actual value in body, header and pre-request script.
func (h *HTTPResponse) setENV(target string) error {
	rx := regexp.MustCompile(`(?s)` + regexp.QuoteMeta("${") + `(.*?)` + regexp.QuoteMeta("}"))
	for _, match := range rx.FindAllStringSubmatch(target, -1) {
		envVar := os.Getenv(match[1])
		if envVar == "" {
			return fmt.Errorf("env: %s not found for %v", match[1], h.URLs)
		}
		h.Script = strings.ReplaceAll(h.Script, match[0], os.Getenv(match[1]))
		h.Body = strings.ReplaceAll(h.Body, match[0], os.Getenv(match[1]))
		for k, v := range h.Headers {
			v = strings.ReplaceAll(v, match[0], os.Getenv(match[1]))
			h.Headers[k] = v
		}
	}
	return nil
}

// HTTPGather gathers all fields and returns any errors it encounters
func (h *HTTPResponse) httpGather(u string) (map[string]interface{}, map[string]string, error) {
	// Prepare fields and tags
	fields := make(map[string]interface{})
	tags := map[string]string{"server": u, "method": h.Method}

	var body io.Reader
	if h.Body != "" {
		body = strings.NewReader(h.Body)
	}
	request, err := http.NewRequest(h.Method, u, body)
	if err != nil {
		return nil, nil, err
	}

	if h.BearerToken != "" {
		token, err := os.ReadFile(h.BearerToken)
		if err != nil {
			return nil, nil, err
		}
		bearer := "Bearer " + strings.Trim(string(token), "\n")
		request.Header.Add("Authorization", bearer)
	}

	for key, val := range h.Headers {
		request.Header.Add(key, val)
		if key == "Host" {
			request.Host = val
		}
	}

	if h.Username != "" || h.Password != "" {
		request.SetBasicAuth(h.Username, h.Password)
	}

	// Start Timer
	start := time.Now()
	resp, err := h.client.Do(request)
	responseTime := time.Since(start).Seconds()

	// If an error in returned, it means we are dealing with a network error, as
	// HTTP error codes do not generate errors in the net/http library
	if err != nil {
		// Log error
		h.Log.Debugf("Network error while polling %s: %s", u, err.Error())

		// Get error details
		if setError(err, fields, tags) == nil {
			// Any error not recognized by `set_error` is considered a "connection_failed"
			setResult("connection_failed", fields, tags)
		}

		return fields, tags, nil
	}

	if _, ok := fields["response_time"]; !ok {
		fields["response_time"] = responseTime
	}

	// This function closes the response body, as
	// required by the net/http library
	defer resp.Body.Close()

	// Add the response headers
	for headerName, tag := range h.HTTPHeaderTags {
		headerValues, foundHeader := resp.Header[headerName]
		if foundHeader && len(headerValues) > 0 {
			tags[tag] = headerValues[0]
		}
	}

	// Set log the HTTP response code
	tags["status_code"] = strconv.Itoa(resp.StatusCode)
	fields["http_response_code"] = resp.StatusCode

	if h.ResponseBodyMaxSize == 0 {
		h.ResponseBodyMaxSize = config.Size(defaultResponseBodyMaxSize)
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(h.ResponseBodyMaxSize)+1))
	// Check first if the response body size exceeds the limit.
	if err == nil && int64(len(bodyBytes)) > int64(h.ResponseBodyMaxSize) {
		h.setBodyReadError("The body of the HTTP Response is too large", bodyBytes, fields, tags)
		return fields, tags, nil
	} else if err != nil {
		h.setBodyReadError(fmt.Sprintf("Failed to read body of HTTP Response : %s", err.Error()), bodyBytes, fields, tags)
		return fields, tags, nil
	}

	// Add the body of the response if expected
	if len(h.ResponseBodyField) > 0 {
		// Check that the content of response contains only valid utf-8 characters.
		if !utf8.Valid(bodyBytes) {
			h.setBodyReadError("The body of the HTTP Response is not a valid utf-8 string", bodyBytes, fields, tags)
			return fields, tags, nil
		}
		fields[h.ResponseBodyField] = string(bodyBytes)
	}
	fields["content_length"] = len(bodyBytes)

	var success = true

	// Check the response for a regex
	if h.ResponseStringMatch != "" {
		if h.compiledStringMatch.Match(bodyBytes) {
			fields["response_string_match"] = 1
		} else {
			success = false
			setResult("response_string_mismatch", fields, tags)
			fields["response_string_match"] = 0
		}
	}

	if h.ResponseSetENV != nil {
		bodyMap := map[string]interface{}{}
		err := json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			return nil, nil, err
		} else {
			for env, responseVar := range h.ResponseSetENV {
				value := h.findField(responseVar, "", bodyMap)
				if value == "" {
					return nil, nil, fmt.Errorf("response field %s not found with body %s", responseVar, string(bodyBytes))
				}
				if value != "" {
					err = os.Setenv(env, value)
					if err != nil {
						return nil, nil, err
					}
				}
			}
		}
	}

	// Check the response status code
	if h.ResponseStatusCode > 0 {
		if resp.StatusCode == h.ResponseStatusCode {
			fields["response_status_code_match"] = 1
		} else {
			success = false
			setResult("response_status_code_mismatch", fields, tags)
			fields["response_status_code_match"] = 0
		}
	}

	if success {
		setResult("success", fields, tags)
	}

	return fields, tags, nil
}

// Set result in case of a body read error
func (h *HTTPResponse) setBodyReadError(errorMsg string, bodyBytes []byte, fields map[string]interface{}, tags map[string]string) {
	h.Log.Debugf(errorMsg)
	setResult("body_read_error", fields, tags)
	fields["content_length"] = len(bodyBytes)
	if h.ResponseStringMatch != "" {
		fields["response_string_match"] = 0
	}
}

// Gather gets all metric fields and tags and returns any errors it encounters
func (h *HTTPResponse) Gather(acc telegraf.Accumulator) error {
	// Compile the body regex if it exist
	if h.compiledStringMatch == nil {
		var err error
		h.compiledStringMatch, err = regexp.Compile(h.ResponseStringMatch)
		if err != nil {
			return fmt.Errorf("failed to compile regular expression %s : %s", h.ResponseStringMatch, err)
		}
	}

	// Save original copy of body
	if h.initalBody == "" {
		h.initalBody = h.Body
	}

	if h.initialHeader == nil {
		h.initialHeader = map[string]string{}
		for k, v := range h.Headers {
			h.initialHeader[k] = v
		}
	}

	if h.initialScript == "" {
		h.initialScript = h.Script
	}
	// Initialize body to original state
	h.Body = h.initalBody

	for k, v := range h.initialHeader {
		h.Headers[k] = v
	}

	h.Script = h.initialScript

	// Set default values
	if h.ResponseTimeout < config.Duration(time.Second) {
		h.ResponseTimeout = config.Duration(time.Second * 5)
	}
	// Check send and expected string
	if h.Method == "" {
		h.Method = "GET"
	}

	if len(h.URLs) == 0 {
		if h.Address == "" {
			h.URLs = []string{"http://localhost"}
		} else {
			h.Log.Warn("'address' deprecated in telegraf 1.12, please use 'urls'")
			h.URLs = []string{h.Address}
		}
	}

	err := h.setENV(h.Script)
	if err != nil {
		return err
	}

	if h.Script != "" {
		_, program, err := starlark.SourceProgram("script", h.Script, h.builtins.Has)
		if err != nil {
			return err
		}
		globals, err := program.Init(h.thread, h.builtins)
		if err != nil {
			return err
		}
		if h.ScriptSetENV != nil {
			for envVar, scriptVar := range h.ScriptSetENV {
				val, ok := globals[scriptVar]
				if !ok {
					return fmt.Errorf("failed to get pre-request value for variable %s", scriptVar)
				}
				valStr, ok := starlark.AsString(val)
				if !ok {
					return fmt.Errorf("failed to convert pre-request result value %v to string", val)
				}
				err = os.Setenv(envVar, starlark.String(valStr).GoString())
				if err != nil {
					return fmt.Errorf("failed to set environment variable %s with value %v, error: %s", envVar, val.String(), err)
				}
			}
		}

	}

	err = h.setENV(h.Body)
	if err != nil {
		return err
	}

	for _, v := range h.Headers {
		err = h.setENV(v)
		if err != nil {
			return err
		}
	}

	if h.client == nil {
		client, err := h.createHTTPClient()
		if err != nil {
			return err
		}
		h.client = client
	}

	for _, u := range h.URLs {
		addr, err := url.Parse(u)
		if err != nil {
			acc.AddError(err)
			continue
		}

		if addr.Scheme != "http" && addr.Scheme != "https" {
			acc.AddError(errors.New("only http and https are supported"))
			continue
		}

		// Prepare data
		var fields map[string]interface{}
		var tags map[string]string

		// Gather data
		fields, tags, err = h.httpGather(u)
		if err != nil {
			acc.AddError(err)
			continue
		}

		// Add metrics
		acc.AddFields("http_response", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("http_response", func() telegraf.Input {
		return &HTTPResponse{}
	})
}
