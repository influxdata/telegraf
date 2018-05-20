package jti_openconfig_telemetry

import (
	"log"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/oc"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var cfg = &OpenConfigTelemetry{
	Servers:         []string{"127.0.0.1:50051"},
	SampleFrequency: internal.Duration{Duration: time.Second * 2},
}

var data = &telemetry.OpenConfigData{
	Path: "/sensor",
	Kv:   []*telemetry.KeyValue{{Key: "/sensor[tag='tagValue']/intKey", Value: &telemetry.KeyValue_IntValue{IntValue: 10}}},
}

var data_with_prefix = &telemetry.OpenConfigData{
	Path: "/sensor_with_prefix",
	Kv: []*telemetry.KeyValue{{Key: "__prefix__", Value: &telemetry.KeyValue_StrValue{StrValue: "/sensor/prefix/"}},
		{Key: "intKey", Value: &telemetry.KeyValue_IntValue{IntValue: 10}}},
}

var data_with_multiple_tags = &telemetry.OpenConfigData{
	Path: "/sensor_with_multiple_tags",
	Kv: []*telemetry.KeyValue{{Key: "__prefix__", Value: &telemetry.KeyValue_StrValue{StrValue: "/sensor/prefix/"}},
		{Key: "tagKey[tag='tagValue']/boolKey", Value: &telemetry.KeyValue_BoolValue{BoolValue: false}},
		{Key: "intKey", Value: &telemetry.KeyValue_IntValue{IntValue: 10}}},
}

var data_with_string_values = &telemetry.OpenConfigData{
	Path: "/sensor_with_string_values",
	Kv: []*telemetry.KeyValue{{Key: "__prefix__", Value: &telemetry.KeyValue_StrValue{StrValue: "/sensor/prefix/"}},
		{Key: "strKey[tag='tagValue']/strValue", Value: &telemetry.KeyValue_StrValue{StrValue: "10"}}},
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
	err := cfg.Start(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"device":       "127.0.0.1",
		"/sensor/@tag": "tagValue",
		"system_id":    "",
		"path":         "/sensor",
	}

	fields := map[string]interface{}{
		"/sensor/intKey":   int64(10),
		"_sequence":        uint64(0),
		"_timestamp":       uint64(0),
		"_component_id":    uint32(0),
		"_subcomponent_id": uint32(0),
	}

	// Give sometime for gRPC channel to be established
	time.Sleep(2 * time.Second)

	acc.AssertContainsTaggedFields(t, "/sensor", fields, tags)
}

func TestOpenConfigTelemetryDataWithPrefix(t *testing.T) {
	var acc testutil.Accumulator
	cfg.Sensors = []string{"/sensor_with_prefix"}
	err := cfg.Start(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"device":    "127.0.0.1",
		"system_id": "",
		"path":      "/sensor_with_prefix",
	}

	fields := map[string]interface{}{
		"/sensor/prefix/intKey": int64(10),
		"_sequence":             uint64(0),
		"_timestamp":            uint64(0),
		"_component_id":         uint32(0),
		"_subcomponent_id":      uint32(0),
	}

	// Give sometime for gRPC channel to be established
	time.Sleep(2 * time.Second)

	acc.AssertContainsTaggedFields(t, "/sensor_with_prefix", fields, tags)
}

func TestOpenConfigTelemetryDataWithMultipleTags(t *testing.T) {
	var acc testutil.Accumulator
	cfg.Sensors = []string{"/sensor_with_multiple_tags"}
	err := cfg.Start(&acc)
	require.NoError(t, err)

	tags1 := map[string]string{
		"/sensor/prefix/tagKey/@tag": "tagValue",
		"device":                     "127.0.0.1",
		"system_id":                  "",
		"path":                       "/sensor_with_multiple_tags",
	}

	fields1 := map[string]interface{}{
		"/sensor/prefix/tagKey/boolKey": false,
		"_sequence":                     uint64(0),
		"_timestamp":                    uint64(0),
		"_component_id":                 uint32(0),
		"_subcomponent_id":              uint32(0),
	}

	tags2 := map[string]string{
		"device":    "127.0.0.1",
		"system_id": "",
		"path":      "/sensor_with_multiple_tags",
	}

	fields2 := map[string]interface{}{
		"/sensor/prefix/intKey": int64(10),
		"_sequence":             uint64(0),
		"_timestamp":            uint64(0),
		"_component_id":         uint32(0),
		"_subcomponent_id":      uint32(0),
	}

	// Give sometime for gRPC channel to be established
	time.Sleep(2 * time.Second)

	acc.AssertContainsTaggedFields(t, "/sensor_with_multiple_tags", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "/sensor_with_multiple_tags", fields2, tags2)
}

func TestOpenConfigTelemetryDataWithStringValues(t *testing.T) {
	var acc testutil.Accumulator
	cfg.Sensors = []string{"/sensor_with_string_values"}
	err := cfg.Start(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"/sensor/prefix/strKey/@tag": "tagValue",
		"device":                     "127.0.0.1",
		"system_id":                  "",
		"path":                       "/sensor_with_string_values",
	}

	fields := map[string]interface{}{
		"/sensor/prefix/strKey/strValue": "10",
		"_sequence":                      uint64(0),
		"_timestamp":                     uint64(0),
		"_component_id":                  uint32(0),
		"_subcomponent_id":               uint32(0),
	}

	// Give sometime for gRPC channel to be established
	time.Sleep(2 * time.Second)

	acc.AssertContainsTaggedFields(t, "/sensor_with_string_values", fields, tags)
}

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	cfg.Servers = []string{lis.Addr().String()}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	telemetry.RegisterOpenConfigTelemetryServer(grpcServer, newServer())
	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()
	os.Exit(m.Run())
}
