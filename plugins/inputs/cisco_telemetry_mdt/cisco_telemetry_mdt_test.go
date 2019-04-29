/**
 * Copyright (c) 2018 Cisco Systems
 * Author: Steven Barth <stbarth@cisco.com>
 */

package cisco_telemetry_mdt

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"

	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf/plugins/inputs/cisco_telemetry_mdt/ems"

	"github.com/golang/protobuf/proto"

	dialout "github.com/influxdata/telegraf/plugins/inputs/cisco_telemetry_mdt/mdt_dialout"
	"github.com/influxdata/telegraf/plugins/inputs/cisco_telemetry_mdt/telemetry"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestHandleTelemetryEmpty(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "dummy"}
	acc := &testutil.Accumulator{}
	c.Start(acc)

	telemetry := &telemetry.Telemetry{
		DataGpbkv: []*telemetry.TelemetryField{
			{},
		},
	}
	data, _ := proto.Marshal(telemetry)

	c.handleTelemetry(data)
	assert.Contains(t, acc.Errors, errors.New("I! Cisco MDT invalid field: encoding path or measurement empty"))
	assert.Empty(t, acc.Metrics)
}

func TestHandleTelemetryTwoSimple(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "dummy"}
	acc := &testutil.Accumulator{}
	c.Start(acc)

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
	assert.Empty(t, acc.Errors)

	tags := map[string]string{"name": "str", "uint64": "1234", "Producer": "hostname", "Target": "subscription"}
	fields := map[string]interface{}{"bool": true}
	acc.AssertContainsTaggedFields(t, "type:model/some/path", fields, tags)

	tags = map[string]string{"name": "str2", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"bool": false}
	acc.AssertContainsTaggedFields(t, "type:model/some/path", fields, tags)
}

func TestHandleTelemetrySingleNested(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "dummy"}
	acc := &testutil.Accumulator{}
	c.Start(acc)

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
	assert.Empty(t, acc.Errors)

	tags := map[string]string{"nested/key/level": "3", "Producer": "hostname", "Target": "subscription"}
	fields := map[string]interface{}{"nested/value/foo": "bar"}
	acc.AssertContainsTaggedFields(t, "type:model/nested/path", fields, tags)
}

func TestTCPDialoutOverflow(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "tcp-dialout", ServiceAddress: "127.0.0.1:57000"}
	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))

	hdr := struct {
		MsgType       uint16
		MsgEncap      uint16
		MsgHdrVersion uint16
		MsgFlags      uint16
		MsgLen        uint32
	}{MsgLen: uint32(1000000000)}

	conn, _ := net.Dial("tcp", "127.0.0.1:57000")
	binary.Write(conn, binary.BigEndian, hdr)
	conn.Close()

	time.Sleep(time.Second)

	c.Stop()

	assert.Contains(t, acc.Errors, errors.New("E! Dialout packet too long: 1000000000"))
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
	c := &CiscoTelemetryMDT{Transport: "tcp-dialout", ServiceAddress: "127.0.0.1:57000"}
	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))

	telemetry := mockTelemetryMessage()

	hdr := struct {
		MsgType       uint16
		MsgEncap      uint16
		MsgHdrVersion uint16
		MsgFlags      uint16
		MsgLen        uint32
	}{}

	conn, _ := net.Dial("tcp", "127.0.0.1:57000")

	data, _ := proto.Marshal(telemetry)
	hdr.MsgLen = uint32(len(data))
	binary.Write(conn, binary.BigEndian, hdr)
	conn.Write(data)

	conn2, _ := net.Dial("tcp", "127.0.0.1:57000")
	telemetry.EncodingPath = "type:model/parallel/path"
	data, _ = proto.Marshal(telemetry)
	hdr.MsgLen = uint32(len(data))
	binary.Write(conn2, binary.BigEndian, hdr)
	conn2.Write(data)
	conn2.Close()

	telemetry.EncodingPath = "type:model/other/path"
	data, _ = proto.Marshal(telemetry)
	hdr.MsgLen = uint32(len(data))
	binary.Write(conn, binary.BigEndian, hdr)
	conn.Write(data)

	time.Sleep(time.Second)

	c.Stop()
	conn.Close()

	assert.Empty(t, acc.Errors)

	tags := map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields := map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/some/path", fields, tags)

	tags = map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/parallel/path", fields, tags)

	tags = map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/other/path", fields, tags)
}

