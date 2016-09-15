package telegraf

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

// ValueType is an enumeration of metric types that represent a simple value.
type ValueType int

// Possible values for the ValueType enum.
const (
	_ ValueType = iota
	Counter
	Gauge
	Untyped
)

type Metric interface {
	// Name returns the measurement name of the metric
	Name() string

	// Name returns the tags associated with the metric
	Tags() map[string]string

	// Time return the timestamp for the metric
	Time() time.Time

	// Type returns the metric type. Can be either telegraf.Gauge or telegraf.Counter
	Type() ValueType

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

	mType ValueType
}

// NewMetric returns an untyped metric.
func NewMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) (Metric, error) {
	pt, err := client.NewPoint(name, tags, fields, t)
	if err != nil {
		return nil, err
	}
	return &metric{
		pt:    pt,
		mType: Untyped,
	}, nil
}

// NewGaugeMetric returns a gauge metric.
// Gauge metrics should be used when the metric is can arbitrarily go up and
// down. ie, temperature, memory usage, cpu usage, etc.
func NewGaugeMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) (Metric, error) {
	pt, err := client.NewPoint(name, tags, fields, t)
	if err != nil {
		return nil, err
	}
	return &metric{
		pt:    pt,
		mType: Gauge,
	}, nil
}

// NewCounterMetric returns a Counter metric.
// Counter metrics should be used when the metric being created is an
// always-increasing counter. ie, net bytes received, requests served, errors, etc.
func NewCounterMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) (Metric, error) {
	pt, err := client.NewPoint(name, tags, fields, t)
	if err != nil {
		return nil, err
	}
	return &metric{
		pt:    pt,
		mType: Counter,
	}, nil
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

func (m *metric) Type() ValueType {
	return m.mType
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
