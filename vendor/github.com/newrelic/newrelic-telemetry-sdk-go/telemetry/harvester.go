// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/internal"
)

// Harvester aggregates and reports metrics and spans.
type Harvester struct {
	// These fields are not modified after Harvester creation.  They may be
	// safely accessed without locking.
	config               Config
	commonAttributesJSON json.RawMessage

	// lock protects the mutable fields below.
	lock              sync.Mutex
	lastHarvest       time.Time
	rawMetrics        []Metric
	aggregatedMetrics map[metricIdentity]*metric
	spans             []Span
}

const (
	// NOTE:  These constant values are used in Config field doc comments.
	defaultHarvestPeriod  = 5 * time.Second
	defaultHarvestTimeout = 15 * time.Second
)

var (
	errAPIKeyUnset = errors.New("APIKey is required")
)

// NewHarvester creates a new harvester.
func NewHarvester(options ...func(*Config)) (*Harvester, error) {
	cfg := Config{
		Client:         &http.Client{},
		HarvestPeriod:  defaultHarvestPeriod,
		HarvestTimeout: defaultHarvestTimeout,
	}
	for _, opt := range options {
		opt(&cfg)
	}

	if cfg.APIKey == "" {
		return nil, errAPIKeyUnset
	}

	h := &Harvester{
		config:            cfg,
		lastHarvest:       time.Now(),
		aggregatedMetrics: make(map[metricIdentity]*metric),
	}

	// Marshal the common attributes to JSON here to avoid doing it on every
	// harvest.  This also has the benefit that it avoids race conditions if
	// the consumer modifies the CommonAttributes map after calling
	// NewHarvester.
	if nil != h.config.CommonAttributes {
		attrs := vetAttributes(h.config.CommonAttributes, h.config.logError)
		attributesJSON, err := json.Marshal(attrs)
		if err != nil {
			h.config.logError(map[string]interface{}{
				"err":     err.Error(),
				"message": "error marshaling common attributes",
			})
		} else {
			h.commonAttributesJSON = attributesJSON
		}
		h.config.CommonAttributes = nil
	}

	h.config.logDebug(map[string]interface{}{
		"event":                  "harvester created",
		"api-key":                h.config.APIKey,
		"harvest-period-seconds": h.config.HarvestPeriod.Seconds(),
		"metrics-url-override":   h.config.MetricsURLOverride,
		"spans-url-override":     h.config.SpansURLOverride,
		"version":                version,
	})

	if 0 != h.config.HarvestPeriod {
		go harvestRoutine(h)
	}

	return h, nil
}

var (
	errSpanIDUnset  = errors.New("span id must be set")
	errTraceIDUnset = errors.New("trace id must be set")
)

// RecordSpan records the given span.
func (h *Harvester) RecordSpan(s Span) error {
	if nil == h {
		return nil
	}
	if "" == s.TraceID {
		return errTraceIDUnset
	}
	if "" == s.ID {
		return errSpanIDUnset
	}
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	h.spans = append(h.spans, s)
	return nil
}

// RecordMetric adds a fully formed metric.  This metric is not aggregated with
// any other metrics and is never dropped.  The Timestamp field must be
// specified on Gauge metrics.  The Timestamp/Interval fields on Count and
// Summary are optional and will be assumed to be the harvester batch times if
// unset.  Use MetricAggregator() instead to aggregate metrics.
func (h *Harvester) RecordMetric(m Metric) {
	if nil == h {
		return
	}
	h.lock.Lock()
	defer h.lock.Unlock()

	if fields := m.validate(); nil != fields {
		h.config.logError(fields)
		return
	}

	h.rawMetrics = append(h.rawMetrics, m)
}

type response struct {
	statusCode int
	body       []byte
	err        error
	retryAfter string
}

var (
	backoffSequenceSeconds = []int{0, 1, 2, 4, 8, 16}
)

