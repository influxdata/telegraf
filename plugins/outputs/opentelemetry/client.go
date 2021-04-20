package opentelemetry

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	// serviceConfig copied from OTel-Go
	// https://github.com/open-telemetry/opentelemetry-go/blob/a2cecb6e80f6a0712187b080a97f8efb5a61082a/exporters/otlp/internal/otlpconfig/options.go#L47.
	serviceConfig = `{
	"methodConfig":[{
		"name":[
			{ "service":"opentelemetry.proto.collector.metrics.v1.MetricsService" },
			{ "service":"opentelemetry.proto.collector.trace.v1.TraceService" }
		],
		"retryPolicy":{
			"MaxAttempts":5,
			"InitialBackoff":"0.3s",
			"MaxBackoff":"5s",
			"BackoffMultiplier":2,
			"RetryableStatusCodes":[
				"CANCELLED",
				"DEADLINE_EXCEEDED",
				"RESOURCE_EXHAUSTED",
				"ABORTED",
				"OUT_OF_RANGE",
				"UNAVAILABLE",
				"DATA_LOSS"
			]
		}
	}]
}`
)

var (
	maxErrorDetailStringLen = 512
	maxTimeseriesPerRequest = 500
)

// client allows reading and writing from/to a remote gRPC endpoint. The
// implementation may hit a single backend, so the application should create a
// number of these clients.
type client struct {
	logger     telegraf.Logger
	url        *url.URL
	timeout    time.Duration
	tlsConfig  *tls.Config
	headers    metadata.MD
	compressor string

	conn *grpc.ClientConn
}

// connect will dial a new connection if one is not set.  When
// dialing, this function uses its a new context and the same timeout
// used for store().
func (c *client) connect(ctx context.Context) (_ *grpc.ClientConn, retErr error) {
	if c.conn != nil {
		return c.conn, nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debugf("new OpenTelemetry connection, url=%s timeout=%s", c.url.String(), c.timeout)

	dopts := []grpc.DialOption{
		grpc.WithBlock(), // Wait for the connection to be established before using it.
		grpc.WithDefaultServiceConfig(serviceConfig),
	}
	if c.url.Scheme != "http" {
		dopts = append(dopts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConfig)))
	} else {
		dopts = append(dopts, grpc.WithInsecure())
	}
	if c.compressor != "" && c.compressor != "none" {
		dopts = append(dopts, grpc.WithDefaultCallOptions(grpc.UseCompressor(c.compressor)))
	}
	address := c.url.Hostname()
	if len(c.url.Port()) > 0 {
		address = net.JoinHostPort(address, c.url.Port())
	}
	conn, err := grpc.DialContext(ctx, address, dopts...)
	c.conn = conn
	if err != nil {
		c.logger.Debug("connection status, address=%s err=%w", address, err)
		return nil, err
	}

	return conn, err
}

// ping sends an empty request the endpoint.
func (c *client) ping(ctx context.Context) error {
	// Loop until the context is canceled, allowing for retryable failures.
	for {
		conn, err := c.connect(ctx)

		if err == nil {
			service := metricsService.NewMetricsServiceClient(conn)
			empty := &metricsService.ExportMetricsServiceRequest{}

			_, err = service.Export(metadata.NewOutgoingContext(ctx, c.headers), empty)
			if err == nil {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if isRecoverable(err) {
				c.logger.Infof("ping recoverable error, still trying, err=%w", err)
				continue
			}
		}
		return fmt.Errorf("non-recoverable failure in ping, err=%w", err)
	}
}

// store sends a batch of samples to the endpoint.
func (c *client) store(req *metricsService.ExportMetricsServiceRequest) error {
	tss := req.ResourceMetrics
	if len(tss) == 0 {
		// Nothing to do, return silently.
		return nil
	}

	// Note the call to connect() applies its own timeout for Dial().
	ctx := context.Background()
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	service := metricsService.NewMetricsServiceClient(conn)

	errs := make(chan error, len(tss)/maxTimeseriesPerRequest+1)
	var wg sync.WaitGroup
	for i := 0; i < len(tss); i += maxTimeseriesPerRequest {
		end := i + maxTimeseriesPerRequest
		if end > len(tss) {
			end = len(tss)
		}
		wg.Add(1)
		go func(begin int, end int) {
			defer wg.Done()
			reqCopy := &metricsService.ExportMetricsServiceRequest{
				ResourceMetrics: req.ResourceMetrics[begin:end],
			}

			var md metadata.MD
			var err error

			if _, err = service.Export(metadata.NewOutgoingContext(ctx, c.headers), reqCopy, grpc.Trailer(&md)); err != nil {
				c.logger.Errorf("export failure, err=%w size=%d trailers=%v recoverable=%t",
					truncateErrorString(err),
					proto.Size(reqCopy),
					md,
					isRecoverable(err),
				)
				errs <- err
				return
			}

			c.logger.Debug("successful write, records=%d size=%d trailers=%v", end-begin, proto.Size(reqCopy), md)
		}(i, end)
	}
	wg.Wait()
	close(errs)
	if err, ok := <-errs; ok {
		return err
	}
	return nil
}

func (c *client) close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// truncateErrorString avoids printing error messages that are very
// large.
func truncateErrorString(err error) string {
	tmp := fmt.Sprint(err)
	if len(tmp) > maxErrorDetailStringLen {
		tmp = fmt.Sprint(tmp[:maxErrorDetailStringLen], " ...")
	}
	return tmp
}

func isRecoverable(err error) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch s.Code() {
	case codes.DeadlineExceeded, codes.Canceled, codes.ResourceExhausted,
		codes.Aborted, codes.OutOfRange, codes.Unavailable, codes.DataLoss:
		// See https://github.com/open-telemetry/opentelemetry-specification/
		// blob/master/specification/protocol/otlp.md#response
		return true
	default:
		return false
	}
}
