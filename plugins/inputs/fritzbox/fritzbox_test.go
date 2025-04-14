package fritzbox

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-tr064/mock"

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
	f, ok := conf.Inputs[0].Input.(*Fritzbox)
	require.True(t, ok)

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify everything is setup according to plugin defaults
	require.ElementsMatch(t, []string{"http://user:password@fritz.box:49000/"}, f.URLs)
	require.Equal(t, []string{"device", "wan", "ppp", "dsl", "wlan"}, f.Collect)
	require.Equal(t, config.Duration(10*time.Second), f.Timeout)
	require.Empty(t, f.TLSKeyPwd)
	require.False(t, f.InsecureSkipVerify)
}

func TestValidCustomConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/valid.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*Fritzbox)
	require.True(t, ok)

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify everything is setup according to the config file
	require.ElementsMatch(t, []string{"http://boxuser:boxpassword@fritz.box:49000/", "http://:repeaterpassword@fritz.repeater:49000/"}, f.URLs)
	require.Equal(t, []string{"device", "wan", "ppp", "dsl", "wlan", "hosts"}, f.Collect)
	require.Equal(t, config.Duration(60*time.Second), f.Timeout)
	require.Equal(t, "secret", f.TLSKeyPwd)
	require.True(t, f.InsecureSkipVerify)
}

func TestInvalidURLsConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/invalid_urls.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*Fritzbox)
	require.True(t, ok)

	// Verify Init failure
	require.EqualError(t, f.Init(), `parsing device URL "::" failed: parse "::": missing protocol scheme`)
}

func TestInvalidCollectConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/invalid_collect.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*Fritzbox)
	require.True(t, ok)

	// Verify Init failure
	require.EqualError(t, f.Init(), `invalid service "undefined" in collect parameter`)
}

func TestCases(t *testing.T) {
	// Get all testcase directories
	testcases, err := os.ReadDir("testdata/testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("fritzbox", func() telegraf.Input {
		return &Fritzbox{Timeout: config.Duration(10 * time.Second)}
	})

	for _, testcase := range testcases {
		// Only handle folders
		if !testcase.IsDir() {
			continue
		}

		t.Run(testcase.Name(), func(t *testing.T) {
			testcaseDir := filepath.Join("testdata", "testcases", testcase.Name())
			configFile := filepath.Join(testcaseDir, "telegraf.conf")
			mockDir := filepath.Join(testcaseDir, "mock")
			expectedMetricsFile := filepath.Join(testcaseDir, "expected.out")

			// Setup the services to mock (one per sub-folder of mockDir)
			services, err := os.ReadDir(mockDir)
			require.NoError(t, err)
			serviceMocks := make([]*mock.ServiceMock, 0, len(services))
			for _, service := range services {
				// Ignore the mock files
				if !testcase.IsDir() {
					continue
				}
				serviceMock := mock.ServiceMockFromFile("/"+service.Name(), filepath.Join(mockDir, service.Name(), "response.xml"))
				serviceMocks = append(serviceMocks, serviceMock)
			}

			// Start testcase mock server
			tr064Server := mock.Start(mockDir, serviceMocks...)
			defer tr064Server.Shutdown()

			// Load plugin from config
			conf := config.NewConfig()
			require.NoError(t, conf.LoadConfig(configFile))
			require.Len(t, conf.Inputs, 1)
			f, ok := conf.Inputs[0].Input.(*Fritzbox)
			require.True(t, ok)

			// Target plugin at mock server
			f.URLs = []string{tr064Server.Server().String()}
			f.Log = &testutil.Logger{Name: "fritzbox"}

			// Verify successful Init
			require.NoError(t, f.Init())

			// Verify successfull Gather
			acc := &testutil.Accumulator{}
			require.NoError(t, acc.GatherError(f.Gather))

			// Load expexected metrics
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expectedMetrics, err := testutil.ParseMetricsFromFile(expectedMetricsFile, parser)
			require.NoError(t, err)

			// Verify metrics are as expected
			testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.IgnoreType(), testutil.SortMetrics())
		})
	}
}
