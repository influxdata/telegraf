package influxdb

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

type APIErrorType int

const (
	_ APIErrorType = iota
	DatabaseNotFound
)

const (
	defaultRequestTimeout = time.Second * 5
	defaultDatabase       = "telegraf"
	defaultUserAgent      = "telegraf"

	errStringDatabaseNotFound      = "database not found"
	errStringHintedHandoffNotEmpty = "hinted handoff queue not empty"
	errStringPartialWrite          = "partial write"
	errStringPointsBeyondRP        = "points beyond retention policy"
	errStringUnableToParse         = "unable to parse"
)

var (

	// Escape an identifier in InfluxQL.
	escapeIdentifier = strings.NewReplacer(
		"\n", `\n`,
		`\`, `\\`,
		`"`, `\"`,
	)
)

// APIError is an error reported by the InfluxDB server
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

// QueryResponse is the response body from the /query endpoint
type QueryResponse struct {
	Results []QueryResult `json:"results"`
}

type QueryResult struct {
	Err string `json:"error,omitempty"`
}

func (r QueryResponse) Error() string {
	if len(r.Results) > 0 {
		return r.Results[0].Err
	}
	return ""
}

// WriteResponse is the response body from the /write endpoint
type WriteResponse struct {
	Err string `json:"error,omitempty"`
}

func (r WriteResponse) Error() string {
	return r.Err
}

type HTTPConfig struct {
	URL             *url.URL
	UserAgent       string
	Timeout         time.Duration
	Username        string
	Password        string
	TLSConfig       *tls.Config
	Proxy           *url.URL
	Headers         map[string]string
	ContentEncoding string
	Database        string
	RetentionPolicy string
	Consistency     string

	InfluxUintSupport bool `toml:"influx_uint_support"`
	Serializer        *influx.Serializer
}

type httpClient struct {
	WriteURL        string
	QueryURL        string
	ContentEncoding string
	Timeout         time.Duration
	Username        string
	Password        string
	Headers         map[string]string

	client     *http.Client
	serializer *influx.Serializer
	url        *url.URL
	database   string
}

func NewHTTPClient(config *HTTPConfig) (*httpClient, error) {
	if config.URL == nil {
		return nil, ErrMissingURL
	}

	database := config.Database
	if database == "" {
		database = defaultDatabase
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	var headers = make(map[string]string, len(config.Headers)+1)
	headers["User-Agent"] = userAgent
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

	writeURL := makeWriteURL(
		config.URL,
		database,
		config.RetentionPolicy,
		config.Consistency)
	queryURL := makeQueryURL(config.URL)

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
					defaultRequestTimeout,
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
		database:        database,
		url:             config.URL,
		WriteURL:        writeURL,
		QueryURL:        queryURL,
		ContentEncoding: config.ContentEncoding,
		Timeout:         timeout,
		Username:        config.Username,
		Password:        config.Password,
		Headers:         headers,
	}
	return client, nil
}

// URL returns the origin URL that this client connects too.
func (c *httpClient) URL() string {
	return c.url.String()
}

// URL returns the database that this client connects too.
func (c *httpClient) Database() string {
	return c.database
}

// CreateDatabase attemps to create a new database in the InfluxDB server.
// Note that some names are not allowed by the server, notably those with
// non-printable characters or slashes.
func (c *httpClient) CreateDatabase(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE DATABASE "%s"`,
		escapeIdentifier.Replace(c.database))

	req, err := c.makeQueryRequest(query)

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	queryResp := &QueryResponse{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(queryResp)

	if err != nil {
		if resp.StatusCode == 200 {
			return nil
		}

		return &APIError{
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}

	// Even with a 200 response there can be an error
	if resp.StatusCode == http.StatusOK && queryResp.Error() == "" {
		return nil
	}

	return &APIError{
		StatusCode:  resp.StatusCode,
		Title:       resp.Status,
		Description: queryResp.Error(),
	}
}

// Write sends the metrics to InfluxDB
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

	writeResp := &WriteResponse{}
	dec := json.NewDecoder(resp.Body)

	var desc string
	err = dec.Decode(writeResp)
	if err == nil {
		desc = writeResp.Err
	}

	if strings.Contains(desc, errStringDatabaseNotFound) {
		return &APIError{
			StatusCode:  resp.StatusCode,
			Title:       resp.Status,
			Description: desc,
			Type:        DatabaseNotFound,
		}
	}

	// This "error" is an informational message about the state of the
	// InfluxDB cluster.
	if strings.Contains(desc, errStringHintedHandoffNotEmpty) {
		return nil
	}

	// Points beyond retention policy is returned when points are immediately
	// discarded for being older than the retention policy.  Usually this not
	// a cause for concern and we don't want to retry.
	if strings.Contains(desc, errStringPointsBeyondRP) {
		log.Printf("W! [outputs.influxdb]: when writing to [%s]: received error %v",
			c.URL(), desc)
		return nil
	}

	// Other partial write errors, such as "field type conflict", are not
	// correctable at this point and so the point is dropped instead of
	// retrying.
	if strings.Contains(desc, errStringPartialWrite) {
		log.Printf("E! [outputs.influxdb]: when writing to [%s]: received error %v; discarding points",
			c.URL(), desc)
		return nil
	}

	// This error indicates a bug in either Telegraf line protocol
	// serialization, retries would not be successful.
	if strings.Contains(desc, errStringUnableToParse) {
		log.Printf("E! [outputs.influxdb]: when writing to [%s]: received error %v; discarding points",
			c.URL(), desc)
		return nil
	}

	return &APIError{
		StatusCode:  resp.StatusCode,
		Title:       resp.Status,
		Description: desc,
	}
}

func (c *httpClient) makeQueryRequest(query string) (*http.Request, error) {
	params := url.Values{}
	params.Set("q", query)
	form := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", c.QueryURL, form)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.addHeaders(req)

	return req, nil
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

func (c *httpClient) addHeaders(req *http.Request) {
	if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	for header, value := range c.Headers {
		req.Header.Set(header, value)
	}
}

func makeWriteURL(loc *url.URL, db, rp, consistency string) string {
	params := url.Values{}
	params.Set("db", db)

	if rp != "" {
		params.Set("rp", rp)
	}

	if consistency != "one" && consistency != "" {
		params.Set("consistency", consistency)
	}

	u := *loc
	switch u.Scheme {
	case "unix":
		u.Scheme = "http"
		u.Host = "127.0.0.1"
		u.Path = "/write"
	case "http":
	case "https":
		u.Path = path.Join(u.Path, "write")
	}
	u.RawQuery = params.Encode()
	return u.String()
}

func makeQueryURL(loc *url.URL) string {
	u := *loc
	switch u.Scheme {
	case "unix":
		u.Scheme = "http"
		u.Host = "127.0.0.1"
		u.Path = "/query"
	case "http":
	case "https":
		u.Path = path.Join(u.Path, "query")
	}
	return u.String()
}
