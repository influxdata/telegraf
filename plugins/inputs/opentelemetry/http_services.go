package opentelemetry

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	pprofileotlp "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf"
)

func (o *OpenTelemetry) startHTTPService(acc telegraf.Accumulator, tlsCfg *tls.Config) {
	mux := http.NewServeMux()
	mux.Handle("POST "+o.srvs.tracesPath, o.handleTraceMessage(o.srvs.traces))
	mux.Handle("POST "+o.srvs.metricsPath, o.handleMetricMessage(o.srvs.metrics))
	mux.Handle("POST "+o.srvs.logsPath, o.handleLogMessage(o.srvs.logs))
	mux.Handle("POST "+o.srvs.profilesPath, o.handleProfileMessage(o.srvs.profiles))

	o.httpServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(o.Timeout),
		ReadTimeout:       time.Duration(o.Timeout),
	}

	// Wrap the listener instead of using ServeTLS to avoid the automatic
	// HTTP/2 setup which fails for cipher suites not supported by HTTP/2.
	if tlsCfg != nil {
		o.listener = tls.NewListener(o.listener, tlsCfg)
	}

	o.wg.Go(func() {
		if err := o.httpServer.Serve(o.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			acc.AddError(fmt.Errorf("serving OpenTelemetry HTTP service failed: %w", err))
		}
	})
}

// Handler function for the traces endpoint
func (o *OpenTelemetry) handleTraceMessage(service *traceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("invalid content type: %w", err))
			return
		}
		if mediatype != "application/x-protobuf" {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("unsupported content type %q", mediatype))
			return
		}

		// Read and decode message
		body, code, err := o.readBody(w, r)
		if err != nil {
			o.writeError(w, r, code, err)
			return
		}

		// Unmarshal protobuf message and use the service to export the traces
		req := ptraceotlp.NewExportRequest()
		if err := req.UnmarshalProto(body); err != nil {
			o.writeError(w, r, http.StatusBadRequest, fmt.Errorf("decoding request failed: %w", err))
			return
		}

		resp, err := service.Export(addPeerInfo(r), req)
		if err != nil {
			o.writeError(w, r, http.StatusBadRequest, err)
			return
		}

		// Send the response
		buf, err := resp.MarshalProto()
		if err != nil {
			o.writeError(w, r, http.StatusInternalServerError, fmt.Errorf("encoding response failed: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf); err != nil {
			o.Log.Errorf("Writing response to %q failed: %v", r.URL.Path, err)
		}
	}
}

// Handler function for the metric endpoint
func (o *OpenTelemetry) handleMetricMessage(service *metricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("invalid content type: %w", err))
			return
		}
		if mediatype != "application/x-protobuf" {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("unsupported content type %q", mediatype))
			return
		}

		// Read and decode message
		body, code, err := o.readBody(w, r)
		if err != nil {
			o.writeError(w, r, code, err)
			return
		}

		// Unmarshal protobuf message and use the service to export the metrics
		req := pmetricotlp.NewExportRequest()
		if err := req.UnmarshalProto(body); err != nil {
			o.writeError(w, r, http.StatusBadRequest, fmt.Errorf("decoding request failed: %w", err))
			return
		}

		resp, err := service.Export(addPeerInfo(r), req)
		if err != nil {
			o.writeError(w, r, http.StatusBadRequest, err)
			return
		}

		// Send the response
		buf, err := resp.MarshalProto()
		if err != nil {
			o.writeError(w, r, http.StatusInternalServerError, fmt.Errorf("encoding response failed: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf); err != nil {
			o.Log.Errorf("Writing response to %q failed: %v", r.URL.Path, err)
		}
	}
}

// Handler function for the logs endpoint
func (o *OpenTelemetry) handleLogMessage(service *logsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("invalid content type: %w", err))
			return
		}
		switch mediatype {
		case "application/x-protobuf":
			// Read and decode message
			body, code, err := o.readBody(w, r)
			if err != nil {
				o.writeError(w, r, code, err)
				return
			}

			// Unmarshal protobuf message and use the service to export the logs
			req := plogotlp.NewExportRequest()
			if err := req.UnmarshalProto(body); err != nil {
				o.writeError(w, r, http.StatusBadRequest, fmt.Errorf("decoding request failed: %w", err))
				return
			}

			resp, err := service.Export(addPeerInfo(r), req)
			if err != nil {
				o.writeError(w, r, http.StatusBadRequest, err)
				return
			}

			// Send the response
			buf, err := resp.MarshalProto()
			if err != nil {
				o.writeError(w, r, http.StatusInternalServerError, fmt.Errorf("encoding response failed: %w", err))
				return
			}

			w.Header().Set("Content-Type", "application/x-protobuf")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(buf); err != nil {
				o.Log.Errorf("Writing response to %q failed: %v", r.URL.Path, err)
			}
		default:
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("unsupported content type %q", mediatype))
			return
		}
	}
}

