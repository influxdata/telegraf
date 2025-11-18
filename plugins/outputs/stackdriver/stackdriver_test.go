package stackdriver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/api/distribution"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitValueInvalid(t *testing.T) {
	plugin := &Stackdriver{
		MetricDataType: "foobar",
	}
	require.ErrorContains(t, plugin.Init(), "unrecognized metric data type")
}

func TestInitMetricNameInvalid(t *testing.T) {
	plugin := &Stackdriver{
		MetricNameFormat: "foobar",
	}
	require.ErrorContains(t, plugin.Init(), "unrecognized metric name format")
}

func TestWrite(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup the plugin and inject the client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    client,
	}
	require.NoError(t, plugin.Init())

	// Start the plugin and write a metric
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(testutil.MockMetrics()))

	// Check the result
	require.Len(t, server.reqs, 1)
	request, ok := server.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T", server.reqs[0])

	require.Len(t, request.TimeSeries, 1)
	ts := request.TimeSeries[0]
	require.Equal(t, "global", ts.Resource.Type)
	require.Equal(t, "projects/[PROJECT]", ts.Resource.Labels["project_id"])
}

func TestWriteResourceTypeAndLabels(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup the plugin and inject the client
	plugin := &Stackdriver{
		Project:      "projects/[PROJECT]",
		Namespace:    "test",
		ResourceType: "foo",
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		Log:    testutil.Logger{},
		client: client,
	}
	require.NoError(t, plugin.Init())

	// Start the plugin and write a metric
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(testutil.MockMetrics()))

	// Check the result
	require.Len(t, server.reqs, 1)
	request, ok := server.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T", server.reqs[0])

	require.Len(t, request.TimeSeries, 1)
	ts := request.TimeSeries[0]
	require.Equal(t, "foo", ts.Resource.Type)
	require.Equal(t, "projects/[PROJECT]", ts.Resource.Labels["project_id"])
	require.Equal(t, "myvalue", ts.Resource.Labels["mylabel"])
}

