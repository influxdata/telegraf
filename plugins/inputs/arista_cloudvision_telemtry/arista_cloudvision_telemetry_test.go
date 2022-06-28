package arista_cloudvision_telemtry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestParsePath(t *testing.T) {
	theSwitches := []string{"leaf1", "leaf2", "spine1", "spine2"}
	path := "/foo/bar/bla[shoo=woo][shoop=/woop/]/z"
	parsed, err := parsePath("theorigin", path, theSwitches)

	require.NoError(t, err)
	require.Equal(t, "theorigin", parsed)
	require.Equal(t, "thetarget", parsed)
	require.Equal(t, []string{"foo", "bar", "bla[shoo=woo][shoop=/woop/]", "z"}, parsed)
	require.Equal(t, []*gnmiLib.PathElem{{Name: "foo"}, {Name: "bar"},
		{Name: "bla", Key: map[string]string{"shoo": "woo", "shoop": "/woop/"}}, {Name: "z"}}, parsed)

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

func TestWaitError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
			return fmt.Errorf("testerror")
		},
		GRPCServer: grpcServer,
	}
	gnmiLib.RegisterGNMIServer(grpcServer, gnmiServer)

	plugin := &CVP{
		Log:        testutil.Logger{},
		Cvpaddress: listener.Addr().String(),
		Encoding:   "proto",
		Redial:     config.Duration(1 * time.Second),
	}

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve(listener)
		require.NoError(t, err)
	}()

	acc.WaitError(1)
	plugin.Stop()
	grpcServer.Stop()
	wg.Wait()

	require.Contains(t, acc.Errors,
		errors.New("aborted gNMI subscription: rpc error: code = Unknown desc = testerror"))
}

func Testwebauthtoken(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
			metadata, ok := metadata.FromIncomingContext(server.Context())
			if !ok {
				return errors.New("failed to get metadata")
			}

			webauthtoken := metadata.Get("webauthtoken")
			if len(webauthtoken) != 1 || webauthtoken[0] != "token123" {
				return errors.New("wrong token")
			}

			return errors.New("success")
		},
		GRPCServer: grpcServer,
	}
	gnmiLib.RegisterGNMIServer(grpcServer, gnmiServer)

	plugin := &CVP{
		Log:        testutil.Logger{},
		Cvpaddress: listener.Addr().String(),
		Cvptoken:   "token123",
		Encoding:   "proto",
		Redial:     config.Duration(1 * time.Second),
	}

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve(listener)
		require.NoError(t, err)
	}()

	acc.WaitError(1)
	plugin.Stop()
	grpcServer.Stop()
	wg.Wait()

	require.Contains(t, acc.Errors,
		errors.New("aborted gNMI subscription: rpc error: code = Unknown desc = success"))
}

func mockGNMINotification() *gnmiLib.Notification {
	return &gnmiLib.Notification{
		Timestamp: 1543236572000000000,
		Prefix: &gnmiLib.Path{
			Origin: "type",
			Elem: []*gnmiLib.PathElem{
				{
					Name: "model",
					Key:  map[string]string{"foo": "bar"},
				},
			},
			Target: "subscription",
		},
		Update: []*gnmiLib.Update{
			{
				Path: &gnmiLib.Path{
					Elem: []*gnmiLib.PathElem{
						{Name: "some"},
						{
							Name: "path",
							Key:  map[string]string{"name": "str", "uint64": "1234"}},
					},
				},
				Val: &gnmiLib.TypedValue{Value: &gnmiLib.TypedValue_IntVal{IntVal: 5678}},
			},
			{
				Path: &gnmiLib.Path{
					Elem: []*gnmiLib.PathElem{
						{Name: "other"},
						{Name: "path"},
					},
				},
				Val: &gnmiLib.TypedValue{Value: &gnmiLib.TypedValue_StringVal{StringVal: "foobar"}},
			},
			{
				Path: &gnmiLib.Path{
					Elem: []*gnmiLib.PathElem{
						{Name: "other"},
						{Name: "this"},
					},
				},
				Val: &gnmiLib.TypedValue{Value: &gnmiLib.TypedValue_StringVal{StringVal: "that"}},
			},
		},
	}
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
