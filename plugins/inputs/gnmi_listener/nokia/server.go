//go:generate protoc --proto_path=../gnmi_protos:. --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative nokia-dialout-telemetry.proto
package nokia

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/influxdata/telegraf"
	common_gnmi "github.com/influxdata/telegraf/plugins/common/gnmi"
)

// Make sure we implement the GRPC interface
var _ DialoutTelemetryServer = &server{}

type server struct {
	acc     telegraf.Accumulator
	handler *common_gnmi.Handler
	log     telegraf.Logger

	UnimplementedDialoutTelemetryServer
}

// New creates a new GRPC server for Nokia devices
func New(acc telegraf.Accumulator, handler *common_gnmi.Handler, log telegraf.Logger) *server {
	return &server{
		acc:     acc,
		handler: handler,
		log:     log,
	}
}

// Register the vendor specific GRPC methods to the server
func (s *server) Register(server *grpc.Server) {
	RegisterDialoutTelemetryServer(server, s)
}

// Publish implements the Nokia dial-out GRPC interface
func (s *server) Publish(srv grpc.BidiStreamingServer[gnmi.SubscribeResponse, gnmi.SubscribeRequest]) error {
	ctx := srv.Context()

	for ctx.Err() == nil {
		// Wait for data
		response, err := srv.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) && ctx.Err() == nil {
				return fmt.Errorf("aborted gNMI listener: %w", err)
			}
			break
		}

		// Determine the message source
		source := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			switch v := p.Addr.(type) {
			case *net.TCPAddr:
				source = v.IP.String()
			case *net.UDPAddr:
				source = v.IP.String()
			case *net.IPAddr:
				source = v.IP.String()
			default:
				source = p.Addr.String()
			}
		}

		// Call the handler
		s.handler.Process(s.acc, source, response)

		// Send an empty response
		if err := srv.Send(&gnmi.SubscribeRequest{}); err != nil {
			s.log.Errorf("Sending GNMI response failed: %v", err)
		}
	}

	return nil
}
