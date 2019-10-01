package prometheus_remote_write

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func mustNew(
	t require.TestingT,
	name string,
	tags map[string]string,
	fields map[string]interface{},
	tm time.Time,
	tp ...telegraf.ValueType,
) telegraf.Metric {
	m, err := metric.New(name, tags, fields, tm, tp...)
	require.NoError(t, err)
	return m
}

func TestWrite(t *testing.T) {
	for i, tc := range []struct {
		metrics  []telegraf.Metric
		expected prompb.WriteRequest
	}{
		{
			metrics:  []telegraf.Metric{},
			expected: prompb.WriteRequest{},
		},

		{
			metrics: []telegraf.Metric{
				mustNew(t, "foo", map[string]string{"bar": "baz"},
					map[string]interface{}{"blip": 0.0}, time.Unix(0, 0), telegraf.Counter),
			},
			expected: prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{{
					Labels: []prompb.Label{
						{Name: "__name__", Value: "foo_blip"},
						{Name: "bar", Value: "baz"},
					},
					Samples: []prompb.Sample{
						{Timestamp: 0, Value: 0.0},
					},
				}},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var actual prompb.WriteRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				buf, err = snappy.Decode(nil, buf)
				require.NoError(t, err)

				err = proto.Unmarshal(buf, &actual)
				require.NoError(t, err)
			}))
			defer server.Close()

			remote := PrometheusRemoteWrite{
				URL: server.URL,
			}
			err := remote.Write(tc.metrics)
			require.NoError(t, err)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestWrite_WithStringField(t *testing.T) {
	for i, tc := range []struct {
		metrics  []telegraf.Metric
		expected prompb.WriteRequest
	}{
		{
			metrics:  []telegraf.Metric{},
			expected: prompb.WriteRequest{},
		},

		{
			metrics: []telegraf.Metric{
				mustNew(t,
					"foo",
					map[string]string{"bar": "baz"},
					map[string]interface{}{
						"blip": "blop",
						"num":  1,
					},
					time.Unix(0, 0),
					telegraf.Counter,
				),
			},
			expected: prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{{
					Labels: []prompb.Label{
						{Name: "__name__", Value: "foo_num"},
						{Name: "bar", Value: "baz"},
						{Name: "blip", Value: "blop"},
					},
					Samples: []prompb.Sample{
						{Timestamp: 0, Value: 1},
					},
				}},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var actual prompb.WriteRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				buf, err = snappy.Decode(nil, buf)
				require.NoError(t, err)

				err = proto.Unmarshal(buf, &actual)
				require.NoError(t, err)
			}))
			defer server.Close()

			remote := PrometheusRemoteWrite{
				URL: server.URL,
			}
			err := remote.Write(tc.metrics)
			require.NoError(t, err)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestWrite_WithStringFieldWithoutAnyNumericField(t *testing.T) {
	for i, tc := range []struct {
		metrics  []telegraf.Metric
		expected prompb.WriteRequest
	}{
		{
			metrics:  []telegraf.Metric{},
			expected: prompb.WriteRequest{},
		},

		{
			metrics: []telegraf.Metric{
				mustNew(
					t,
					"foo",
					map[string]string{"bar": "baz"},
					map[string]interface{}{"blip": "blop"},
					time.Unix(0, 0),
					telegraf.Counter,
				),
			},
			expected: prompb.WriteRequest{},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var actual prompb.WriteRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				buf, err = snappy.Decode(nil, buf)
				require.NoError(t, err)

				err = proto.Unmarshal(buf, &actual)
				require.NoError(t, err)
			}))
			defer server.Close()

			remote := PrometheusRemoteWrite{
				URL: server.URL,
			}
			err := remote.Write(tc.metrics)
			require.NoError(t, err)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestWriteWithHistogram(t *testing.T) {
	for i, tc := range []struct {
		metrics  []telegraf.Metric
		expected prompb.WriteRequest
	}{
		{
			metrics:  []telegraf.Metric{},
			expected: prompb.WriteRequest{},
		},

		{
			metrics: []telegraf.Metric{
				mustNew(t, "foo", map[string]string{"bar": "baz"},
					map[string]interface{}{"99": 1.0}, time.Unix(0, 0), telegraf.Histogram),
			},
			expected: prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{{
					Labels: []prompb.Label{
						{Name: "__name__", Value: "foo"},
						{Name: "bar", Value: "baz"},
						{Name: "le", Value: "99"},
					},
					Samples: []prompb.Sample{
						{Timestamp: 0, Value: 1.0},
					},
				}},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var actual prompb.WriteRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				buf, err = snappy.Decode(nil, buf)
				require.NoError(t, err)

				err = proto.Unmarshal(buf, &actual)
				require.NoError(t, err)
			}))
			defer server.Close()

			remote := PrometheusRemoteWrite{
				URL: server.URL,
			}
			err := remote.Write(tc.metrics)
			require.NoError(t, err)
			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestWriteWithSummary(t *testing.T) {
	for i, tc := range []struct {
		metrics  []telegraf.Metric
		expected prompb.WriteRequest
	}{
		{
			metrics:  []telegraf.Metric{},
			expected: prompb.WriteRequest{},
		},

		{
			metrics: []telegraf.Metric{
				mustNew(t, "foo", map[string]string{"bar": "baz"},
					map[string]interface{}{"99": 1.0}, time.Unix(0, 0), telegraf.Summary),
			},
			expected: prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{{
					Labels: []prompb.Label{
						{Name: "__name__", Value: "foo"},
						{Name: "bar", Value: "baz"},
						{Name: "quantile", Value: "99"},
					},
					Samples: []prompb.Sample{
						{Timestamp: 0, Value: 1.0},
					},
				}},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var actual prompb.WriteRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				buf, err = snappy.Decode(nil, buf)
				require.NoError(t, err)

				err = proto.Unmarshal(buf, &actual)
				require.NoError(t, err)
			}))
			defer server.Close()

			remote := PrometheusRemoteWrite{
				URL: server.URL,
			}
			err := remote.Write(tc.metrics)
			require.NoError(t, err)
			assert.Equal(t, actual, tc.expected)
		})
	}
}
