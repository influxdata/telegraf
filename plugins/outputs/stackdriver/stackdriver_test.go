package stackdriver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
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

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var mockMetric mockMetricServer

type mockMetricServer struct {
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

func (s *mockMetricServer) CreateTimeSeries(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
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

func TestMain(m *testing.M) {
	serv := grpc.NewServer()
	monitoringpb.RegisterMetricServiceServer(serv, &mockMetric)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}

	go serv.Serve(lis) //nolint:errcheck // Ignore the returned error as the tests will fail anyway

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(lis.Addr().String(), opt)
	if err != nil {
		log.Fatal(err)
	}
	clientOpt = option.WithGRPCConn(conn)

	os.Exit(m.Run())
}

func TestWrite(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:   "projects/" + "[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    c,
	}

	err = s.Connect()
	require.NoError(t, err)
	err = s.Write(testutil.MockMetrics())
	require.NoError(t, err)

	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Equal(t, "global", request.TimeSeries[0].Resource.Type)
	require.Equal(t, "projects/[PROJECT]", request.TimeSeries[0].Resource.Labels["project_id"])
}

func TestWriteResourceTypeAndLabels(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:      "projects/" + "[PROJECT]",
		Namespace:    "test",
		ResourceType: "foo",
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		Log:    testutil.Logger{},
		client: c,
	}

	err = s.Connect()
	require.NoError(t, err)
	err = s.Write(testutil.MockMetrics())
	require.NoError(t, err)

	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Equal(t, "foo", request.TimeSeries[0].Resource.Type)
	require.Equal(t, "projects/[PROJECT]", request.TimeSeries[0].Resource.Labels["project_id"])
	require.Equal(t, "myvalue", request.TimeSeries[0].Resource.Labels["mylabel"])
}

func TestWriteTagsAsResourceLabels(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:              "projects/" + "[PROJECT]",
		Namespace:            "test",
		ResourceType:         "foo",
		TagsAsResourceLabels: []string{"job_name"},
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		Log:    testutil.Logger{},
		client: c,
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"job_name": "cpu",
				"mytag":    "foo",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		testutil.MustMetric("mem",
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

	require.NoError(t, s.Connect())
	require.NoError(t, s.Write(metrics))
	require.Len(t, mockMetric.reqs, 1)

	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 2)
	for _, ts := range request.TimeSeries {
		switch ts.Metric.Type {
		case "test_mem_value/unknown":
			require.Equal(t, "mem", ts.Resource.Labels["job_name"])
		case "test_cpu_value/unknown":
			require.Equal(t, "cpu", ts.Resource.Labels["job_name"])
		default:
			require.False(t, true, "Unknown metric type")
		}
	}
}

func TestWriteMetricTypesOfficial(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:   "projects/" + "[PROJECT]",
		Namespace: "test",
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		MetricNameFormat: "official",
		MetricCounter:    []string{"mem_c"},
		MetricGauge:      []string{"mem_g"},
		MetricHistogram:  []string{"mem_h"},
		Log:              testutil.Logger{},
		client:           c,
	}
	require.NoError(t, s.Init())

	metrics := []telegraf.Metric{
		testutil.MustMetric("mem_g",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("mem_c",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("mem_h",
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

	require.NoError(t, s.Connect())
	require.NoError(t, s.Write(metrics))
	require.Len(t, mockMetric.reqs, 1)

	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
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
			require.False(t, true, "Unknown metric type", ts.Metric.Type)
		}
	}
}

func TestWriteMetricTypesPath(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:   "projects/" + "[PROJECT]",
		Namespace: "test",
		ResourceLabels: map[string]string{
			"mylabel": "myvalue",
		},
		MetricNameFormat: "path",
		MetricCounter:    []string{"mem_c"},
		MetricGauge:      []string{"mem_g"},
		Log:              testutil.Logger{},
		client:           c,
	}
	require.NoError(t, s.Init())

	metrics := []telegraf.Metric{
		testutil.MustMetric("mem_g",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("mem_c",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(3, 0),
		),
	}

	require.NoError(t, s.Connect())
	require.NoError(t, s.Write(metrics))
	require.Len(t, mockMetric.reqs, 1)

	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 2)
	for _, ts := range request.TimeSeries {
		switch ts.Metric.Type {
		case "custom.googleapis.com/test/mem_c/value":
			require.Equal(t, metricpb.MetricDescriptor_CUMULATIVE, ts.MetricKind)
		case "custom.googleapis.com/test/mem_g/value":
			require.Equal(t, metricpb.MetricDescriptor_GAUGE, ts.MetricKind)
		default:
			require.False(t, true, "Unknown metric type", ts.Metric.Type)
		}
	}
}

