package otlp

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/stretchr/testify/require"
	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type metricServiceServer struct {
	status *status.Status
	metricsService.UnimplementedMetricsServiceServer
}

func (s *metricServiceServer) Export(ctx context.Context, req *metricsService.ExportMetricsServiceRequest) (*metricsService.ExportMetricsServiceResponse, error) {
	var emptyValue = metricsService.ExportMetricsServiceResponse{}

	if s.status == nil {
		return &emptyValue, nil
	}

	return nil, s.status.Err()
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

var listener net.Listener

func TestMain(m *testing.M) {
	listener = newLocalListener()
	grpcServer := grpc.NewServer()
	metricsService.RegisterMetricsServiceServer(grpcServer, &metricServiceServer{
		status: nil,
	})
	go grpcServer.Serve(listener)
	defer grpcServer.Stop()
	os.Exit(m.Run())
}
func TestConfigOptions(t *testing.T) {
	o := OTLP{
		Endpoint: ":::::",
	}
	err := o.Connect()
	require.EqualError(t, err, "invalid endpoint configured")

	o = OTLP{
		Timeout: "9zzz",
	}
	err = o.Connect()
	require.EqualError(t, err, "invalid timeout configured")

	o = OTLP{
		Endpoint: "http://" + listener.Addr().String(),
	}
	err = o.Connect()
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, o.grpcTimeout, defaultTimeout)
	require.Equal(t, o.Headers, map[string]string{"telemetry-reporting-agent": fmt.Sprint(
		"telegraf/",
		internal.Version(),
	)})

	attributes := map[string]string{
		"service.name":    "test",
		"service.version": "0.0.1",
	}
	o = OTLP{
		Endpoint:   "http://" + listener.Addr().String(),
		Timeout:    "10s",
		Attributes: attributes,
	}
	err = o.Connect()
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, o.grpcTimeout, time.Second*10)
	require.Equal(t, len(o.resourceTags), 2)
	for _, tag := range o.resourceTags {
		require.Equal(t, attributes[tag.Key], tag.Value)
	}
}

// func TestWrite(t *testing.T) {
// 	expectedResponse := &emptypb.Empty{}
// 	mockMetric.err = nil
// 	mockMetric.reqs = nil
// 	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

// 	c, err := monitoring.NewMetricClient(context.Background(), clientOpt)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	s := &OTLP{
// 		Endpoint:  "localhost:4317",
// 		Namespace: "test",
// 		client:    c,
// 	}

// 	err = s.Connect()
// 	require.NoError(t, err)
// 	err = s.Write(testutil.MockMetrics())
// 	require.NoError(t, err)

// 	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
// 	require.Equal(t, request.TimeSeries[0].Resource.Type, "global")
// 	require.Equal(t, request.TimeSeries[0].Resource.Labels["project_id"], "projects/[PROJECT]")
// }

// func TestWriteResourceLabels(t *testing.T) {
// 	expectedResponse := &emptypb.Empty{}
// 	mockMetric.err = nil
// 	mockMetric.reqs = nil
// 	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

// 	c, err := monitoring.NewMetricClient(context.Background(), clientOpt)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	s := &OTLP{
// 		Endpoint: "localhost:4317",
// 		Headers: map[string]string{
// 			"mylabel": "myvalue",
// 		},
// 		Namespace: "test",
// 		client:    c,
// 	}

// 	err = s.Connect()
// 	require.NoError(t, err)
// 	err = s.Write(testutil.MockMetrics())
// 	require.NoError(t, err)

// 	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
// 	require.Equal(t, request.TimeSeries[0].Resource.Type, "foo")
// 	require.Equal(t, request.TimeSeries[0].Resource.Labels["project_id"], "projects/[PROJECT]")
// 	require.Equal(t, request.TimeSeries[0].Resource.Labels["mylabel"], "myvalue")
// }

// func TestWriteAscendingTime(t *testing.T) {
// 	expectedResponse := &emptypb.Empty{}
// 	mockMetric.err = nil
// 	mockMetric.reqs = nil
// 	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

// 	c, err := monitoring.NewMetricClient(context.Background(), clientOpt)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	s := &OTLP{
// 		Endpoint:  "localhost:4317",
// 		Namespace: "test",
// 		client:    c,
// 	}

// 	// Metrics in descending order of timestamp
// 	metrics := []telegraf.Metric{
// 		testutil.MustMetric("cpu",
// 			map[string]string{},
// 			map[string]interface{}{
// 				"value": 42,
// 			},
// 			time.Unix(2, 0),
// 		),
// 		testutil.MustMetric("cpu",
// 			map[string]string{},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(1, 0),
// 		),
// 	}

// 	err = s.Connect()
// 	require.NoError(t, err)
// 	err = s.Write(metrics)
// 	require.NoError(t, err)

