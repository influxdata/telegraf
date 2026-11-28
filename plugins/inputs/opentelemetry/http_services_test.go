package opentelemetry

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/influxdata/influxdb-observability/otel2influx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestParseServiceAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		protocol string
		address  string
		prefix   string
	}{
		{name: "empty", input: "", protocol: "grpc", address: "0.0.0.0:4317"},
		{name: "no scheme", input: "localhost:1234", protocol: "grpc", address: "localhost:1234"},
		{name: "no scheme no host", input: ":1234", protocol: "grpc", address: ":1234"},
		{name: "no scheme no port", input: "localhost", protocol: "grpc", address: "localhost:4317"},
		{name: "grpc", input: "grpc://127.0.0.1:1234", protocol: "grpc", address: "127.0.0.1:1234"},
		{name: "grpc no port", input: "grpc://127.0.0.1", protocol: "grpc", address: "127.0.0.1:4317"},
		{name: "grpc ipv6", input: "grpc://[::1]:1234", protocol: "grpc", address: "[::1]:1234"},
		{name: "grpc ipv6 no port", input: "grpc://[::1]", protocol: "grpc", address: "[::1]:4317"},
		{name: "grpc slash", input: "grpc://127.0.0.1:1234/", protocol: "grpc", address: "127.0.0.1:1234"},
		{name: "http", input: "http://127.0.0.1:1234", protocol: "http", address: "127.0.0.1:1234"},
		{name: "http no port", input: "http://0.0.0.0", protocol: "http", address: "0.0.0.0:4318"},
		{name: "http uppercase scheme", input: "HTTP://0.0.0.0", protocol: "http", address: "0.0.0.0:4318"},
		{name: "http root path", input: "http://0.0.0.0/", protocol: "http", address: "0.0.0.0:4318"},
		{name: "http base path", input: "http://0.0.0.0/otlp", protocol: "http", address: "0.0.0.0:4318", prefix: "/otlp"},
		{name: "http base path slash", input: "http://0.0.0.0/otlp/", protocol: "http", address: "0.0.0.0:4318", prefix: "/otlp"},
		{name: "http nested path", input: "http://0.0.0.0/otlp/api", protocol: "http", address: "0.0.0.0:4318", prefix: "/otlp/api"},
		{name: "http escaped space", input: "http://0.0.0.0/my%20path", protocol: "http", address: "0.0.0.0:4318", prefix: "/my%20path"},
		{name: "http braces", input: "http://0.0.0.0/{foo}", protocol: "http", address: "0.0.0.0:4318", prefix: "/%7Bfoo%7D"},
		{name: "http escaped slash", input: "http://0.0.0.0/a%2Fb", protocol: "http", address: "0.0.0.0:4318", prefix: "/a%2Fb"},
		{name: "http non-ascii", input: "http://0.0.0.0/\u00e4", protocol: "http", address: "0.0.0.0:4318", prefix: "/%C3%A4"},
		{name: "http dot-dot path", input: "http://0.0.0.0/a/../b", protocol: "http", address: "0.0.0.0:4318", prefix: "/b"},
		{name: "http dot path", input: "http://0.0.0.0/a/./b", protocol: "http", address: "0.0.0.0:4318", prefix: "/a/b"},
		{name: "http double slash path", input: "http://0.0.0.0/a//b", protocol: "http", address: "0.0.0.0:4318", prefix: "/a/b"},
		{name: "http escape root", input: "http://0.0.0.0/../..", protocol: "http", address: "0.0.0.0:4318"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &OpenTelemetry{
				ServiceAddress: tt.input,
			}
			require.NoError(t, plugin.Init())
			require.Equal(t, tt.protocol, plugin.endpoint.Scheme)
			require.Equal(t, tt.address, plugin.endpoint.Host)
			if tt.protocol == "grpc" {
				require.Empty(t, plugin.srvs.tracesPath)
				require.Empty(t, plugin.srvs.metricsPath)
				require.Empty(t, plugin.srvs.logsPath)
				require.Empty(t, plugin.srvs.profilesPath)
				return
			}
			require.Equal(t, tt.prefix+"/v1/traces", plugin.srvs.tracesPath)
			require.Equal(t, tt.prefix+"/v1/metrics", plugin.srvs.metricsPath)
			require.Equal(t, tt.prefix+"/v1/logs", plugin.srvs.logsPath)
			require.Equal(t, tt.prefix+"/v1development/profiles", plugin.srvs.profilesPath)
		})
	}
}