func TestGRPCDialoutError(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "grpc-dialout", ServiceAddress: "127.0.0.1:57001"}
	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))

	conn, _ := grpc.Dial("127.0.0.1:57001", grpc.WithInsecure())
	client := dialout.NewGRPCMdtDialoutClient(conn)
	stream, _ := client.MdtDialout(context.Background())

	args := &dialout.MdtDialoutArgs{Errors: "foobar"}
	stream.Send(args)

	time.Sleep(time.Second)

	c.Stop()

	assert.Contains(t, acc.Errors, errors.New("E! GRPC dialout error: foobar"))
}

func TestGRPCDialoutMultiple(t *testing.T) {
	c := &CiscoTelemetryMDT{Transport: "grpc-dialout", ServiceAddress: "127.0.0.1:57001"}
	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))
	telemetry := mockTelemetryMessage()

	conn, _ := grpc.Dial("127.0.0.1:57001", grpc.WithInsecure(), grpc.WithBlock())
	client := dialout.NewGRPCMdtDialoutClient(conn)
	stream, _ := client.MdtDialout(context.TODO())

	data, _ := proto.Marshal(telemetry)
	args := &dialout.MdtDialoutArgs{Data: data, ReqId: 456}
	stream.Send(args)

	conn2, _ := grpc.Dial("127.0.0.1:57001", grpc.WithInsecure(), grpc.WithBlock())
	client2 := dialout.NewGRPCMdtDialoutClient(conn2)
	stream2, _ := client2.MdtDialout(context.TODO())

	telemetry.EncodingPath = "type:model/parallel/path"
	data, _ = proto.Marshal(telemetry)
	args = &dialout.MdtDialoutArgs{Data: data}
	stream2.Send(args)
	stream2.CloseSend()

	time.Sleep(time.Second)
	conn2.Close()

	telemetry.EncodingPath = "type:model/other/path"
	data, _ = proto.Marshal(telemetry)
	args = &dialout.MdtDialoutArgs{Data: data}
	stream.Send(args)
	time.Sleep(time.Second)

	c.Stop()
	conn.Close()

	assert.Empty(t, acc.Errors)

	tags := map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields := map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/some/path", fields, tags)

	tags = map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/parallel/path", fields, tags)

	tags = map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/other/path", fields, tags)

}

type mockDialinServer struct {
	t        *testing.T
	scenario int
}

func (m *mockDialinServer) GetConfig(*ems.ConfigGetArgs, ems.GRPCConfigOper_GetConfigServer) error {
	return nil
}

func (m *mockDialinServer) MergeConfig(context.Context, *ems.ConfigArgs) (*ems.ConfigReply, error) {
	return nil, nil
}

func (m *mockDialinServer) DeleteConfig(context.Context, *ems.ConfigArgs) (*ems.ConfigReply, error) {
	return nil, nil
}

func (m *mockDialinServer) ReplaceConfig(context.Context, *ems.ConfigArgs) (*ems.ConfigReply, error) {
	return nil, nil
}

func (m *mockDialinServer) CliConfig(context.Context, *ems.CliConfigArgs) (*ems.CliConfigReply, error) {
	return nil, nil
}

func (m *mockDialinServer) CommitReplace(context.Context, *ems.CommitReplaceArgs) (*ems.CommitReplaceReply, error) {
	return nil, nil
}

func (m *mockDialinServer) CommitConfig(context.Context, *ems.CommitArgs) (*ems.CommitReply, error) {
	return nil, nil
}

