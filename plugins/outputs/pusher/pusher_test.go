package pusher

import (
	"fmt"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	// Set up test server
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(res, `{"results":[{}]}`)
	}))
	defer server.Close()

	s, _ := serializers.NewJsonSerializer()
	p := &Pusher{
		AppId:       "test_app",
		AppKey:      "test_key",
		AppSecret:   "test_secret",
		ChannelName: "test_channel",
		Host:        server.Listener.Addr().String(),
		Secure:      false,
		serializer:  s,
	}
	err := p.Connect()
	require.NoError(t, err)

	metrics := testutil.MockMetrics()

	err = p.Write(metrics)
	require.NoError(t, err)

}
