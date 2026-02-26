package influxdb

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	defaultRequestTimeout            = time.Second * 5
	defaultDatabase                  = "telegraf"
	errStringDatabaseNotFound        = "database not found"
	errStringRetentionPolicyNotFound = "retention policy not found"
	errStringHintedHandoffNotEmpty   = "hinted handoff queue not empty"
	errStringPartialWrite            = "partial write"
	errStringPointsBeyondRP          = "points beyond retention policy"
	errStringUnableToParse           = "unable to parse"
)

var (
	// Escape an identifier in InfluxQL.
	escapeIdentifier = strings.NewReplacer(
		"\n", `\n`,
		`\`, `\\`,
		`"`, `\"`,
	)
)

// APIError is a general error reported by the InfluxDB server
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

type DatabaseNotFoundError struct {
	APIError
	Database string
}

// queryResponseError is the response body from the /query endpoint
type queryResponseError struct {
	Results []queryResult `json:"results"`
}

type queryResult struct {
	Err string `json:"error,omitempty"`
}

func (r queryResponseError) Error() string {
	if len(r.Results) > 0 {
		return r.Results[0].Err
	}
	return ""
}

// writeResponseError is the response body from the /write endpoint
type writeResponseError struct {
	Err string `json:"error,omitempty"`
}

func (r writeResponseError) Error() string {
	return r.Err
}

type HTTPConfig struct {
	URL                       *url.URL
	LocalAddr                 *net.TCPAddr
	UserAgent                 string
	Timeout                   time.Duration
	Username                  config.Secret
	Password                  config.Secret
	TLSConfig                 *tls.Config
	Proxy                     *url.URL
	Headers                   map[string]string
	ContentEncoding           string
	Database                  string
	DatabaseTag               string
	ExcludeDatabaseTag        bool
	RetentionPolicy           string
	RetentionPolicyTag        string
	ExcludeRetentionPolicyTag bool
	Consistency               string
	SkipDatabaseCreation      bool

	InfluxUintSupport bool `toml:"influx_uint_support"`
	Serializer        *influx.Serializer
	Log               telegraf.Logger

	BytesWritten selfstat.Stat
}

type httpClient struct {
	client *http.Client
	config HTTPConfig
	// Tracks that the 'create database` statement was executed for the
	// database.  An attempt to create the database is made each time a new
	// database is encountered in the database_tag and after a "database not
	// found" error occurs.
	createDatabaseExecuted map[string]bool

	log telegraf.Logger
}

func NewHTTPClient(cfg HTTPConfig) (*httpClient, error) {
	if cfg.URL == nil {
		return nil, ErrMissingURL
	}

	if cfg.Database == "" {
		cfg.Database = defaultDatabase
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = defaultRequestTimeout
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = internal.ProductToken()
	}

	if cfg.Headers == nil {
		cfg.Headers = make(map[string]string)
	}
	cfg.Headers["User-Agent"] = userAgent
	for k, v := range cfg.Headers {
		cfg.Headers[k] = v
	}

	var proxy func(*http.Request) (*url.URL, error)
	if cfg.Proxy != nil {
		proxy = http.ProxyURL(cfg.Proxy)
	} else {
		proxy = http.ProxyFromEnvironment
	}

	if cfg.Serializer == nil {
		cfg.Serializer = &influx.Serializer{}
		if err := cfg.Serializer.Init(); err != nil {
			return nil, err
		}
	}

	var transport *http.Transport
	switch cfg.URL.Scheme {
	case "http", "https":
		var dialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)
		if cfg.LocalAddr != nil {
			dialer := &net.Dialer{LocalAddr: cfg.LocalAddr}
			dialerFunc = dialer.DialContext
		}
		transport = &http.Transport{
			Proxy:           proxy,
			TLSClientConfig: cfg.TLSConfig,
			DialContext:     dialerFunc,
		}
	case "unix":
		transport = &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.DialTimeout(
					cfg.URL.Scheme,
					cfg.URL.Path,
					defaultRequestTimeout,
				)
			},
		}
	default:
		return nil, fmt.Errorf("unsupported scheme %q", cfg.URL.Scheme)
	}

	client := &httpClient{
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		createDatabaseExecuted: make(map[string]bool),
		config:                 cfg,
		log:                    cfg.Log,
	}
	return client, nil
}

// URL returns the origin URL that this client connects too.
func (c *httpClient) URL() string {
	return c.config.URL.String()
}

// Database returns the default database that this client connects too.
func (c *httpClient) Database() string {
	return c.config.Database
}