// 	require.Len(t, mockMetric.reqs, 2)
// 	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
// 	require.Len(t, request.TimeSeries, 1)
// 	ts := request.TimeSeries[0]
// 	require.Len(t, ts.Points, 1)
// 	require.Equal(t, ts.Points[0].Interval, &monitoringpb.TimeInterval{
// 		EndTime: &googlepb.Timestamp{
// 			Seconds: 1,
// 		},
// 	})
// 	require.Equal(t, ts.Points[0].Value, &monitoringpb.TypedValue{
// 		Value: &monitoringpb.TypedValue_Int64Value{
// 			Int64Value: int64(43),
// 		},
// 	})

// 	request = mockMetric.reqs[1].(*monitoringpb.CreateTimeSeriesRequest)
// 	require.Len(t, request.TimeSeries, 1)
// 	ts = request.TimeSeries[0]
// 	require.Len(t, ts.Points, 1)
// 	require.Equal(t, ts.Points[0].Interval, &monitoringpb.TimeInterval{
// 		EndTime: &googlepb.Timestamp{
// 			Seconds: 2,
// 		},
// 	})
// 	require.Equal(t, ts.Points[0].Value, &monitoringpb.TypedValue{
// 		Value: &monitoringpb.TypedValue_Int64Value{
// 			Int64Value: int64(42),
// 		},
// 	})
// }

// func TestWriteBatchable(t *testing.T) {
// 	expectedResponse := &emptypb.Empty{}
// 	mockMetric.err = nil
// 	mockMetric.reqs = nil
// 	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

// 	c, err := monitoring.NewMetricClient(context.Background(), clientOpt)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	s := &OTLP{
// 		Endpoint:  "localhost:4317",
// 		Namespace: "test",
// 		client:    c,
// 	}

// 	// Metrics in descending order of timestamp
// 	metrics := []telegraf.Metric{
// 		testutil.MustMetric("cpu",
// 			map[string]string{
// 				"foo": "bar",
// 			},
// 			map[string]interface{}{
// 				"value": 42,
// 			},
// 			time.Unix(2, 0),
// 		),
// 		testutil.MustMetric("cpu",
// 			map[string]string{
// 				"foo": "foo",
// 			},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(3, 0),
// 		),
// 		testutil.MustMetric("cpu",
// 			map[string]string{
// 				"foo": "bar",
// 			},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(1, 0),
// 		),
// 		testutil.MustMetric("ram",
// 			map[string]string{
// 				"foo": "bar",
// 			},
// 			map[string]interface{}{
// 				"value": 42,
// 			},
// 			time.Unix(4, 0),
// 		),
// 		testutil.MustMetric("ram",
// 			map[string]string{
// 				"foo": "foo",
// 			},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(5, 0),
// 		),
// 		testutil.MustMetric("ram",
// 			map[string]string{
// 				"foo": "bar",
// 			},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(3, 0),
// 		),
// 		testutil.MustMetric("disk",
// 			map[string]string{
// 				"foo": "foo",
// 			},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(3, 0),
// 		),
// 		testutil.MustMetric("disk",
// 			map[string]string{
// 				"foo": "bar",
// 			},
// 			map[string]interface{}{
// 				"value": 43,
// 			},
// 			time.Unix(1, 0),
// 		),
// 	}

// 	err = s.Connect()
// 	require.NoError(t, err)
// 	err = s.Write(metrics)
// 	require.NoError(t, err)

// 	require.Len(t, mockMetric.reqs, 2)
// 	request := mockMetric.reqs[0].(*monitoringpb.CreateTimeSeriesRequest)
// 	require.Len(t, request.TimeSeries, 6)
// 	ts := request.TimeSeries[0]
// 	require.Len(t, ts.Points, 1)
// 	require.Equal(t, ts.Points[0].Interval, &monitoringpb.TimeInterval{
// 		EndTime: &googlepb.Timestamp{
// 			Seconds: 3,
// 		},
// 	})
// 	require.Equal(t, ts.Points[0].Value, &monitoringpb.TypedValue{
// 		Value: &monitoringpb.TypedValue_Int64Value{
// 			Int64Value: int64(43),
// 		},
// 	})

// 	ts = request.TimeSeries[1]
// 	require.Len(t, ts.Points, 1)
// 	require.Equal(t, ts.Points[0].Interval, &monitoringpb.TimeInterval{
// 		EndTime: &googlepb.Timestamp{
// 			Seconds: 1,
// 		},
// 	})
// 	require.Equal(t, ts.Points[0].Value, &monitoringpb.TypedValue{
// 		Value: &monitoringpb.TypedValue_Int64Value{
// 			Int64Value: int64(43),
// 		},
// 	})

// 	ts = request.TimeSeries[2]
// 	require.Len(t, ts.Points, 1)
// 	require.Equal(t, ts.Points[0].Interval, &monitoringpb.TimeInterval{
// 		EndTime: &googlepb.Timestamp{
// 			Seconds: 3,
// 		},
// 	})
// 	require.Equal(t, ts.Points[0].Value, &monitoringpb.TypedValue{
// 		Value: &monitoringpb.TypedValue_Int64Value{
// 			Int64Value: int64(43),
// 		},
// 	})

