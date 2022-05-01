package gnmi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestParsePath(t *testing.T) {
	path := "/foo/bar/bla[shoo=woo][shoop=/woop/]/z"
	parsed, err := parsePath("theorigin", path, "thetarget")

	require.NoError(t, err)
	require.Equal(t, "theorigin", parsed.Origin)
	require.Equal(t, "thetarget", parsed.Target)
	require.Equal(t, []*gnmiLib.PathElem{{Name: "foo"}, {Name: "bar"},
		{Name: "bla", Key: map[string]string{"shoo": "woo", "shoop": "/woop/"}}, {Name: "z"}}, parsed.Elem)

	parsed, err = parsePath("", "", "")
	require.NoError(t, err)
	require.Equal(t, &gnmiLib.Path{}, parsed)

	parsed, err = parsePath("", "/foo[[", "")
	require.Nil(t, parsed)
	require.NotNil(t, err)
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

	plugin := &GNMI{
		Log:       testutil.Logger{},
		Addresses: []string{listener.Addr().String()},
		Encoding:  "proto",
		Redial:    config.Duration(1 * time.Second),
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

func TestUsernamePassword(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
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
	gnmiLib.RegisterGNMIServer(grpcServer, gnmiServer)

	plugin := &GNMI{
		Log:       testutil.Logger{},
		Addresses: []string{listener.Addr().String()},
		Username:  "theusername",
		Password:  "thepassword",
		Encoding:  "proto",
		Redial:    config.Duration(1 * time.Second),
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

func TestNotification(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *GNMI
		server   *MockServer
		expected []telegraf.Metric
	}{
		{
			name: "multiple metrics",
			plugin: &GNMI{
				Log:      testutil.Logger{},
				Encoding: "proto",
				Redial:   config.Duration(1 * time.Second),
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
				SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
					notification := mockGNMINotification()
					err := server.Send(&gnmiLib.SubscribeResponse{Response: &gnmiLib.SubscribeResponse_Update{Update: notification}})
					if err != nil {
						return err
					}
					err = server.Send(&gnmiLib.SubscribeResponse{Response: &gnmiLib.SubscribeResponse_SyncResponse{SyncResponse: true}})
					if err != nil {
						return err
					}
					notification.Prefix.Elem[0].Key["foo"] = "bar2"
					notification.Update[0].Path.Elem[1].Key["name"] = "str2"
					notification.Update[0].Val = &gnmiLib.TypedValue{Value: &gnmiLib.TypedValue_JsonVal{JsonVal: []byte{'"', '1', '2', '3', '"'}}}
					return server.Send(&gnmiLib.SubscribeResponse{Response: &gnmiLib.SubscribeResponse_Update{Update: notification}})
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
			plugin: &GNMI{
				Log:      testutil.Logger{},
				Encoding: "proto",
				Redial:   config.Duration(1 * time.Second),
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
				SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
					response := &gnmiLib.SubscribeResponse{
						Response: &gnmiLib.SubscribeResponse_Update{
							Update: &gnmiLib.Notification{
								Timestamp: 1543236572000000000,
								Prefix: &gnmiLib.Path{
									Origin: "type",
									Elem: []*gnmiLib.PathElem{
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
								Update: []*gnmiLib.Update{
									{
										Path: &gnmiLib.Path{},
										Val: &gnmiLib.TypedValue{
											Value: &gnmiLib.TypedValue_IntVal{IntVal: 42},
										},
									},
								},
							},
						},
					}
					return server.Send(response)
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
		{
			name: "tagged update pair",
			plugin: &GNMI{
				Log:      testutil.Logger{},
				Encoding: "proto",
				Redial:   config.Duration(1 * time.Second),
				Subscriptions: []Subscription{
					{
						Name:             "oc-intf-desc",
						Origin:           "openconfig-interfaces",
						Path:             "/interfaces/interface/state/description",
						SubscriptionMode: "on_change",
						TagOnly:          true,
					},
					{
						Name:             "oc-intf-counters",
						Origin:           "openconfig-interfaces",
						Path:             "/interfaces/interface/state/counters",
						SubscriptionMode: "sample",
					},
				},
			},
			server: &MockServer{
				SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
					tagResponse := &gnmiLib.SubscribeResponse{
						Response: &gnmiLib.SubscribeResponse_Update{
							Update: &gnmiLib.Notification{
								Timestamp: 1543236571000000000,
								Prefix:    &gnmiLib.Path{},
								Update: []*gnmiLib.Update{
									{
										Path: &gnmiLib.Path{
											Origin: "",
											Elem: []*gnmiLib.PathElem{
												{
													Name: "interfaces",
												},
												{
													Name: "interface",
													Key:  map[string]string{"name": "Ethernet1"},
												},
												{
													Name: "state",
												},
												{
													Name: "description",
												},
											},
											Target: "",
										},
										Val: &gnmiLib.TypedValue{
											Value: &gnmiLib.TypedValue_StringVal{StringVal: "foo"},
										},
									},
								},
							},
						},
					}
					if err := server.Send(tagResponse); err != nil {
						return err
					}
					if err := server.Send(&gnmiLib.SubscribeResponse{Response: &gnmiLib.SubscribeResponse_SyncResponse{SyncResponse: true}}); err != nil {
						return err
					}
					taggedResponse := &gnmiLib.SubscribeResponse{
						Response: &gnmiLib.SubscribeResponse_Update{
							Update: &gnmiLib.Notification{
								Timestamp: 1543236572000000000,
								Prefix:    &gnmiLib.Path{},
								Update: []*gnmiLib.Update{
									{
										Path: &gnmiLib.Path{
											Origin: "",
											Elem: []*gnmiLib.PathElem{
												{
													Name: "interfaces",
												},
												{
													Name: "interface",
													Key:  map[string]string{"name": "Ethernet1"},
												},
												{
													Name: "state",
												},
												{
													Name: "counters",
												},
												{
													Name: "in-broadcast-pkts",
												},
											},
											Target: "",
										},
										Val: &gnmiLib.TypedValue{
											Value: &gnmiLib.TypedValue_IntVal{IntVal: 42},
										},
									},
								},
							},
						},
					}
					return server.Send(taggedResponse)
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"oc-intf-counters",
					map[string]string{
						"path":                     "",
						"source":                   "127.0.0.1",
						"name":                     "Ethernet1",
						"oc-intf-desc/description": "foo",
					},
					map[string]interface{}{
						"in_broadcast_pkts": 42,
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
			gnmiLib.RegisterGNMIServer(grpcServer, tt.server)

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
	plugin := &GNMI{Log: ml}
	// TODO: FIX SA1019: gnmi.Error is deprecated: Do not use.
	errorResponse := &gnmiLib.SubscribeResponse_Error{Error: &gnmiLib.Error{Message: me, Code: mc}}
	plugin.handleSubscribeResponse("127.0.0.1:0", &gnmiLib.SubscribeResponse{Response: errorResponse})
	require.NotEmpty(t, ml.lastFormat)
	require.Equal(t, []interface{}{mc, me}, ml.lastArgs)
}

func TestRedial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := &GNMI{
		Log:       testutil.Logger{},
		Addresses: []string{listener.Addr().String()},
		Encoding:  "proto",
		Redial:    config.Duration(10 * time.Millisecond),
	}

	grpcServer := grpc.NewServer()
	gnmiServer := &MockServer{
		SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
			notification := mockGNMINotification()
			return server.Send(&gnmiLib.SubscribeResponse{Response: &gnmiLib.SubscribeResponse_Update{Update: notification}})
		},
		GRPCServer: grpcServer,
	}
	gnmiLib.RegisterGNMIServer(grpcServer, gnmiServer)

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

	// Restart gNMI server at the same address
	listener, err = net.Listen("tcp", listener.Addr().String())
	require.NoError(t, err)

	grpcServer = grpc.NewServer()
	gnmiServer = &MockServer{
		SubscribeF: func(server gnmiLib.GNMI_SubscribeServer) error {
			notification := mockGNMINotification()
			notification.Prefix.Elem[0].Key["foo"] = "bar2"
			notification.Update[0].Path.Elem[1].Key["name"] = "str2"
			notification.Update[0].Val = &gnmiLib.TypedValue{Value: &gnmiLib.TypedValue_BoolVal{BoolVal: false}}
			return server.Send(&gnmiLib.SubscribeResponse{Response: &gnmiLib.SubscribeResponse_Update{Update: notification}})
		},
		GRPCServer: grpcServer,
	}
	gnmiLib.RegisterGNMIServer(grpcServer, gnmiServer)

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