func TestWriteTagsAsResourceLabels(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:              "projects/[PROJECT]",
		Namespace:            "test",
		ResourceType:         "foo",
		TagsAsResourceLabels: []string{"job_name"},
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		Log:    testutil.Logger{},
		client: client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write the metrics
	input := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"job_name": "cpu",
				"mytag":    "foo",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		metric.New(
			"mem",
			map[string]string{
				"job_name": "mem",
				"mytag":    "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	// Check the result
	require.Len(t, server.reqs, 1)
	request, ok := server.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T", server.reqs[0])

	require.Len(t, request.TimeSeries, 2)
	for _, ts := range request.TimeSeries {
		switch ts.Metric.Type {
		case "custom.googleapis.com/test/cpu/value":
			require.Equal(t, "cpu", ts.Resource.Labels["job_name"])
		case "custom.googleapis.com/test/mem/value":
			require.Equal(t, "mem", ts.Resource.Labels["job_name"])
		default:
			require.Failf(t, "Wrong metric type", "Unknown metric type: %v", ts.Metric.Type)
		}
	}
}

func TestWriteMetricTypesOfficial(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		MetricNameFormat: "official",
		MetricCounter:    []string{"mem_c"},
		MetricGauge:      []string{"mem_g"},
		MetricHistogram:  []string{"mem_h"},
		Log:              testutil.Logger{},
		client:           client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write the metrics
	input := []telegraf.Metric{
		metric.New("mem_g",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
		metric.New("mem_c",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
		metric.New("mem_h",
			map[string]string{},
			map[string]interface{}{
				"sum":   1,
				"count": 1,
				"5.0":   0.0,
				"10.0":  0.0,
				"15.0":  1.0,
				"+Inf":  1.0,
			},
			time.Unix(3, 0),
			telegraf.Histogram,
		),
	}
	require.NoError(t, plugin.Write(input))

	// Check the result
	require.Len(t, server.reqs, 1)
	request, ok := server.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T", server.reqs[0])

	require.Len(t, request.TimeSeries, 3)
	for _, ts := range request.TimeSeries {
		switch ts.Metric.Type {
		case "custom.googleapis.com/test_mem_c_value/counter":
			require.Equal(t, metricpb.MetricDescriptor_CUMULATIVE, ts.MetricKind)
		case "custom.googleapis.com/test_mem_g_value/gauge":
			require.Equal(t, metricpb.MetricDescriptor_GAUGE, ts.MetricKind)
		case "custom.googleapis.com/test_mem_h/histogram":
			require.Equal(t, metricpb.MetricDescriptor_CUMULATIVE, ts.MetricKind)
		default:
			require.Failf(t, "Wrong metric type", "Unknown metric type: %v", ts.Metric.Type)
		}
	}
}

func TestWriteMetricTypesPath(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		MetricNameFormat: "path",
		MetricCounter:    []string{"mem_c"},
		MetricGauge:      []string{"mem_g"},
		Log:              testutil.Logger{},
		client:           client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write the metrics
	input := []telegraf.Metric{
		metric.New("mem_g",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
		metric.New("mem_c",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	// Check the result
	require.Len(t, server.reqs, 1)
	request, ok := server.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T", server.reqs[0])

	require.Len(t, request.TimeSeries, 2)
	for _, ts := range request.TimeSeries {
		switch ts.Metric.Type {
		case "custom.googleapis.com/test/mem_c/value":
			require.Equal(t, metricpb.MetricDescriptor_CUMULATIVE, ts.MetricKind)
		case "custom.googleapis.com/test/mem_g/value":
			require.Equal(t, metricpb.MetricDescriptor_GAUGE, ts.MetricKind)
		default:
			require.Failf(t, "Wrong metric type", "Unknown metric type: %v", ts.Metric.Type)
		}
	}
}

func TestWriteAscendingTime(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write the metrics in descending order of timestamp
	input := []telegraf.Metric{
		metric.New("cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		metric.New("cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(1, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	// Check the result
	require.Len(t, server.reqs, 2)
	request, ok := server.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T for first request", server.reqs[0])

	require.Len(t, request.TimeSeries, 1)
	ts := request.TimeSeries[0]
	require.Len(t, ts.Points, 1)
	require.Equal(t, &monitoringpb.TimeInterval{EndTime: &timestamppb.Timestamp{Seconds: 1}}, ts.Points[0].Interval)
	require.Equal(t, &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{Int64Value: int64(43)}}, ts.Points[0].Value)

	request, ok = server.reqs[1].(*monitoringpb.CreateTimeSeriesRequest)
	require.Truef(t, ok, "Invalid request type %T for second request", server.reqs[1])

	require.Len(t, request.TimeSeries, 1)
	ts = request.TimeSeries[0]
	require.Len(t, ts.Points, 1)
	require.Equal(t, &monitoringpb.TimeInterval{EndTime: &timestamppb.Timestamp{Seconds: 2}}, ts.Points[0].Interval)
	require.Equal(t, &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{Int64Value: int64(42)}}, ts.Points[0].Value)
}

func TestWriteBatchable(t *testing.T) {
	// Start the test-server
	server := &mockServer{
		resps: []proto.Message{&emptypb.Empty{}},
	}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write the metrics in descending order of timestamp
	input := []telegraf.Metric{
		metric.New("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		metric.New("cpu",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		metric.New("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 41,
			},
			time.Unix(1, 0),
		),
		metric.New("ram",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 44,
			},
			time.Unix(4, 0),
		),
		metric.New("ram",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 45,
			},
			time.Unix(5, 0),
		),
		metric.New("ram",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		metric.New("disk",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		metric.New("disk",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 41,
			},
			time.Unix(1, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	// Check the result
	expected := [][]*monitoringpb.TimeSeries{
		{
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 1,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(41),
							},
						},
					},
				},
			},
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 1,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(41),
							},
						},
					},
				},
			},
		},
		{
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 2,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(42),
							},
						},
					},
				},
			},
		},
		{
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 3,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(43),
							},
						},
					},
				},
			},
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 3,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(43),
							},
						},
					},
				},
			},
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 3,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(43),
							},
						},
					},
				},
			},
		},
		{
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 4,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(44),
							},
						},
					},
				},
			},
		},
		{
			{
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: 5,
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(45),
							},
						},
					},
				},
			},
		},
	}

	require.Len(t, server.reqs, len(expected))
	for i, r := range server.reqs {
		request, ok := r.(*monitoringpb.CreateTimeSeriesRequest)
		require.Truef(t, ok, "Invalid request type %T for request %d", r, i)

		expectedTS := expected[i]
		require.Lenf(t, request.TimeSeries, len(expectedTS), "Mismatch for number of timeseries in request %d", i)
		for j, ets := range expectedTS {
			ts := request.TimeSeries[j]
			require.Lenf(t, ts.Points, len(ets.Points), "Mismatch for number of point in request %d in timeseries %d", i, j)
			for k, epoint := range ets.Points {
				point := ts.Points[k]
				require.Equalf(t, epoint, point, "Mismatch for request %d in timeseries %d and point %d", i, j, k)
			}
		}
	}
}

