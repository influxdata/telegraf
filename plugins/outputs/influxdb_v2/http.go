package influxdb_v2

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alitto/pond/v2"
	"golang.org/x/net/http2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/ratelimiter"
)

const (
	defaultMaxWaitSeconds           = 60
	defaultMaxWaitRetryAfterSeconds = 10 * 60
)

type httpClient struct {
	url              *url.URL
	localAddr        *net.TCPAddr
	token            config.Secret
	organization     string
	bucket           string
	bucketTag        string
	excludeBucketTag bool
	timeout          time.Duration
	headers          map[string]*config.Secret
	proxy            *url.URL
	userAgent        string
	contentEncoding  string
	pingTimeout      config.Duration
	readIdleTimeout  config.Duration
	tlsConfig        *tls.Config
	encoder          internal.ContentEncoder
	serializer       ratelimiter.Serializer
	rateLimiter      *ratelimiter.RateLimiter
	client           *http.Client
	params           url.Values
	retryTime        time.Time
	retryCount       atomic.Int64
	concurrent       uint64
	log              telegraf.Logger

	// Mutex to protect the retry-time field
	sync.Mutex

	pool pond.Pool
}

func (c *httpClient) Init() error {
	if c.headers == nil {
		c.headers = make(map[string]*config.Secret, 1)
	}

	if _, ok := c.headers["User-Agent"]; !ok {
		sec := config.NewSecret([]byte(c.userAgent))
		c.headers["User-Agent"] = &sec
	}

	var proxy func(*http.Request) (*url.URL, error)
	if c.proxy != nil {
		proxy = http.ProxyURL(c.proxy)
	} else {
		proxy = http.ProxyFromEnvironment
	}

	var transport *http.Transport
	switch c.url.Scheme {
	case "http", "https":
		var dialerFunc func(ctx context.Context, network, addr string) (net.Conn, error)
		if c.localAddr != nil {
			dialer := &net.Dialer{LocalAddr: c.localAddr}
			dialerFunc = dialer.DialContext
		}
		transport = &http.Transport{
			Proxy:           proxy,
			TLSClientConfig: c.tlsConfig,
			DialContext:     dialerFunc,
		}
		if c.readIdleTimeout != 0 || c.pingTimeout != 0 {
			http2Trans, err := http2.ConfigureTransports(transport)
			if err == nil {
				http2Trans.ReadIdleTimeout = time.Duration(c.readIdleTimeout)
				http2Trans.PingTimeout = time.Duration(c.pingTimeout)
			}
		}
	case "unix":
		transport = &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.DialTimeout(
					c.url.Scheme,
					c.url.Path,
					c.timeout,
				)
			},
		}
	default:
		return fmt.Errorf("unsupported scheme %q", c.url.Scheme)
	}

	preppedURL, params, err := prepareWriteURL(*c.url, c.organization)
	if err != nil {
		return err
	}

	c.url = preppedURL
	c.client = &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}
	c.params = params

	// Use single-threaded writing by default.
	if c.concurrent < 1 {
		c.concurrent = 1
	}
	c.pool = pond.NewPool(int(c.concurrent))
	return nil
}

