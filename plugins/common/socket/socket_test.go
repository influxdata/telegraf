package socket

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestListenData(t *testing.T) {
	messages := [][]byte{
		[]byte("test,foo=bar v=1i 123456789\ntest,foo=baz v=2i 123456790\n"),
		[]byte("test,foo=zab v=3i 123456791\n"),
	}
	expectedTemplates := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{"v": int64(1)},
			time.Unix(0, 123456789),
		),
		metric.New(
			"test",
			map[string]string{"foo": "baz"},
			map[string]interface{}{"v": int64(2)},
			time.Unix(0, 123456790),
		),
		metric.New(
			"test",
			map[string]string{"foo": "zab"},
			map[string]interface{}{"v": int64(3)},
			time.Unix(0, 123456791),
		),
	}

	tests := []struct {
		name       string
		schema     string
		buffersize config.Size
		encoding   string
	}{
		{
			name:       "TCP",
			schema:     "tcp",
			buffersize: config.Size(1024),
		},
		{
			name:   "TCP with TLS",
			schema: "tcp+tls",
		},
		{
			name:       "TCP with gzip encoding",
			schema:     "tcp",
			buffersize: config.Size(1024),
			encoding:   "gzip",
		},
		{
			name:       "UDP",
			schema:     "udp",
			buffersize: config.Size(1024),
		},
		{
			name:       "UDP with gzip encoding",
			schema:     "udp",
			buffersize: config.Size(1024),
			encoding:   "gzip",
		},
		{
			name:       "unix socket",
			schema:     "unix",
			buffersize: config.Size(1024),
		},
		{
			name:   "unix socket with TLS",
			schema: "unix+tls",
		},
		{
			name:     "unix socket with gzip encoding",
			schema:   "unix",
			encoding: "gzip",
		},
		{
			name:       "unixgram socket",
			schema:     "unixgram",
			buffersize: config.Size(1024),
		},
	}

	serverTLS := pki.TLSServerConfig()
	clientTLS := pki.TLSClientConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := strings.TrimSuffix(tt.schema, "+tls")

			// Prepare the address and socket if needed
			var sockPath string
			var serviceAddress string
			var tlsCfg *tls.Config
			switch proto {
			case "tcp", "udp":
				serviceAddress = proto + "://" + "127.0.0.1:0"
			case "unix", "unixgram":
				if runtime.GOOS == "windows" {
					t.Skip("Skipping on Windows, as unixgram sockets are not supported")
				}

				// Create a socket
				sockPath = testutil.TempSocket(t)
				f, err := os.Create(sockPath)
				require.NoError(t, err)
				defer f.Close()
				serviceAddress = proto + "://" + sockPath
			}

			// Setup the configuration according to test specification
			cfg := &Config{
				ContentEncoding: tt.encoding,
				ReadBufferSize:  tt.buffersize,
			}
			if strings.HasSuffix(tt.schema, "tls") {
				cfg.ServerConfig = *serverTLS
				var err error
				tlsCfg, err = clientTLS.TLSConfig()
				require.NoError(t, err)
			}

			// Create the socket
			sock, err := cfg.NewSocket(serviceAddress, &SplitConfig{}, &testutil.Logger{})
			require.NoError(t, err)

			// Create callbacks
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			var acc testutil.Accumulator
			onData := func(remote net.Addr, data []byte) {
				m, err := parser.Parse(data)
				require.NoError(t, err)
				addr, _, err := net.SplitHostPort(remote.String())
				if err != nil {
					addr = remote.String()
				}
				for i := range m {
					m[i].AddTag("source", addr)
				}
				acc.AddMetrics(m)
			}
			onError := func(err error) {
				acc.AddError(err)
			}

			// Start the listener
			require.NoError(t, sock.Setup())
			sock.Listen(onData, onError)
			defer sock.Close()

			addr := sock.Address()

			// Create a noop client
			// Server is async, so verify no errors at the end.
			client, err := createClient(serviceAddress, addr, tlsCfg)
			require.NoError(t, err)
			require.NoError(t, client.Close())

			// Setup the client for submitting data
			client, err = createClient(serviceAddress, addr, tlsCfg)
			require.NoError(t, err)

			// Conditionally add the source address to the expectation
			expected := make([]telegraf.Metric, 0, len(expectedTemplates))
			for _, tmpl := range expectedTemplates {
				m := tmpl.Copy()
				switch proto {
				case "tcp", "udp":
					laddr := client.LocalAddr().String()
					addr, _, err := net.SplitHostPort(laddr)
					if err != nil {
						addr = laddr
					}
					m.AddTag("source", addr)
				case "unix", "unixgram":
					m.AddTag("source", sockPath)
				}
				expected = append(expected, m)
			}

			// Send the data with the correct encoding
			encoder, err := internal.NewContentEncoder(tt.encoding)
			require.NoError(t, err)

			for i, msg := range messages {
				m, err := encoder.Encode(msg)
				require.NoErrorf(t, err, "encoding failed for msg %d", i)
				_, err = client.Write(m)
				require.NoErrorf(t, err, "sending msg %d failed", i)
			}

			// Test the resulting metrics and compare against expected results
			require.Eventuallyf(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return acc.NMetrics() >= uint64(len(expected))
			}, time.Second, 100*time.Millisecond, "did not receive metrics (%d)", acc.NMetrics())

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}