func TestWriteIgnoreError(t *testing.T) {
	// Start the test-server
	server := &mockServer{err: errors.New("invalid argument")}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write a metric, the resulting error should be ignored
	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestWritePassthroughErrors(t *testing.T) {
	// Start the test-server
	server := &mockServer{err: errors.New("unknown")}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Write the metrics in descending order of timestamp
	require.ErrorContains(t, plugin.Write(testutil.MockMetrics()), "desc = unknown")
}

func TestIntervalEndpoints(t *testing.T) {
	// Start the test-server
	server := &mockServer{err: errors.New("invalid argument")}
	srv, client := startServer(t, server)
	defer srv.GracefulStop()

	// Setup and start the plugin with the injected client
	plugin := &Stackdriver{
		Project:   "projects/[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    client,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Prepare the ime for the metrics and the margins to check the sent
	// request against. Subtracting a millisecond is required because the first
	// metrics will create a time-range where the upper and lower bound is the
	// same, however such time-ranges cannot be sent to Stackdriver and
	// therefore the plugin slightly lowers the lower bound by one millisecond.
	now := time.Now().UTC()
	earlier := now.Add(-1 * time.Millisecond)
	later := time.Now().UTC().Add(time.Second * 10)

	// Metrics in descending order of timestamp
	metrics := []telegraf.Metric{
		metric.New("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			now,
			telegraf.Gauge,
		),
		metric.New("cpu",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			later,
			telegraf.Gauge,
		),
		metric.New("uptime",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			now,
			telegraf.Counter,
		),
		metric.New("uptime",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			later,
			telegraf.Counter,
		),
	}

	for idx, m := range metrics {
		for _, f := range m.FieldList() {
			value, err := plugin.getStackdriverTypedValue(f.Value)
			require.NoError(t, err)
			require.NotNilf(t, value, "Got nil value for metric %q field %q", m, f)

			metricKind, err := getStackdriverMetricKind(m.Type())
			require.NoErrorf(t, err, "Get kind for metric %q (%T) field %q failed: %v", m.Name(), m.Type(), f, err)

			startTime, endTime := getStackdriverIntervalEndpoints(metricKind, value, m, f, plugin.counterCache)

			// we only generate startTimes for counters
			if metricKind != metricpb.MetricDescriptor_CUMULATIVE {
				require.Nilf(t, startTime, "startTime for non-counter metric %q (%T) field %q should be nil, was: %v", m.Name(), m.Type(), f, startTime)
			} else {
				if idx%2 == 0 {
					// We require greater-or-equal because we might pass a
					// second boundary while the test is running and new start
					// times are backdated 1ms from the end time.
					require.GreaterOrEqual(t, startTime.AsTime().UTC().Unix(), earlier.UTC().Unix())
				} else {
					require.LessOrEqual(t, startTime.AsTime().UTC().Unix(), later.UTC().Unix())
				}
			}

			if idx%2 == 0 {
				require.Equal(t, now, endTime.AsTime())
			} else {
				require.Equal(t, later, endTime.AsTime())
			}
		}
	}
}

func TestTypedValuesSource(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected interface{}
		value    any
	}{
		{
			name:     "float",
			key:      "key",
			expected: &monitoringpb.TypedValue_DoubleValue{},
			value:    float64(44.0),
		},
		{
			name:     "int64",
			key:      "key",
			expected: &monitoringpb.TypedValue_Int64Value{},
			value:    int64(46),
		},
		{
			name:     "uint",
			key:      "key",
			expected: &monitoringpb.TypedValue_Int64Value{},
			value:    uint64(46),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Stackdriver{
				Namespace:        "namespace",
				MetricTypePrefix: "foo",
				MetricDataType:   "source",
			}

			value, err := plugin.getStackdriverTypedValue(tt.value)
			require.NoError(t, err)
			require.IsType(t, tt.expected, value.Value)
		})
	}
}

