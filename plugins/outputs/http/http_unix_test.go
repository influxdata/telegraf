package http

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/stretchr/testify/require"
)

func TestUnixStatusCode(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.NotFoundHandler())
	defer ts.Close()

	socket := "./test.sock"
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatalf("Failed to create unix socket: %s", socket)
	}

	ts.Listener = listener
	ts.Start()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name       string
		plugin     *HTTP
		statusCode int
		errFunc    func(t *testing.T, err error)
	}{
		{
			name: "success",
			plugin: &HTTP{
				URL: u.String(),
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					UnixSocket: socket,
				},
			},
			statusCode: http.StatusOK,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "1xx status is an error",
			plugin: &HTTP{
				URL: u.String(),
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					UnixSocket: socket,
				},
			},
			statusCode: 103,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "3xx status is an error",
			plugin: &HTTP{
				URL: u.String(),
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					UnixSocket: socket,
				},
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "4xx status is an error",
			plugin: &HTTP{
				URL: u.String(),
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					UnixSocket: socket,
				},
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			tt.errFunc(t, err)
		})
	}
}
