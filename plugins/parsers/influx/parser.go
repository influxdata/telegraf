package influx

import (
	"bytes"
	"fmt"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/influxdb/models"
)

// InfluxParser is an object for Parsing incoming metrics.
type InfluxParser struct {
	// DefaultTags will be added to every parsed metric
	DefaultTags map[string]string
}

// ParseMetrics returns a slice of Metrics from a text representation of a
// metric (in line-protocol format)
// with each metric separated by newlines. If any metrics fail to parse,
// a non-nil error will be returned in addition to the metrics that parsed
// successfully.
func (p *InfluxParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	points, err := models.ParsePoints(buf)
	metrics := make([]telegraf.Metric, len(points))
	for i, point := range points {
		tags := point.Tags()
		for k, v := range p.DefaultTags {
			// Only set tags not in parsed metric
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}
		// Ignore error here because it's impossible that a model.Point
		// wouldn't parse into client.Point properly
		metrics[i], _ = telegraf.NewMetric(point.Name(), tags,
			point.Fields(), point.Time())
	}
	return metrics, err
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