func TestTypedValuesInt64(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected interface{}
		value    any
	}{
		{
			name:     "int",
			key:      "key",
			expected: &monitoringpb.TypedValue_DoubleValue{},
			value:    42,
		},
		{
			name:     "float",
			key:      "key",
			expected: &monitoringpb.TypedValue_DoubleValue{},
			value:    float64(44.0),
		},
		{
			name:     "int64",
			key:      "key",
			expected: &monitoringpb.TypedValue_DoubleValue{},
			value:    int64(46),
		},
		{
			name:     "uint",
			key:      "key",
			expected: &monitoringpb.TypedValue_DoubleValue{},
			value:    uint64(46),
		},
		{
			name:     "numeric string",
			key:      "key",
			expected: &monitoringpb.TypedValue_DoubleValue{},
			value:    "3.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Stackdriver{
				Namespace:        "namespace",
				MetricTypePrefix: "foo",
				MetricDataType:   "double",
			}
			value, err := plugin.getStackdriverTypedValue(tt.value)
			require.NoError(t, err)
			require.IsType(t, tt.expected, value.Value)
		})
	}
}

func TestMetricNamePath(t *testing.T) {
	s := &Stackdriver{
		Namespace:        "namespace",
		MetricTypePrefix: "foo",
		MetricNameFormat: "path",
	}
	m := metric.New("uptime",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 42,
		},
		time.Now(),
		telegraf.Gauge,
	)
	require.Equal(t, "foo/namespace/uptime/key", s.generateMetricName(m, m.Type(), "key"))
}

func TestMetricNameOfficial(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
		metric   telegraf.Metric
	}{
		{
			name:     "gauge",
			key:      "key",
			expected: "prometheus.googleapis.com/namespace_uptime_key/gauge",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Now(),
				telegraf.Gauge,
			),
		},
		{
			name:     "untyped",
			key:      "key",
			expected: "prometheus.googleapis.com/namespace_uptime_key/unknown",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Now(),
				telegraf.Untyped,
			),
		},
		{
			name:     "histogram",
			key:      "key",
			expected: "prometheus.googleapis.com/namespace_uptime_key/histogram",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Now(),
				telegraf.Histogram,
			),
		},
		{
			name:     "counter",
			key:      "key",
			expected: "prometheus.googleapis.com/namespace_uptime_key/counter",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Now(),
				telegraf.Counter,
			),
		},
		{
			name:     "summary",
			key:      "key",
			expected: "prometheus.googleapis.com/namespace_uptime_key",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Now(),
				telegraf.Summary,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Stackdriver{
				Namespace:        "namespace",
				MetricTypePrefix: "prometheus.googleapis.com",
				MetricNameFormat: "official",
			}

			name := plugin.generateMetricName(tt.metric, tt.metric.Type(), tt.key)
			require.Equal(t, tt.expected, name)
		})
	}
}