func (m *mockDialinServer) ConfigDiscardChanges(context.Context, *ems.DiscardChangesArgs) (*ems.DiscardChangesReply, error) {
	return nil, nil
}

func (m *mockDialinServer) GetOper(*ems.GetOperArgs, ems.GRPCConfigOper_GetOperServer) error {
	return nil
}

func (m *mockDialinServer) CreateSubs(args *ems.CreateSubsArgs, server ems.GRPCConfigOper_CreateSubsServer) error {
	assert.Equal(m.t, args.GetSubidstr(), "thesubscription")
	assert.Equal(m.t, args.GetEncode(), grpcEncodeGPBKV)

	metadata, ok := metadata.FromIncomingContext(server.Context())
	assert.Equal(m.t, ok, true)
	assert.Equal(m.t, metadata.Get("username"), []string{"theuser"})
	assert.Equal(m.t, metadata.Get("password"), []string{"thepassword"})

	if m.scenario == 0 {
		telemetry := mockTelemetryMessage()
		data, _ := proto.Marshal(telemetry)
		server.Send(&ems.CreateSubsReply{Data: data})

		telemetry.EncodingPath = "type:model/parallel/path"
		data, _ = proto.Marshal(telemetry)
		server.Send(&ems.CreateSubsReply{Data: data})
	} else if m.scenario == 1 {
		telemetry := mockTelemetryMessage()
		telemetry.EncodingPath = "type:model/other/path"
		data, _ := proto.Marshal(telemetry)
		server.Send(&ems.CreateSubsReply{Data: data})
	} else if m.scenario == 2 {
		server.Send(&ems.CreateSubsReply{Errors: "testerror"})
	}

	return nil
}

func TestGRPCDialinError(t *testing.T) {
	m := &mockDialinServer{t: t, scenario: 2}
	listener, _ := net.Listen("tcp", "127.0.0.1:57002")
	server := grpc.NewServer()
	ems.RegisterGRPCConfigOperServer(server, m)
	go server.Serve(listener)

	c := &CiscoTelemetryMDT{Transport: "grpc-dialin", ServiceAddress: "127.0.0.1:57002",
		Username: "theuser", Password: "thepassword", Subscription: "thesubscription",
		Redial: internal.Duration{Duration: 1 * time.Second}}
	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))

	time.Sleep(1 * time.Second)

	server.Stop()
	c.Stop()

	assert.Equal(t, acc.Errors, []error{errors.New("E! GRPC dialin error: testerror")})
}

func TestGRPCDialinMultipleRedial(t *testing.T) {
	m := &mockDialinServer{t: t}
	listener, _ := net.Listen("tcp", "127.0.0.1:57002")
	server := grpc.NewServer()
	ems.RegisterGRPCConfigOperServer(server, m)
	go server.Serve(listener)

	c := &CiscoTelemetryMDT{Transport: "grpc-dialin", ServiceAddress: "127.0.0.1:57002",
		Username: "theuser", Password: "thepassword", Subscription: "thesubscription",
		Redial: internal.Duration{Duration: 1 * time.Second}}
	acc := &testutil.Accumulator{}
	assert.Nil(t, c.Start(acc))

	time.Sleep(1 * time.Second)

	server.Stop()
	m.scenario = 1

	listener, _ = net.Listen("tcp", "127.0.0.1:57002")
	server = grpc.NewServer()
	ems.RegisterGRPCConfigOperServer(server, m)
	go server.Serve(listener)

	time.Sleep(1 * time.Second)

	server.Stop()
	c.Stop()

	assert.Empty(t, acc.Errors)

	tags := map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields := map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/some/path", fields, tags)

	tags = map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/parallel/path", fields, tags)

	tags = map[string]string{"name": "str", "Producer": "hostname", "Target": "subscription"}
	fields = map[string]interface{}{"value": int64(-1)}
	acc.AssertContainsTaggedFields(t, "type:model/other/path", fields, tags)
}