func TestListenConnection(t *testing.T) {
	messages := [][]byte{
		[]byte("test,foo=bar v=1i 123456789\ntest,foo=baz v=2i 123456790\n"),
		[]byte("test,foo=zab v=3i 123456791\n"),
	}
	expectedTemplates := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{"v": int64(1)},
			time.Unix(0, 123456789),
		),
		metric.New(
			"test",
			map[string]string{"foo": "baz"},
			map[string]interface{}{"v": int64(2)},
			time.Unix(0, 123456790),
		),
		metric.New(
			"test",
			map[string]string{"foo": "zab"},
			map[string]interface{}{"v": int64(3)},
			time.Unix(0, 123456791),
		),
	}

	tests := []struct {
		name       string
		schema     string
		buffersize config.Size
		encoding   string
	}{
		{
			name:       "TCP",
			schema:     "tcp",
			buffersize: config.Size(1024),
		},
		{
			name:   "TCP with TLS",
			schema: "tcp+tls",
		},
		{
			name:       "TCP with gzip encoding",
			schema:     "tcp",
			buffersize: config.Size(1024),
			encoding:   "gzip",
		},
		{
			name:       "UDP",
			schema:     "udp",
			buffersize: config.Size(1024),
		},
		{
			name:       "UDP with gzip encoding",
			schema:     "udp",
			buffersize: config.Size(1024),
			encoding:   "gzip",
		},
		{
			name:       "unix socket",
			schema:     "unix",
			buffersize: config.Size(1024),
		},
		{
			name:   "unix socket with TLS",
			schema: "unix+tls",
		},
		{
			name:     "unix socket with gzip encoding",
			schema:   "unix",
			encoding: "gzip",
		},
		{
			name:       "unixgram socket",
			schema:     "unixgram",
			buffersize: config.Size(1024),
		},
	}

	serverTLS := pki.TLSServerConfig()
	clientTLS := pki.TLSClientConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := strings.TrimSuffix(tt.schema, "+tls")

			// Prepare the address and socket if needed
			var sockPath string
			var serviceAddress string
			var tlsCfg *tls.Config
			switch proto {
			case "tcp", "udp":
				serviceAddress = proto + "://" + "127.0.0.1:0"
			case "unix", "unixgram":
				if runtime.GOOS == "windows" {
					t.Skip("Skipping on Windows, as unixgram sockets are not supported")
				}

				// Create a socket
				sockPath = testutil.TempSocket(t)
				f, err := os.Create(sockPath)
				require.NoError(t, err)
				defer f.Close()
				serviceAddress = proto + "://" + sockPath
			}

			// Setup the configuration according to test specification
			cfg := &Config{
				ContentEncoding: tt.encoding,
				ReadBufferSize:  tt.buffersize,
			}
			if strings.HasSuffix(tt.schema, "tls") {
				cfg.ServerConfig = *serverTLS
				var err error
				tlsCfg, err = clientTLS.TLSConfig()
				require.NoError(t, err)
			}

			// Create the socket
			sock, err := cfg.NewSocket(serviceAddress, &SplitConfig{}, &testutil.Logger{})
			require.NoError(t, err)

			// Create callbacks
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			var acc testutil.Accumulator
			onConnection := func(remote net.Addr, reader io.ReadCloser) {
				data, err := io.ReadAll(reader)
				require.NoError(t, err)
				m, err := parser.Parse(data)
				require.NoError(t, err)
				addr, _, err := net.SplitHostPort(remote.String())
				if err != nil {
					addr = remote.String()
				}
				for i := range m {
					m[i].AddTag("source", addr)
				}
				acc.AddMetrics(m)
			}
			onError := func(err error) {
				acc.AddError(err)
			}

			// Start the listener
			require.NoError(t, sock.Setup())
			sock.ListenConnection(onConnection, onError)
			defer sock.Close()

			addr := sock.Address()

			// Create a noop client
			// Server is async, so verify no errors at the end.
			client, err := createClient(serviceAddress, addr, tlsCfg)
			require.NoError(t, err)
			require.NoError(t, client.Close())

			// Setup the client for submitting data
			client, err = createClient(serviceAddress, addr, tlsCfg)
			require.NoError(t, err)

			// Conditionally add the source address to the expectation
			expected := make([]telegraf.Metric, 0, len(expectedTemplates))
			for _, tmpl := range expectedTemplates {
				m := tmpl.Copy()
				switch proto {
				case "tcp", "udp":
					laddr := client.LocalAddr().String()
					addr, _, err := net.SplitHostPort(laddr)
					if err != nil {
						addr = laddr
					}
					m.AddTag("source", addr)
				case "unix", "unixgram":
					m.AddTag("source", sockPath)
				}
				expected = append(expected, m)
			}

			// Send the data with the correct encoding
			encoder, err := internal.NewContentEncoder(tt.encoding)
			require.NoError(t, err)

			for i, msg := range messages {
				m, err := encoder.Encode(msg)
				require.NoErrorf(t, err, "encoding failed for msg %d", i)
				_, err = client.Write(m)
				require.NoErrorf(t, err, "sending msg %d failed", i)
			}
			client.Close()

			// Test the resulting metrics and compare against expected results
			require.Eventuallyf(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return acc.NMetrics() >= uint64(len(expected))
			}, time.Second, 100*time.Millisecond, "did not receive metrics (%d)", acc.NMetrics())

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}

