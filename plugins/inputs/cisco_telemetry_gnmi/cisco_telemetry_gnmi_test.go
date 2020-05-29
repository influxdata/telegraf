package cisco_telemetry_gnmi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
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

type MockServer struct {
	SubscribeF func(gnmi.GNMI_SubscribeServer) error
	GRPCServer *grpc.Server
}

func (s *MockServer) Capabilities(context.Context, *gnmi.CapabilityRequest) (*gnmi.CapabilityResponse, error) {
	return nil, nil
}

func (s *MockServer) Get(context.Context, *gnmi.GetRequest) (*gnmi.GetResponse, error) {
	return nil, nil
}

func (s *MockServer) Set(context.Context, *gnmi.SetRequest) (*gnmi.SetResponse, error) {
	return nil, nil
}

func (s *MockServer) Subscribe(server gnmi.GNMI_SubscribeServer) error {
	return s.SubscribeF(server)
}

func TestWaitError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmi.GNMI_SubscribeServer) error {
			return fmt.Errorf("testerror")
		},
		GRPCServer: grpcServer,
	}
	gnmi.RegisterGNMIServer(grpcServer, gnmiServer)

	plugin := &CiscoTelemetryGNMI{
		Log:       testutil.Logger{},
		Addresses: []string{listener.Addr().String()},
		Encoding:  "proto",
		Redial:    internal.Duration{Duration: 1 * time.Second},
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
		errors.New("aborted GNMI subscription: rpc error: code = Unknown desc = testerror"))
}

func TestUsernamePassword(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmi.GNMI_SubscribeServer) error {
			metadata, ok := metadata.FromIncomingContext(server.Context())
			if !ok {
				return errors.New("failed to get metadata")
			}

			username := metadata.Get("username")
			if len(username) != 1 || username[0] != "theusername" {
				return errors.New("wrong username")
			}

			password := metadata.Get("password")
			if len(password) != 1 || password[0] != "thepassword" {
				return errors.New("wrong password")
			}

			return errors.New("success")
		},
		GRPCServer: grpcServer,
	}
	gnmi.RegisterGNMIServer(grpcServer, gnmiServer)

	plugin := &CiscoTelemetryGNMI{
		Log:       testutil.Logger{},
		Addresses: []string{listener.Addr().String()},
		Username:  "theusername",
		Password:  "thepassword",
		Encoding:  "proto",
		Redial:    internal.Duration{Duration: 1 * time.Second},
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
		errors.New("aborted GNMI subscription: rpc error: code = Unknown desc = success"))
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

