package arista_cloudvision_telemtry

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/telegraf"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestParsePath(t *testing.T) {
	theSwitches := []string{"leaf1", "leaf2", "spine1", "spine2"}
	path := "/interfaces/interface/state/counters"
	parsed, err := parsePath("theorigin", path, theSwitches)

	require.NoError(t, err)
	parsed, err = parsePath("", "/foo[[", theSwitches)
	require.Nil(t, parsed)
	require.Equal(t, errors.New("Invalid gNMI path: /foo[[/"), err)
}

type MockServer struct {
	SubscribeF func(gnmiLib.GNMI_SubscribeServer) error
	GRPCServer *grpc.Server
}

func (s *MockServer) Capabilities(context.Context, *gnmiLib.CapabilityRequest) (*gnmiLib.CapabilityResponse, error) {
	return nil, nil
}

func (s *MockServer) Get(context.Context, *gnmiLib.GetRequest) (*gnmiLib.GetResponse, error) {
	return nil, nil
}

func (s *MockServer) Set(context.Context, *gnmiLib.SetRequest) (*gnmiLib.SetResponse, error) {
	return nil, nil
}

func (s *MockServer) Subscribe(server gnmiLib.GNMI_SubscribeServer) error {
	return s.SubscribeF(server)
}

type MockLogger struct {
	telegraf.Logger
	lastFormat string
	lastArgs   []interface{}
}

func (l *MockLogger) Errorf(format string, args ...interface{}) {
	l.lastFormat = format
	l.lastArgs = args
}

func TestSubscribeResponseError(t *testing.T) {
	me := "mock error message"
	var mc uint32 = 7
	ml := &MockLogger{}
	plugin := &CVP{Log: ml}
	// TODO: FIX SA1019: gnmi.Error is deprecated: Do not use.
	errorResponse := &gnmiLib.SubscribeResponse_Error{Error: &gnmiLib.Error{Message: me, Code: mc}}
	plugin.handleSubscribeResponse("127.0.0.1:0", &gnmiLib.SubscribeResponse{Response: errorResponse})
	require.NotEmpty(t, ml.lastFormat)
	require.Equal(t, []interface{}{mc, me}, ml.lastArgs)
}