// Handler function for the profiles endpoint
func (o *OpenTelemetry) handleProfileMessage(service *profileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify content type
		mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("invalid content type: %w", err))
			return
		}
		if mediatype != "application/x-protobuf" {
			o.writeError(w, r, http.StatusUnsupportedMediaType, fmt.Errorf("unsupported content type %q", mediatype))
			return
		}

		// Read and decode message
		body, code, err := o.readBody(w, r)
		if err != nil {
			o.writeError(w, r, code, err)
			return
		}

		// Unmarshal protobuf message and use the service to export the profiles
		var req pprofileotlp.ExportProfilesServiceRequest
		req.Reset()
		if err := proto.Unmarshal(body, &req); err != nil {
			o.writeError(w, r, http.StatusBadRequest, fmt.Errorf("decoding request failed: %w", err))
			return
		}

		resp, err := service.Export(addPeerInfo(r), &req)
		if err != nil {
			o.writeError(w, r, http.StatusBadRequest, err)
			return
		}

		// Send the response
		buf, err := proto.Marshal(resp)
		if err != nil {
			o.writeError(w, r, http.StatusInternalServerError, fmt.Errorf("encoding response failed: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf); err != nil {
			o.Log.Errorf("Writing response to %q failed: %v", r.URL.Path, err)
		}
	}
}

// readBody reads the (optionally compressed) request body while limiting the
// size of the decompressed data. In case of an error the HTTP status code
// to respond with is returned alongside the error.
func (o *OpenTelemetry) readBody(w http.ResponseWriter, r *http.Request) ([]byte, int, error) {
	var reader io.Reader = http.MaxBytesReader(w, r.Body, int64(o.MaxMsgSize))
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Encoding"))) {
	case "", "identity":
	case "gzip":
		gz, err := gzip.NewReader(reader)
		if err != nil {
			if mbe := (*http.MaxBytesError)(nil); errors.As(err, &mbe) {
				return nil, http.StatusRequestEntityTooLarge, fmt.Errorf("request exceeds the limit of %d bytes", int64(o.MaxMsgSize))
			}
			return nil, http.StatusBadRequest, fmt.Errorf("decompressing request failed: %w", err)
		}
		defer gz.Close()
		reader = gz
	default:
		return nil, http.StatusUnsupportedMediaType, fmt.Errorf("unsupported content encoding %q", r.Header.Get("Content-Encoding"))
	}

	body, err := io.ReadAll(io.LimitReader(reader, int64(o.MaxMsgSize)+1))
	if err != nil {
		if mbe := (*http.MaxBytesError)(nil); errors.As(err, &mbe) {
			return nil, http.StatusRequestEntityTooLarge, fmt.Errorf("request exceeds the limit of %d bytes", int64(o.MaxMsgSize))
		}
		return nil, http.StatusBadRequest, fmt.Errorf("reading request failed: %w", err)
	}
	if int64(len(body)) > int64(o.MaxMsgSize) {
		return nil, http.StatusRequestEntityTooLarge, fmt.Errorf("request exceeds the limit of %d bytes", int64(o.MaxMsgSize))
	}

	return body, 0, nil
}

// writeError responds with a protobuf encoded google.rpc.Status message as
// required by the specification.
func (o *OpenTelemetry) writeError(w http.ResponseWriter, r *http.Request, code int, reason error) {
	// Client errors are reported back to the client so only log them at
	// trace level to avoid flooding the log with misbehaving clients
	if code < 500 {
		o.Log.Tracef("Request to %q failed: %v", r.URL.Path, reason)
	} else {
		o.Log.Errorf("Request to %q failed: %v", r.URL.Path, reason)
	}

	grpcCode := codes.InvalidArgument
	if code >= 500 {
		grpcCode = codes.Internal
	}
	buf, err := proto.Marshal(status.New(grpcCode, reason.Error()).Proto())
	if err != nil {
		http.Error(w, reason.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(code)
	if _, err := w.Write(buf); err != nil {
		o.Log.Errorf("Writing error response to %q failed: %v", r.URL.Path, err)
	}
}

func addPeerInfo(r *http.Request) context.Context {
	var p peer.Peer

	if addr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr); err == nil {
		p.Addr = addr
	}
	if v := r.Context().Value(http.LocalAddrContextKey); v != nil {
		if laddr, ok := v.(net.Addr); ok {
			p.LocalAddr = laddr
		}
	}
	if r.TLS != nil {
		p.AuthInfo = credentials.TLSInfo{State: *r.TLS}
	}

	return peer.NewContext(r.Context(), &p)
}
