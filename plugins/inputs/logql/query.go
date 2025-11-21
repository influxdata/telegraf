package logql

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

type query struct {
	Name    string `toml:"name"`
	Query   string `toml:"query"`
	Sorting string `toml:"sorting"`
	Limit   uint64 `toml:"limit"`

	client *client
	url    *url.URL
	values *url.Values
	log    telegraf.Logger
}

func (q *query) init(c *client, log telegraf.Logger) error {
	if q.Query == "" {
		return fmt.Errorf("'query' cannot be empty for %q", c.url)
	}

	switch q.Sorting {
	case "", "forward", "backward":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("invalid sorting direction %q", q.Sorting)
	}

	// Prepare the query information from the URL and given parameters
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parsing URL %q failed: %w", c.url, err)
	}

	q.client = c
	q.url = u
	q.log = log

	q.values = &url.Values{}
	q.values.Set("query", q.Query)
	if q.Sorting != "" {
		q.values.Set("direction", q.Sorting)
	}
	if q.Limit > 0 {
		q.values.Set("limit", strconv.FormatUint(q.Limit, 10))
	}
	return nil
}

type InstantQuery struct {
	query
}

func (q *InstantQuery) init(c *client, log telegraf.Logger) error {
	// Init the underlying query
	if err := q.query.init(c, log); err != nil {
		return err
	}

	// Set instant-query specific values
	q.query.url = q.query.url.JoinPath("loki", "api", "v1", "query")

	return nil
}

func (q *InstantQuery) execute(ctx context.Context, acc telegraf.Accumulator, t time.Time) error {
	q.query.values.Set("time", t.Format(time.RFC3339Nano))
	q.query.url.RawQuery = q.query.values.Encode()

	result, err := q.client.execute(ctx, q.query.url.String())
	if err != nil {
		return fmt.Errorf("querying %q failed: %w", q.client.url, err)
	}

	return q.convertResult(acc, result)
}

type RangeQuery struct {
	query
	Start    config.Duration `toml:"start"`
	End      config.Duration `toml:"end"`
	Step     config.Duration `toml:"step"`
	Interval config.Duration `toml:"interval"`
}

func (q *RangeQuery) init(c *client, log telegraf.Logger) error {
	// Check the parameters
	if q.Start <= q.End {
		return fmt.Errorf("invalid range %v to %v for query %q", q.Start, q.End, q.Query)
	}
	if q.Step < 0 {
		return fmt.Errorf("'step' must be non-negative for query %q", q.query.Query)
	}
	if q.Interval < 0 {
		return fmt.Errorf("'interval' must be non-negative for query %q", q.query.Query)
	}

	// Init the underlying query
	if err := q.query.init(c, log); err != nil {
		return err
	}

	// Set range-query specific values
	q.query.url = q.query.url.JoinPath("loki", "api", "v1", "query_range")
	if q.Step > 0 {
		q.query.values.Set("step", time.Duration(q.Step).String())
	}
	if q.Interval > 0 {
		q.query.values.Set("interval", time.Duration(q.Interval).String())
	}

	return nil
}

func (q *RangeQuery) execute(ctx context.Context, acc telegraf.Accumulator, t time.Time) error {
	q.query.values.Set("start", t.Add(-time.Duration(q.Start)).Format(time.RFC3339Nano))
	q.query.values.Set("end", t.Add(-time.Duration(q.End)).Format(time.RFC3339Nano))
	q.query.url.RawQuery = q.query.values.Encode()

	result, err := q.client.execute(ctx, q.query.url.String())
	if err != nil {
		return fmt.Errorf("querying %q failed: %w", q.client.url, err)
	}

	return q.convertResult(acc, result)
}

func (q *query) convertResult(acc telegraf.Accumulator, result interface{}) error {
	// Determine the default name
	name := "logql"
	if q.Name != "" {
		name = q.Name
	}

	switch entries := result.(type) {
	case []vector:
		for _, r := range entries {
			// Cleanup labels
			maps.DeleteFunc(r.Labels, isInternal)

			fields := map[string]interface{}{"value": r.Value.value}
			acc.AddFields(name, fields, r.Labels, r.Value.timestamp)
		}
	case []matrix:
		for _, r := range entries {
			// Cleanup labels
			maps.DeleteFunc(r.Labels, isInternal)

			for _, v := range r.Values {
				fields := map[string]interface{}{"value": v.value}
				acc.AddFields(name, fields, r.Labels, v.timestamp)
			}
		}
	case []stream:
		for _, r := range entries {
			// Cleanup labels
			maps.DeleteFunc(r.Labels, isInternal)

			for _, v := range r.Lines {
				fields := map[string]interface{}{"message": v.message}
				acc.AddFields(name, fields, r.Labels, v.timestamp)
			}
		}
	default:
		return fmt.Errorf("unknown result type %T", entries)
	}

	return nil
}

func isInternal(label, _ string) bool {
	return strings.HasPrefix(label, "__") && strings.HasSuffix(label, "__")
}
