package influxdb

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

const (
	defaultRequestTimeout          = time.Second * 5
	defaultDatabase                = "telegraf"
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
	URL                       *url.URL
	UserAgent                 string
	Timeout                   time.Duration
	Username                  string
	Password                  string
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

func NewHTTPClient(config HTTPConfig) (*httpClient, error) {
	if config.URL == nil {
		return nil, ErrMissingURL
	}

	if config.Database == "" {
		config.Database = defaultDatabase
	}

	if config.Timeout == 0 {
		config.Timeout = defaultRequestTimeout
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = internal.ProductToken()
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}
	config.Headers["User-Agent"] = userAgent
	for k, v := range config.Headers {
		config.Headers[k] = v
	}

	var proxy func(*http.Request) (*url.URL, error)
	if config.Proxy != nil {
		proxy = http.ProxyURL(config.Proxy)
	} else {
		proxy = http.ProxyFromEnvironment
	}

	if config.Serializer == nil {
		config.Serializer = influx.NewSerializer()
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
					defaultRequestTimeout,
				)
			},
		}
	default:
		return nil, fmt.Errorf("unsupported scheme %q", config.URL.Scheme)
	}

	client := &httpClient{
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		createDatabaseExecuted: make(map[string]bool),
		config:                 config,
		log:                    config.Log,
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
		return err
	}

	reader, err := c.requestBodyReader(metrics)
	if err != nil {
		return err
	}
	defer reader.Close()

	req, err := c.makeWriteRequest(loc, reader)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		internal.OnClientError(c.client, err)
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
		return &DatabaseNotFoundError{
			APIError: APIError{
				StatusCode:  resp.StatusCode,
				Title:       resp.Status,
				Description: desc,
			},
			Database: db,
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
	c.addHeaders(req)

	return req, nil
}

func (c *httpClient) makeWriteRequest(url string, body io.Reader) (*http.Request, error) {
	var err error

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	c.addHeaders(req)

	if c.config.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	return req, nil
}

// requestBodyReader warp io.Reader from influx.NewReader to io.ReadCloser, which is usefully to fast close the write
// side of the connection in case of error
func (c *httpClient) requestBodyReader(metrics []telegraf.Metric) (io.ReadCloser, error) {
	reader := influx.NewReader(metrics, c.config.Serializer)

	if c.config.ContentEncoding == "gzip" {
		rc, err := internal.CompressWithGzip(reader)
		if err != nil {
			return nil, err
		}

		return rc, nil
	}

	return ioutil.NopCloser(reader), nil
}

func (c *httpClient) addHeaders(req *http.Request) {
	if c.config.Username != "" || c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	for header, value := range c.config.Headers {
		req.Header.Set(header, value)
	}
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