func (c *httpClient) Write(ctx context.Context, metrics []telegraf.Metric) error {
	if c.retryTime.After(time.Now()) {
		return errors.New("retry time has not elapsed")
	}

	// Create the batches for sending
	workers := int(c.concurrent)
	batchSize := len(metrics) / workers
	if len(metrics)%workers > 0 {
		batchSize++
	}
	var batches []*batch
	if c.bucketTag == "" {
		batches = createBatches(metrics, c.bucket, batchSize)
	} else {
		batches = createBatchesFromTag(metrics, c.bucketTag, c.bucket, batchSize, c.excludeBucketTag)
	}

	// Serialize the data in the batches
	ratets := time.Now()
	defer c.rateLimiter.Release()

	limitReached := -1
	var writeErr internal.PartialWriteError
	for i, batch := range batches {
		// Get the current limit for the outbound data
		limit := c.rateLimiter.Remaining(ratets)

		// Serialize the metrics with the remaining limit, exit early if nothing was serialized
		used, err := batch.serialize(c.serializer, limit, c.encoder)
		if err != nil {
			var werr *internal.PartialWriteError
			if errors.As(err, &werr) {
				writeErr.MetricsReject = append(writeErr.MetricsReject, werr.MetricsReject...)
				writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, werr.MetricsRejectErrors...)
				writeErr.Err = werr.Err
			} else {
				writeErr.Err = err
				batch.err = err
			}
		}
		c.rateLimiter.Reserve(used)

		if errors.Is(batch.err, internal.ErrSizeLimitReached) {
			limitReached = i
			// If we serialized at least one metric in this batch the limit
			// should include the current batch, otherwise we stop before this
			// batch.
			if used > 0 {
				limitReached++
			}
			break
		}
	}

	// Skip all non-serialized batches
	if limitReached > 0 && limitReached < len(batches) {
		batches = batches[:limitReached]
	}

	// Send the batches
	var splitMu sync.Mutex
	var split []int
	var throttle atomic.Bool
	tasks := c.pool.NewGroupContext(ctx)
	defer tasks.Stop()
	for i, batch := range batches {
		// Stop writes as soon as we encounter a throttling request of the
		// server to not cause more overload
		if throttle.Load() {
			break
		}
		tasks.Submit(func() {
			// Stop writes as soon as we encounter a throttling request of the
			// server to not cause more overload
			if throttle.Load() {
				return
			}
			c.rateLimiter.Accept(ratets, int64(len(batch.payload)))
			batch.processed = true
			if err := c.writeBatch(ctx, batch); err != nil {
				var terr *ThrottleError
				if errors.As(err, &terr) {
					if terr.StatusCode == http.StatusRequestEntityTooLarge {
						splitMu.Lock()
						split = append(split, i)
						splitMu.Unlock()
					} else {
						throttle.Store(true)

						// Remember when we can send again
						// To be on the safe side use the latest time we encounter
						retryAfter := time.Now().Add(terr.RetryAfter)
						c.Lock()
						if retryAfter.After(c.retryTime) {
							c.retryTime = retryAfter
						}
						c.Unlock()
					}
				}
				batch.err = err
			}
		})
	}
	if err := tasks.Wait(); err != nil {
		if writeErr.Err != nil {
			return &writeErr
		}
		return err
	}

	// Explicitly release all reserved rate portions here as we finished the
	// first sending stage. Below we may also reserve rate portions but those
	// are released using the deferred statement above later on.
	c.rateLimiter.Release()

	// Handle the batches that need resending and remove the split instances
	if !throttle.Load() {
		slices.Reverse(split)
		for _, idx := range split {
			// Delete the split patch
			batch := batches[idx]
			s := c.splitAndWrite(ctx, batch)
			batches = append(batches, s...)
			batches = slices.Delete(batches, idx, idx+1)
		}
	}

	// Check the errors
	allProcessed := true
	for _, batch := range batches {
		allProcessed = allProcessed && batch.processed
		err := batch.err

		// Mark all metrics as accepted if sending was OK
		if err == nil {
			writeErr.MetricsAccept = append(writeErr.MetricsAccept, batch.indices...)
			continue
		}

		// Propagate the error
		writeErr.Err = err
		c.log.Error(err)

		// API errors might be retyable depending on what the server says
		var apiErr *APIError
		if errors.As(err, &apiErr) && !apiErr.Retryable {
			// If the error is retryable, we simply do not mark any metric of
			// that batch and the metrics will be re-queued.
			writeErr.MetricsReject = append(writeErr.MetricsReject, batch.indices...)
			writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, err)
		}
	}
	if writeErr.Err == nil && !allProcessed {
		writeErr.Err = errors.New("not all metrics have been sent")
	}
	if writeErr.Err != nil {
		return &writeErr
	}
	return nil
}