// CreateDatabase attempts to create a new database in the InfluxDB server.
// Note that some names are not allowed by the server, notably those with
// non-printable characters or slashes.
func (c *httpClient) CreateDatabase(ctx context.Context, database string) error {
	//nolint:gocritic // sprintfQuotedString - "%s" used by purpose, string escaping is done by special function
	query := fmt.Sprintf(`CREATE DATABASE "%s"`, escapeIdentifier.Replace(database))

	req, err := c.makeQueryRequest(query)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		internal.OnClientError(c.client, err)
		return err
	}
	defer resp.Body.Close()

	body, err := validateResponse(resp.Body)

	// Check for poorly formatted response (can't be decoded)
	if err != nil {
		return &APIError{
			StatusCode:  resp.StatusCode,
			Title:       resp.Status,
			Description: "An error response was received while attempting to create the following database: " + database + ". Error: " + err.Error(),
		}
	}

	queryResp := &queryResponseError{}
	dec := json.NewDecoder(body)
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

	// Even with a 200 status code there can be an error in the response body.
	// If there is also no error string then the operation was successful.
	if resp.StatusCode == http.StatusOK && queryResp.Error() == "" {
		c.createDatabaseExecuted[database] = true
		return nil
	}

	// Don't attempt to recreate the database after a 403 Forbidden error.
	// This behavior exists only to maintain backwards compatibility.
	if resp.StatusCode == http.StatusForbidden {
		c.createDatabaseExecuted[database] = true
	}

	return &APIError{
		StatusCode:  resp.StatusCode,
		Title:       resp.Status,
		Description: queryResp.Error(),
	}
}

type dbrp struct {
	Database        string
	RetentionPolicy string
}

