package influxdb_v3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/ratelimiter"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

const (
	defaultMaxWaitSeconds           = 60
	defaultMaxWaitRetryAfterSeconds = 10 * 60
)

type client struct {
	url *url.URL
	cfg *clientConfig
	log telegraf.Logger

	client      *http.Client
	headers     map[string]string
	encoder     internal.ContentEncoder
	serializer  ratelimiter.Serializer
	rateLimiter *ratelimiter.RateLimiter
	retryTime   time.Time
	retryCount  int

	cancel context.CancelFunc
}

func newClient(addr string, cfg *clientConfig, log telegraf.Logger) (*client, error) {
	// Parse the URL
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing URL %q failed: %w", addr, err)
	}

	switch u.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("invalid scheme in URL %q", addr)
	}

	c := &client{
		url: u.JoinPath("api", "v3", "write_lp"),
		cfg: cfg,
		log: log,
	}

	// Prepare the headers including authorization
	c.headers = map[string]string{
		"User-Agent": cfg.UserAgent,
	}

	if _, ok := c.headers["Authorization"]; !ok && !cfg.Token.Empty() {
		token, err := cfg.Token.Get()
		if err != nil {
			return nil, fmt.Errorf("getting token failed: %w", err)
		}
		c.headers["Authorization"] = "Bearer " + token.String()
		token.Destroy()
	}

	// Init encoding if configured
	switch cfg.ContentEncoding {
	case "gzip":
		enc, err := internal.NewGzipEncoder()
		if err != nil {
			return nil, fmt.Errorf("setting up gzip encoder failed: %w", err)
		}
		c.encoder = enc
		c.headers["Content-Encoding"] = "gzip"
	case "none", "identity":
	default:
		return nil, fmt.Errorf("invalid content encoding %q", cfg.ContentEncoding)
	}

	// Setup the limited serializer
	serializer := &influx.Serializer{
		UintSupport:   cfg.UintSupport,
		OmitTimestamp: cfg.OmitTimestamp,
	}
	if err := serializer.Init(); err != nil {
		return nil, fmt.Errorf("setting up serializer failed: %w", err)
	}
	c.serializer = ratelimiter.NewIndividualSerializer(serializer)

	// Create ratelimiter
	limiter, err := cfg.RateLimitConfig.CreateRateLimiter()
	if err != nil {
		return nil, fmt.Errorf("creating rate-limiter failed: %w", err)
	}
	c.rateLimiter = limiter

	return c, nil
}

func (c *client) connect() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	// Setup the HTTP client
	httpClient, err := c.cfg.HTTPClientConfig.CreateClient(ctx, c.log)
	if err != nil {
		return fmt.Errorf("creating HTTP client failed: %w", err)
	}
	c.client = httpClient

	return nil
}

func (c *client) close() {
	c.cancel()
	c.client.CloseIdleConnections()
}

func (c *client) write(ctx context.Context, metrics []telegraf.Metric) error {
	if c.retryTime.After(time.Now()) {
		return errors.New("retry time has not elapsed")
	}

	batches := make(map[string][]telegraf.Metric)
	batchIndices := make(map[string][]int)
	if c.cfg.DatabaseTag == "" {
		batches[c.cfg.Database] = metrics
		batchIndices[c.cfg.Database] = make([]int, len(metrics))
		for i := range metrics {
			batchIndices[c.cfg.Database][i] = i
		}
	} else {
		for i, metric := range metrics {
			database, ok := metric.GetTag(c.cfg.DatabaseTag)
			if !ok {
				database = c.cfg.Database
			} else if c.cfg.ExcludeDatabaseTag {
				// Avoid modifying the metric if we do remove the tag
				metric = metric.Copy()
				metric.Accept()
				metric.RemoveTag(c.cfg.DatabaseTag)
			}

			batches[database] = append(batches[database], metric)
			batchIndices[database] = append(batchIndices[database], i)
		}
	}

	var wErr internal.PartialWriteError
	for database, batch := range batches {
		err := c.writeBatch(ctx, database, batch)
		if err == nil {
			wErr.MetricsAccept = append(wErr.MetricsAccept, batchIndices[database]...)
			continue
		}

		// Check if the request was too large and split it
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode == http.StatusRequestEntityTooLarge {
				// TODO: Need a testcase to verify rejected metrics are not retried...
				return c.splitAndWriteBatch(ctx, c.cfg.Database, metrics)
			}
			wErr.Err = err
			if !apiErr.Retryable {
				wErr.MetricsReject = append(wErr.MetricsReject, batchIndices[database]...)
			}
			// TODO: Clarify if we should continue here to try the remaining databases?
			return &wErr
		}

		// Check if we got a write error and if so, translate the returned
		// metric indices to return the original indices in case of dynamic
		// database targets
		var writeErr *internal.PartialWriteError
		if errors.As(err, &writeErr) {
			wErr.Err = writeErr.Err
			for _, idx := range writeErr.MetricsAccept {
				wErr.MetricsAccept = append(wErr.MetricsAccept, batchIndices[database][idx])
			}
			for _, idx := range writeErr.MetricsReject {
				wErr.MetricsReject = append(wErr.MetricsReject, batchIndices[database][idx])
			}
			if !errors.Is(writeErr.Err, internal.ErrSizeLimitReached) {
				continue
			}
			return &wErr
		}

		// Return the error without special treatment
		wErr.Err = err
		return &wErr
	}
	return nil
}

