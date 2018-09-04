package rpc // import "collectd.org/rpc"

import (
	"context"
	"io"

	pb "collectd.org/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// RegisterServer registers the implementation srv with the gRPC instance s.
func RegisterServer(s *grpc.Server, srv Interface) {
	pb.RegisterCollectdServer(s, &server{
		Interface: srv,
	})
}

// Type server implements pb.CollectdServer using the Go implementation of
// rpc.Interface.
type server struct {
	Interface
}

// PutValues reads ValueLists from stream and calls the Write() implementation
// on each one.
func (s *server) PutValues(stream pb.Collectd_PutValuesServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		vl, err := UnmarshalValueList(req.GetValueList())
		if err != nil {
			return err
		}

		if err := s.Write(stream.Context(), vl); err != nil {
			return grpc.Errorf(codes.Internal, "Write(%v): %v", vl, err)
		}
	}

	return stream.SendAndClose(&pb.PutValuesResponse{})
}

// QueryValues calls the Query() implementation and streams all ValueLists from
// the channel back to the client.
func (s *server) QueryValues(req *pb.QueryValuesRequest, stream pb.Collectd_QueryValuesServer) error {
	id := UnmarshalIdentifier(req.GetIdentifier())

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	ch, err := s.Query(ctx, id)
	if err != nil {
		return grpc.Errorf(codes.Internal, "Query(%v): %v", id, err)
	}

	for vl := range ch {
		pbVL, err := MarshalValueList(vl)
		if err != nil {
			return err
		}

		res := &pb.QueryValuesResponse{
			ValueList: pbVL,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}