func TestGenerateHistogramName(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		namespace string
		format    string
		expected  string

		metric telegraf.Metric
	}{
		{
			name:      "path",
			prefix:    "",
			namespace: "",
			format:    "path",
			expected:  "uptime",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{"value": 42},
				time.Now(),
				telegraf.Histogram,
			),
		},
		{
			name:      "path with namespace",
			prefix:    "",
			namespace: "name",
			format:    "path",
			expected:  "name/uptime",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{"value": 42},
				time.Now(),
				telegraf.Histogram,
			),
		},
		{
			name:      "path with namespace+prefix",
			prefix:    "prefix",
			namespace: "name",
			format:    "path",
			expected:  "prefix/name/uptime",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{"value": 42},
				time.Now(),
				telegraf.Histogram,
			),
		},
		{
			name:      "official",
			prefix:    "",
			namespace: "",
			format:    "official",
			expected:  "uptime/histogram",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{"value": 42},
				time.Now(),
				telegraf.Histogram,
			),
		},
		{
			name:      "official with namespace",
			prefix:    "",
			namespace: "name",
			format:    "official",
			expected:  "name_uptime/histogram",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{"value": 42},
				time.Now(),
				telegraf.Histogram,
			),
		},
		{
			name:      "official with prefix+namespace",
			prefix:    "prefix",
			namespace: "name",
			format:    "official",
			expected:  "prefix/name_uptime/histogram",
			metric: metric.New(
				"uptime",
				map[string]string{},
				map[string]interface{}{"value": 42},
				time.Now(),
				telegraf.Histogram,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Stackdriver{
				Namespace:        tt.namespace,
				MetricTypePrefix: tt.prefix,
				MetricNameFormat: tt.format,
			}

			name := plugin.generateHistogramName(tt.metric)
			require.Equal(t, tt.expected, name)
		})
	}
}

func TestGetStackdriverLabels(t *testing.T) {
	tags := []*telegraf.Tag{
		{Key: "project", Value: "bar"},
		{Key: "discuss", Value: "revolutionary"},
		{Key: "marble", Value: "discount"},
		{Key: "applied", Value: "falsify"},
		{Key: "test", Value: "foo"},
		{Key: "porter", Value: "discount"},
		{Key: "play", Value: "tiger"},
		{Key: "fireplace", Value: "display"},
		{Key: "host", Value: "this"},
		{Key: "name", Value: "bat"},
		{Key: "device", Value: "local"},
		{Key: "foo", Value: "bar"},
		{Key: "hostname", Value: "local"},
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
		{Key: "c", Value: "3"},
		{Key: "d", Value: "4"},
		{Key: "e", Value: "5"},
		{Key: "f", Value: "6"},
		{Key: "g", Value: "7"},
		{Key: "h", Value: "8"},
		{Key: "i", Value: "9"},
		{Key: "j", Value: "10"},
		{Key: "k", Value: "11"},
		{Key: "l", Value: "12"},
		{Key: "m", Value: "13"},
		{Key: "n", Value: "14"},
		{Key: "o", Value: "15"},
		{Key: "p", Value: "16"},
		{Key: "reserve", Value: "publication"},
		{Key: "xpfqacltlmpguimhtjlou2qlmf9uqqwk3teajwlwqkoxtsppbnjksaxvzc1aa973pho9m96gfnl5op8ku7sv93rexyx42qe3zty12ityv", Value: "keyquota"},
		{
			Key: "valuequota",
			Value: "icym5wcpejnhljcvy2vwk15svmhrtueoppwlvix61vlbaeedufn1g6u4jgwjoekwew9s2dboxtgrkiyuircnl8h1lbzntt9gzcf60qunhxurhiz0g2bynzy1v6eyn4ravnde" +
				"iiugobsrsj2bfaguahg4gxn7nx4irwfknunhkk6jdlldevawj8levebjajcrcbeugewd14fa8o34ycfwx2ymalyeqxhfqrsksxnii2deqq6cghrzi6qzwmittkzdtye3imoygqm" +
				"jjshiskvnzz1e4ipd9c6wfor5jsygn1kvcg6jm4clnsl1fnxotbei9xp4swrkjpgursmfmkyvxcgq9hoy435nwnolo3ipnvdlhk6pmlzpdjn6gqi3v9gv7jn5ro2p1t5ufxzfsv" +
				"qq1fyrgoi7gvmttil1banh3cftkph1dcoaqfhl7y0wkvhwwvrmslmmxp1wedyn8bacd7akmjgfwdvcmrymbzvmrzfvq1gs1xnmmg8rsfxci2h6r1ralo3splf4f3bdg4c7cy0yy" +
				"9qbxzxhcmdpwekwc7tdjs8uj6wmofm2aor4hum8nwyfwwlxy3yvsnbjy32oucsrmhcnu6l2i8laujkrhvsr9fcix5jflygznlydbqw5uhw1rg1g5wiihqumwmqgggemzoaivm3u" +
				"t41vjaff4uqtqyuhuwblmuiphfkd7si49vgeeswzg7tpuw0oxmkesgibkcjtev2h9ouxzjs3eb71jffhdacyiuyhuxwvm5bnrjewbm4x2kmhgbirz3eoj7ijgplggdkx5vixufg" +
				"65ont8zi1jabsuxx0vsqgprunwkugqkxg2r7iy6fmgs4lob4dlseinowkst6gp6x1ejreauyzjz7atzm3hbmr5rbynuqp4lxrnhhcbuoun69mavvaaki0bdz5ybmbbbz5qdv0od" +
				"tpjo2aezat5uosjuhzbvic05jlyclikynjgfhencdkz3qcqzbzhnsynj1zdke0sk4zfpvfyryzsxv9pu0qm",
		},
	}

	plugin := &Stackdriver{Log: testutil.Logger{}}
	labels := plugin.getStackdriverLabels(tags)
	require.Len(t, labels, QuotaLabelsPerMetricDescriptor)
}

