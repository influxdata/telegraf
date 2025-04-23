package prometheus

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

type FakeHTTPClient struct {
	testing *testing.T
}

func (client *FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp := http.Response{
		StatusCode: http.StatusOK,
	}
	targetsResponse := []httpSDOutput{{
		Targets: []string{"localhost:8081", "localhost:8082", "localhost:8083"},
		Labels:  map[string]string{"host": "localhost"},
	}}
	data, err := json.Marshal(targetsResponse)
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		return &resp, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(data))

	require.Equal(client.testing, http.MethodGet, req.Method)
	return &resp, nil
}

func TestHttpSD(t *testing.T) {
	client := FakeHTTPClient{
		testing: t,
	}
	p := &Prometheus{
		Log: testutil.Logger{},
		HTTPSDConfig: HTTPSDConfig{
			Enabled: true,
			URL:     "localhost:8080/service-discovery",
		},
	}
	err := p.Init()
	require.NoError(t, err)

	httpSDUrl, err := url.Parse(p.HTTPSDConfig.URL)
	require.NoError(t, err)

	err = p.refreshHTTPServices(httpSDUrl, &client)
	require.NoError(t, err)

	// there should be 3 targets returned by the service discovery endpoint
	require.Len(t, p.httpServices, 3)
}
