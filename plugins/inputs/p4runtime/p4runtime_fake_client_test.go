package p4runtime

import (
	"context"

	"github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

type fakeP4RuntimeClient struct {
	writeFn func(
		ctx context.Context,
		in *v1.WriteRequest,
		opts ...grpc.CallOption,
	) (*v1.WriteResponse, error)

	readFn func(
		in *v1.ReadRequest,
	) (v1.P4Runtime_ReadClient, error)

	setForwardingPipelineConfigFn func(
		ctx context.Context,
		in *v1.SetForwardingPipelineConfigRequest,
		opts ...grpc.CallOption,
	) (*v1.SetForwardingPipelineConfigResponse, error)

	getForwardingPipelineConfigFn func() (*v1.GetForwardingPipelineConfigResponse, error)

	streamChannelFn func(
		ctx context.Context,
		opts ...grpc.CallOption,
	) (v1.P4Runtime_StreamChannelClient, error)

	capabilitiesFn func(
		ctx context.Context,
		in *v1.CapabilitiesRequest,
		opts ...grpc.CallOption,
	) (*v1.CapabilitiesResponse, error)
}

// fakeP4RuntimeClient implements the v1.P4RuntimeClient interface
var _ v1.P4RuntimeClient = &fakeP4RuntimeClient{}

func (c *fakeP4RuntimeClient) Write(
	ctx context.Context,
	in *v1.WriteRequest,
	opts ...grpc.CallOption,
) (*v1.WriteResponse, error) {
	if c.writeFn == nil {
		panic("No mock defined for Write RPC")
	}
	return c.writeFn(ctx, in, opts...)
}

func (c *fakeP4RuntimeClient) Read(
	_ context.Context,
	in *v1.ReadRequest,
	_ ...grpc.CallOption,
) (v1.P4Runtime_ReadClient, error) {
	if c.readFn == nil {
		panic("No mock defined for Read RPC")
	}
	return c.readFn(in)
}

func (c *fakeP4RuntimeClient) SetForwardingPipelineConfig(
	ctx context.Context,
	in *v1.SetForwardingPipelineConfigRequest,
	opts ...grpc.CallOption,
) (*v1.SetForwardingPipelineConfigResponse, error) {
	if c.setForwardingPipelineConfigFn == nil {
		panic("No mock defined for SetForwardingPipelineConfig RPC")
	}
	return c.setForwardingPipelineConfigFn(ctx, in, opts...)
}

func (c *fakeP4RuntimeClient) GetForwardingPipelineConfig(
	context.Context,
	*v1.GetForwardingPipelineConfigRequest,
	...grpc.CallOption,
) (*v1.GetForwardingPipelineConfigResponse, error) {
	if c.getForwardingPipelineConfigFn == nil {
		panic("No mock defined for GetForwardingPipelineConfig RPC")
	}
	return c.getForwardingPipelineConfigFn()
}

func (c *fakeP4RuntimeClient) StreamChannel(
	ctx context.Context,
	opts ...grpc.CallOption,
) (v1.P4Runtime_StreamChannelClient, error) {
	if c.streamChannelFn == nil {
		panic("No mock defined for StreamChannel")
	}
	return c.streamChannelFn(ctx, opts...)
}

func (c *fakeP4RuntimeClient) Capabilities(
	ctx context.Context,
	in *v1.CapabilitiesRequest,
	opts ...grpc.CallOption,
) (*v1.CapabilitiesResponse, error) {
	if c.capabilitiesFn == nil {
		panic("No mock defined for Capabilities RPC")
	}
	return c.capabilitiesFn(ctx, in, opts...)
}

type fakeP4RuntimeReadClient struct {
	grpc.ClientStream
	recvFn func() (*v1.ReadResponse, error)
}

// fakeP4RuntimeReadClient implements the v1.P4Runtime_ReadClient interface
var _ v1.P4Runtime_ReadClient = &fakeP4RuntimeReadClient{}

func (c *fakeP4RuntimeReadClient) Recv() (*v1.ReadResponse, error) {
	if c.recvFn == nil {
		panic("No mock provided for Recv function")
	}
	return c.recvFn()
}
