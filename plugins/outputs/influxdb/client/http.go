package client

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	defaultRequestTimeout = time.Second * 5
)

//
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

	wu := writeURL(u, defaultWP)
	return &httpClient{
		writeURL: []byte(wu),
		config:   config,
		url:      u,
		client: &fasthttp.Client{
			TLSConfig: config.TLSConfig,
		},
	}, nil
}

type HTTPConfig struct {
	// URL should be of the form "http://host:port" (REQUIRED)
	URL string

	// UserAgent sets the User-Agent header.
	UserAgent string

	// Timeout is the time to wait for a response to each HTTP request (writes
	// and queries).
	Timeout time.Duration

	// Username is the basic auth username for the server.
	Username string
	// Password is the basic auth password for the server.
	Password string

	// TLSConfig is the tls auth settings to use for each request.
	TLSConfig *tls.Config

	// Gzip, if true, compresses each payload using gzip.
	// TODO
	// Gzip bool
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
	writeURL []byte
	config   HTTPConfig
	client   *fasthttp.Client
	url      *url.URL
}

func (c *httpClient) Query(command string) error {
	req := c.makeRequest()
	req.Header.SetRequestURI(queryURL(c.url, command))

	return c.doRequest(req, fasthttp.StatusOK)
}

func (c *httpClient) Write(b []byte) (int, error) {
	req := c.makeWriteRequest(len(b), c.writeURL)
	req.SetBody(b)

	err := c.doRequest(req, fasthttp.StatusNoContent)
	if err == nil {
		return len(b), nil
	}
	return 0, err
}

func (c *httpClient) WriteWithParams(b []byte, wp WriteParams) (int, error) {
	req := c.makeWriteRequest(len(b), []byte(writeURL(c.url, wp)))
	req.SetBody(b)

	err := c.doRequest(req, fasthttp.StatusNoContent)
	if err == nil {
		return len(b), nil
	}
	return 0, err
}

func (c *httpClient) WriteStream(r io.Reader, contentLength int) (int, error) {
	req := c.makeWriteRequest(contentLength, c.writeURL)
	req.SetBodyStream(r, contentLength)

	err := c.doRequest(req, fasthttp.StatusNoContent)
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
	req := c.makeWriteRequest(contentLength, []byte(writeURL(c.url, wp)))
	req.SetBodyStream(r, contentLength)

	err := c.doRequest(req, fasthttp.StatusNoContent)
	if err == nil {
		return contentLength, nil
	}
	return 0, err
}

func (c *httpClient) doRequest(
	req *fasthttp.Request,
	expectedCode int,
) error {
	resp := fasthttp.AcquireResponse()

	err := c.client.DoTimeout(req, resp, c.config.Timeout)

	code := resp.StatusCode()
	// If it's a "no content" response, then release and return nil
	if code == fasthttp.StatusNoContent {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
		return nil
	}

	// not a "no content" response, so parse the result:
	var response Response
	decErr := json.Unmarshal(resp.Body(), &response)

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

	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)

	return err
}

func (c *httpClient) makeWriteRequest(
	contentLength int,
	writeURL []byte,
) *fasthttp.Request {
	req := c.makeRequest()
	req.Header.SetContentLength(contentLength)
	req.Header.SetRequestURIBytes(writeURL)
	// TODO
	// if gzip {
	// 	req.Header.SetBytesKV([]byte("Content-Encoding"), []byte("gzip"))
	// }
	return req
}

func (c *httpClient) makeRequest() *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.Header.SetContentTypeBytes([]byte("text/plain"))
	req.Header.SetMethodBytes([]byte("POST"))
	req.Header.SetUserAgent(c.config.UserAgent)
	if c.config.Username != "" && c.config.Password != "" {
		req.Header.Set("Authorization", "Basic "+basicAuth(c.config.Username, c.config.Password))
	}
	return req
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

// See 2 (end of page 4) http://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the httpClient sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
