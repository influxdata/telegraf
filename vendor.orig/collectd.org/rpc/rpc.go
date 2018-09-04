/*
Package rpc implements an idiomatic Go interface to collectd's gRPC server.

The functions and types in this package aim to make it easy and convenient to
use collectd's gRPC interface. It supports both client and server code.

Client code

Synopsis:

  conn, err := grpc.Dial(*addr, opts...)
  if err != nil {
	  // handle error
  }

  c := rpc.NewClient(conn)

  // Send a ValueList to the server.
  if err := c.Write(context.Background(), vl); err != nil {
	  // handle error
  }

  // Retrieve matching ValueLists from the server.
  ch, err := c.Query(context.Background(), api.Identifier{
	  Host: "*",
	  Plugin: "golang",
  })
  if err != nil {
	  // handle error
  }

  for vl := range ch {
	  // consume ValueList
  }

Server code

Synopsis:

  type myServer struct {
	  rpc.Interface
  }

  func (s *myServer) Write(ctx context.Context, vl *api.ValueList) error {
	  // implementation
  }

  func (s *myServer) Query(ctx context.Context, id *api.Identifier) (<-chan *api.ValueList, error) {
	  // implementation
  }

  func main() {
	  sock, err := net.Listen("tcp", ":12345")
	  if err != nil {
		  // handle error
	  }

	  srv := grpc.NewServer(opts...)
	  rpc.RegisterServer(srv, &myServer{})
	  srv.Serve(sock)
  }
*/
package rpc // import "collectd.org/rpc"

import (
	"context"

	"collectd.org/api"
)

// Interface is an idiomatic Go interface for the Collectd gRPC service.
//
// To implement a client, pass a client connection to NewClient() to get back
// an object implementing this interface.
//
// To implement a server, use RegisterServer() to hook an object, which
// implements Interface, up to a gRPC server.
type Interface interface {
	api.Writer
	Query(context.Context, *api.Identifier) (<-chan *api.ValueList, error)
}
