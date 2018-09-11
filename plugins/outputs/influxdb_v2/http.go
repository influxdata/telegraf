package influxdb_v2

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

type APIErrorType int

type APIError struct {
	StatusCode  int
	Title       string
	Description string
	Type        APIErrorType
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Description)
	}
	return e.Title
}

const (
	defaultRequestTimeout = time.Second * 5
	defaultDatabase       = "telegraf"
	defaultUserAgent      = "telegraf"
)

type HTTPConfig struct {
	URL             *url.URL
	Token           string
	Organization    string
	Bucket          string
	Precision       string
	Timeout         time.Duration
	Headers         map[string]string
	Proxy           *url.URL
	UserAgent       string
	ContentEncoding string
	TLSConfig       *tls.Config

	Serializer *influx.Serializer
}

type httpClient struct {
	WriteURL        string
	ContentEncoding string
	Timeout         time.Duration
	Headers         map[string]string

	client     *http.Client
	serializer *influx.Serializer
	url        *url.URL
}

func NewHTTPClient(config *HTTPConfig) (*httpClient, error) {
	if config.URL == nil {
		return nil, ErrMissingURL
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	var headers = make(map[string]string, len(config.Headers)+2)
	headers["User-Agent"] = userAgent
	headers["Authorization"] = "Token " + config.Token
	for k, v := range config.Headers {
		headers[k] = v
	}

	var proxy func(*http.Request) (*url.URL, error)
	if config.Proxy != nil {
		proxy = http.ProxyURL(config.Proxy)
	} else {
		proxy = http.ProxyFromEnvironment
	}

	serializer := config.Serializer
	if serializer == nil {
		serializer = influx.NewSerializer()
	}

	writeURL, err := makeWriteURL(
		*config.URL,
		config.Organization,
		config.Bucket,
		config.Precision)
	if err != nil {
		return nil, err
	}

	var transport *http.Transport
	switch config.URL.Scheme {
	case "http", "https":
		transport = &http.Transport{
			Proxy:           proxy,
			TLSClientConfig: config.TLSConfig,
		}
	case "unix":
		transport = &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.DialTimeout(
					config.URL.Scheme,
					config.URL.Path,
					timeout,
				)
			},
		}
	default:
		return nil, fmt.Errorf("unsupported scheme %q", config.URL.Scheme)
	}

	client := &httpClient{
		serializer: serializer,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		url:             config.URL,
		WriteURL:        writeURL,
		ContentEncoding: config.ContentEncoding,
		Timeout:         timeout,
		Headers:         headers,
	}
	return client, nil
}

// URL returns the origin URL that this client connects too.
func (c *httpClient) URL() string {
	return c.url.String()
}

type genericRespError struct {
	Code      string
	Message   string
	Op        string
	Err       string
	Line      int32
	MaxLength int32
}

func (g genericRespError) String() string {
	return fmt.Sprintf("%s: %s", g.Code, g.Message)
}

func (c *httpClient) Write(ctx context.Context, metrics []telegraf.Metric) error {
	var err error

	reader := influx.NewReader(metrics, c.serializer)
	req, err := c.makeWriteRequest(reader)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	writeResp := &genericRespError{}
	json.NewDecoder(resp.Body).Decode(writeResp)
	var desc string

	switch resp.StatusCode {
	case http.StatusBadRequest: // 400
		// LineProtocolError
		desc = fmt.Sprintf("%s - %s;%d;%s", writeResp, writeResp.Op, writeResp.Line, writeResp.Err)
	case http.StatusUnauthorized, http.StatusForbidden: // 401, 403
		// Error
		desc = fmt.Sprintf("%s - %s;%s", writeResp, writeResp.Op, writeResp.Err)
	case http.StatusRequestEntityTooLarge: // 413
		// LineProtocolLengthError
		desc = fmt.Sprintf("%s - %s;%d", writeResp, writeResp.Op, writeResp.MaxLength)
	case http.StatusTooManyRequests, http.StatusServiceUnavailable: // 429, 503
		retryAfter := resp.Header.Get("Retry-After")
		retry, err := strconv.Atoi(retryAfter)
		if err != nil {
			return fmt.Errorf("Bad value for 'Retry-After': %s", err.Error())
		}
		time.Sleep(time.Second * time.Duration(retry))
		c.Write(ctx, metrics)
	}

	if xErr := resp.Header.Get("X-Influx-Error"); xErr != "" {
		desc = fmt.Sprintf("%s - %s", desc, xErr)
	}

	return &APIError{
		StatusCode:  resp.StatusCode,
		Title:       resp.Status,
		Description: desc,
	}
}

func (c *httpClient) makeWriteRequest(body io.Reader) (*http.Request, error) {
	var err error
	if c.ContentEncoding == "gzip" {
		body, err = compressWithGzip(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("POST", c.WriteURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	c.addHeaders(req)

	if c.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	return req, nil
}

func (c *httpClient) addHeaders(req *http.Request) {
	for header, value := range c.Headers {
		req.Header.Set(header, value)
	}
}

func compressWithGzip(data io.Reader) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)
	var err error

	go func() {
		_, err = io.Copy(gzipWriter, data)
		gzipWriter.Close()
		pipeWriter.Close()
	}()

	return pipeReader, err
}

func makeWriteURL(loc url.URL, org, bucket, precision string) (string, error) {
	params := url.Values{}
	params.Set("bucket", bucket)
	params.Set("org", org)
	if precision != "" {
		params.Set("precision", precision)
	}

	switch loc.Scheme {
	case "unix":
		loc.Scheme = "http"
		loc.Host = "127.0.0.1"
		loc.Path = "v2/write"
	case "http", "https":
		loc.Path = path.Join(loc.Path, "v2/write")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	loc.RawQuery = params.Encode()
	return loc.String(), nil
}
