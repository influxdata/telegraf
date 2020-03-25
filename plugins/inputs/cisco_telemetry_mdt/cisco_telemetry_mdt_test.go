package cisco_telemetry_mdt

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"

	dialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/golang/protobuf/proto"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestHandleTelemetryTwoSimple(t *testing.T) {
	c := &CiscoTelemetryMDT{Log: testutil.Logger{}, Transport: "dummy", Aliases: map[string]string{"alias": "type:model/some/path"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	// error is expected since we are passing in dummy transport
	require.Error(t, err)

	telemetry := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/some/path",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "str"},
							},
							{
								Name:        "uint64",
								ValueByType: &telemetry.TelemetryField_Uint64Value{Uint64Value: 1234},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "bool",
								ValueByType: &telemetry.TelemetryField_BoolValue{BoolValue: true},
							},
						},
					},
				},
			},
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "str2"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "bool",
								ValueByType: &telemetry.TelemetryField_BoolValue{BoolValue: false},
							},
						},
					},
				},
			},
		},
	}
	data, _ := proto.Marshal(telemetry)

	c.handleTelemetry(data)
	require.Empty(t, acc.Errors)

	tags := map[string]string{"path": "type:model/some/path", "name": "str", "uint64": "1234", "source": "hostname", "subscription": "subscription"}
	fields := map[string]interface{}{"bool": true}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)

	tags = map[string]string{"path": "type:model/some/path", "name": "str2", "source": "hostname", "subscription": "subscription"}
	fields = map[string]interface{}{"bool": false}
	acc.AssertContainsTaggedFields(t, "alias", fields, tags)
}

