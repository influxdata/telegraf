// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"encoding/json"
	"errors"
	"math"
	"time"
)

var (
	errFloatInfinity = errors.New("invalid float is infinity")
	errFloatNaN      = errors.New("invalid float is NaN")
)

func isFloatValid(f float64) error {
	if math.IsInf(f, 0) {
		return errFloatInfinity
	}
	if math.IsNaN(f) {
		return errFloatNaN
	}
	return nil
}

// AggregatedCount is the metric type that counts the number of times an event occurred.
// This counter is reset every time the data is reported, meaning the value
// reported represents the difference in count over the reporting time window.
//
// Example possible uses:
//
//  * the number of messages put on a topic
//  * the number of HTTP requests
//  * the number of errors thrown
//  * the number of support tickets answered
//
type AggregatedCount struct{ metricHandle }

// Increment increases the Count value by one.
func (c *AggregatedCount) Increment() {
	c.Increase(1)
}

// Increase increases the Count value by the number given.  The value must be
// non-negative.
func (c *AggregatedCount) Increase(val float64) {
	if nil == c {
		return
	}
	if val < 0 {
		return
	}

	h := c.harvester
	if nil == h {
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if err := isFloatValid(val); err != nil {
		h.config.logError(map[string]interface{}{
			"message": "invalid aggregated count value",
			"err":     err.Error(),
		})
		return
	}

	m := h.findOrCreateMetric(c.metricIdentity)
	if nil == m.c {
		m.c = &Count{
			Name:           c.Name,
			AttributesJSON: json.RawMessage(c.attributesJSON),
		}
	}
	m.c.Value += val
}

// AggregatedGauge is the metric type that records a value that can increase or decrease.
// It generally represents the value for something at a particular moment in
// time.  One typically records a AggregatedGauge value on a set interval.
//
// Only the most recent AggregatedGauge metric value is reported over a given harvest
// period, all others are dropped.
//
// Example possible uses:
//
//  * the temperature in a room
//  * the amount of memory currently in use for a process
//  * the bytes per second flowing into Kafka at this exact moment in time
//  * the current speed of your car
//
type AggregatedGauge struct{ metricHandle }

// valueNow facilitates testing.
func (g *AggregatedGauge) valueNow(val float64, now time.Time) {
	if nil == g {
		return
	}
	h := g.harvester
	if nil == h {
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if err := isFloatValid(val); err != nil {
		h.config.logError(map[string]interface{}{
			"message": "invalid aggregated gauge value",
			"err":     err.Error(),
		})
		return
	}

	m := h.findOrCreateMetric(g.metricIdentity)
	if nil == m.g {
		m.g = &Gauge{
			Name:           g.Name,
			AttributesJSON: json.RawMessage(g.attributesJSON),
			Value:          val,
		}
	}
	m.g.Value = val
	m.g.Timestamp = now
}

// Value records the value given.
func (g *AggregatedGauge) Value(val float64) {
	g.valueNow(val, time.Now())
}

// AggregatedSummary is the metric type used for reporting aggregated information about
// discrete events.   It provides the count, average, sum, min and max values
// over time.  All fields are reset to 0 every reporting interval.
//
// The final metric reported at the end of a harvest period is an aggregation.
// Values reported are the count of the number of metrics recorded, sum of
// all their values, minimum value recorded, and maximum value recorded.
//
// Example possible uses:
//
//  * the duration and count of spans
//  * the duration and count of transactions
//  * the time each message spent in a queue
//
type AggregatedSummary struct{ metricHandle }

// Record adds an observation to a summary.
func (s *AggregatedSummary) Record(val float64) {
	if nil == s {
		return
	}
	h := s.harvester
	if nil == h {
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if err := isFloatValid(val); err != nil {
		h.config.logError(map[string]interface{}{
			"message": "invalid aggregated summary value",
			"err":     err.Error(),
		})
		return
	}

	m := h.findOrCreateMetric(s.metricIdentity)
	if nil == m.s {
		m.s = &Summary{
			Name:           s.Name,
			AttributesJSON: json.RawMessage(s.attributesJSON),
			Count:          1,
			Sum:            val,
			Min:            val,
			Max:            val,
		}
		return
	}
	m.s.Sum += val
	m.s.Count++
	if val < m.s.Min {
		m.s.Min = val
	}
	if val > m.s.Max {
		m.s.Max = val
	}
}

// RecordDuration adds a duration observation to a summary.  It records the
// value in milliseconds, New Relic's recommended duration units.
func (s *AggregatedSummary) RecordDuration(val time.Duration) {
	s.Record(val.Seconds() * 1000.0)
}