func TestBuildHistogram(t *testing.T) {
	m := metric.New(
		"http_server_duration",
		map[string]string{},
		map[string]interface{}{
			"sum":   1,
			"count": 2,
			"5.0":   0.0,
			"10.0":  1.0,
			"15.0":  1.0,
			"20.0":  2.0,
			"+Inf":  3.0,
			"foo":   4.0,
		},
		time.Unix(0, 0),
	)

	expected := &distribution.Distribution{
		Count:        2,
		Mean:         0.5,
		BucketCounts: []int64{0, 1, 0, 1, 1},
		BucketOptions: &distribution.Distribution_BucketOptions{
			Options: &distribution.Distribution_BucketOptions_ExplicitBuckets{
				ExplicitBuckets: &distribution.Distribution_BucketOptions_Explicit{
					Bounds: []float64{5.0, 10.0, 15.0, 20.0},
				},
			},
		},
	}

	value, err := buildHistogram(m)
	require.NoError(t, err)
	require.Equal(t, expected, value.GetDistributionValue())
}

func startServer(t *testing.T, mock *mockServer) (*grpc.Server, *monitoring.MetricClient) {
	t.Helper()

	// Setup the server
	server := grpc.NewServer()
	monitoringpb.RegisterMetricServiceServer(server, mock)

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "creating listener")

	//nolint:errcheck // Ignore the returned error as the tests will fail anyway
	go server.Serve(listener)

	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "starting GRPC client")
	options := option.WithGRPCConn(conn)
	client, err := monitoring.NewMetricClient(t.Context(), options)
	require.NoError(t, err, "starting monitoring client")

	return server, client
}

type mockServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	monitoringpb.MetricServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockServer) CreateTimeSeries(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}

	s.reqs = append(s.reqs, req)
	if s.err != nil {
		var statusResp *status.Status
		switch s.err.Error() {
		case "invalid argument":
			statusResp = status.New(codes.InvalidArgument, s.err.Error())
		default:
			statusResp = status.New(codes.Unknown, s.err.Error())
		}

		return nil, statusResp.Err()
	}
	return s.resps[0].(*emptypb.Empty), nil
}