func TestHandleTelemetrySingleNested(t *testing.T) {
	c := &CiscoTelemetryMDT{Log: testutil.Logger{}, Transport: "dummy", Aliases: map[string]string{"nested": "type:model/nested/path"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	// error is expected since we are passing in dummy transport
	require.Error(t, err)

	telemetry := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/nested/path",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "nested",
								Fields: []*telemetry.TelemetryField{
									{
										Name: "key",
										Fields: []*telemetry.TelemetryField{
											{
												Name:        "level",
												ValueByType: &telemetry.TelemetryField_DoubleValue{DoubleValue: 3},
											},
										},
									},
								},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "nested",
								Fields: []*telemetry.TelemetryField{
									{
										Name: "value",
										Fields: []*telemetry.TelemetryField{
											{
												Name:        "foo",
												ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "bar"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	data, _ := proto.Marshal(telemetry)

	c.handleTelemetry(data)
	require.Empty(t, acc.Errors)

	tags := map[string]string{"path": "type:model/nested/path", "level": "3", "source": "hostname", "subscription": "subscription"}
	fields := map[string]interface{}{"nested/value/foo": "bar"}
	acc.AssertContainsTaggedFields(t, "nested", fields, tags)
}

func TestHandleEmbeddedTags(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "dummy", Aliases: map[string]string{"extra": "type:model/extra"}, EmbeddedTags: []string{"type:model/extra/list/name"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	// error is expected since we are passing in dummy transport
	require.Error(t, err)

	telemetry := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/extra",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "foo",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "bar"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "list",
								Fields: []*telemetry.TelemetryField{
									{
										Name:        "name",
										ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "entry1"},
									},
									{
										Name:        "test",
										ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "foo"},
									},
								},
							},
							{
								Name: "list",
								Fields: []*telemetry.TelemetryField{
									{
										Name:        "name",
										ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "entry2"},
									},
									{
										Name:        "test",
										ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "bar"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	data, _ := proto.Marshal(telemetry)

	c.handleTelemetry(data)
	require.Empty(t, acc.Errors)

	tags1 := map[string]string{"path": "type:model/extra", "foo": "bar", "source": "hostname", "subscription": "subscription", "list/name": "entry1"}
	fields1 := map[string]interface{}{"list/test": "foo"}
	tags2 := map[string]string{"path": "type:model/extra", "foo": "bar", "source": "hostname", "subscription": "subscription", "list/name": "entry2"}
	fields2 := map[string]interface{}{"list/test": "bar"}
	acc.AssertContainsTaggedFields(t, "extra", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "extra", fields2, tags2)
}

func TestHandleNXAPI(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "dummy", Aliases: map[string]string{"nxapi": "show nxapi"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	// error is expected since we are passing in dummy transport
	require.Error(t, err)

	telemetry := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "show nxapi",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "foo",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "bar"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Fields: []*telemetry.TelemetryField{
									{
										Name: "TABLE_nxapi",
										Fields: []*telemetry.TelemetryField{
											{
												Fields: []*telemetry.TelemetryField{
													{
														Name: "ROW_nxapi",
														Fields: []*telemetry.TelemetryField{
															{
																Fields: []*telemetry.TelemetryField{
																	{
																		Name:        "index",
																		ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "i1"},
																	},
																	{
																		Name:        "value",
																		ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "foo"},
																	},
																},
															},
															{
																Fields: []*telemetry.TelemetryField{
																	{
																		Name:        "index",
																		ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "i2"},
																	},
																	{
																		Name:        "value",
																		ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "bar"},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	data, _ := proto.Marshal(telemetry)

	c.handleTelemetry(data)
	require.Empty(t, acc.Errors)

	tags1 := map[string]string{"path": "show nxapi", "foo": "bar", "TABLE_nxapi": "i1", "source": "hostname", "subscription": "subscription"}
	fields1 := map[string]interface{}{"value": "foo"}
	tags2 := map[string]string{"path": "show nxapi", "foo": "bar", "TABLE_nxapi": "i2", "source": "hostname", "subscription": "subscription"}
	fields2 := map[string]interface{}{"value": "bar"}
	acc.AssertContainsTaggedFields(t, "nxapi", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "nxapi", fields2, tags2)
}

func TestHandleNXDME(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "dummy", Aliases: map[string]string{"dme": "sys/dme"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	// error is expected since we are passing in dummy transport
	require.Error(t, err)

	telemetry := &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "sys/dme",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "foo",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "bar"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Fields: []*telemetry.TelemetryField{
									{
										Name: "fooEntity",
										Fields: []*telemetry.TelemetryField{
											{
												Fields: []*telemetry.TelemetryField{
													{
														Name: "attributes",
														Fields: []*telemetry.TelemetryField{
															{
																Fields: []*telemetry.TelemetryField{
																	{
																		Name:        "rn",
																		ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "some-rn"},
																	},
																	{
																		Name:        "value",
																		ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "foo"},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	data, _ := proto.Marshal(telemetry)

	c.handleTelemetry(data)
	require.Empty(t, acc.Errors)

	tags1 := map[string]string{"path": "sys/dme", "foo": "bar", "fooEntity": "some-rn", "source": "hostname", "subscription": "subscription"}
	fields1 := map[string]interface{}{"value": "foo"}
	acc.AssertContainsTaggedFields(t, "dme", fields1, tags1)
}

func TestTCPDialoutOverflow(t *testing.T) {
	c := &CiscoTelemetryMDT{Log: testutil.Logger{}, Transport: "tcp", ServiceAddress: "127.0.0.1:0"}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	require.NoError(t, err)

	hdr := struct {
		MsgType       uint16
		MsgEncap      uint16
		MsgHdrVersion uint16
		MsgFlags      uint16
		MsgLen        uint32
	}{MsgLen: uint32(1000000000)}

	addr := c.Address()
	conn, err := net.Dial(addr.Network(), addr.String())
	require.NoError(t, err)
	binary.Write(conn, binary.BigEndian, hdr)
	conn.Read([]byte{0})
	conn.Close()

	c.Stop()

	require.Contains(t, acc.Errors, errors.New("dialout packet too long: 1000000000"))
}

func mockTelemetryMessage() *telemetry.Telemetry {
	return &telemetry.Telemetry{
		MsgTimestamp: 1543236572000,
		EncodingPath: "type:model/some/path",
		NodeId:       &telemetry.Telemetry_NodeIdStr{NodeIdStr: "hostname"},
		Subscription: &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "subscription"},
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "str"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "value",
								ValueByType: &telemetry.TelemetryField_Sint64Value{Sint64Value: -1},
							},
						},
					},
				},
			},
		},
	}
}

func TestTCPDialoutMultiple(t *testing.T) {
	c := &CiscoTelemetryMDT{Log: testutil.Logger{}, Transport: "tcp", ServiceAddress: "127.0.0.1:0", Aliases: map[string]string{
		"some": "type:model/some/path", "parallel": "type:model/parallel/path", "other": "type:model/other/path"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	require.NoError(t, err)

	telemetry := mockTelemetryMessage()

	hdr := struct {
		MsgType       uint16
		MsgEncap      uint16
		MsgHdrVersion uint16
		MsgFlags      uint16
		MsgLen        uint32
	}{}

	addr := c.Address()
	conn, err := net.Dial(addr.Network(), addr.String())
	require.NoError(t, err)

	data, _ := proto.Marshal(telemetry)
	hdr.MsgLen = uint32(len(data))
	binary.Write(conn, binary.BigEndian, hdr)
	conn.Write(data)

	conn2, err := net.Dial(addr.Network(), addr.String())
	require.NoError(t, err)

	telemetry.EncodingPath = "type:model/parallel/path"
	data, _ = proto.Marshal(telemetry)
	hdr.MsgLen = uint32(len(data))
	binary.Write(conn2, binary.BigEndian, hdr)
	conn2.Write(data)
	conn2.Write([]byte{0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0})
	conn2.Read([]byte{0})
	conn2.Close()

	telemetry.EncodingPath = "type:model/other/path"
	data, _ = proto.Marshal(telemetry)
	hdr.MsgLen = uint32(len(data))
	binary.Write(conn, binary.BigEndian, hdr)
	conn.Write(data)
	conn.Write([]byte{0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0})
	conn.Read([]byte{0})
	c.Stop()
	conn.Close()

	// We use the invalid dialout flags to let the server close the connection
	require.Equal(t, acc.Errors, []error{errors.New("invalid dialout flags: 257"), errors.New("invalid dialout flags: 257")})

	tags := map[string]string{"path": "type:model/some/path", "name": "str", "source": "hostname", "subscription": "subscription"}
	fields := map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "some", fields, tags)

	tags = map[string]string{"path": "type:model/parallel/path", "name": "str", "source": "hostname", "subscription": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "parallel", fields, tags)

	tags = map[string]string{"path": "type:model/other/path", "name": "str", "source": "hostname", "subscription": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "other", fields, tags)
}

func TestGRPCDialoutError(t *testing.T) {
	c := &CiscoTelemetryMDT{Log: testutil.Logger{}, Transport: "grpc", ServiceAddress: "127.0.0.1:0"}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	require.NoError(t, err)

	addr := c.Address()
	conn, _ := grpc.Dial(addr.String(), grpc.WithInsecure())
	client := dialout.NewGRPCMdtDialoutClient(conn)
	stream, _ := client.MdtDialout(context.Background())

	args := &dialout.MdtDialoutArgs{Errors: "foobar"}
	stream.Send(args)

	// Wait for the server to close
	stream.Recv()
	c.Stop()

	require.Equal(t, acc.Errors, []error{errors.New("GRPC dialout error: foobar")})
}

func TestGRPCDialoutMultiple(t *testing.T) {
	c := &CiscoTelemetryMDT{Log: testutil.Logger{}, Transport: "grpc", ServiceAddress: "127.0.0.1:0", Aliases: map[string]string{
		"some": "type:model/some/path", "parallel": "type:model/parallel/path", "other": "type:model/other/path"}}
	acc := &testutil.Accumulator{}
	err := c.Start(acc)
	require.NoError(t, err)
	telemetry := mockTelemetryMessage()

	addr := c.Address()
	conn, _ := grpc.Dial(addr.String(), grpc.WithInsecure(), grpc.WithBlock())
	client := dialout.NewGRPCMdtDialoutClient(conn)
	stream, _ := client.MdtDialout(context.TODO())

	data, _ := proto.Marshal(telemetry)
	args := &dialout.MdtDialoutArgs{Data: data, ReqId: 456}
	stream.Send(args)

	conn2, _ := grpc.Dial(addr.String(), grpc.WithInsecure(), grpc.WithBlock())
	client2 := dialout.NewGRPCMdtDialoutClient(conn2)
	stream2, _ := client2.MdtDialout(context.TODO())

	telemetry.EncodingPath = "type:model/parallel/path"
	data, _ = proto.Marshal(telemetry)
	args = &dialout.MdtDialoutArgs{Data: data}
	stream2.Send(args)
	stream2.Send(&dialout.MdtDialoutArgs{Errors: "testclose"})
	stream2.Recv()
	conn2.Close()

	telemetry.EncodingPath = "type:model/other/path"
	data, _ = proto.Marshal(telemetry)
	args = &dialout.MdtDialoutArgs{Data: data}
	stream.Send(args)
	stream.Send(&dialout.MdtDialoutArgs{Errors: "testclose"})
	stream.Recv()

	c.Stop()
	conn.Close()

	require.Equal(t, acc.Errors, []error{errors.New("GRPC dialout error: testclose"), errors.New("GRPC dialout error: testclose")})

	tags := map[string]string{"path": "type:model/some/path", "name": "str", "source": "hostname", "subscription": "subscription"}
	fields := map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "some", fields, tags)

	tags = map[string]string{"path": "type:model/parallel/path", "name": "str", "source": "hostname", "subscription": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "parallel", fields, tags)

	tags = map[string]string{"path": "type:model/other/path", "name": "str", "source": "hostname", "subscription": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "other", fields, tags)

}
