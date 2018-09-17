package influxdb_v2

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	defaultMaxWait        = 10 // seconds
	defaultDatabase       = "telegraf"
	defaultUserAgent      = "telegraf"
)

type HTTPConfig struct {
	URL             *url.URL
	Token           string
	Organization    string
	Bucket          string
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
	retryTime  time.Time
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
		config.Bucket)
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
	Line      *int32
	MaxLength *int32
}

func (g genericRespError) Error() string {
	errString := fmt.Sprintf("%s: %s", g.Code, g.Message)
	if g.Line != nil {
		return fmt.Sprintf("%s - line[%d]", errString, g.Line)
	} else if g.MaxLength != nil {
		return fmt.Sprintf("%s - maxlen[%d]", errString, g.MaxLength)
	}
	return errString
}

func (c *httpClient) Write(ctx context.Context, metrics []telegraf.Metric) error {
	if c.retryTime.After(time.Now()) {
		return errors.New("Retry time has not elapsed")
	}
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
	err = json.NewDecoder(resp.Body).Decode(writeResp)
	desc := writeResp.Error()
	if err != nil {
		desc = err.Error()
	}

	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusUnauthorized,
		http.StatusForbidden, http.StatusRequestEntityTooLarge:
		log.Printf("E! [outputs.influxdb_v2] Failed to write metric: %s\n", desc)
		return nil
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		retryAfter := resp.Header.Get("Retry-After")
		retry, err := strconv.Atoi(retryAfter)
		if err != nil {
			retry = 0
		}
		if retry > defaultMaxWait {
			retry = defaultMaxWait
		}
		c.retryTime = time.Now().Add(time.Duration(retry) * time.Second)
		return fmt.Errorf("Waiting %ds for server before sending metric again", retry)
	}

	// This is only until platform spec is fully implemented. As of the
	// time of writing, there is no error body returned.
	if xErr := resp.Header.Get("X-Influx-Error"); xErr != "" {
		desc = fmt.Sprintf("%s; %s", desc, xErr)
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

func makeWriteURL(loc url.URL, org, bucket string) (string, error) {
	params := url.Values{}
	params.Set("bucket", bucket)
	params.Set("org", org)

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