func TestClosingConnections(t *testing.T) {
	// Setup the configuration
	cfg := &Config{
		ReadBufferSize: 1024,
	}

	// Create the socket
	serviceAddress := "tcp://127.0.0.1:0"
	logger := &testutil.CaptureLogger{}
	sock, err := cfg.NewSocket(serviceAddress, &SplitConfig{}, logger)
	require.NoError(t, err)

	// Create callbacks
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	var acc testutil.Accumulator
	onData := func(_ net.Addr, data []byte) {
		m, err := parser.Parse(data)
		require.NoError(t, err)
		acc.AddMetrics(m)
	}
	onError := func(err error) {
		acc.AddError(err)
	}

	// Start the listener
	require.NoError(t, sock.Setup())
	sock.Listen(onData, onError)
	defer sock.Close()

	addr := sock.Address()

	// Create a noop client
	client, err := createClient(serviceAddress, addr, nil)
	require.NoError(t, err)

	_, err = client.Write([]byte("test value=42i\n"))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= 1
	}, time.Second, 100*time.Millisecond, "did not receive metric")

	// This has to be a stream-listener...
	listener, ok := sock.listener.(*streamListener)
	require.True(t, ok)
	listener.Lock()
	conns := len(listener.connections)
	listener.Unlock()
	require.NotZero(t, conns)

	sock.Close()

	// Verify that plugin.Stop() closed the client's connection
	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	buf := []byte{1}
	_, err = client.Read(buf)
	require.Equal(t, err, io.EOF)

	require.Empty(t, logger.Errors())
	require.Empty(t, logger.Warnings())
}