// Write sends the metrics to InfluxDB
func (c *httpClient) Write(ctx context.Context, metrics []telegraf.Metric) error {
	// If these options are not used, we can skip in plugin batching and send
	// the full batch in a single request.
	if c.config.DatabaseTag == "" && c.config.RetentionPolicyTag == "" {
		return c.writeBatch(ctx, c.config.Database, c.config.RetentionPolicy, metrics)
	}

	batches := make(map[dbrp][]telegraf.Metric)
	for _, metric := range metrics {
		db, ok := metric.GetTag(c.config.DatabaseTag)
		if !ok {
			db = c.config.Database
		}

		rp, ok := metric.GetTag(c.config.RetentionPolicyTag)
		if !ok {
			rp = c.config.RetentionPolicy
		}

		dbrp := dbrp{
			Database:        db,
			RetentionPolicy: rp,
		}

		if c.config.ExcludeDatabaseTag || c.config.ExcludeRetentionPolicyTag {
			// Avoid modifying the metric in case we need to retry the request.
			metric = metric.Copy()
			metric.Accept()
			if c.config.ExcludeDatabaseTag {
				metric.RemoveTag(c.config.DatabaseTag)
			}
			if c.config.ExcludeRetentionPolicyTag {
				metric.RemoveTag(c.config.RetentionPolicyTag)
			}
		}

		batches[dbrp] = append(batches[dbrp], metric)
	}

	for dbrp, batch := range batches {
		if !c.config.SkipDatabaseCreation && !c.createDatabaseExecuted[dbrp.Database] {
			err := c.CreateDatabase(ctx, dbrp.Database)
			if err != nil {
				c.log.Warnf("When writing to [%s]: database %q creation failed: %v",
					c.config.URL, dbrp.Database, err)
			}
		}

		err := c.writeBatch(ctx, dbrp.Database, dbrp.RetentionPolicy, batch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *httpClient) writeBatch(ctx context.Context, db, rp string, metrics []telegraf.Metric) error {
	loc, err := makeWriteURL(c.config.URL, db, rp, c.config.Consistency)
	if err != nil {
		return fmt.Errorf("failed making write url: %w", err)
	}

	reader := c.requestBodyReader(metrics)
	defer reader.Close()
	defer func() { c.config.BytesWritten.Incr(reader.bytesWritten.Load()) }()

	req, err := c.makeWriteRequest(loc, reader)
	if err != nil {
		return fmt.Errorf("failed making write req: %w", err)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		internal.OnClientError(c.client, err)
		return fmt.Errorf("failed doing req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, err := validateResponse(resp.Body)

	// Check for poorly formatted response that can't be decoded
	if err != nil {
		return &APIError{
			StatusCode:  resp.StatusCode,
			Title:       resp.Status,
			Description: "An error response was received while attempting to write metrics. Error: " + err.Error(),
		}
	}

	writeResp := &writeResponseError{}
	dec := json.NewDecoder(body)

	var desc string
	err = dec.Decode(writeResp)
	if err == nil {
		desc = writeResp.Err
	}
	if strings.Contains(desc, errStringDatabaseNotFound) {
		return &DatabaseNotFoundError{
			APIError: APIError{
				StatusCode:  resp.StatusCode,
				Title:       resp.Status,
				Description: desc,
			},
			Database: db,
		}
	}

	// checks for any 4xx code and drops metric and retrying will not make the request work
	if len(resp.Status) > 0 && resp.Status[0] == '4' {
		c.log.Errorf("E! [outputs.influxdb] Failed to write metric (will be dropped: %s): %s\n", resp.Status, desc)
		return nil
	}

	// This error handles if there is an invalid or missing retention policy
	if strings.Contains(desc, errStringRetentionPolicyNotFound) {
		c.log.Errorf("When writing to [%s]: received error %v", c.URL(), desc)
		return nil
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
		c.log.Warnf("When writing to [%s]: received error %v",
			c.URL(), desc)
		return nil
	}

	// Other partial write errors, such as "field type conflict", are not
	// correctable at this point and so the point is dropped instead of
	// retrying.
	if strings.Contains(desc, errStringPartialWrite) {
		c.log.Errorf("When writing to [%s]: received error %v; discarding points",
			c.URL(), desc)
		return nil
	}

	// This error indicates a bug in either Telegraf line protocol
	// serialization, retries would not be successful.
	if strings.Contains(desc, errStringUnableToParse) {
		c.log.Errorf("When writing to [%s]: received error %v; discarding points",
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
	queryURL, err := makeQueryURL(c.config.URL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("q", query)
	form := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", queryURL, form)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := c.addHeaders(req); err != nil {
		return nil, err
	}

	return req, err
}

func (c *httpClient) makeWriteRequest(address string, body io.Reader) (*http.Request, error) {
	var err error

	req, err := http.NewRequest("POST", address, body)
	if err != nil {
		return nil, fmt.Errorf("failed creating new request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if err := c.addHeaders(req); err != nil {
		return nil, err
	}

	if c.config.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	return req, nil
}

// requestBodyReader warp io.Reader from influx.NewReader to io.ReadCloser, which is useful to fast close the write
// side of the connection in case of error
func (c *httpClient) requestBodyReader(metrics []telegraf.Metric) *wrappedReader {
	reader := influx.NewReader(metrics, c.config.Serializer)
	var rc io.ReadCloser
	if c.config.ContentEncoding == "gzip" {
		rc = internal.CompressWithGzip(reader)
	} else {
		rc = io.NopCloser(reader)
	}

	// Create a wrapper to be able to able to extract the number of bytes written
	return &wrappedReader{r: rc}
}

func (c *httpClient) addHeaders(req *http.Request) error {
	if !c.config.Username.Empty() || !c.config.Password.Empty() {
		username, err := c.config.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := c.config.Password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		req.SetBasicAuth(username.String(), password.String())
		username.Destroy()
		password.Destroy()
	}

	for header, value := range c.config.Headers {
		if strings.EqualFold(header, "host") {
			req.Host = value
		} else {
			req.Header.Set(header, value)
		}
	}

	return nil
}

func validateResponse(response io.ReadCloser) (io.ReadCloser, error) {
	bodyBytes, err := io.ReadAll(response)
	if err != nil {
		return nil, err
	}
	defer response.Close()

	originalResponse := io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Empty response is valid.
	if response == http.NoBody || len(bodyBytes) == 0 || bodyBytes == nil {
		return originalResponse, nil
	}

	if valid := json.Valid(bodyBytes); !valid {
		err = errors.New(string(bodyBytes))
	}

	return originalResponse, err
}

func makeWriteURL(loc *url.URL, db, rp, consistency string) (string, error) {
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
	case "http", "https":
		u.Path = path.Join(u.Path, "write")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}

func makeQueryURL(loc *url.URL) (string, error) {
	u := *loc
	switch u.Scheme {
	case "unix":
		u.Scheme = "http"
		u.Host = "127.0.0.1"
		u.Path = "/query"
	case "http", "https":
		u.Path = path.Join(u.Path, "query")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	return u.String(), nil
}

func (c *httpClient) Close() {
	c.client.CloseIdleConnections()
}

type wrappedReader struct {
	r            io.ReadCloser
	bytesWritten atomic.Int64
}

func (r *wrappedReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)

	// Atomically update bytesWritten to ensure thread-safe tracking during
	// concurrent reads
	r.bytesWritten.Add(int64(n))

	return n, err
}

func (r *wrappedReader) Close() error {
	return r.r.Close()
}