func (r response) needsRetry(cfg *Config, attempts int) (bool, time.Duration) {
	if attempts >= len(backoffSequenceSeconds) {
		attempts = len(backoffSequenceSeconds) - 1
	}
	backoff := time.Duration(backoffSequenceSeconds[attempts]) * time.Second

	switch r.statusCode {
	case 202, 200:
		// success
		return false, 0
	case 400, 403, 404, 405, 411, 413:
		// errors that should not retry
		return false, 0
	case 429:
		// special retry backoff time
		if "" != r.retryAfter {
			// Honor Retry-After header value in seconds
			if d, err := time.ParseDuration(r.retryAfter + "s"); nil == err {
				if d > backoff {
					return true, d
				}
			}
		}
		return true, backoff
	default:
		// all other errors should retry
		return true, backoff
	}
}

func postData(req *http.Request, client *http.Client) response {
	resp, err := client.Do(req)
	if nil != err {
		return response{err: fmt.Errorf("error posting data: %v", err)}
	}
	defer resp.Body.Close()

	r := response{
		statusCode: resp.StatusCode,
		retryAfter: resp.Header.Get("Retry-After"),
	}

	// On success, metrics ingest returns 202, span ingest returns 200.
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		r.body, _ = ioutil.ReadAll(resp.Body)
	} else {
		r.err = fmt.Errorf("unexpected post response code: %d: %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return r
}

func (h *Harvester) swapOutMetrics(now time.Time) []request {
	h.lock.Lock()
	lastHarvest := h.lastHarvest
	h.lastHarvest = now
	rawMetrics := h.rawMetrics
	h.rawMetrics = nil
	aggregatedMetrics := h.aggregatedMetrics
	h.aggregatedMetrics = make(map[metricIdentity]*metric, len(aggregatedMetrics))
	h.lock.Unlock()

	for _, m := range aggregatedMetrics {
		if nil != m.c {
			rawMetrics = append(rawMetrics, m.c)
		}
		if nil != m.s {
			rawMetrics = append(rawMetrics, m.s)
		}
		if nil != m.g {
			rawMetrics = append(rawMetrics, m.g)
		}
	}

	if 0 == len(rawMetrics) {
		return nil
	}

	batch := &metricBatch{
		Timestamp:      lastHarvest,
		Interval:       now.Sub(lastHarvest),
		AttributesJSON: h.commonAttributesJSON,
		Metrics:        rawMetrics,
	}
	reqs, err := newRequests(batch, h.config.APIKey, h.config.metricURL(), h.config.userAgent())
	if nil != err {
		h.config.logError(map[string]interface{}{
			"err":     err.Error(),
			"message": "error creating requests for metrics",
		})
		return nil
	}
	return reqs
}

func (h *Harvester) swapOutSpans() []request {
	h.lock.Lock()
	sps := h.spans
	h.spans = nil
	h.lock.Unlock()

	if nil == sps {
		return nil
	}
	batch := &spanBatch{
		AttributesJSON: h.commonAttributesJSON,
		Spans:          sps,
	}
	reqs, err := newRequests(batch, h.config.APIKey, h.config.spanURL(), h.config.userAgent())
	if nil != err {
		h.config.logError(map[string]interface{}{
			"err":     err.Error(),
			"message": "error creating requests for spans",
		})
		return nil
	}
	return reqs
}

func harvestRequest(req request, cfg *Config) {
	var attempts int
	for {
		cfg.logDebug(map[string]interface{}{
			"event":       "data post",
			"url":         req.Request.URL.String(),
			"body-length": req.compressedBodyLength,
		})
		// Check if the audit log is enabled to prevent unnecessarily
		// copying UncompressedBody.
		if cfg.auditLogEnabled() {
			cfg.logAudit(map[string]interface{}{
				"event": "uncompressed request body",
				"url":   req.Request.URL.String(),
				"data":  jsonString(req.UncompressedBody),
			})
		}

		resp := postData(req.Request, cfg.Client)

		if nil != resp.err {
			cfg.logError(map[string]interface{}{
				"err": resp.err.Error(),
			})
		} else {
			cfg.logDebug(map[string]interface{}{
				"event":  "data post response",
				"status": resp.statusCode,
				"body":   jsonOrString(resp.body),
			})
		}
		retry, backoff := resp.needsRetry(cfg, attempts)
		if !retry {
			return
		}

		tmr := time.NewTimer(backoff)
		select {
		case <-tmr.C:
			break
		case <-req.Request.Context().Done():
			tmr.Stop()
			return
		}
		attempts++
	}
}

// HarvestNow sends metric and span data to New Relic.  This method blocks until
// all data has been sent successfully or the Config.HarvestTimeout timeout has
// elapsed. This method can be used with a zero Config.HarvestPeriod value to
// control exactly when data is sent to New Relic servers.
func (h *Harvester) HarvestNow(ct context.Context) {
	if nil == h {
		return
	}

	ctx, cancel := context.WithTimeout(ct, h.config.HarvestTimeout)
	defer cancel()

	var reqs []request
	reqs = append(reqs, h.swapOutMetrics(time.Now())...)
	reqs = append(reqs, h.swapOutSpans()...)

	for _, req := range reqs {
		req.Request = req.Request.WithContext(ctx)
		harvestRequest(req, &h.config)
		if err := ctx.Err(); err != nil {
			// NOTE: It is possible that the context was
			// cancelled/timedout right after the request
			// successfully finished.  In that case, we will
			// erroneously log a message.  I (will) don't think
			// that's worth trying to engineer around.
			h.config.logError(map[string]interface{}{
				"event":         "harvest cancelled or timed out",
				"message":       "dropping data",
				"context-error": err.Error(),
			})
			return
		}
	}
}

func minDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

func harvestRoutine(h *Harvester) {
	// Introduce a small jitter to ensure the backend isn't hammered if many
	// harvesters start at once.
	d := minDuration(h.config.HarvestPeriod, 3*time.Second)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	jitter := time.Nanosecond * time.Duration(rnd.Int63n(d.Nanoseconds()))
	time.Sleep(jitter)

	ticker := time.NewTicker(h.config.HarvestPeriod)
	for range ticker.C {
		go h.HarvestNow(context.Background())
	}
}

type metricIdentity struct {
	// Note that the type is not a field here since a single 'metric' type
	// may contain a count, gauge, and summary.
	Name           string
	attributesJSON string
}

type metric struct {
	s *Summary
	c *Count
	g *Gauge
}

type metricHandle struct {
	metricIdentity
	harvester *Harvester
}

func newMetricHandle(h *Harvester, name string, attributes map[string]interface{}) metricHandle {
	return metricHandle{
		harvester: h,
		metricIdentity: metricIdentity{
			attributesJSON: string(internal.MarshalOrderedAttributes(attributes)),
			Name:           name,
		},
	}
}

// findOrCreateMetric finds or creates the metric associated with the given
// identity.  This function assumes the Harvester is locked.
func (h *Harvester) findOrCreateMetric(identity metricIdentity) *metric {
	m := h.aggregatedMetrics[identity]
	if nil == m {
		// this happens the first time we update the value,
		// or after a harvest when the metric is removed.
		m = &metric{}
		h.aggregatedMetrics[identity] = m
	}
	return m
}

// MetricAggregator is used to aggregate individual data points into metrics.
type MetricAggregator struct {
	harvester *Harvester
}

// MetricAggregator returns a metric aggregator.  Use this instead of
// RecordMetric if you have individual data points that you would like to
// combine into metrics.
func (h *Harvester) MetricAggregator() *MetricAggregator {
	if nil == h {
		return nil
	}
	return &MetricAggregator{harvester: h}
}

// Count creates a new AggregatedCount metric.
func (ag *MetricAggregator) Count(name string, attributes map[string]interface{}) *AggregatedCount {
	if nil == ag {
		return nil
	}
	return &AggregatedCount{metricHandle: newMetricHandle(ag.harvester, name, attributes)}
}

// Gauge creates a new AggregatedGauge metric.
func (ag *MetricAggregator) Gauge(name string, attributes map[string]interface{}) *AggregatedGauge {
	if nil == ag {
		return nil
	}
	return &AggregatedGauge{metricHandle: newMetricHandle(ag.harvester, name, attributes)}
}

// Summary creates a new AggregatedSummary metric.
func (ag *MetricAggregator) Summary(name string, attributes map[string]interface{}) *AggregatedSummary {
	if nil == ag {
		return nil
	}
	return &AggregatedSummary{metricHandle: newMetricHandle(ag.harvester, name, attributes)}
}
