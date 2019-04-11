package loggregator_rlp_test

import (
	"crypto/tls"
	"log"
	"net"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type MockRLP struct {
	server *grpc.Server
	Addr   string

	envelopeResponse []*loggregator_v2.Envelope
	tlsConfig        *tls.Config

	mu                 sync.Mutex
	connectionAttempts int
	actualReq          *loggregator_v2.EgressBatchRequest
}

func NewMockRlp(envelopeResponse []*loggregator_v2.Envelope, tlsConfig *tls.Config) *MockRLP {
	f := &MockRLP{
		envelopeResponse: envelopeResponse,
		tlsConfig:        tlsConfig,
	}

	return f
}

func (f *MockRLP) Receiver(
	*loggregator_v2.EgressRequest,
	loggregator_v2.Egress_ReceiverServer,
) error {
	return status.Errorf(codes.Unimplemented, "use BatchedReceiver instead")
}

func (f *MockRLP) BatchedReceiver(
	req *loggregator_v2.EgressBatchRequest,
	srv loggregator_v2.Egress_BatchedReceiverServer,
) error {
	f.mu.Lock()
	f.connectionAttempts++
	f.actualReq = req
	f.mu.Unlock()
	var i int
	for range time.Tick(10 * time.Millisecond) {
		srv.Send(&loggregator_v2.EnvelopeBatch{
			Batch: f.envelopeResponse,
		})
		i++
	}
	return nil
}

func (f *MockRLP) Start() {
	addr := f.Addr
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	var lis net.Listener
	for i := 0; ; i++ {
		var err error
		lis, err = net.Listen("tcp", addr)
		if err != nil {
			// This can happen if the port is already in use...
			if i < 50 {
				log.Printf("failed to bind for fake producer. Trying again (%d/50)...: %s", i+1, err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			panic(err)
		}
		break
	}
	f.Addr = lis.Addr().String()

	opt := grpc.Creds(credentials.NewTLS(f.tlsConfig))
	f.server = grpc.NewServer(opt)
	loggregator_v2.RegisterEgressServer(f.server, f)

	go f.listen(lis)
}

func (f *MockRLP) listen(lis net.Listener) {
	_ = f.server.Serve(lis)
}

func (f *MockRLP) Stop() {
	if f.server == nil {
		return
	}

	f.server.Stop()
	f.server = nil
	return
}

func (f *MockRLP) ActualReq() *loggregator_v2.EgressBatchRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.actualReq
}

func (f *MockRLP) ConnectionAttempts() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connectionAttempts
}
