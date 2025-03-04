package p4runtime

import (
	"context"

	p4 "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

type fakeP4RuntimeClient struct {
	writeFn func(
		ctx context.Context,
		in *p4.WriteRequest,
		opts ...grpc.CallOption,
	) (*p4.WriteResponse, error)

	readFn func(
		in *p4.ReadRequest,
	) (p4.P4Runtime_ReadClient, error)

	setForwardingPipelineConfigFn func(
		ctx context.Context,
		in *p4.SetForwardingPipelineConfigRequest,
		opts ...grpc.CallOption,
	) (*p4.SetForwardingPipelineConfigResponse, error)

	getForwardingPipelineConfigFn func() (*p4.GetForwardingPipelineConfigResponse, error)

	streamChannelFn func(
		ctx context.Context,
		opts ...grpc.CallOption,
	) (p4.P4Runtime_StreamChannelClient, error)

	capabilitiesFn func(
		ctx context.Context,
		in *p4.CapabilitiesRequest,
		opts ...grpc.CallOption,
	) (*p4.CapabilitiesResponse, error)
}

// fakeP4RuntimeClient implements the v1.P4RuntimeClient interface
var _ p4.P4RuntimeClient = &fakeP4RuntimeClient{}

func (c *fakeP4RuntimeClient) Write(
	ctx context.Context,
	in *p4.WriteRequest,
	opts ...grpc.CallOption,
) (*p4.WriteResponse, error) {
	if c.writeFn == nil {
		panic("No mock defined for Write RPC")
	}
	return c.writeFn(ctx, in, opts...)
}

func (c *fakeP4RuntimeClient) Read(
	_ context.Context,
	in *p4.ReadRequest,
	_ ...grpc.CallOption,
) (p4.P4Runtime_ReadClient, error) {
	if c.readFn == nil {
		panic("No mock defined for Read RPC")
	}
	return c.readFn(in)
}

func (c *fakeP4RuntimeClient) SetForwardingPipelineConfig(
	ctx context.Context,
	in *p4.SetForwardingPipelineConfigRequest,
	opts ...grpc.CallOption,
) (*p4.SetForwardingPipelineConfigResponse, error) {
	if c.setForwardingPipelineConfigFn == nil {
		panic("No mock defined for SetForwardingPipelineConfig RPC")
	}
	return c.setForwardingPipelineConfigFn(ctx, in, opts...)
}

func (c *fakeP4RuntimeClient) GetForwardingPipelineConfig(
	context.Context,
	*p4.GetForwardingPipelineConfigRequest,
	...grpc.CallOption,
) (*p4.GetForwardingPipelineConfigResponse, error) {
	if c.getForwardingPipelineConfigFn == nil {
		panic("No mock defined for GetForwardingPipelineConfig RPC")
	}
	return c.getForwardingPipelineConfigFn()
}

func (c *fakeP4RuntimeClient) StreamChannel(
	ctx context.Context,
	opts ...grpc.CallOption,
) (p4.P4Runtime_StreamChannelClient, error) {
	if c.streamChannelFn == nil {
		panic("No mock defined for StreamChannel")
	}
	return c.streamChannelFn(ctx, opts...)
}

func (c *fakeP4RuntimeClient) Capabilities(
	ctx context.Context,
	in *p4.CapabilitiesRequest,
	opts ...grpc.CallOption,
) (*p4.CapabilitiesResponse, error) {
	if c.capabilitiesFn == nil {
		panic("No mock defined for Capabilities RPC")
	}
	return c.capabilitiesFn(ctx, in, opts...)
}

type fakeP4RuntimeReadClient struct {
	grpc.ClientStream
	recvFn func() (*p4.ReadResponse, error)
}

// fakeP4RuntimeReadClient implements the v1.P4Runtime_ReadClient interface
var _ p4.P4Runtime_ReadClient = &fakeP4RuntimeReadClient{}

func (c *fakeP4RuntimeReadClient) Recv() (*p4.ReadResponse, error) {
	if c.recvFn == nil {
		panic("No mock provided for Recv function")
	}
	return c.recvFn()
}