func TestWriteAscendingTime(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:   "projects/" + "[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    c,
	}

	// Metrics in descending order of timestamp
	metrics := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(1, 0),
		),
	}

	err = s.Connect()
	require.NoError(t, err)
	err = s.Write(metrics)
	require.NoError(t, err)

	require.Len(t, mockMetric.reqs, 2)
	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 1)
	ts := request.TimeSeries[0]
	require.Len(t, ts.Points, 1)
	require.Equal(t,
		&monitoringpb.TimeInterval{
			EndTime: &timestamppb.Timestamp{
				Seconds: 1,
			},
		}, ts.Points[0].Interval)
	require.Equal(t,
		&monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: int64(43),
			},
		}, ts.Points[0].Value)

	request = mockMetric.reqs[1].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 1)
	ts = request.TimeSeries[0]
	require.Len(t, ts.Points, 1)
	require.Equal(t,
		&monitoringpb.TimeInterval{
			EndTime: &timestamppb.Timestamp{
				Seconds: 2,
			},
		}, ts.Points[0].Interval)
	require.Equal(t, &monitoringpb.TypedValue{
		Value: &monitoringpb.TypedValue_Int64Value{
			Int64Value: int64(42),
		},
	}, ts.Points[0].Value)
}

func TestWriteBatchable(t *testing.T) {
	expectedResponse := &emptypb.Empty{}
	mockMetric.err = nil
	mockMetric.reqs = nil
	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:   "projects/" + "[PROJECT]",
		Namespace: "test",
		Log:       testutil.Logger{},
		client:    c,
	}

	// Metrics in descending order of timestamp
	metrics := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(1, 0),
		),
		testutil.MustMetric("ram",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(4, 0),
		),
		testutil.MustMetric("ram",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(5, 0),
		),
		testutil.MustMetric("ram",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("disk",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("disk",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(1, 0),
		),
	}

	err = s.Connect()
	require.NoError(t, err)
	err = s.Write(metrics)
	require.NoError(t, err)

	require.Len(t, mockMetric.reqs, 5)

	// Request 1 with two time series
	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 2)
	ts := request.TimeSeries[0]
	require.Len(t, ts.Points, 1)
	require.Equal(t,
		&monitoringpb.TimeInterval{
			EndTime: &timestamppb.Timestamp{
				Seconds: 1,
			},
		}, ts.Points[0].Interval)
	require.Equal(t,
		&monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: int64(43),
			},
		}, ts.Points[0].Value)

	ts = request.TimeSeries[1]
	require.Len(t, ts.Points, 1)
	require.Equal(t,
		&monitoringpb.TimeInterval{
			EndTime: &timestamppb.Timestamp{
				Seconds: 1,
			},
		}, ts.Points[0].Interval)
	require.Equal(t,
		&monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: int64(43),
			},
		}, ts.Points[0].Value)

	// Request 2 with 1 time series
	request = mockMetric.reqs[1].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 1)
	require.Len(t, ts.Points, 1)
	require.Equal(t, &monitoringpb.TimeInterval{
		EndTime: &timestamppb.Timestamp{
			Seconds: 2,
		},
	}, request.TimeSeries[0].Points[0].Interval)

	// Request 3 with 1 time series with 1 point
	request = mockMetric.reqs[2].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 3)
	require.Len(t, request.TimeSeries[0].Points, 1)
	require.Len(t, request.TimeSeries[1].Points, 1)
	require.Len(t, request.TimeSeries[2].Points, 1)
	require.Equal(t, &monitoringpb.TimeInterval{
		EndTime: &timestamppb.Timestamp{
			Seconds: 3,
		},
	}, request.TimeSeries[0].Points[0].Interval)

	// Request 4 with 1 time series with 1 point
	request = mockMetric.reqs[3].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 1)
	require.Len(t, request.TimeSeries[0].Points, 1)
	require.Equal(t, &monitoringpb.TimeInterval{
		EndTime: &timestamppb.Timestamp{
			Seconds: 4,
		},
	}, request.TimeSeries[0].Points[0].Interval)

	// Request 5 with 1 time series with 1 point
	request = mockMetric.reqs[4].(*monitoringpb.CreateTimeSeriesRequest)
	require.Len(t, request.TimeSeries, 1)
	require.Len(t, request.TimeSeries[0].Points, 1)
	require.Equal(t, &monitoringpb.TimeInterval{
		EndTime: &timestamppb.Timestamp{
			Seconds: 5,
		},
	}, request.TimeSeries[0].Points[0].Interval)
}

func TestWriteIgnoredErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedErr bool
	}{
		{
			name:        "other errors reported",
			err:         errors.New("unknown"),
			expectedErr: true,
		},
		{
			name:        "invalid argument",
			err:         errors.New("invalid argument"),
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetric.err = tt.err
			mockMetric.reqs = nil

			c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
			if err != nil {
				t.Fatal(err)
			}

			s := &Stackdriver{
				Project:   "projects/" + "[PROJECT]",
				Namespace: "test",
				Log:       testutil.Logger{},
				client:    c,
			}

			require.NoError(t, s.Connect())
			if tt.expectedErr {
				require.Error(t, s.Write(testutil.MockMetrics()))
			} else {
				require.NoError(t, s.Write(testutil.MockMetrics()))
			}
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

	s := &Stackdriver{
		Log: testutil.Logger{},
	}

	labels := s.getStackdriverLabels(tags)
	require.Len(t, labels, QuotaLabelsPerMetricDescriptor)
}

func TestGetStackdriverIntervalEndpoints(t *testing.T) {
	c, err := monitoring.NewMetricClient(t.Context(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	s := &Stackdriver{
		Project:      "projects/" + "[PROJECT]",
		Namespace:    "test",
		Log:          testutil.Logger{},
		client:       c,
		counterCache: NewCounterCache(testutil.Logger{}),
	}

	now := time.Now().UTC()
	later := time.Now().UTC().Add(time.Second * 10)

	// Metrics in descending order of timestamp
	metrics := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			now,
			telegraf.Gauge,
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			later,
			telegraf.Gauge,
		),
		testutil.MustMetric("uptime",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			now,
			telegraf.Counter,
		),
		testutil.MustMetric("uptime",
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
			value, err := s.getStackdriverTypedValue(f.Value)
			require.NoError(t, err)
			require.NotNilf(t, value, "Got nil value for metric %q field %q", m, f)

			metricKind, err := getStackdriverMetricKind(m.Type())
			require.NoErrorf(t, err, "Get kind for metric %q (%T) field %q failed: %v", m.Name(), m.Type(), f, err)

			startTime, endTime := getStackdriverIntervalEndpoints(metricKind, value, m, f, s.counterCache)

			// we only generate startTimes for counters
			if metricKind != metricpb.MetricDescriptor_CUMULATIVE {
				require.Nilf(t, startTime, "startTime for non-counter metric %q (%T) field %q should be nil, was: %v", m.Name(), m.Type(), f, startTime)
			} else {
				if idx%2 == 0 {
					// greaterorequal because we might pass a second boundary while the test is running
					// and new startTimes are backdated 1ms from the endTime.
					require.GreaterOrEqual(t, startTime.AsTime().UTC().Unix(), now.UTC().Unix())
				} else {
					require.GreaterOrEqual(t, startTime.AsTime().UTC().Unix(), later.UTC().Unix())
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

func TestStackdriverTypedValuesSource(t *testing.T) {
	s := &Stackdriver{
		Namespace:        "namespace",
		MetricTypePrefix: "foo",
		MetricDataType:   "source",
	}

	tests := []struct {
		name     string
		key      string
		expected string
		value    any
	}{
		{
			name:     "float",
			key:      "key",
			expected: "*monitoringpb.TypedValue_DoubleValue",
			value:    float64(44.0),
		},
		{
			name:     "int64",
			key:      "key",
			expected: "*monitoringpb.TypedValue_Int64Value",
			value:    int64(46),
		},
		{
			name:     "uint",
			key:      "key",
			expected: "*monitoringpb.TypedValue_Int64Value",
			value:    uint64(46),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typedValue, err := s.getStackdriverTypedValue(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.expected, reflect.TypeOf(typedValue.Value).String())
		})
	}
}

func TestStackdriverTypedValuesInt64(t *testing.T) {
	s := &Stackdriver{
		Namespace:        "namespace",
		MetricTypePrefix: "foo",
		MetricDataType:   "double",
	}

	tests := []struct {
		name     string
		key      string
		expected string
		value    any
	}{
		{
			name:     "int",
			key:      "key",
			expected: "*monitoringpb.TypedValue_DoubleValue",
			value:    42,
		},
		{
			name:     "float",
			key:      "key",
			expected: "*monitoringpb.TypedValue_DoubleValue",
			value:    float64(44.0),
		},
		{
			name:     "int64",
			key:      "key",
			expected: "*monitoringpb.TypedValue_DoubleValue",
			value:    int64(46),
		},
		{
			name:     "uint",
			key:      "key",
			expected: "*monitoringpb.TypedValue_DoubleValue",
			value:    uint64(46),
		},
		{
			name:     "numeric string",
			key:      "key",
			expected: "*monitoringpb.TypedValue_DoubleValue",
			value:    "3.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typedValue, err := s.getStackdriverTypedValue(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.expected, reflect.TypeOf(typedValue.Value).String())
		})
	}
}

func TestStackdriverMetricNamePath(t *testing.T) {
	s := &Stackdriver{
		Namespace:        "namespace",
		MetricTypePrefix: "foo",
		MetricNameFormat: "path",
	}
	m := testutil.MustMetric("uptime",
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

func TestStackdriverMetricNameOfficial(t *testing.T) {
	s := &Stackdriver{
		Namespace:        "namespace",
		MetricTypePrefix: "prometheus.googleapis.com",
		MetricNameFormat: "official",
	}

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
			require.Equal(t, tt.expected, s.generateMetricName(tt.metric, tt.metric.Type(), tt.key))
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
			s := &Stackdriver{
				Namespace:        tt.namespace,
				MetricTypePrefix: tt.prefix,
				MetricNameFormat: tt.format,
			}

			require.Equal(t, tt.expected, s.generateHistogramName(tt.metric))
		})
	}
}

func TestBuildHistogram(t *testing.T) {
	m := testutil.MustMetric(
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
	value, err := buildHistogram(m)
	require.NoError(t, err)

	dist := value.GetDistributionValue()
	require.NotNil(t, dist)
	require.Equal(t, int64(2), dist.Count)
	require.InDelta(t, 0.5, dist.Mean, testutil.DefaultDelta)
	require.Len(t, dist.BucketCounts, 5)
	require.Equal(t, []int64{0, 1, 0, 1, 1}, dist.BucketCounts)
	require.Len(t, dist.BucketOptions.GetExplicitBuckets().Bounds, 4)
	require.Equal(t, []float64{5.0, 10.0, 15.0, 20.0}, dist.BucketOptions.GetExplicitBuckets().Bounds)
}

func TestStackdriverValueInvalid(t *testing.T) {
	s := &Stackdriver{
		MetricDataType: "foobar",
	}
	require.Error(t, s.Init())
}

func TestStackdriverMetricNameInvalid(t *testing.T) {
	s := &Stackdriver{
		MetricNameFormat: "foobar",
	}
	require.Error(t, s.Init())
}
