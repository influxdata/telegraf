package influxdb_v2

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/limited"
)

type APIError struct {
	StatusCode  int
	Title       string
	Description string
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Description)
	}
	return e.Title
}

const (
	defaultMaxWaitSeconds           = 60
	defaultMaxWaitRetryAfterSeconds = 10 * 60
)

type httpClient struct {
	URL              *url.URL
	LocalAddr        *net.TCPAddr
	Token            config.Secret
	Organization     string
	Bucket           string
	BucketTag        string
	ExcludeBucketTag bool
	Timeout          time.Duration
	Headers          map[string]string
	Proxy            *url.URL
	UserAgent        string
	ContentEncoding  string
	PingTimeout      config.Duration
	ReadIdleTimeout  config.Duration
	TLSConfig        *tls.Config
	Serializer       limited.Serializer
	Encoder          internal.ContentEncoder
	Log              telegraf.Logger

	client     *http.Client
	params     url.Values
	retryTime  time.Time
	retryCount int
}

func (c *httpClient) Init() error {
	token, err := c.Token.Get()
	if err != nil {
		return fmt.Errorf("getting token failed: %w", err)
	}
	if c.Headers == nil {
		c.Headers = make(map[string]string, 2)
	}
	c.Headers["Authorization"] = "Token " + token.String()
	token.Destroy()
	c.Headers["User-Agent"] = c.UserAgent

	var proxy func(*http.Request) (*url.URL, error)
	if c.Proxy != nil {
		proxy = http.ProxyURL(c.Proxy)
	} else {
		proxy = http.ProxyFromEnvironment
	}

	var transport *http.Transport
	switch c.URL.Scheme {
	case "http", "https":
		var dialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)
		if c.LocalAddr != nil {
			dialer := &net.Dialer{LocalAddr: c.LocalAddr}
			dialerFunc = dialer.DialContext
		}
		transport = &http.Transport{
			Proxy:           proxy,
			TLSClientConfig: c.TLSConfig,
			DialContext:     dialerFunc,
		}
		if c.ReadIdleTimeout != 0 || c.PingTimeout != 0 {
			http2Trans, err := http2.ConfigureTransports(transport)
			if err == nil {
				http2Trans.ReadIdleTimeout = time.Duration(c.ReadIdleTimeout)
				http2Trans.PingTimeout = time.Duration(c.PingTimeout)
			}
		}
	case "unix":
		transport = &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.DialTimeout(
					c.URL.Scheme,
					c.URL.Path,
					c.Timeout,
				)
			},
		}
	default:
		return fmt.Errorf("unsupported scheme %q", c.URL.Scheme)
	}

	preppedURL, params, err := prepareWriteURL(*c.URL, c.Organization)
	if err != nil {
		return err
	}

	c.URL = preppedURL
	c.client = &http.Client{
		Timeout:   c.Timeout,
		Transport: transport,
	}
	c.params = params

	return nil
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
		return errors.New("retry time has not elapsed")
	}

	batches := make(map[string][]telegraf.Metric)
	if c.BucketTag == "" {
		err := c.writeBatch(ctx, c.Bucket, metrics)
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) {
				if apiErr.StatusCode == http.StatusRequestEntityTooLarge {
					return c.splitAndWriteBatch(ctx, c.Bucket, metrics)
				}
			}

			return err
		}
	} else {
		for _, metric := range metrics {
			bucket, ok := metric.GetTag(c.BucketTag)
			if !ok {
				bucket = c.Bucket
			}

			if _, ok := batches[bucket]; !ok {
				batches[bucket] = make([]telegraf.Metric, 0)
			}

			if c.ExcludeBucketTag {
				// Avoid modifying the metric in case we need to retry the request.
				metric = metric.Copy()
				metric.Accept()
				metric.RemoveTag(c.BucketTag)
			}

			batches[bucket] = append(batches[bucket], metric)
		}

		for bucket, batch := range batches {
			err := c.writeBatch(ctx, bucket, batch)
			if err != nil {
				var apiErr *APIError
				if errors.As(err, &apiErr) {
					if apiErr.StatusCode == http.StatusRequestEntityTooLarge {
						return c.splitAndWriteBatch(ctx, c.Bucket, metrics)
					}
				}

				return err
			}
		}
	}
	return nil
}

func (c *httpClient) splitAndWriteBatch(ctx context.Context, bucket string, metrics []telegraf.Metric) error {
	c.Log.Warnf("Retrying write after splitting metric payload in half to reduce batch size")
	midpoint := len(metrics) / 2

	if err := c.writeBatch(ctx, bucket, metrics[:midpoint]); err != nil {
		return err
	}

	return c.writeBatch(ctx, bucket, metrics[midpoint:])
}

