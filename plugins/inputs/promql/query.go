package promql

import (
	"context"
	"fmt"
	"strconv"
	"time"

	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

type query struct {
	Query string `toml:"query"`
	Limit uint64 `toml:"limit"`

	client  *client
	options []apiv1.Option
	log     telegraf.Logger
}

func (q *query) init(c *client, log telegraf.Logger, options ...apiv1.Option) {
	q.client = c
	q.options = append(options, apiv1.WithLimit(q.Limit))
	q.log = log
}

type InstantQuery struct {
	query
}

func (q *InstantQuery) init(c *client, log telegraf.Logger, options ...apiv1.Option) {
	q.query.init(c, log, options...)
}

func (q *InstantQuery) execute(ctx context.Context, acc telegraf.Accumulator, t time.Time) error {
	results, warnings, err := q.client.Query(ctx, q.Query, t, q.options...)
	if err != nil {
		return fmt.Errorf("querying %q failed: %w", q.client.url, err)
	}
	for _, w := range warnings {
		q.log.Warnf("query %q produced warning: %s", q.Query, w)
	}

	return q.convertModelValue(acc, results)
}

type RangeQuery struct {
	query
	Start config.Duration `toml:"start"`
	End   config.Duration `toml:"end"`
	Step  config.Duration `toml:"step"`
}

func (q *RangeQuery) init(c *client, log telegraf.Logger, options ...apiv1.Option) {
	q.query.init(c, log, options...)
}

func (q *RangeQuery) execute(ctx context.Context, acc telegraf.Accumulator, t time.Time) error {
	r := apiv1.Range{
		Start: t.Add(-time.Duration(q.Start)),
		End:   t.Add(-time.Duration(q.End)),
		Step:  time.Duration(q.Step),
	}
	results, warnings, err := q.client.QueryRange(ctx, q.Query, r, q.options...)
	if err != nil {
		return fmt.Errorf("querying %q failed: %w", q.client.url, err)
	}
	for _, w := range warnings {
		q.log.Warnf("query %q produced warning: %s", q.Query, w)
	}
	return q.convertModelValue(acc, results)
}

func (q *query) convertModelValue(acc telegraf.Accumulator, results model.Value) error {
	switch result := results.(type) {
	case *model.Scalar:
		tags := make(map[string]string)
		fields := map[string]interface{}{"value": float64(result.Value)}
		acc.AddGauge("promql", fields, tags, result.Timestamp.Time())
	case *model.String:
		tags := make(map[string]string)
		fields := map[string]interface{}{"value": result.Value}
		acc.AddFields("promql", fields, tags, result.Timestamp.Time())
	case model.Vector:
		if result.Len() == 0 {
			q.log.Debugf("Query %q returned no result", q.Query)
			return nil
		}
		for _, sample := range result {
			var name string
			tags := make(map[string]string, len(sample.Metric))
			for k, v := range sample.Metric {
				if k == "__name__" {
					name = string(v)
					continue
				}
				tags[string(k)] = string(v)
			}
			if sample.Histogram != nil {
				hist := sample.Histogram
				fields := make(map[string]interface{}, 2+len(hist.Buckets))
				fields["count"] = hist.Count
				fields["sum"] = hist.Sum
				for _, b := range hist.Buckets {
					fields[strconv.FormatFloat(float64(b.Upper), 'g', -1, 64)] = float64(b.Count)
				}
				acc.AddHistogram(name, fields, tags, sample.Timestamp.Time())
			} else {
				fields := map[string]interface{}{"value": float64(sample.Value)}
				acc.AddGauge(name, fields, tags, sample.Timestamp.Time())
			}
		}
	case model.Matrix:
		if result.Len() == 0 {
			q.log.Debugf("Query %q returned no result", q.Query)
			return nil
		}
		for _, stream := range result {
			var name string
			tags := make(map[string]string, len(stream.Metric))
			for k, v := range stream.Metric {
				if k == "__name__" {
					name = string(v)
					continue
				}
				tags[string(k)] = string(v)
			}
			for _, v := range stream.Values {
				fields := map[string]interface{}{"value": float64(v.Value)}
				acc.AddGauge(name, fields, tags, v.Timestamp.Time())
			}
			for _, h := range stream.Histograms {
				hist := h.Histogram
				fields := make(map[string]interface{}, 2+len(hist.Buckets))
				fields["count"] = hist.Count
				fields["sum"] = hist.Sum
				for _, b := range hist.Buckets {
					fields[strconv.FormatFloat(float64(b.Upper), 'g', -1, 64)] = float64(b.Count)
				}
				acc.AddHistogram(name, fields, tags, h.Timestamp.Time())
			}
		}
	default:
		return fmt.Errorf("unknown result type %T", result)
	}

	return nil
}
