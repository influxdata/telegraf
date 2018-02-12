package influx

import (
	"bytes"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// InfluxParser is an object for Parsing incoming metrics.
type InfluxParser struct {
	// DefaultTags will be added to every parsed metric
	DefaultTags map[string]string
}

func (p *InfluxParser) ParseWithDefaultTimePrecision(buf []byte, t time.Time, precision string) ([]telegraf.Metric, error) {
	if !bytes.HasSuffix(buf, []byte("\n")) {
		buf = append(buf, '\n')
	}
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	metrics, err := metric.ParseWithDefaultTimePrecision(buf, t, precision)
	if len(p.DefaultTags) > 0 {
		for _, m := range metrics {
			for k, v := range p.DefaultTags {
				// only set the default tag if it doesn't already exist:
				if !m.HasTag(k) {
					m.AddTag(k, v)
				}
			}
		}
	}
	return metrics, err
}

// Parse returns a slice of Metrics from a text representation of a
// metric (in line-protocol format)
// with each metric separated by newlines. If any metrics fail to parse,
// a non-nil error will be returned in addition to the metrics that parsed
// successfully.
func (p *InfluxParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	return p.ParseWithDefaultTimePrecision(buf, time.Now(), "")
}

func (p *InfluxParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf(
			"Can not parse the line: %s, for data format: influx ", line)
	}

	return metrics[0], nil
}

func (p *InfluxParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}
