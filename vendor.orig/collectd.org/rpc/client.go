package rpc // import "collectd.org/rpc"

import (
	"context"
	"io"
	"log"

	"collectd.org/api"
	pb "collectd.org/rpc/proto"
	"google.golang.org/grpc"
)

// Type client implements rpc.Interface using a gRPC stub.
type client struct {
	pb.CollectdClient
}

// Newclient returns a wrapper around the gRPC client connection that maps
// between the Go interface and the gRPC interface.
func NewClient(conn *grpc.ClientConn) Interface {
	return &client{
		CollectdClient: pb.NewCollectdClient(conn),
	}
}

// Query maps its arguments to a QueryValuesRequest object and calls
// QueryValues. The response is parsed by a goroutine and written to the
// returned channel.
func (c *client) Query(ctx context.Context, id *api.Identifier) (<-chan *api.ValueList, error) {
	stream, err := c.QueryValues(ctx, &pb.QueryValuesRequest{
		Identifier: MarshalIdentifier(id),
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan *api.ValueList, 16)

	go func() {
		defer close(ch)

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error while receiving value lists: %v", err)
				return
			}

			vl, err := UnmarshalValueList(res.GetValueList())
			if err != nil {
				log.Printf("received malformed response: %v", err)
				continue
			}

			select {
			case ch <- vl:
				continue
			case <-stream.Context().Done():
				break
			}
		}
	}()

	return ch, nil
}

// Write maps its arguments to a PutValuesRequest and calls PutValues.
func (c *client) Write(ctx context.Context, vl *api.ValueList) error {
	pbVL, err := MarshalValueList(vl)
	if err != nil {
		return err
	}

	stream, err := c.PutValues(ctx)
	if err != nil {
		return err
	}

	req := &pb.PutValuesRequest{
		ValueList: pbVL,
	}
	if err := stream.Send(req); err != nil {
		stream.CloseSend()
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}