func (c *httpClient) writeBatch(ctx context.Context, bucket string, metrics []telegraf.Metric) error {
	// Serialize the metrics with the remaining limit, exit early if nothing was serialized
	body, werr := c.Serializer.SerializeBatch(metrics, 0)
	if werr != nil && !errors.Is(werr, internal.ErrSizeLimitReached) || len(body) == 0 {
		return werr
	}

	// Encode the content if requested
	if c.Encoder != nil {
		var err error
		if body, err = c.Encoder.Encode(body); err != nil {
			return fmt.Errorf("encoding failed: %w", err)
		}
	}

	// Setup the request
	address := makeWriteURL(*c.URL, c.params, bucket)
	req, err := http.NewRequest("POST", address, io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}
	if c.Encoder != nil {
		req.Header.Set("Content-Encoding", c.ContentEncoding)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	c.addHeaders(req)

	// Execute the request
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		internal.OnClientError(c.client, err)
		return err
	}
	defer resp.Body.Close()

	// Check for success
	switch resp.StatusCode {
	case
		// this is the expected response:
		http.StatusNoContent,
		// but if we get these we should still accept it as delivered:
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusPartialContent,
		http.StatusMultiStatus,
		http.StatusAlreadyReported:
		c.retryCount = 0
		return nil
	}

	// We got an error and now try to decode further
	writeResp := &genericRespError{}
	err = json.NewDecoder(resp.Body).Decode(writeResp)
	desc := writeResp.Error()
	if err != nil {
		desc = resp.Status
	}

	switch resp.StatusCode {
	// request was too large, send back to try again
	case http.StatusRequestEntityTooLarge:
		c.Log.Errorf("Failed to write metric to %s, request was too large (413)", bucket)
		return &APIError{
			StatusCode:  resp.StatusCode,
			Title:       resp.Status,
			Description: desc,
		}
	case
		// request was malformed:
		http.StatusBadRequest,
		// request was received but server refused to process it due to a semantic problem with the request.
		// for example, submitting metrics outside the retention period.
		// Clients should *not* repeat the request and the metrics should be dropped.
		http.StatusUnprocessableEntity,
		http.StatusNotAcceptable:
		c.Log.Errorf("Failed to write metric to %s (will be dropped: %s): %s\n", bucket, resp.Status, desc)
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("failed to write metric to %s (%s): %s", bucket, resp.Status, desc)
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusBadGateway,
		http.StatusGatewayTimeout:
		// ^ these handle the cases where the server is likely overloaded, and may not be able to say so.
		c.retryCount++
		retryDuration := c.getRetryDuration(resp.Header)
		c.retryTime = time.Now().Add(retryDuration)
		c.Log.Warnf("Failed to write to %s; will retry in %s. (%s)\n", bucket, retryDuration, resp.Status)
		return fmt.Errorf("waiting %s for server (%s) before sending metric again", retryDuration, bucket)
	}

	// if it's any other 4xx code, the client should not retry as it's the client's mistake.
	// retrying will not make the request magically work.
	if len(resp.Status) > 0 && resp.Status[0] == '4' {
		c.Log.Errorf("Failed to write metric to %s (will be dropped: %s): %s\n", bucket, resp.Status, desc)
		return nil
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

// retryDuration takes the longer of the Retry-After header and our own back-off calculation
func (c *httpClient) getRetryDuration(headers http.Header) time.Duration {
	// basic exponential backoff (x^2)/40 (denominator to widen the slope)
	// at 40 denominator, it'll take 49 retries to hit the max defaultMaxWait of 60s
	backoff := math.Pow(float64(c.retryCount), 2) / 40
	backoff = math.Min(backoff, defaultMaxWaitSeconds)

	// get any value from the header, if available
	retryAfterHeader := float64(0)
	retryAfterHeaderString := headers.Get("Retry-After")
	if len(retryAfterHeaderString) > 0 {
		var err error
		retryAfterHeader, err = strconv.ParseFloat(retryAfterHeaderString, 64)
		if err != nil {
			// there was a value but we couldn't parse it? guess minimum 10 sec
			retryAfterHeader = 10
		}
		// protect against excessively large retry-after
		retryAfterHeader = math.Min(retryAfterHeader, defaultMaxWaitRetryAfterSeconds)
	}
	// take the highest value of backoff and retry-after.
	retry := math.Max(backoff, retryAfterHeader)
	return time.Duration(retry*1000) * time.Millisecond
}

func (c *httpClient) addHeaders(req *http.Request) {
	for header, value := range c.Headers {
		if strings.EqualFold(header, "host") {
			req.Host = value
		} else {
			req.Header.Set(header, value)
		}
	}
}

func makeWriteURL(loc url.URL, params url.Values, bucket string) string {
	params.Set("bucket", bucket)
	loc.RawQuery = params.Encode()
	return loc.String()
}

func prepareWriteURL(loc url.URL, org string) (*url.URL, url.Values, error) {
	switch loc.Scheme {
	case "unix":
		loc.Scheme = "http"
		loc.Host = "127.0.0.1"
		loc.Path = "/api/v2/write"
	case "http", "https":
		loc.Path = path.Join(loc.Path, "/api/v2/write")
	default:
		return nil, nil, fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}

	params := loc.Query()
	params.Set("org", org)

	return &loc, params, nil
}

func (c *httpClient) Close() {
	c.client.CloseIdleConnections()
}