func TestParseServiceAddressFail(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "grpc path",
			address:  "grpc://0.0.0.0/otlp",
			expected: "paths are not supported for gRPC",
		},
		{
			name:     "https without tls",
			address:  "https://:8080",
			expected: "missing TLS settings for https",
		},
		{
			name:     "unknown scheme",
			address:  "udp://0.0.0.0",
			expected: `unsupported scheme "udp"`,
		},
		{
			name:     "query",
			address:  "http://0.0.0.0/?foo=bar",
			expected: "query and fragment are not supported",
		},
		{
			name:     "user",
			address:  "http://user@0.0.0.0",
			expected: "credentials are not supported in service address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &OpenTelemetry{
				ServiceAddress: tt.address,
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestHTTPTraces(t *testing.T) {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", "test")
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetTraceID(pcommon.TraceID{0x5b, 0x8e, 0xff, 0xf7, 0x98, 0x03, 0x81, 0x03, 0xd2, 0x69, 0xb6, 0x33, 0x81, 0x3f, 0xc6, 0x0c})
	span.SetSpanID(pcommon.SpanID{0xee, 0xe1, 0x9b, 0x7e, 0xc3, 0xc1, 0xb1, 0x74})
	span.SetName("testspan")
	span.SetKind(ptrace.SpanKindServer)
	span.SetStartTimestamp(pcommon.Timestamp(1544712660000000000))
	span.SetEndTimestamp(pcommon.Timestamp(1544712661000000000))
	request, err := ptraceotlp.NewExportRequestFromTraces(traces).MarshalProto()
	require.NoError(t, err)

	plugin := &OpenTelemetry{
		ServiceAddress:      "http://127.0.0.1:0",
		SpanDimensions:      otel2influx.DefaultOtelTracesToLineProtocolConfig().SpanDimensions,
		LogRecordDimensions: otel2influx.DefaultOtelLogsToLineProtocolConfig().LogRecordDimensions,
		MaxMsgSize:          config.Size(1024),
		Timeout:             config.Duration(5 * time.Second),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Send the request uncompressed and gzip compressed
	endpoint := "http://" + plugin.listener.Addr().String() + "/v1/traces"
	resp := postHTTP(t, endpoint, "", request)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.Body))
	require.Equal(t, "application/x-protobuf", resp.ContentType)
	require.NoError(t, ptraceotlp.NewExportResponse().UnmarshalProto(resp.Body))

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err = w.Write(request)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	resp = postHTTP(t, endpoint, "gzip", buf.Bytes())
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.Body))
	plugin.Stop()

	require.Empty(t, acc.Errors)
	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 2)
	for _, m := range metrics {
		require.Equal(t, "spans", m.Name())
		tags := m.Tags()
		require.Equal(t, "5b8efff798038103d269b633813fc60c", tags["trace_id"])
		require.Equal(t, "eee19b7ec3c1b174", tags["span_id"])
		require.Equal(t, "test", tags["service.name"])
		require.Equal(t, "testspan", m.Fields()["span.name"])
	}
}

