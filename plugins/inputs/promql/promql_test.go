package promql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/prometheusremotewrite"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitSuccess(t *testing.T) {
	username := config.NewSecret([]byte("john"))
	password := config.NewSecret([]byte("secret"))
	token := config.NewSecret([]byte("a token"))
	defer username.Destroy()
	defer password.Destroy()
	defer token.Destroy()

	tests := []struct {
		name   string
		plugin *PromQL
	}{
		{
			name: "no authentication",
			plugin: &PromQL{
				URL:            "http://localhost:9090",
				InstantQueries: []InstantQuery{{query: query{Query: "prometheus_http_requests_total"}}},
			},
		},
		{
			name: "basic auth without password",
			plugin: &PromQL{
				URL:            "http://localhost:9090",
				Username:       username,
				InstantQueries: []InstantQuery{{query: query{Query: "prometheus_http_requests_total"}}},
			},
		},
		{
			name: "basic auth with password",
			plugin: &PromQL{
				URL:            "http://localhost:9090",
				Username:       username,
				Password:       password,
				InstantQueries: []InstantQuery{{query: query{Query: "prometheus_http_requests_total"}}},
			},
		},
		{
			name: "token auth",
			plugin: &PromQL{
				URL:            "http://localhost:9090",
				Token:          token,
				InstantQueries: []InstantQuery{{query: query{Query: "prometheus_http_requests_total"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())
		})
	}
}

func TestInitFail(t *testing.T) {
	username := config.NewSecret([]byte("john"))
	password := config.NewSecret([]byte("secret"))
	token := config.NewSecret([]byte("a token"))
	defer username.Destroy()
	defer password.Destroy()
	defer token.Destroy()

	tests := []struct {
		name     string
		plugin   *PromQL
		expected string
	}{
		{
			name:     "all empty",
			plugin:   &PromQL{},
			expected: "'url' cannot be empty",
		},
		{
			name: "no queries",
			plugin: &PromQL{
				URL: "http://localhost:9090",
			},
			expected: "no queries configured",
		},
		{
			name: "password without username",
			plugin: &PromQL{
				URL:            "http://localhost:9090",
				Password:       password,
				InstantQueries: []InstantQuery{{query: query{Query: "prometheus_http_requests_total"}}},
			},
			expected: "expecting username for basic authentication",
		},
		{
			name: "basic and token auth",
			plugin: &PromQL{
				URL:            "http://localhost:9090",
				Username:       username,
				Token:          token,
				InstantQueries: []InstantQuery{{query: query{Query: "prometheus_http_requests_total"}}},
			},
			expected: "cannot use both basic and bearer authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestInstantQueries(t *testing.T) {
	ts := int64(1758808909)

	tests := []struct {
		name     string
		data     model.Value
		expected []telegraf.Metric
	}{
		{
			name: "scalar",
			data: &model.Scalar{
				Value:     model.SampleValue(3.14),
				Timestamp: model.TimeFromUnix(ts),
			},
			expected: []telegraf.Metric{
				metric.New(
					"promql",
					map[string]string{},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		/*
			 * NOT SUPPORTED by the Prometheus Go library yet
			 * see https://github.com/prometheus/common/issues/423
			{
				name: "string",
				data: &model.String{
					Value:     "foobar",
					Timestamp: model.TimeFromUnix(ts),
				},
				expected: []telegraf.Metric{
					metric.New(
						"promql",
						map[string]string{},
						map[string]interface{}{"value": "foobar"},
						time.Unix(ts, 0),
						telegraf.Gauge,
					),
				},
			},
		*/
		{
			name: "vector sample",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "vector multiple samples",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "staging",
					},
					Value:     model.SampleValue(23.0),
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "production",
					},
					Value:     model.SampleValue(42.42),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"vector_metric",
					map[string]string{"job": "staging"},
					map[string]interface{}{"value": float64(23.0)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"vector_metric",
					map[string]string{"job": "production"},
					map[string]interface{}{"value": float64(42.42)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "vector histogram",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Timestamp: model.TimeFromUnix(ts),
					Histogram: &model.SampleHistogram{
						Count: 5,
						Sum:   100.0,
						Buckets: model.HistogramBuckets{
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      0.0,
								Upper:      2.0,
								Count:      10.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      2.0,
								Upper:      4.0,
								Count:      20.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      4.0,
								Upper:      6.0,
								Count:      30.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      6.0,
								Upper:      8.0,
								Count:      40.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      8.0,
								Upper:      10.0,
								Count:      100.0,
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
			},
		},
		/*
		 * Mixed vector responses with both sample value AND histograms are not
		 * possible according to https://prometheus.io/docs/prometheus/latest/querying/api/#instant-vectors
		 */
		{
			name: "matrix samples",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "matrix multiple streams",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
					},
				},
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "staging",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
					},
				},
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "production",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "staging"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "production"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "matrix histograms",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Histograms: []model.SampleHistogramPair{
						{
							Timestamp: model.TimeFromUnix(ts),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      10.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      20.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      30.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      40.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      100.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      110.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      120.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      130.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      140.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      190.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 4,
								Sum:   10.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      5.0,
										Count:      210.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      5.0,
										Upper:      10.0,
										Count:      220.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      10.0,
										Upper:      15.0,
										Count:      230.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      15.0,
										Upper:      20.0,
										Count:      240.0,
									},
								},
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(110),
						"4":  float64(120),
						"6":  float64(130),
						"8":  float64(140),
						"10": float64(190),
					},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"5":  float64(210),
						"10": float64(220),
						"15": float64(230),
						"20": float64(240),
					},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Histogram,
				),
			},
		},
		{
			name: "matrix mixed within stream",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "sampling",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
					Histograms: []model.SampleHistogramPair{
						{
							Timestamp: model.TimeFromUnix(ts),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      10.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      20.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      30.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      40.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      100.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      110.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      120.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      130.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      140.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      190.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 4,
								Sum:   10.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      5.0,
										Count:      210.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      5.0,
										Upper:      10.0,
										Count:      220.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      10.0,
										Upper:      15.0,
										Count:      230.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      15.0,
										Upper:      20.0,
										Count:      240.0,
									},
								},
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{
						"2":  float64(110),
						"4":  float64(120),
						"6":  float64(130),
						"8":  float64(140),
						"10": float64(190),
					},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{
						"5":  float64(210),
						"10": float64(220),
						"15": float64(230),
						"20": float64(240),
					},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Histogram,
				),
			},
		},
		{
			name: "matrix mixed streams",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "sampling",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Histograms: []model.SampleHistogramPair{
						{
							Timestamp: model.TimeFromUnix(ts),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      10.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      20.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      30.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      40.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      100.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      110.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      120.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      130.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      140.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      190.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 4,
								Sum:   10.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      5.0,
										Count:      210.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      5.0,
										Upper:      10.0,
										Count:      220.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      10.0,
										Upper:      15.0,
										Count:      230.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      15.0,
										Upper:      20.0,
										Count:      240.0,
									},
								},
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(110),
						"4":  float64(120),
						"6":  float64(130),
						"8":  float64(140),
						"10": float64(190),
					},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"5":  float64(210),
						"10": float64(220),
						"15": float64(230),
						"20": float64(240),
					},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Histogram,
				),
			},
		},
		{
			name: "result without name property",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"job": "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct the response
			response := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": tt.data.Type().String(),
					"result":     tt.data,
				},
			}
			buf, err := json.Marshal(response)
			require.NoError(t, err, "marshalling response")

			// Setup the mocked Prometheus endpoint
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check the expected request properties
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					t.Errorf("Unexpected method %q", r.Method)
					return
				}
				if h := r.Header.Get("User-Agent"); h != internal.ProductToken() {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Unexpected user agent %q", h)
					return
				}
				if h := r.Header.Get("Content-Type"); h != "application/x-www-form-urlencoded" {
					w.WriteHeader(http.StatusUnsupportedMediaType)
					t.Errorf("Unexpected content type %q", h)
					return
				}

				// Only support queries
				if r.URL.Path != "/api/v1/query" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Construct the response and write it
				if _, err := w.Write(buf); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Writing response failed: %v", err)
					return
				}
			}))
			defer server.Close()

			// Setup the plugin and start it
			plugin := &PromQL{
				URL:            server.URL,
				InstantQueries: []InstantQuery{{query: query{Query: "dummy"}}},
				Timeout:        config.Duration(1 * time.Second),
				Log:            testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(nil))
			defer plugin.Stop()

			// Call gather and check for errors and metrics
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors, "found accumulated errors")
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestRangeQueries(t *testing.T) {
	ts := int64(1758808909)

	tests := []struct {
		name     string
		data     model.Value
		expected []telegraf.Metric
	}{
		{
			name: "scalar",
			data: &model.Scalar{
				Value:     model.SampleValue(3.14),
				Timestamp: model.TimeFromUnix(ts),
			},
			expected: []telegraf.Metric{
				metric.New(
					"promql",
					map[string]string{},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		/*
			 * NOT SUPPORTED by the Prometheus Go library yet
			 * see https://github.com/prometheus/common/issues/423
			{
				name: "string",
				data: &model.String{
					Value:     "foobar",
					Timestamp: model.TimeFromUnix(ts),
				},
				expected: []telegraf.Metric{
					metric.New(
						"promql",
						map[string]string{},
						map[string]interface{}{"value": "foobar"},
						time.Unix(ts, 0),
						telegraf.Gauge,
					),
				},
			},
		*/
		{
			name: "vector sample",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "vector multiple samples",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "staging",
					},
					Value:     model.SampleValue(23.0),
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "production",
					},
					Value:     model.SampleValue(42.42),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"vector_metric",
					map[string]string{"job": "staging"},
					map[string]interface{}{"value": float64(23.0)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"vector_metric",
					map[string]string{"job": "production"},
					map[string]interface{}{"value": float64(42.42)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "vector histogram",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Timestamp: model.TimeFromUnix(ts),
					Histogram: &model.SampleHistogram{
						Count: 5,
						Sum:   100.0,
						Buckets: model.HistogramBuckets{
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      0.0,
								Upper:      2.0,
								Count:      10.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      2.0,
								Upper:      4.0,
								Count:      20.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      4.0,
								Upper:      6.0,
								Count:      30.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      6.0,
								Upper:      8.0,
								Count:      40.0,
							},
							&model.HistogramBucket{
								Boundaries: 2,
								Lower:      8.0,
								Upper:      10.0,
								Count:      100.0,
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
			},
		},
		/*
		 * Mixed vector responses with both sample value AND histograms are not
		 * possible according to https://prometheus.io/docs/prometheus/latest/querying/api/#instant-vectors
		 */
		{
			name: "matrix samples",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "matrix multiple streams",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
					},
				},
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "staging",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
					},
				},
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "production",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "staging"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "production"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "matrix histograms",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Histograms: []model.SampleHistogramPair{
						{
							Timestamp: model.TimeFromUnix(ts),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      10.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      20.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      30.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      40.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      100.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      110.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      120.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      130.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      140.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      190.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 4,
								Sum:   10.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      5.0,
										Count:      210.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      5.0,
										Upper:      10.0,
										Count:      220.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      10.0,
										Upper:      15.0,
										Count:      230.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      15.0,
										Upper:      20.0,
										Count:      240.0,
									},
								},
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(110),
						"4":  float64(120),
						"6":  float64(130),
						"8":  float64(140),
						"10": float64(190),
					},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"5":  float64(210),
						"10": float64(220),
						"15": float64(230),
						"20": float64(240),
					},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Histogram,
				),
			},
		},
		{
			name: "matrix mixed within stream",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "sampling",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
					Histograms: []model.SampleHistogramPair{
						{
							Timestamp: model.TimeFromUnix(ts),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      10.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      20.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      30.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      40.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      100.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      110.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      120.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      130.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      140.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      190.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 4,
								Sum:   10.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      5.0,
										Count:      210.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      5.0,
										Upper:      10.0,
										Count:      220.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      10.0,
										Upper:      15.0,
										Count:      230.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      15.0,
										Upper:      20.0,
										Count:      240.0,
									},
								},
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{
						"2":  float64(110),
						"4":  float64(120),
						"6":  float64(130),
						"8":  float64(140),
						"10": float64(190),
					},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{
						"5":  float64(210),
						"10": float64(220),
						"15": float64(230),
						"20": float64(240),
					},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Histogram,
				),
			},
		},
		{
			name: "matrix mixed streams",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "sampling",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Histograms: []model.SampleHistogramPair{
						{
							Timestamp: model.TimeFromUnix(ts),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      10.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      20.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      30.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      40.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      100.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 5,
								Sum:   100.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      2.0,
										Count:      110.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      2.0,
										Upper:      4.0,
										Count:      120.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      4.0,
										Upper:      6.0,
										Count:      130.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      6.0,
										Upper:      8.0,
										Count:      140.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      8.0,
										Upper:      10.0,
										Count:      190.0,
									},
								},
							},
						},
						{
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
							Histogram: &model.SampleHistogram{
								Count: 4,
								Sum:   10.0,
								Buckets: model.HistogramBuckets{
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      0.0,
										Upper:      5.0,
										Count:      210.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      5.0,
										Upper:      10.0,
										Count:      220.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      10.0,
										Upper:      15.0,
										Count:      230.0,
									},
									&model.HistogramBucket{
										Boundaries: 2,
										Lower:      15.0,
										Upper:      20.0,
										Count:      240.0,
									},
								},
							},
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "sampling"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(10),
						"4":  float64(20),
						"6":  float64(30),
						"8":  float64(40),
						"10": float64(100),
					},
					time.Unix(ts, 0),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"2":  float64(110),
						"4":  float64(120),
						"6":  float64(130),
						"8":  float64(140),
						"10": float64(190),
					},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Histogram,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{
						"5":  float64(210),
						"10": float64(220),
						"15": float64(230),
						"20": float64(240),
					},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Histogram,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct the response
			response := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": tt.data.Type().String(),
					"result":     tt.data,
				},
			}
			buf, err := json.Marshal(response)
			require.NoError(t, err, "marshalling response")

			// Setup the mocked Prometheus endpoint
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check the expected request properties
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					t.Errorf("Unexpected method %q", r.Method)
					return
				}
				if h := r.Header.Get("User-Agent"); h != internal.ProductToken() {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Unexpected user agent %q", h)
					return
				}
				if h := r.Header.Get("Content-Type"); h != "application/x-www-form-urlencoded" {
					w.WriteHeader(http.StatusUnsupportedMediaType)
					t.Errorf("Unexpected content type %q", h)
					return
				}

				// Only support queries
				if r.URL.Path != "/api/v1/query_range" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Check the submitted parameters
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Reading request body failed: %v", err)
					return
				}
				params := make(map[string]time.Time, 2)
				for _, e := range strings.Split(string(body), "&") {
					key, value, found := strings.Cut(e, "=")
					if !found {
						w.WriteHeader(http.StatusInternalServerError)
						t.Errorf("Malformed parameter %q", e)
						return
					}
					switch key {
					case "start", "end":
						x, err := strconv.ParseFloat(value, 64)
						if err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							t.Errorf("Parsing %q failed: %v", e, err)
							return
						}
						params[key] = time.Unix(0, int64(x*1e9))
					case "step":
						if value != "60" {
							w.WriteHeader(http.StatusBadRequest)
							t.Errorf("Invalid stepping %q", value)
							return
						}
					case "query":
						if value != "dummy" {
							w.WriteHeader(http.StatusBadRequest)
							t.Errorf("Invalid query %q", value)
							return
						}
					case "timeout":
						if value != "1s" {
							w.WriteHeader(http.StatusBadRequest)
							t.Errorf("Invalid timeout %q", value)
							return
						}
					default:
						w.WriteHeader(http.StatusInternalServerError)
						t.Errorf("Invalid  paramter %q", e)
						return
					}
				}
				if diff := params["end"].Sub(params["start"]); diff != 5*time.Minute {
					w.WriteHeader(http.StatusBadRequest)
					t.Errorf("Invalid time range %v -> %v", params, diff)
					return
				}

				// Construct the response and write it
				if _, err := w.Write(buf); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Writing response failed: %v", err)
					return
				}
			}))
			defer server.Close()

			// Setup the plugin and start it
			plugin := &PromQL{
				URL: server.URL,
				RangeQueries: []RangeQuery{
					{
						query: query{Query: "dummy"},
						Start: config.Duration(6 * time.Minute),
						End:   config.Duration(1 * time.Minute),
						Step:  config.Duration(1 * time.Minute),
					},
				},
				Timeout: config.Duration(1 * time.Second),
				Log:     testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(nil))
			defer plugin.Stop()

			// Call gather and check for errors and metrics
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors, "found accumulated errors")
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestMetricNameOverride(t *testing.T) {
	ts := int64(1758808909)

	tests := []struct {
		name      string
		queryName string
		data      model.Value
		expected  []telegraf.Metric
	}{
		{
			name: "scalar with default query name",
			data: &model.Scalar{
				Value:     model.SampleValue(3.14),
				Timestamp: model.TimeFromUnix(ts),
			},
			expected: []telegraf.Metric{
				metric.New(
					"promql",
					map[string]string{},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:      "scalar with query name",
			queryName: "foobar",
			data: &model.Scalar{
				Value:     model.SampleValue(3.14),
				Timestamp: model.TimeFromUnix(ts),
			},
			expected: []telegraf.Metric{
				metric.New(
					"foobar",
					map[string]string{},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "vector with default query name and result name",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "vector with default query name and no result name",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"job": "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:      "vector with query name and result name",
			queryName: "foobar",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__": model.LabelValue("vector_metric"),
						"job":      "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"vector_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:      "vector with query name and no result name",
			queryName: "foobar",
			data: &model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"job": "testing",
					},
					Value:     model.SampleValue(3.14),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"foobar",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.14)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "matrix with default query name and result name",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name: "matrix with default query name and no result name",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"job": "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"promql",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name:      "matrix with query name and result name",
			queryName: "foobar",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"__name__": model.LabelValue("matrix_metric"),
						"job":      "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"matrix_metric",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
		{
			name:      "matrix with query name and no result name",
			queryName: "foobar",
			data: &model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{
						"job": "testing",
					},
					Values: []model.SamplePair{
						{
							Value:     model.SampleValue(1.1),
							Timestamp: model.TimeFromUnix(ts),
						},
						{
							Value:     model.SampleValue(2.2),
							Timestamp: model.TimeFromUnix(ts).Add(1 * time.Second),
						},
						{
							Value:     model.SampleValue(3.3),
							Timestamp: model.TimeFromUnix(ts).Add(2 * time.Second),
						},
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"foobar",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(1.1)},
					time.Unix(ts, 0),
					telegraf.Gauge,
				),
				metric.New(
					"foobar",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(2.2)},
					time.Unix(ts, 0).Add(1*time.Second),
					telegraf.Gauge,
				),
				metric.New(
					"foobar",
					map[string]string{"job": "testing"},
					map[string]interface{}{"value": float64(3.3)},
					time.Unix(ts, 0).Add(2*time.Second),
					telegraf.Gauge,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct the response
			response := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"resultType": tt.data.Type().String(),
					"result":     tt.data,
				},
			}
			buf, err := json.Marshal(response)
			require.NoError(t, err, "marshalling response")

			// Setup the mocked Prometheus endpoint
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check the expected request properties
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					t.Errorf("Unexpected method %q", r.Method)
					return
				}
				if h := r.Header.Get("User-Agent"); h != internal.ProductToken() {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Unexpected user agent %q", h)
					return
				}
				if h := r.Header.Get("Content-Type"); h != "application/x-www-form-urlencoded" {
					w.WriteHeader(http.StatusUnsupportedMediaType)
					t.Errorf("Unexpected content type %q", h)
					return
				}

				// Only support queries
				if r.URL.Path != "/api/v1/query" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Construct the response and write it
				if _, err := w.Write(buf); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Writing response failed: %v", err)
					return
				}
			}))
			defer server.Close()

			// Setup the plugin and start it
			plugin := &PromQL{
				URL: server.URL,
				InstantQueries: []InstantQuery{
					{
						query: query{
							Query: "dummy",
							Name:  tt.queryName,
						},
					},
				},
				Timeout: config.Duration(1 * time.Second),
				Log:     testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(nil))
			defer plugin.Stop()

			// Call gather and check for errors and metrics
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors, "found accumulated errors")
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond)
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestWarnings(t *testing.T) {
	// Construct the response
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "scalar",
			"result": &model.Scalar{
				Value:     model.SampleValue(3.14),
				Timestamp: model.Now(),
			},
		},
		"warnings": []string{
			"element A is not queryable",
			"node B cannot be scraped",
		},
	}
	buf, err := json.Marshal(response)
	require.NoError(t, err, "marshalling response")

	// Setup the mocked Prometheus endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(buf); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Writing response failed: %v", err)
			return
		}
	}))
	defer server.Close()

	// Setup the plugin and start it
	logger := &testutil.CaptureLogger{Name: "inputs.promql"}
	plugin := &PromQL{
		URL:            server.URL,
		InstantQueries: []InstantQuery{{query: query{Query: "dummy"}}},
		Timeout:        config.Duration(1 * time.Second),
		Log:            logger,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(nil))
	defer plugin.Stop()

	// Define the expected warnings
	expected := []string{
		`W! [inputs.promql] query "dummy" produced warning: element A is not queryable`,
		`W! [inputs.promql] query "dummy" produced warning: node B cannot be scraped`,
	}

	// Call gather and check for errors and metrics
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors, "found accumulated errors")
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= 1
	}, 3*time.Second, 100*time.Millisecond)

	require.Eventually(t, func() bool {
		return len(logger.Warnings()) >= len(expected)
	}, 3*time.Second, 100*time.Millisecond)

	require.ElementsMatch(t, expected, logger.Warnings(), "warnings do not match")
}

