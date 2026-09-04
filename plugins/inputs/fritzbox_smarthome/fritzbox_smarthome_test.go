package fritzbox_smarthome

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-fritzsmarthome/mock"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestValidDefaultConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("sample.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*FritzboxSmarthome)
	require.True(t, ok)

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify everything is setup according to plugin defaults
	require.ElementsMatch(t, []string{"http://user:password@fritz.box/"}, f.URLs)
	require.Equal(t, config.Duration(10*time.Second), f.Timeout)
	require.Empty(t, f.TLSKeyPwd)
	require.False(t, f.InsecureSkipVerify)
}

func TestValidCustomConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/valid.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*FritzboxSmarthome)
	require.True(t, ok)

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify everything is setup according to the config file
	require.ElementsMatch(t, []string{"http://boxuser1:boxpassword1@fritz1.box/", "http://:boxpassword2@fritz2.box/"}, f.URLs)
	require.Equal(t, config.Duration(60*time.Second), f.Timeout)
	require.Equal(t, "secret", f.TLSKeyPwd)
	require.True(t, f.InsecureSkipVerify)
}

func TestInvalidURLsConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/invalid_urls.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*FritzboxSmarthome)
	require.True(t, ok)

	// Verify Init failure
	require.EqualError(t, f.Init(), `parsing device URL "::" failed: parse "::": missing protocol scheme`)
}

func TestGather(t *testing.T) {
	// Start mock server
	serverMock := mock.Start("testdata/gather/mock")
	defer serverMock.Stop(t.Context())

	// Register the plugin
	inputs.Add("fritzbox_smarthome", func() telegraf.Input {
		return &FritzboxSmarthome{Timeout: config.Duration(10 * time.Second)}
	})

	// Load plugin from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/gather/telegraf.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*FritzboxSmarthome)
	require.True(t, ok)

	// Target plugin at mock server
	f.URLs = []string{serverMock.ConnectURL().String()}
	f.Log = &testutil.Logger{Name: "fritzbox_smarthome"}

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify successfull Gather
	acc := &testutil.Accumulator{}
	require.NoError(t, acc.GatherError(f.Gather))

	// Load expexected metrics
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	expectedMetrics, err := testutil.ParseMetricsFromFile("testdata/gather/expected.out", parser)
	require.NoError(t, err)

	// Verify metrics are as expected
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.IgnoreType(), testutil.SortMetrics())
}

func TestUnexpectedResponse(t *testing.T) {
	// Start dummy API backend
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/smarthome/overview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body>Please login</body></html>`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Register the plugin
	inputs.Add("fritzbox_smarthome", func() telegraf.Input {
		return &FritzboxSmarthome{Timeout: config.Duration(10 * time.Second)}
	})

	// Load plugin from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/gather/telegraf.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*FritzboxSmarthome)
	require.True(t, ok)

	// Target plugin at dummy server
	f.URLs = []string{server.URL}
	f.Log = &testutil.Logger{Name: "fritzbox_smarthome"}

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify Gather fails with error
	acc := &testutil.Accumulator{}
	err := acc.GatherError(f.Gather)
	require.ErrorContains(t, err, "empty or unexpected response")
}
