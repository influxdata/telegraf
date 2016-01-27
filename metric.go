package telegraf

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
)

type Metric interface {
	// Name returns the measurement name of the metric
	Name() string

	// Name returns the tags associated with the metric
	Tags() map[string]string

	// Time return the timestamp for the metric
	Time() time.Time

	// UnixNano returns the unix nano time of the metric
	UnixNano() int64

	// Fields returns the fields for the metric
	Fields() map[string]interface{}

	// String returns a line-protocol string of the metric
	String() string

	// PrecisionString returns a line-protocol string of the metric, at precision
	PrecisionString(precison string) string

	// Point returns a influxdb client.Point object
	Point() *client.Point
}

// metric is a wrapper of the influxdb client.Point struct
type metric struct {
	pt *client.Point
}

// NewMetric returns a metric with the given timestamp. If a timestamp is not
// given, then data is sent to the database without a timestamp, in which case
// the server will assign local time upon reception. NOTE: it is recommended to
// send data with a timestamp.
func NewMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) (Metric, error) {
	var T time.Time
	if len(t) > 0 {
		T = t[0]
	}

	pt, err := client.NewPoint(name, tags, fields, T)
	if err != nil {
		return nil, err
	}
	return &metric{
		pt: pt,
	}, nil
}

// ParseMetrics returns a slice of Metrics from a text representation of a
// metric (in line-protocol format)
// with each metric separated by newlines. If any metrics fail to parse,
// a non-nil error will be returned in addition to the metrics that parsed
// successfully.
func ParseMetrics(buf []byte) ([]Metric, error) {
	points, err := models.ParsePoints(buf)
	metrics := make([]Metric, len(points))
	for i, point := range points {
		// Ignore error here because it's impossible that a model.Point
		// wouldn't parse into client.Point properly
		metrics[i], _ = NewMetric(point.Name(), point.Tags(),
			point.Fields(), point.Time())
	}
	return metrics, err
}

func (m *metric) Name() string {
	return m.pt.Name()
}

func (m *metric) Tags() map[string]string {
	return m.pt.Tags()
}

func (m *metric) Time() time.Time {
	return m.pt.Time()
}

func (m *metric) UnixNano() int64 {
	return m.pt.UnixNano()
}

func (m *metric) Fields() map[string]interface{} {
	return m.pt.Fields()
}

func (m *metric) String() string {
	return m.pt.String()
}

func (m *metric) PrecisionString(precison string) string {
	return m.pt.PrecisionString(precison)
}

func (m *metric) Point() *client.Point {
	return m.pt
}
