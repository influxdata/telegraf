package cisco_telemetry_gnmi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestParsePath(t *testing.T) {
	path := "/foo/bar/bla[shoo=woo][shoop=/woop/]/z"
	parsed, err := parsePath("theorigin", path, "thetarget")

	assert.Nil(t, err)
	assert.Equal(t, parsed.Origin, "theorigin")
	assert.Equal(t, parsed.Target, "thetarget")
	assert.Equal(t, parsed.Element, []string{"foo", "bar", "bla[shoo=woo][shoop=/woop/]", "z"})
	assert.Equal(t, parsed.Elem, []*gnmi.PathElem{{Name: "foo"}, {Name: "bar"},
		{Name: "bla", Key: map[string]string{"shoo": "woo", "shoop": "/woop/"}}, {Name: "z"}})

	parsed, err = parsePath("", "", "")
	assert.Nil(t, err)
	assert.Equal(t, *parsed, gnmi.Path{})

	parsed, err = parsePath("", "/foo[[", "")
	assert.Nil(t, parsed)
	assert.Equal(t, errors.New("Invalid GNMI path: /foo[[/"), err)
}

type mockGNMIServer struct {
	t        *testing.T
	acc      *testutil.Accumulator
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
	metadata, ok := metadata.FromIncomingContext(server.Context())
	require.Equal(m.t, ok, true)
	require.Equal(m.t, metadata.Get("username"), []string{"theuser"})
	require.Equal(m.t, metadata.Get("password"), []string{"thepassword"})

	// Must read request before sending a response; even though we don't check
	// the request itself currently.
	_, err := server.Recv()
	if err != nil {
		panic(err)
	}

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
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	acc := &testutil.Accumulator{}
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 0, server: server, acc: acc})

	c := &CiscoTelemetryGNMI{Addresses: []string{listener.Addr().String()},
		Username: "theuser", Password: "thepassword", Encoding: "proto",
		Redial: internal.Duration{Duration: 1 * time.Second}}

	require.NoError(t, c.Start(acc))
	go func() {
		err := server.Serve(listener)
		require.NoError(t, err)
	}()
	acc.WaitError(1)
	c.Stop()
	server.Stop()

	require.Contains(t, acc.Errors, errors.New("aborted GNMI subscription: rpc error: code = Unknown desc = testerror"))
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
			{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "other"},
						{Name: "this"},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "that"}},
			},
		},
	}
}

func TestGNMIMultiple(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	acc := &testutil.Accumulator{}
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 1, server: server, acc: acc})

	c := &CiscoTelemetryGNMI{Addresses: []string{listener.Addr().String()},
		Username: "theuser", Password: "thepassword", Encoding: "proto",
		Redial:        internal.Duration{Duration: 1 * time.Second},
		Subscriptions: []Subscription{{Name: "alias", Origin: "type", Path: "/model", SubscriptionMode: "sample"}},
	}

	require.NoError(t, c.Start(acc))
	go func() {
		err := server.Serve(listener)
		require.NoError(t, err)
	}()
	acc.Wait(4)
	c.Stop()
	server.Stop()

	require.Empty(t, acc.Errors)

	tags := map[string]string{"path": "type:/model", "source": "127.0.0.1", "foo": "bar", "name": "str", "uint64": "1234"}
	fields := map[string]interface{}{"some/path": int64(5678)}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:/model", "source": "127.0.0.1", "foo": "bar"}
	fields = map[string]interface{}{"other/path": "foobar", "other/this": "that"}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:/model", "foo": "bar2", "source": "127.0.0.1", "name": "str2", "uint64": "1234"}
	fields = map[string]interface{}{"some/path": "123"}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:/model", "source": "127.0.0.1", "foo": "bar2"}
	fields = map[string]interface{}{"other/path": "foobar", "other/this": "that"}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)
}

func TestGNMIMultipleRedial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	acc := &testutil.Accumulator{}
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 2, server: server, acc: acc})

	c := &CiscoTelemetryGNMI{Addresses: []string{listener.Addr().String()},
		Username: "theuser", Password: "thepassword", Encoding: "proto",
		Redial:        internal.Duration{Duration: 10 * time.Millisecond},
		Subscriptions: []Subscription{{Name: "alias", Origin: "type", Path: "/model", SubscriptionMode: "sample"}},
	}

	require.NoError(t, c.Start(acc))
	go func() {
		err := server.Serve(listener)
		require.NoError(t, err)
	}()
	acc.Wait(2)
	server.Stop()

	listener, _ = net.Listen("tcp", listener.Addr().String())
	server = grpc.NewServer()
	gnmi.RegisterGNMIServer(server, &mockGNMIServer{t: t, scenario: 3, server: server, acc: acc})

	go func() {
		err := server.Serve(listener)
		require.NoError(t, err)
	}()
	acc.Wait(4)
	c.Stop()
	server.Stop()

	tags := map[string]string{"path": "type:/model", "source": "127.0.0.1", "foo": "bar", "name": "str", "uint64": "1234"}
	fields := map[string]interface{}{"some/path": int64(5678)}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:/model", "source": "127.0.0.1", "foo": "bar"}
	fields = map[string]interface{}{"other/path": "foobar", "other/this": "that"}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:/model", "foo": "bar2", "source": "127.0.0.1", "name": "str2", "uint64": "1234"}
	fields = map[string]interface{}{"some/path": false}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:/model", "source": "127.0.0.1", "foo": "bar2"}
	fields = map[string]interface{}{"other/path": "foobar", "other/this": "that"}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)
}
