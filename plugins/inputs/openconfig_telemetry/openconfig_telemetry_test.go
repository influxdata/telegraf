package openconfig_telemetry

import (
	"log"
	"net"
	"os"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/influxdata/telegraf/plugins/inputs/openconfig_telemetry/oc"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var cfg = &OpenConfigTelemetry{
	Server:          "127.0.0.1:50051",
	SampleFrequency: 2000,
}

var data = &telemetry.OpenConfigData{
	Path: "/sensor",
	Kv: []*telemetry.KeyValue{{"testKey", &telemetry.KeyValue_StrValue{"testValue"}},
		{"intKey", &telemetry.KeyValue_IntValue{10}}},
}

var data_with_prefix = &telemetry.OpenConfigData{
	Path: "/sensor_with_prefix",
	Kv: []*telemetry.KeyValue{{"__prefix__", &telemetry.KeyValue_StrValue{"prefixValue"}},
		{"intKey", &telemetry.KeyValue_IntValue{10}}},
}

var data_with_multiple_tags = &telemetry.OpenConfigData{
	Path: "/sensor_with_multiple_tags",
	Kv: []*telemetry.KeyValue{{"__prefix__", &telemetry.KeyValue_StrValue{"prefixValue"}},
		{"strKey", &telemetry.KeyValue_StrValue{"strValue"}},
		{"intKey", &telemetry.KeyValue_IntValue{10}}},
}

var data_with_string_values = &telemetry.OpenConfigData{
	Path: "/sensor_with_string_values",
	Kv: []*telemetry.KeyValue{{"__prefix__", &telemetry.KeyValue_StrValue{"prefixValue"}},
		{"strKey", &telemetry.KeyValue_StrValue{"10"}}},
}

type openConfigTelemetryServer struct {
}

func (s *openConfigTelemetryServer) TelemetrySubscribe(req *telemetry.SubscriptionRequest, stream telemetry.OpenConfigTelemetry_TelemetrySubscribeServer) error {
	path := req.PathList[0].Path
	if path == "/sensor" {
		stream.Send(data)
	} else if path == "/sensor_with_prefix" {
		stream.Send(data_with_prefix)
	} else if path == "/sensor_with_multiple_tags" {
		stream.Send(data_with_multiple_tags)
	} else if path == "/sensor_with_string_values" {
		stream.Send(data_with_string_values)
	}
	return nil
}

func (s *openConfigTelemetryServer) CancelTelemetrySubscription(ctx context.Context, req *telemetry.CancelSubscriptionRequest) (*telemetry.CancelSubscriptionReply, error) {
	return nil, nil
}

func (s *openConfigTelemetryServer) GetTelemetrySubscriptions(ctx context.Context, req *telemetry.GetSubscriptionsRequest) (*telemetry.GetSubscriptionsReply, error) {
	return nil, nil
}

func (s *openConfigTelemetryServer) GetTelemetryOperationalState(ctx context.Context, req *telemetry.GetOperationalStateRequest) (*telemetry.GetOperationalStateReply, error) {
	return nil, nil
}

func (s *openConfigTelemetryServer) GetDataEncodings(ctx context.Context, req *telemetry.DataEncodingRequest) (*telemetry.DataEncodingReply, error) {
	return nil, nil
}

func newServer() *openConfigTelemetryServer {
	s := new(openConfigTelemetryServer)
	return s
}

func TestOpenConfigTelemetryData(t *testing.T) {
	var acc testutil.Accumulator

	cfg.Sensors = []string{"/sensor"}
	err := cfg.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"testKey": "testValue",
		"device":  "127.0.0.1",
	}

	fields := map[string]interface{}{
		"intKey": int64(10),
	}

	acc.AssertContainsTaggedFields(t, "/sensor", fields, tags)
}

func TestOpenConfigTelemetryDataWithPrefix(t *testing.T) {
	var acc testutil.Accumulator
	cfg.Sensors = []string{"/sensor_with_prefix"}
	err := cfg.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"__prefix__": "prefixValue",
		"device":     "127.0.0.1",
	}

	fields := map[string]interface{}{
		"intKey": int64(10),
	}

	acc.AssertContainsTaggedFields(t, "/sensor_with_prefix", fields, tags)
}

func TestOpenConfigTelemetryDataWithMultipleTags(t *testing.T) {
	var acc testutil.Accumulator
	cfg.Sensors = []string{"/sensor_with_multiple_tags"}
	err := cfg.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"__prefix__": "prefixValue",
		"strKey":     "strValue",
		"device":     "127.0.0.1",
	}

	fields := map[string]interface{}{
		"intKey": int64(10),
	}

	acc.AssertContainsTaggedFields(t, "/sensor_with_multiple_tags", fields, tags)
}

func TestOpenConfigTelemetryDataWithStringValues(t *testing.T) {
	var acc testutil.Accumulator
	cfg.Sensors = []string{"/sensor_with_string_values"}
	err := cfg.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"__prefix__": "prefixValue",
		"device":     "127.0.0.1",
	}

	fields := map[string]interface{}{
		"strKey": int64(10),
	}

	acc.AssertContainsTaggedFields(t, "/sensor_with_string_values", fields, tags)
}

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	telemetry.RegisterOpenConfigTelemetryServer(grpcServer, newServer())
	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()
	os.Exit(m.Run())
}