func TestHTTPS(t *testing.T) {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", "test")
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetTraceID(pcommon.TraceID{0x5b, 0x8e, 0xff, 0xf7, 0x98, 0x03, 0x81, 0x03, 0xd2, 0x69, 0xb6, 0x33, 0x81, 0x3f, 0xc6, 0x0c})
	span.SetSpanID(pcommon.SpanID{0xee, 0xe1, 0x9b, 0x7e, 0xc3, 0xc1, 0xb1, 0x74})
	span.SetName("testspan")
	span.SetStartTimestamp(pcommon.Timestamp(1544712660000000000))
	span.SetEndTimestamp(pcommon.Timestamp(1544712661000000000))
	request, err := ptraceotlp.NewExportRequestFromTraces(traces).MarshalProto()
	require.NoError(t, err)

	// Setup and start the plugin with TLS enabled
	pki := testutil.NewPKI("../../../testutil/pki")
	plugin := &OpenTelemetry{
		ServiceAddress:      "https://127.0.0.1:0",
		SpanDimensions:      otel2influx.DefaultOtelTracesToLineProtocolConfig().SpanDimensions,
		LogRecordDimensions: otel2influx.DefaultOtelLogsToLineProtocolConfig().LogRecordDimensions,
		Timeout:             config.Duration(5 * time.Second),
		ServerConfig:        *pki.TLSServerConfig(),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	// Setup a client using the client certificates
	tlsCfg, err := pki.TLSClientConfig().TLSConfig()
	require.NoError(t, err)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsCfg}}
	defer client.CloseIdleConnections()

	// Send the request via HTTPS
	addr := plugin.listener.Addr().String()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "https://"+addr+"/v1/traces", bytes.NewReader(request))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-protobuf")
	resp, err := client.Do(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.NoError(t, ptraceotlp.NewExportResponse().UnmarshalProto(body))

	// Plain HTTP requests must be rejected
	plain := postHTTP(t, "http://"+addr+"/v1/traces", "", request)
	require.Equal(t, http.StatusBadRequest, plain.StatusCode)

	plugin.Stop()
	require.Empty(t, acc.Errors)
	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 1)
	require.Equal(t, "spans", metrics[0].Name())
	require.Equal(t, "testspan", metrics[0].Fields()["span.name"])
}

func TestHTTPBasePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		expected string
	}{
		{name: "none", basePath: "", expected: "/v1/logs"},
		{name: "root", basePath: "/", expected: "/v1/logs"},
		{name: "simple", basePath: "/otlp", expected: "/otlp/v1/logs"},
		{name: "trailing slash", basePath: "/otlp/", expected: "/otlp/v1/logs"},
		{name: "nested", basePath: "/otlp/api", expected: "/otlp/api/v1/logs"},
		{name: "escaped space", basePath: "/my%20path", expected: "/my%20path/v1/logs"},
		{name: "braces", basePath: "/{foo}", expected: "/%7Bfoo%7D/v1/logs"},
		{name: "escaped slash", basePath: "/a%2Fb", expected: "/a%2Fb/v1/logs"},
		{name: "unclean", basePath: "/a/../b//c/.", expected: "/b/c/v1/logs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &OpenTelemetry{
				ServiceAddress:      "http://127.0.0.1:0" + tt.basePath,
				SpanDimensions:      otel2influx.DefaultOtelTracesToLineProtocolConfig().SpanDimensions,
				LogRecordDimensions: otel2influx.DefaultOtelLogsToLineProtocolConfig().LogRecordDimensions,
				MaxMsgSize:          config.Size(1024),
				Timeout:             config.Duration(5 * time.Second),
				Log:                 testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			addr := "http://" + plugin.listener.Addr().String()

			resp := postHTTP(t, addr+tt.expected, "", nil)
			require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.Body))

			if tt.expected != "/v1/logs" {
				resp = postHTTP(t, addr+"/v1/logs", "", nil)
				require.Equal(t, http.StatusNotFound, resp.StatusCode)
			}

			plugin.Stop()
			require.Empty(t, acc.Errors)
		})
	}
}

