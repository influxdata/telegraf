package gnmi

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/influxdata/telegraf"
	common_gnmi "github.com/influxdata/telegraf/plugins/common/gnmi"
	"github.com/influxdata/telegraf/selfstat"
)

type subscriber struct {
	handler    *common_gnmi.Handler
	host       string
	port       string
	maxMsgSize int
	log        telegraf.Logger
	keepalive.ClientParameters
}

func (s *subscriber) subscribe(ctx context.Context, acc telegraf.Accumulator, tlscfg *tls.Config, request *gnmi.SubscribeRequest) error {
	var creds credentials.TransportCredentials
	if tlscfg != nil {
		creds = credentials.NewTLS(tlscfg)
	} else {
		creds = insecure.NewCredentials()
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	if s.maxMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(s.maxMsgSize),
		))
	}

	if s.ClientParameters.Time > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(s.ClientParameters))
	}

	// Used to report the status of the TCP connection to the device. If the
	// GNMI connection goes down, but TCP is still up this will still report
	// connected until the TCP connection times out.
	connectStat := selfstat.Register("gnmi", "grpc_connection_status", map[string]string{"source": s.host})
	defer connectStat.Set(0)

	address := net.JoinHostPort(s.host, s.port)
	client, err := grpc.NewClient(address, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	subscribeClient, err := gnmi.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %w", err)
	}

	// If io.EOF is returned, the stream may have ended and stream status
	// can be determined by calling Recv.
	if err := subscribeClient.Send(request); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}
	connectStat.Set(1)
	s.log.Debugf("Connection to gNMI device %s established", address)

	defer s.log.Debugf("Connection to gNMI device %s closed", address)
	for ctx.Err() == nil {
		var reply *gnmi.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if !errors.Is(err, io.EOF) && ctx.Err() == nil {
				return fmt.Errorf("aborted gNMI subscription: %w", err)
			}
			break
		}

		s.handler.Process(acc, s.host, reply)
	}
	return nil
}
