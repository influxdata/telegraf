package socket_listener

import (
	"crypto/tls"
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
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestSocketListener(t *testing.T) {
	messages := [][]byte{
		[]byte("test,foo=bar v=1i 123456789\ntest,foo=baz v=2i 123456790\n"),
		[]byte("test,foo=zab v=3i 123456791\n"),
	}
	expected := []telegraf.Metric{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := strings.TrimSuffix(tt.schema, "+tls")

			// Prepare the address and socket if needed
			var serverAddr string
			switch proto {
			case "tcp", "udp":
				serverAddr = "127.0.0.1:0"
			case "unix", "unixgram":
				if runtime.GOOS == "windows" {
					t.Skip("Skipping on Windows, as unixgram sockets are not supported")
				}

				// Create a socket
				sock, err := os.CreateTemp("", "sock-")
				require.NoError(t, err)
				defer sock.Close()
				defer os.Remove(sock.Name())
				serverAddr = sock.Name()
			}

			// Setup plugin according to test specification
			plugin := &SocketListener{
				Log:             &testutil.Logger{},
				ServiceAddress:  proto + "://" + serverAddr,
				ContentEncoding: tt.encoding,
				ReadBufferSize:  tt.buffersize,
			}
			if strings.HasSuffix(tt.schema, "tls") {
				plugin.ServerConfig = *pki.TLSServerConfig()
			}
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			plugin.SetParser(parser)

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Setup the client for submitting data
			var client net.Conn
			switch tt.schema {
			case "tcp":
				var err error
				addr := plugin.listener.addr().String()
				client, err = net.Dial("tcp", addr)
				require.NoError(t, err)
			case "tcp+tls":
				addr := plugin.listener.addr().String()
				tlscfg, err := pki.TLSClientConfig().TLSConfig()
				require.NoError(t, err)
				client, err = tls.Dial("tcp", addr, tlscfg)
				require.NoError(t, err)
			case "udp":
				var err error
				addr := plugin.listener.addr().String()
				client, err = net.Dial("udp", addr)
				require.NoError(t, err)
			case "unix":
				var err error
				client, err = net.Dial("unix", serverAddr)
				require.NoError(t, err)
			case "unix+tls":
				tlscfg, err := pki.TLSClientConfig().TLSConfig()
				require.NoError(t, err)
				tlscfg.InsecureSkipVerify = true
				client, err = tls.Dial("unix", serverAddr, tlscfg)
				require.NoError(t, err)
			case "unixgram":
				var err error
				client, err = net.Dial("unixgram", serverAddr)
				require.NoError(t, err)
			default:
				require.Failf(t, "schema %q not supported in test", tt.schema)
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
			require.Eventually(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return acc.NMetrics() >= uint64(len(expected))
			}, time.Second, 100*time.Millisecond, "did not receive metrics")
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}