func TestHTTPErrors(t *testing.T) {
	plugin := &OpenTelemetry{
		ServiceAddress:      "http://127.0.0.1:0",
		SpanDimensions:      otel2influx.DefaultOtelTracesToLineProtocolConfig().SpanDimensions,
		LogRecordDimensions: otel2influx.DefaultOtelLogsToLineProtocolConfig().LogRecordDimensions,
		MaxMsgSize:          config.Size(1024),
		Timeout:             config.Duration(5 * time.Second),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	addr := "http://" + plugin.listener.Addr().String()

	tests := []struct {
		name            string
		method          string
		path            string
		contentType     string
		contentEncoding string
		body            []byte
		expectedCode    int
		expectedStatus  bool
	}{
		{
			name:         "wrong method",
			method:       http.MethodGet,
			path:         "/v1/metrics",
			contentType:  "application/x-protobuf",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "unknown path",
			path:         "/v1/foo",
			contentType:  "application/x-protobuf",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing content type",
			path:         "/v1/metrics",
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:         "unsupported content type",
			path:         "/v1/metrics",
			contentType:  "text/plain",
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:         "json content type",
			path:         "/v1/traces",
			contentType:  "application/json",
			body:         []byte(`{}`),
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:            "unsupported content encoding",
			path:            "/v1/metrics",
			contentType:     "application/x-protobuf",
			contentEncoding: "br",
			expectedCode:    http.StatusUnsupportedMediaType,
			expectedStatus:  true,
		},
		{
			name:            "invalid gzip",
			path:            "/v1/metrics",
			contentType:     "application/x-protobuf",
			contentEncoding: "gzip",
			body:            []byte("not compressed"),
			expectedCode:    http.StatusBadRequest,
			expectedStatus:  true,
		},
		{
			name:           "invalid traces",
			path:           "/v1/traces",
			contentType:    "application/x-protobuf",
			body:           []byte{0xff, 0xff, 0xff},
			expectedCode:   http.StatusBadRequest,
			expectedStatus: true,
		},
		{
			name:           "invalid profiles",
			path:           "/v1development/profiles",
			contentType:    "application/x-protobuf",
			body:           []byte{0xff, 0xff, 0xff},
			expectedCode:   http.StatusBadRequest,
			expectedStatus: true,
		},
		{
			name:           "too large",
			path:           "/v1/traces",
			contentType:    "application/x-protobuf",
			body:           make([]byte, 2048),
			expectedCode:   http.StatusRequestEntityTooLarge,
			expectedStatus: true,
		},
		{
			name:            "too large after decompression",
			path:            "/v1/traces",
			contentType:     "application/x-protobuf",
			contentEncoding: "gzip",
			body:            make([]byte, 2048),
			expectedCode:    http.StatusRequestEntityTooLarge,
			expectedStatus:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}

			body := tt.body
			if tt.contentEncoding == "gzip" {
				var buf bytes.Buffer
				w := gzip.NewWriter(&buf)
				_, err := w.Write(tt.body)
				require.NoError(t, err)
				require.NoError(t, w.Close())
				body = buf.Bytes()
			}

			req, err := http.NewRequestWithContext(t.Context(), method, addr+tt.path, bytes.NewReader(body))
			require.NoError(t, err)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.expectedCode, resp.StatusCode)
			if !tt.expectedStatus {
				return
			}

			// Errors must be reported as protobuf encoded google.rpc.Status
			require.Equal(t, "application/x-protobuf", resp.Header.Get("Content-Type"))
			buf, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var st status.Status
			require.NoError(t, proto.Unmarshal(buf, &st))
			require.NotEmpty(t, st.GetMessage())
			require.NotZero(t, st.GetCode())
		})
	}
}

type httpResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

func postHTTP(t *testing.T, endpoint, contentEncoding string, body []byte) httpResponse {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-protobuf")
	if contentEncoding != "" {
		req.Header.Set("Content-Encoding", contentEncoding)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return httpResponse{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        buf,
	}
}