func TestIntegrationInstant(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup container with prometheus server
	container := testutil.Container{
		Image:        "prom/prometheus",
		ExposedPorts: []string{"9090"},
		Cmd: []string{
			"--config.file=/etc/prometheus/prometheus.yml",
			"--storage.tsdb.path=/prometheus",
			"--web.enable-remote-write-receiver",
		},
		WaitingFor: wait.ForAll(
			wait.ForMappedPort(nat.Port("9090")),
			wait.ForLog("Server is ready to receive web requests."),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := "http://" + container.Address + ":" + container.Ports["9090"]

	// Define the input and expected metrics structure
	ts := time.Now()
	input := []telegraf.Metric{
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 1440},
			ts.Add(-300*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 890},
			ts.Add(-270*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 550},
			ts.Add(-240*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 340},
			ts.Add(-210*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 210},
			ts.Add(-180*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 130},
			ts.Add(-150*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 80},
			ts.Add(-120*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 50},
			ts.Add(-90*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 30},
			ts.Add(-60*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 20},
			ts.Add(-30*time.Second),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests_total": 10},
			ts,
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests_total": 5},
			ts.Add(-5*time.Minute),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests_total": 2},
			ts.Add(-4*time.Minute),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests_total": 3},
			ts.Add(-3*time.Minute),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests_total": 2},
			ts.Add(-2*time.Minute),
			telegraf.Counter,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests_total": 1},
			ts,
			telegraf.Counter,
		),
	}
	expected := []telegraf.Metric{
		metric.New(
			"test_http_requests_total",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(10)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests_total",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(1)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	// Write the expected metrics to the Prometheus instance using remote-write
	w := newWriter(addr)
	require.NoError(t, w.write(input))

	// Setup the plugin and start it
	plugin := &PromQL{
		URL: addr,
		InstantQueries: []InstantQuery{
			{
				query: query{Query: `test_http_requests_total`},
			},
		},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	require.NoError(t, plugin.Start(nil))
	defer plugin.Stop()

	// Collect the metrics
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors, "found accumulated errors")

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	// Check the returned metric
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestIntegrationRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup container with prometheus server
	container := testutil.Container{
		Image:        "prom/prometheus",
		ExposedPorts: []string{"9090"},
		Cmd: []string{
			"--config.file=/etc/prometheus/prometheus.yml",
			"--storage.tsdb.path=/prometheus",
			"--web.enable-remote-write-receiver",
		},
		WaitingFor: wait.ForAll(
			wait.ForMappedPort(nat.Port("9090")),
			wait.ForLog("Server is ready to receive web requests."),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := "http://" + container.Address + ":" + container.Ports["9090"]

	// Define the input and expected metrics structure
	ts := time.Now()
	input := []telegraf.Metric{
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 1440},
			ts.Add(-300*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 890},
			ts.Add(-270*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 550},
			ts.Add(-240*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 340},
			ts.Add(-210*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 210},
			ts.Add(-180*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 130},
			ts.Add(-150*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 80},
			ts.Add(-120*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 50},
			ts.Add(-90*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 30},
			ts.Add(-60*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 20},
			ts.Add(-30*time.Second),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"requests": 10},
			ts,
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests": 5},
			ts.Add(-5*time.Minute),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests": 2},
			ts.Add(-4*time.Minute),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests": 3},
			ts.Add(-3*time.Minute),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests": 2},
			ts.Add(-2*time.Minute),
			telegraf.Gauge,
		),
		metric.New(
			"test_http",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"requests": 1},
			ts,
			telegraf.Gauge,
		),
	}
	expected := []telegraf.Metric{
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(10)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(30)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(80)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(210)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(550)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "localhost"},
			map[string]interface{}{"value": float64(1440)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(1)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(2)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(2)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(2)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(3)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"test_http_requests",
			map[string]string{"instance": "remote"},
			map[string]interface{}{"value": float64(5)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	// Write the expected metrics to the Prometheus instance using remote-write
	w := newWriter(addr)
	for _, m := range input {
		require.NoError(t, w.writeSingle(m))
	}

	// Setup the plugin and start it
	plugin := &PromQL{
		URL: addr,
		RangeQueries: []RangeQuery{
			{
				query: query{Query: `test_http_requests`},
				Start: config.Duration(6 * time.Minute),
				Step:  config.Duration(1 * time.Minute),
			},
		},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	require.NoError(t, plugin.Start(nil))
	defer plugin.Stop()

	// Collect the metrics
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors, "found accumulated errors")

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	// Check the returned metric
	// We have to ignore the timestamp as it will be relative to the query
	// timing.
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

// Internal implementations
type writer struct {
	addr       string
	serializer *prometheusremotewrite.Serializer
}

func newWriter(addr string) *writer {
	return &writer{
		addr:       strings.TrimRight(addr, "/") + "/api/v1/write",
		serializer: &prometheusremotewrite.Serializer{Log: testutil.Logger{Name: "serializer"}},
	}
}

func (w *writer) write(metrics []telegraf.Metric) error {
	buf, err := w.serializer.SerializeBatch(metrics)
	if err != nil {
		return fmt.Errorf("serializing metrics failed: %w", err)
	}

	return w.send(buf)
}

func (w *writer) writeSingle(m telegraf.Metric) error {
	buf, err := w.serializer.Serialize(m)
	if err != nil {
		return fmt.Errorf("serializing metrics failed: %w", err)
	}

	return w.send(buf)
}

func (w *writer) send(buf []byte) error {
	// Setup HTTP request with the required headers
	req, err := http.NewRequest(http.MethodPost, w.addr, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	req.Header.Add("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("User-Agent", "Telegraf test writer")
	req.Header.Add("X-Prometheus-Remote-Write-Version", "0.1.0")

	// Do the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("received status %s ", resp.Status)
	}

	return nil
}
