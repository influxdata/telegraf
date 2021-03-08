package prometheusremotewrite

import (
	"fmt"
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

type Parser struct {
	DefaultTags map[string]string
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	var err error
	var req prompb.WriteRequest

	if err := proto.Unmarshal(buf, &req); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request body: %s", err)
	}

	now := time.Now()

	for _, ts := range req.Timeseries {
		tags := map[string]string{}
		for key, value := range p.DefaultTags {
			tags[key] = value
		}

		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}

		metricName := tags[model.MetricNameLabel]
		delete(tags, model.MetricNameLabel)

		for _, s := range ts.Samples {
			fields := getNameAndValue(&s, metricName)

			// converting to telegraf metric
			if len(fields) > 0 {
				t := getTimestamp(&s, now)
				metric, err := metric.New("prometheusremotewrite", tags, fields, t)
				if err != nil {
					return nil, fmt.Errorf("unable to convert to telegraf metric: %s", err)
				}
				if err == nil {
					metrics = append(metrics, metric)
				}
			}
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
		return nil, fmt.Errorf("No metrics in line")
	}

	if len(metrics) > 1 {
		return nil, fmt.Errorf("More than one metric in line")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

// Get name and value from metric
func getNameAndValue(s *prompb.Sample, metricName string) map[string]interface{} {
	fields := make(map[string]interface{})
	if !math.IsNaN(s.Value) {
		fields[metricName] = s.Value
	}
	return fields
}

func getTimestamp(s *prompb.Sample, now time.Time) time.Time {
	var t time.Time
	if s.Timestamp > 0 {
		t = time.Unix(0, s.Timestamp*1000000)
	} else {
		t = now
	}
	return t
}