func (c *httpClient) writeBatch(ctx context.Context, b *batch) error {
	// Setup the request
	address := makeWriteURL(*c.url, c.params, b.bucket)
	req, err := http.NewRequest("POST", address, io.NopCloser(bytes.NewBuffer(b.payload)))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}
	if c.encoder != nil {
		req.Header.Set("Content-Encoding", c.contentEncoding)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	// Set authorization
	token, err := c.token.Get()
	if err != nil {
		return fmt.Errorf("getting token failed: %w", err)
	}
	req.Header.Set("Authorization", "Token "+token.String())
	token.Destroy()

	if err := c.addHeaders(req); err != nil {
		return fmt.Errorf("adding headers failed: %w", err)
	}

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
		c.retryCount.Store(0)
		return nil
	}

	// We got an error and now try to decode further
	var desc string
	writeResp := &genericRespError{}
	if json.NewDecoder(resp.Body).Decode(writeResp) == nil {
		desc = ": " + writeResp.Error()
	}

	switch resp.StatusCode {
	// request was too large, send back to try again
	case http.StatusRequestEntityTooLarge:
		c.log.Errorf("Failed to write metrics with size %d bytes to %s, request was too large (413)", len(b.payload), b.bucket)
		return &ThrottleError{
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
			Err:        fmt.Errorf("failed to write metrics to %s (will be dropped: %s)%s", b.bucket, resp.Status, desc),
			StatusCode: resp.StatusCode,
		}
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("failed to write metrics to %s (%s)%s", b.bucket, resp.Status, desc)
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusBadGateway,
		http.StatusGatewayTimeout:
		// ^ these handle the cases where the server is likely overloaded, and may not be able to say so.
		retryDuration := getRetryDuration(resp.Header, c.retryCount.Add(1))
		return &ThrottleError{
			Err:        fmt.Errorf("failed to write to %s; will retry in %s. (%s)", b.bucket, retryDuration, resp.Status),
			StatusCode: resp.StatusCode,
			RetryAfter: retryDuration,
		}
	}

	// if it's any other 4xx code, the client should not retry as it's the client's mistake.
	// retrying will not make the request magically work.
	if len(resp.Status) > 0 && resp.Status[0] == '4' {
		return &APIError{
			Err:        fmt.Errorf("failed to write metrics to %s (will be dropped: %s)%s", b.bucket, resp.Status, desc),
			StatusCode: resp.StatusCode,
		}
	}

	// This is only until platform spec is fully implemented. As of the
	// time of writing, there is no error body returned.
	if xErr := resp.Header.Get("X-Influx-Error"); xErr != "" {
		desc = fmt.Sprintf(": %s; %s", desc, xErr)
	}

	return &APIError{
		Err:        fmt.Errorf("failed to write metrics to bucket %q: %s%s", b.bucket, resp.Status, desc),
		StatusCode: resp.StatusCode,
		Retryable:  true,
	}
}

func (c *httpClient) splitAndWrite(ctx context.Context, b *batch) []*batch {
	// Ignore the rate-limit for now and serialize what we have. The resulting
	// batch should _always_ be smaller than before splitting so we should be
	// able to make progress here.
	limit := int64(math.MaxInt64)

	// Split the batch and resend both parts
	first, second := b.split()

	// Serialize each element and send it
	var splits []*batch
	for _, current := range []*batch{first, second} {
		if _, err := current.serialize(c.serializer, limit, c.encoder); err != nil {
			current.err = err
			splits = append(splits, current)
		} else {
			if err := c.writeBatch(ctx, current); err != nil {
				current.err = err

				var terr *ThrottleError
				if errors.As(err, &terr) && terr.StatusCode == http.StatusRequestEntityTooLarge && len(b.metrics) > 1 {
					s := c.splitAndWrite(ctx, current)
					splits = append(splits, s...)
				} else {
					splits = append(splits, current)
				}
			} else {
				splits = append(splits, current)
			}
		}
	}

	return splits
}

// retryDuration takes the longer of the Retry-After header and our own back-off calculation
func getRetryDuration(headers http.Header, count int64) time.Duration {
	// basic exponential backoff (x^2)/40 (denominator to widen the slope)
	// at 40 denominator, it'll take 49 retries to hit the max defaultMaxWait of 60s
	backoff := math.Pow(float64(count), 2) / 40
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

func (c *httpClient) addHeaders(req *http.Request) error {
	for header, value := range c.headers {
		secret, err := value.Get()
		if err != nil {
			return err
		}

		headerVal := secret.String()
		secret.Destroy()
		if strings.EqualFold(header, "host") {
			req.Host = headerVal
		} else {
			req.Header.Set(header, headerVal)
		}
	}

	return nil
}

func makeWriteURL(loc url.URL, params url.Values, bucket string) string {
	p := maps.Clone(params)
	p.Set("bucket", bucket)
	loc.RawQuery = p.Encode()
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