func TestNotification(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *CiscoTelemetryGNMI
		server   *MockServer
		expected []telegraf.Metric
	}{
		{
			name: "multiple metrics",
			plugin: &CiscoTelemetryGNMI{
				Log:      testutil.Logger{},
				Encoding: "proto",
				Redial:   internal.Duration{Duration: 1 * time.Second},
				Subscriptions: []Subscription{
					{
						Name:             "alias",
						Origin:           "type",
						Path:             "/model",
						SubscriptionMode: "sample",
					},
				},
			},
			server: &MockServer{
				SubscribeF: func(server gnmi.GNMI_SubscribeServer) error {
					notification := mockGNMINotification()
					server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
					server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true}})
					notification.Prefix.Elem[0].Key["foo"] = "bar2"
					notification.Update[0].Path.Elem[1].Key["name"] = "str2"
					notification.Update[0].Val = &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: []byte{'"', '1', '2', '3', '"'}}}
					server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
					return nil
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"alias",
					map[string]string{
						"path":   "type:/model",
						"source": "127.0.0.1",
						"foo":    "bar",
						"name":   "str",
						"uint64": "1234",
					},
					map[string]interface{}{
						"some/path": int64(5678),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"alias",
					map[string]string{
						"path":   "type:/model",
						"source": "127.0.0.1",
						"foo":    "bar",
					},
					map[string]interface{}{
						"other/path": "foobar",
						"other/this": "that",
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"alias",
					map[string]string{
						"path":   "type:/model",
						"foo":    "bar2",
						"source": "127.0.0.1",
						"name":   "str2",
						"uint64": "1234",
					},
					map[string]interface{}{
						"some/path": "123",
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"alias",
					map[string]string{
						"path":   "type:/model",
						"source": "127.0.0.1",
						"foo":    "bar2",
					},
					map[string]interface{}{
						"other/path": "foobar",
						"other/this": "that",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "full path field key",
			plugin: &CiscoTelemetryGNMI{
				Log:      testutil.Logger{},
				Encoding: "proto",
				Redial:   internal.Duration{Duration: 1 * time.Second},
				Subscriptions: []Subscription{
					{
						Name:             "PHY_COUNTERS",
						Origin:           "type",
						Path:             "/state/port[port-id=*]/ethernet/oper-speed",
						SubscriptionMode: "sample",
					},
				},
			},
			server: &MockServer{
				SubscribeF: func(server gnmi.GNMI_SubscribeServer) error {
					response := &gnmi.SubscribeResponse{
						Response: &gnmi.SubscribeResponse_Update{
							Update: &gnmi.Notification{
								Timestamp: 1543236572000000000,
								Prefix: &gnmi.Path{
									Origin: "type",
									Elem: []*gnmi.PathElem{
										{
											Name: "state",
										},
										{
											Name: "port",
											Key:  map[string]string{"port-id": "1"},
										},
										{
											Name: "ethernet",
										},
										{
											Name: "oper-speed",
										},
									},
									Target: "subscription",
								},
								Update: []*gnmi.Update{
									{
										Path: &gnmi.Path{},
										Val: &gnmi.TypedValue{
											Value: &gnmi.TypedValue_IntVal{IntVal: 42},
										},
									},
								},
							},
						},
					}
					server.Send(response)
					return nil
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"PHY_COUNTERS",
					map[string]string{
						"path":    "type:/state/port/ethernet/oper-speed",
						"source":  "127.0.0.1",
						"port_id": "1",
					},
					map[string]interface{}{
						"oper_speed": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)

			tt.plugin.Addresses = []string{listener.Addr().String()}

			grpcServer := grpc.NewServer()
			tt.server.GRPCServer = grpcServer
			gnmi.RegisterGNMIServer(grpcServer, tt.server)

			var acc testutil.Accumulator
			err = tt.plugin.Start(&acc)
			require.NoError(t, err)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := grpcServer.Serve(listener)
				require.NoError(t, err)
			}()

			acc.Wait(len(tt.expected))
			tt.plugin.Stop()
			grpcServer.Stop()
			wg.Wait()

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(),
				testutil.IgnoreTime())
		})
	}
}

func TestRedial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := &CiscoTelemetryGNMI{
		Log:       testutil.Logger{},
		Addresses: []string{listener.Addr().String()},
		Encoding:  "proto",
		Redial:    internal.Duration{Duration: 10 * time.Millisecond},
	}

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmi.GNMI_SubscribeServer) error {
			notification := mockGNMINotification()
			server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
			return nil
		},
		GRPCServer: grpcServer,
	}
	gnmi.RegisterGNMIServer(grpcServer, gnmiServer)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve(listener)
		require.NoError(t, err)
	}()

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	acc.Wait(2)
	grpcServer.Stop()
	wg.Wait()

	// Restart GNMI server at the same address
	listener, err = net.Listen("tcp", listener.Addr().String())
	require.NoError(t, err)

	grpcServer = grpc.NewServer()
	gnmiServer = &MockServer{
		SubscribeF: func(server gnmi.GNMI_SubscribeServer) error {
			notification := mockGNMINotification()
			notification.Prefix.Elem[0].Key["foo"] = "bar2"
			notification.Update[0].Path.Elem[1].Key["name"] = "str2"
			notification.Update[0].Val = &gnmi.TypedValue{Value: &gnmi.TypedValue_BoolVal{BoolVal: false}}
			server.Send(&gnmi.SubscribeResponse{Response: &gnmi.SubscribeResponse_Update{Update: notification}})
			return nil
		},
		GRPCServer: grpcServer,
	}
	gnmi.RegisterGNMIServer(grpcServer, gnmiServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve(listener)
		require.NoError(t, err)
	}()

	acc.Wait(4)
	plugin.Stop()
	grpcServer.Stop()
	wg.Wait()
}
