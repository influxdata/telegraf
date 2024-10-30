package prometheusremotewrite

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	DefaultTags map[string]string
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var err error
	var metrics []telegraf.Metric
	var req prompb.WriteRequest

	if err := req.Unmarshal(buf); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request body: %w", err)
	}

	now := time.Now()

	for _, ts := range req.Timeseries {
		tags := make(map[string]string, len(p.DefaultTags)+len(ts.Labels))
		for key, value := range p.DefaultTags {
			tags[key] = value
		}

		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}

		metricName := tags[model.MetricNameLabel]
		if metricName == "" {
			return nil, fmt.Errorf("metric name %q not found in tag-set or empty", model.MetricNameLabel)
		}
		delete(tags, model.MetricNameLabel)
		t := now
		for _, s := range ts.Samples {
			fields := make(map[string]interface{})
			if !math.IsNaN(s.Value) {
				fields[metricName] = s.Value
			}
			// converting to telegraf metric
			if len(fields) > 0 {
				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
				}
				m := metric.New("prometheus_remote_write", tags, fields, t)
				metrics = append(metrics, m)
			}
		}

		for _, hp := range ts.Histograms {
			if hp.Timestamp > 0 {
				t = time.Unix(0, hp.Timestamp*1000000)
			}

			// instead of parsing into several metrics we should just parse into ONE Telegraf metric
			// ideally, we parse histograms into various fields into a Telegraf metric
			// but for PoC we just marshall the histogram struct into a json string
			serialized, err := json.Marshal(hp)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal histogram: %w", err)
			}
			fields := map[string]any{
				metricName: string(serialized),
			}

			m := metric.New("prometheus_remote_write", tags, fields, t, telegraf.Histogram)
			metrics = append(metrics, m)
		}
	}
	return metrics, err
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, errors.New("no metrics in line")
	}

	if len(metrics) > 1 {
		return nil, errors.New("more than one metric in line")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func init() {
	parsers.Add("prometheusremotewrite",
		func(string) telegraf.Parser {
			return &Parser{}
		})
}
