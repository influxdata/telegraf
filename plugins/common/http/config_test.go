package httpconfig

import (
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/testutil"
)

func TestCreateTransportSettings(t *testing.T) {
	cfg := &TransportConfig{
		IdleConnTimeout:       config.Duration(30 * time.Second),
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
		ResponseHeaderTimeout: config.Duration(5 * time.Second),
	}
	transport, err := cfg.CreateTransport()
	require.NoError(t, err)

	require.Equal(t, 30*time.Second, transport.IdleConnTimeout)
	require.Equal(t, 10, transport.MaxIdleConns)
	require.Equal(t, 2, transport.MaxIdleConnsPerHost)
	require.Equal(t, 5*time.Second, transport.ResponseHeaderTimeout)
}

func TestCreateTransportProxy(t *testing.T) {
	requested := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case requested <- r.URL.String():
		default:
		}
	}))
	defer ts.Close()

	cfg := &TransportConfig{
		HTTPProxy: proxy.HTTPProxy{HTTPProxyURL: ts.URL},
	}
	transport, err := cfg.CreateTransport()
	require.NoError(t, err)

	client := &http.Client{Transport: transport}
	resp, err := client.Get("http://telegraf.invalid/metrics")
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, "http://telegraf.invalid/metrics", <-requested)
}

func TestCreateTransportLocalAddress(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Binding to 127.0.0.2 requires a loopback alias outside of Linux")
	}

	remote := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case remote <- r.RemoteAddr:
		default:
		}
	}))
	defer ts.Close()

	cfg := &TransportConfig{LocalAddress: "127.0.0.2"}
	transport, err := cfg.CreateTransport()
	require.NoError(t, err)

	client := &http.Client{Transport: transport}
	resp, err := client.Get(ts.URL)
	require.NoError(t, err)
	resp.Body.Close()

	host, _, err := net.SplitHostPort(<-remote)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.2", host)
}

func TestCreateTransportInvalidLocalAddress(t *testing.T) {
	cfg := &TransportConfig{LocalAddress: "::1"}
	_, err := cfg.CreateTransport()
	require.ErrorContains(t, err, "invalid local address")
}

func TestCreateClientTimeout(t *testing.T) {
	cfg := &HTTPClientConfig{}
	client, err := cfg.CreateClient(t.Context(), testutil.Logger{})
	require.NoError(t, err)
	require.Equal(t, 5*time.Second, client.Timeout)

	cfg.Timeout = config.Duration(20 * time.Second)
	client, err = cfg.CreateClient(t.Context(), testutil.Logger{})
	require.NoError(t, err)
	require.Equal(t, 20*time.Second, client.Timeout)
}