func TestNoSplitter(t *testing.T) {
	messages := [][]byte{
		[]byte("test,foo=bar v"),
		[]byte("=1i 123456789\ntest,foo=baz v=2i 123456790\ntest,foo=zab v=3i 123456791\n"),
	}
	expectedTemplates := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{"v": int64(1)},
			time.Unix(0, 123456789),
		),
		metric.New(
			"test",
			map[string]string{"foo": "baz"},
			map[string]interface{}{"v": int64(2)},
			time.Unix(0, 123456790),
		),
		metric.New(
			"test",
			map[string]string{"foo": "zab"},
			map[string]interface{}{"v": int64(3)},
			time.Unix(0, 123456791),
		),
	}

	// Prepare the address and socket if needed
	serviceAddress := "tcp://127.0.0.1:0"

	// Setup the configuration according to test specification
	cfg := &Config{}

	// Create the socket
	sock, err := cfg.NewSocket(serviceAddress, nil, &testutil.Logger{})
	require.NoError(t, err)

	// Create callbacks
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	var acc testutil.Accumulator
	onConnection := func(remote net.Addr, reader io.ReadCloser) {
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		m, err := parser.Parse(data)
		require.NoError(t, err)
		addr, _, err := net.SplitHostPort(remote.String())
		if err != nil {
			addr = remote.String()
		}
		for i := range m {
			m[i].AddTag("source", addr)
		}
		acc.AddMetrics(m)
	}
	onError := func(err error) {
		acc.AddError(err)
	}

	// Start the listener
	require.NoError(t, sock.Setup())
	sock.ListenConnection(onConnection, onError)
	defer sock.Close()

	addr := sock.Address()

	// Create a noop client
	// Server is async, so verify no errors at the end.
	client, err := createClient(serviceAddress, addr, nil)
	require.NoError(t, err)
	require.NoError(t, client.Close())

	// Setup the client for submitting data
	client, err = createClient(serviceAddress, addr, nil)
	require.NoError(t, err)

	// Conditionally add the source address to the expectation
	expected := make([]telegraf.Metric, 0, len(expectedTemplates))
	for _, tmpl := range expectedTemplates {
		m := tmpl.Copy()
		laddr := client.LocalAddr().String()
		addr, _, err := net.SplitHostPort(laddr)
		if err != nil {
			addr = laddr
		}
		m.AddTag("source", addr)
		expected = append(expected, m)
	}

	// Send the data
	for i, msg := range messages {
		_, err = client.Write(msg)
		time.Sleep(100 * time.Millisecond)
		require.NoErrorf(t, err, "sending msg %d failed", i)
	}
	client.Close()

	// Test the resulting metrics and compare against expected results
	require.Eventuallyf(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= uint64(len(expected))
	}, time.Second, 100*time.Millisecond, "did not receive metrics (%d)", acc.NMetrics())

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func createClient(endpoint string, addr net.Addr, tlsCfg *tls.Config) (net.Conn, error) {
	// Determine the protocol in a crude fashion
	parts := strings.SplitN(endpoint, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid endpoint %q", endpoint)
	}
	protocol := parts[0]

	if tlsCfg == nil {
		return net.Dial(protocol, addr.String())
	}

	if protocol == "unix" {
		tlsCfg.InsecureSkipVerify = true
	}
	return tls.Dial(protocol, addr.String(), tlsCfg)
}
