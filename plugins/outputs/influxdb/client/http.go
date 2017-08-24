package client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	defaultRequestTimeout = time.Second * 5
)

func NewHTTP(config HTTPConfig, defaultWP WriteParams) (Client, error) {
	// validate required parameters:
	if len(config.URL) == 0 {
		return nil, fmt.Errorf("config.URL is required to create an HTTP client")
	}
	if len(defaultWP.Database) == 0 {
		return nil, fmt.Errorf("A default database is required to create an HTTP client")
	}

	// set defaults:
	if config.Timeout == 0 {
		config.Timeout = defaultRequestTimeout
	}

	// parse URL:
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing config.URL: %s", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("config.URL scheme must be http(s), got %s", u.Scheme)
	}

	var transport http.Transport
	if len(config.HTTPProxy) > 0 {
		proxyURL, err := url.Parse(config.HTTPProxy)
		if err != nil {
			return nil, fmt.Errorf("error parsing config.HTTPProxy: %s", err)
		}

		transport = http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: config.TLSConfig,
		}
	} else {
		transport = http.Transport{
			TLSClientConfig: config.TLSConfig,
		}
	}

	return &httpClient{
		writeURL: writeURL(u, defaultWP),
		config:   config,
		url:      u,
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: &transport,
		},
	}, nil
}

type HTTPConfig struct {
	// URL should be of the form "http://host:port" (REQUIRED)
	URL string

	// UserAgent sets the User-Agent header.
	UserAgent string

	// Timeout specifies a time limit for requests made by this
	// Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	//
	// A Timeout of zero means no timeout.
	Timeout time.Duration

	// Username is the basic auth username for the server.
	Username string
	// Password is the basic auth password for the server.
	Password string

	// TLSConfig is the tls auth settings to use for each request.
	TLSConfig *tls.Config

	// Proxy URL should be of the form "http://host:port"
	HTTPProxy string

	// The content encoding mechanism to use for each request.
	ContentEncoding string
}

// Response represents a list of statement results.
type Response struct {
	// ignore Results:
	Results []interface{} `json:"-"`
	Err     string        `json:"error,omitempty"`
}

// Error returns the first error from any statement.
// Returns nil if no errors occurred on any statements.
func (r *Response) Error() error {
	if r.Err != "" {
		return fmt.Errorf(r.Err)
	}
	return nil
}

type httpClient struct {
	writeURL string
	config   HTTPConfig
	client   *http.Client
	url      *url.URL
}

func (c *httpClient) Query(command string) error {
	req, err := c.makeRequest(queryURL(c.url, command), bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}
	return c.doRequest(req, http.StatusOK)
}

func (c *httpClient) Write(b []byte) (int, error) {
	req, err := c.makeWriteRequest(bytes.NewReader(b), len(b), c.writeURL)
	if err != nil {
		return 0, nil
	}

	err = c.doRequest(req, http.StatusNoContent)
	if err == nil {
		return len(b), nil
	}
	return 0, err
}

func (c *httpClient) WriteWithParams(b []byte, wp WriteParams) (int, error) {
	req, err := c.makeWriteRequest(bytes.NewReader(b), len(b), writeURL(c.url, wp))
	if err != nil {
		return 0, nil
	}

	err = c.doRequest(req, http.StatusNoContent)
	if err == nil {
		return len(b), nil
	}
	return 0, err
}

func (c *httpClient) WriteStream(r io.Reader, contentLength int) (int, error) {
	req, err := c.makeWriteRequest(r, contentLength, c.writeURL)
	if err != nil {
		return 0, nil
	}

	err = c.doRequest(req, http.StatusNoContent)
	if err == nil {
		return contentLength, nil
	}
	return 0, err
}

func (c *httpClient) WriteStreamWithParams(
	r io.Reader,
	contentLength int,
	wp WriteParams,
) (int, error) {
	req, err := c.makeWriteRequest(r, contentLength, writeURL(c.url, wp))
	if err != nil {
		return 0, nil
	}

	err = c.doRequest(req, http.StatusNoContent)
	if err == nil {
		return contentLength, nil
	}
	return 0, err
}

func (c *httpClient) doRequest(
	req *http.Request,
	expectedCode int,
) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	code := resp.StatusCode
	// If it's a "no content" response, then release and return nil
	if code == http.StatusNoContent {
		return nil
	}

	// not a "no content" response, so parse the result:
	var response Response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Fatal error reading body: %s", err)
	}
	decErr := json.Unmarshal(body, &response)

	// If we got a JSON decode error, send that back
	if decErr != nil {
		err = fmt.Errorf("Unable to decode json: received status code %d err: %s", code, decErr)
	}
	// Unexpected response code OR error in JSON response body overrides
	// a JSON decode error:
	if code != expectedCode || response.Error() != nil {
		err = fmt.Errorf("Response Error: Status Code [%d], expected [%d], [%v]",
			code, expectedCode, response.Error())
	}

	return err
}

func (c *httpClient) makeWriteRequest(
	body io.Reader,
	contentLength int,
	writeURL string,
) (*http.Request, error) {
	req, err := c.makeRequest(writeURL, body)
	if err != nil {
		return nil, err
	}
	if c.config.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req.Header.Set("Content-Length", fmt.Sprint(contentLength))
	}
	return req, nil
}

func (c *httpClient) makeRequest(uri string, body io.Reader) (*http.Request, error) {
	var req *http.Request
	var err error
	if c.config.ContentEncoding == "gzip" {
		body, err = compressWithGzip(body)
		if err != nil {
			return nil, err
		}
	}
	req, err = http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
	return req, nil
}

func compressWithGzip(data io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	gw := gzip.NewWriter(pw)
	var err error

	go func() {
		_, err = io.Copy(gw, data)
		gw.Close()
		pw.Close()
	}()

	return pr, err
}

func (c *httpClient) Close() error {
	// Nothing to do.
	return nil
}

func writeURL(u *url.URL, wp WriteParams) string {
	params := url.Values{}
	params.Set("db", wp.Database)
	if wp.RetentionPolicy != "" {
		params.Set("rp", wp.RetentionPolicy)
	}
	if wp.Precision != "n" && wp.Precision != "" {
		params.Set("precision", wp.Precision)
	}
	if wp.Consistency != "one" && wp.Consistency != "" {
		params.Set("consistency", wp.Consistency)
	}

	u.RawQuery = params.Encode()
	u.Path = "write"
	return u.String()
}

func queryURL(u *url.URL, command string) string {
	params := url.Values{}
	params.Set("q", command)

	u.RawQuery = params.Encode()
	u.Path = "query"
	return u.String()
}
