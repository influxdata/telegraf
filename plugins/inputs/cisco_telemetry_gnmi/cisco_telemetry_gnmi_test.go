package cisco_telemetry_gnmi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"google.golang.org/grpc"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
)

func TestParsePath(t *testing.T) {
	path := "/foo/bar/bla[shoo=woo][shoop=/woop/]/z"
	parsed := parsePath("theorigin", path, "thetarget")

	assert.Equal(t, parsed.Origin, "theorigin")
	assert.Equal(t, parsed.Target, "thetarget")
	assert.Equal(t, parsed.Element, []string{"foo", "bar", "bla[shoo=woo][shoop=/woop/]", "z"})
	assert.Equal(t, parsed.Elem, []*gnmi.PathElem{{Name: "foo"}, {Name: "bar"},
		{Name: "bla", Key: map[string]string{"shoo": "woo", "shoop": "/woop/"}}, {Name: "z"}})

	parsed = parsePath("", "", "")
	assert.Equal(t, *parsed, gnmi.Path{})
}

type mockGNMIServer struct {
	t        *testing.T
	server   *grpc.Server
	scenario int
}

func (m *mockGNMIServer) Capabilities(context.Context, *gnmi.CapabilityRequest) (*gnmi.CapabilityResponse, error) {
	return nil, nil
}

func (m *mockGNMIServer) Get(context.Context, *gnmi.GetRequest) (*gnmi.GetResponse, error) {
	return nil, nil
}

func (m *mockGNMIServer) Set(context.Context, *gnmi.SetRequest) (*gnmi.SetResponse, error) {
	return nil, nil
}

func (m *mockGNMIServer) Subscribe(server gnmi.GNMI_SubscribeServer) error {
	defer func() {
		if m.scenario >= 0 {
			m.scenario = -1
			time.AfterFunc(100*time.Millisecond, m.server.Stop)
		}
	}()

	metadata, ok := metadata.FromIncomingContext(server.Context())
	assert.Equal(m.t, ok, true)
	assert.Equal(m.t, metadata.Get("username"), []string{"theuser"})
	assert.Equal(m.t, metadata.Get("password"), []string{"thepassword"})

	switch m.scenario {
	case 0:
		return fmt.Errorf("testerror")
	case 1:
		notification := mockGNMINotification()
		server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
		server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}})
		notification.Prefix.Elem[0].Key["foo"] = "bar2"
		notification.Update[0].Path.Elem[1].Key["name"] = "str2"
		notification.Update[0].Val = &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte{'"', '1', '2', '3', '"'}}}
		server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
		return nil
	case 2:
		notification := mockGNMINotification()
		server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
		return nil
	case 3:
		notification := mockGNMINotification()
		notification.Prefix.Elem[0].Key["foo"] = "bar2"
		notification.Update[0].Path.Elem[1].Key["name"] = "str2"
		notification.Update[0].Val = &gnmi.TypedValue{Value: &gnmi.TypedValue_BoolVal{BoolVal: false}}
		server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
		return nil
	default:
		return fmt.Errorf("test not implemented ;)")
	}
}

func TestGNMIError(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:57003")
	server := grpc.NewServer()
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 0, server: server})

	c := &CiscoTelemetryGNMI{Address: "127.0.0.1:57003",
		Username: "theuser", Password: "thepassword",
		Redial: internal.Duration{Duration: 1 * time.Second}}

	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))
	server.Serve(listener)
	c.Stop()

	assert.Contains(t, acc.Errors, errors.New("GNMI subscription aborted: rpc error: code = Unknown desc = testerror"))
}

func mockGNMINotification() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1543236572000000000,
		Prefix: &gnmi.Path{
			Origin: "type",
			Elem: []*gnmi.PathElem{
				{
					Name: "model",
					Key:  map[string]string{"foo": "bar"},
				},
			},
			Target: "subscription",
		},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "some"},
						{
							Name: "path",
							Key:  map[string]string{"name": "str", "uint64": "1234"}},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 5678}},
			},
			{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "other"},
						{Name: "path"},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "foobar"}},
			},
		},
	}
}

func TestGNMIMultiple(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:57004")
	server := grpc.NewServer()
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 1, server: server})

	c := &CiscoTelemetryGNMI{Address: "127.0.0.1:57004",
		Username: "theuser", Password: "thepassword",
		Redial: internal.Duration{Duration: 1 * time.Second},
	}

	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))

	server.Serve(listener)
	c.Stop()

	assert.Empty(t, acc.Errors)

	tags := map[string]string{"source": "127.0.0.1", "foo": "bar"}
	fields := map[string]interface{}{"some/path[name=str][uint64=1234]": int64(5678), "other/path": "foobar"}
	acc.AssertContainsTaggedFields(t, "type:/model", fields, tags)

	tags = map[string]string{"foo": "bar2", "source": "127.0.0.1"}
	fields = map[string]interface{}{"some/path[name=str2][uint64=1234]": "123", "other/path": "foobar"}
	acc.AssertContainsTaggedFields(t, "type:/model", fields, tags)
}

func TestGNMIMultipleRedial(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:57004")
	server := grpc.NewServer()
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 2, server: server})

	c := &CiscoTelemetryGNMI{Address: "127.0.0.1:57004",
		Username: "theuser", Password: "thepassword",
		Redial: internal.Duration{Duration: 200 * time.Millisecond}}

	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))
	server.Serve(listener)

	listener, _ = net.Listen("tcp", "127.0.0.1:57004")
	server = grpc.NewServer()
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 3, server: server})

	server.Serve(listener)
	c.Stop()

	assert.Empty(t, acc.Errors)

	tags := map[string]string{"source": "127.0.0.1", "foo": "bar"}
	fields := map[string]interface{}{"other/path": "foobar", "some/path[name=str][uint64=1234]": int64(5678)}
	acc.AssertContainsTaggedFields(t, "type:/model", fields, tags)

	tags = map[string]string{"foo": "bar2", "source": "127.0.0.1"}
	fields = map[string]interface{}{"some/path[name=str2][uint64=1234]": false, "other/path": "foobar"}
	acc.AssertContainsTaggedFields(t, "type:/model", fields, tags)
}