// 	ts = request.TimeSeries[4]
// 	require.Len(t, ts.Points, 1)
// 	require.Equal(t, ts.Points[0].Interval, &monitoringpb.TimeInterval{
// 		EndTime: &googlepb.Timestamp{
// 			Seconds: 5,
// 		},
// 	})
// 	require.Equal(t, ts.Points[0].Value, &monitoringpb.TypedValue{
// 		Value: &monitoringpb.TypedValue_Int64Value{
// 			Int64Value: int64(43),
// 		},
// 	})
// }

// func TestWriteIgnoredErrors(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		err         error
// 		expectedErr bool
// 	}{
// 		{
// 			name: "points too old",
// 			err:  errors.New(errStringPointsTooOld),
// 		},
// 		{
// 			name: "points out of order",
// 			err:  errors.New(errStringPointsOutOfOrder),
// 		},
// 		{
// 			name: "points too frequent",
// 			err:  errors.New(errStringPointsTooFrequent),
// 		},
// 		{
// 			name:        "other errors reported",
// 			err:         errors.New("test"),
// 			expectedErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockMetric.err = tt.err
// 			mockMetric.reqs = nil

// 			c, err := monitoring.NewMetricClient(context.Background(), clientOpt)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			s := &OTLP{
// 				Endpoint:  "localhost:4317",
// 				Namespace: "test",
// 				client:    c,
// 			}

// 			err = s.Connect()
// 			require.NoError(t, err)
// 			err = s.Write(testutil.MockMetrics())
// 			if tt.expectedErr {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestGetStackdriverLabels(t *testing.T) {
// 	tags := []*telegraf.Tag{
// 		{Key: "project", Value: "bar"},
// 		{Key: "discuss", Value: "revolutionary"},
// 		{Key: "marble", Value: "discount"},
// 		{Key: "applied", Value: "falsify"},
// 		{Key: "test", Value: "foo"},
// 		{Key: "porter", Value: "discount"},
// 		{Key: "play", Value: "tiger"},
// 		{Key: "fireplace", Value: "display"},
// 		{Key: "host", Value: "this"},
// 		{Key: "name", Value: "bat"},
// 		{Key: "device", Value: "local"},
// 		{Key: "reserve", Value: "publication"},
// 		{Key: "xpfqacltlmpguimhtjlou2qlmf9uqqwk3teajwlwqkoxtsppbnjksaxvzc1aa973pho9m96gfnl5op8ku7sv93rexyx42qe3zty12ityv", Value: "keyquota"},
// 		{Key: "valuequota", Value: "icym5wcpejnhljcvy2vwk15svmhrtueoppwlvix61vlbaeedufn1g6u4jgwjoekwew9s2dboxtgrkiyuircnl8h1lbzntt9gzcf60qunhxurhiz0g2bynzy1v6eyn4ravndeiiugobsrsj2bfaguahg4gxn7nx4irwfknunhkk6jdlldevawj8levebjajcrcbeugewd14fa8o34ycfwx2ymalyeqxhfqrsksxnii2deqq6cghrzi6qzwmittkzdtye3imoygqmjjshiskvnzz1e4ipd9c6wfor5jsygn1kvcg6jm4clnsl1fnxotbei9xp4swrkjpgursmfmkyvxcgq9hoy435nwnolo3ipnvdlhk6pmlzpdjn6gqi3v9gv7jn5ro2p1t5ufxzfsvqq1fyrgoi7gvmttil1banh3cftkph1dcoaqfhl7y0wkvhwwvrmslmmxp1wedyn8bacd7akmjgfwdvcmrymbzvmrzfvq1gs1xnmmg8rsfxci2h6r1ralo3splf4f3bdg4c7cy0yy9qbxzxhcmdpwekwc7tdjs8uj6wmofm2aor4hum8nwyfwwlxy3yvsnbjy32oucsrmhcnu6l2i8laujkrhvsr9fcix5jflygznlydbqw5uhw1rg1g5wiihqumwmqgggemzoaivm3ut41vjaff4uqtqyuhuwblmuiphfkd7si49vgeeswzg7tpuw0oxmkesgibkcjtev2h9ouxzjs3eb71jffhdacyiuyhuxwvm5bnrjewbm4x2kmhgbirz3eoj7ijgplggdkx5vixufg65ont8zi1jabsuxx0vsqgprunwkugqkxg2r7iy6fmgs4lob4dlseinowkst6gp6x1ejreauyzjz7atzm3hbmr5rbynuqp4lxrnhhcbuoun69mavvaaki0bdz5ybmbbbz5qdv0odtpjo2aezat5uosjuhzbvic05jlyclikynjgfhencdkz3qcqzbzhnsynj1zdke0sk4zfpvfyryzsxv9pu0qm"},
// 	}

// 	labels := getStackdriverLabels(tags)
// 	require.Equal(t, QuotaLabelsPerMetricDescriptor, len(labels))
// }