func (c *client) splitAndWriteBatch(ctx context.Context, database string, metrics []telegraf.Metric) error {
	c.log.Warnf("Retrying write after splitting metric payload in half to reduce batch size")
	midpoint := len(metrics) / 2

	if err := c.writeBatch(ctx, database, metrics[:midpoint]); err != nil {
		return err
	}

	return c.writeBatch(ctx, database, metrics[midpoint:])
}

func (c *client) writeBatch(ctx context.Context, database string, metrics []telegraf.Metric) error {
	// Get the current limit for the outbound data
	ratets := time.Now()
	limit := c.rateLimiter.Remaining(ratets)

	// Serialize the metrics with the remaining limit, exit early if nothing was serialized
	body, werr := c.serializer.SerializeBatch(metrics, limit)
	if werr != nil && !errors.Is(werr, internal.ErrSizeLimitReached) || len(body) == 0 {
		return werr
	}
	used := int64(len(body))

	// Encode the content if requested
	if c.encoder != nil {
		var err error
		if body, err = c.encoder.Encode(body); err != nil {
			return fmt.Errorf("encoding failed: %w", err)
		}
	}

	// Setup the request
	params := c.url.Query()
	params.Set("db", database)
	c.url.RawQuery = params.Encode()
	req, err := http.NewRequest("POST", c.url.String(), io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	for header, value := range c.headers {
		if strings.EqualFold(header, "host") {
			req.Host = value
		} else {
			req.Header.Set(header, value)
		}
	}

	// Execute the request
	c.rateLimiter.Accept(ratets, used)
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
		return werr
	}

	// We got an error and now try to decode further
	var desc string
	if responseBody, err := io.ReadAll(resp.Body); err == nil && len(responseBody) > 0 {
		var respErr apiErrorBody
		if json.Unmarshal(responseBody, &respErr) == nil {
			if respErr.ErrorMsg != "" {
				desc = respErr.ErrorMsg
			}
			if c.log.Level().Includes(telegraf.Debug) {
				for _, d := range respErr.Data {
					c.log.Debugf("  writing %q failed: %s)", d.Metric, d.Message)
				}
			}
		} else {
			desc = string(responseBody)
		}
		desc = ": " + desc
	}

	switch resp.StatusCode {
	// request was too large, send back to try again
	case http.StatusRequestEntityTooLarge:
		c.log.Errorf("Failed to write metric to %s, request was too large (413)", database)
		return &APIError{
			Err:        fmt.Errorf("%s: %s", resp.Status, desc),
			StatusCode: resp.StatusCode,
		}
	case
		// request was malformed:
		http.StatusBadRequest,
		// request was received but server refused to process it due to a semantic problem with the request.
		// for example, submitting metrics outside the retention period.
		http.StatusUnprocessableEntity,
		http.StatusNotAcceptable:

		// Clients should *not* repeat the request and the metrics should be rejected.
		return &APIError{
			Err:        fmt.Errorf("failed to write metric to %s (will be dropped: %s)%s", database, resp.Status, desc),
			StatusCode: resp.StatusCode,
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("failed to write metric to %s (%s)%s", database, resp.Status, desc)
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusBadGateway,
		http.StatusGatewayTimeout:
		// ^ these handle the cases where the server is likely overloaded, and may not be able to say so.
		c.retryCount++
		retryDuration := c.getRetryDuration(resp.Header)
		c.retryTime = time.Now().Add(retryDuration)
		c.log.Warnf("Failed to write to %s; will retry in %s. (%s)\n", database, retryDuration, resp.Status)
		return fmt.Errorf("waiting %s for server (%s) before sending metric again", retryDuration, database)
	}

	// if it's any other 4xx code, the client should not retry as it's the client's mistake.
	// retrying will not make the request magically work.
	if len(resp.Status) > 0 && resp.Status[0] == '4' {
		return &APIError{
			Err:        fmt.Errorf("failed to write metric to %s (will be dropped: %s)%s", database, resp.Status, desc),
			StatusCode: resp.StatusCode,
		}
	}

	return &APIError{
		Err:        fmt.Errorf("failed to write metric to database %q: %s%s", database, resp.Status, desc),
		StatusCode: resp.StatusCode,
		Retryable:  true,
	}
}

// retryDuration takes the longer of the Retry-After header and our own back-off calculation
func (c *client) getRetryDuration(headers http.Header) time.Duration {
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
