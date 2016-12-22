package influx

import (
	"bytes"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/influxdb/models"
)

// InfluxParser is an object for Parsing incoming metrics.
type InfluxParser struct {
	// DefaultTags will be added to every parsed metric
	DefaultTags map[string]string
}

func (p *InfluxParser) ParseWithDefaultTime(buf []byte, t time.Time) ([]telegraf.Metric, error) {
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	points, err := models.ParsePointsWithPrecision(buf, t, "n")
	metrics := make([]telegraf.Metric, len(points))
	for i, point := range points {
		for k, v := range p.DefaultTags {
			// only set the default tag if it doesn't already exist:
			if tmp := point.Tags().GetString(k); tmp == "" {
				point.AddTag(k, v)
			}
		}
		// Ignore error here because it's impossible that a model.Point
		// wouldn't parse into client.Point properly
		metrics[i] = telegraf.NewMetricFromPoint(point)
	}
	return metrics, err
}

// Parse returns a slice of Metrics from a text representation of a
// metric (in line-protocol format)
// with each metric separated by newlines. If any metrics fail to parse,
// a non-nil error will be returned in addition to the metrics that parsed
// successfully.
func (p *InfluxParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	return p.ParseWithDefaultTime(buf, time.Now())
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
