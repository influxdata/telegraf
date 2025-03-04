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
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConfig(t *testing.T) {
	// Verify plugin can be loaded from config
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/fritzbox.conf"))
	require.Len(t, conf.Inputs, 1)
	f, ok := conf.Inputs[0].Input.(*Fritzbox)
	require.True(t, ok)

	// Verify successful Init
	require.NoError(t, f.Init())

	// Verify everything is setup according to config file
	require.ElementsMatch(t, []string{"http://boxuser:boxpassword@fritz.box:49000/", "http://:repeaterpassword@fritz.repeater:49000/"}, f.URLs)
	require.Equal(t, []string{"device", "wan", "ppp", "dsl", "wlan", "hosts"}, f.Collect)
	require.Equal(t, config.Duration(60*time.Second), f.Timeout)
	require.Equal(t, "secret", f.TLSKeyPwd)
	require.True(t, f.InsecureSkipVerify)
}

func TestGather(t *testing.T) {
	// Start mock server
	tr064Server := mock.Start("testdata/mock",
		mock.ServiceMockFromFile("/deviceinfo", "testdata/mock/DeviceInfo.xml"),
		mock.ServiceMockFromFile("/wancommonifconfig", "testdata/mock/WANCommonInterfaceConfig1.xml"),
		mock.ServiceMockFromFile("/WANCommonIFC1", "testdata/mock/WANCommonInterfaceConfig2.xml"),
		mock.ServiceMockFromFile("/wanpppconn", "testdata/mock/WANPPPConnection.xml"),
		mock.ServiceMockFromFile("/wandslifconfig", "testdata/mock/WANDSLInterfaceConfig.xml"),
		mock.ServiceMockFromFile("/wlanconfig", "testdata/mock/WLANConfiguration.xml"),
		mock.ServiceMockFromFile("/hosts", "testdata/mock/Hosts.xml"),
	)
	defer tr064Server.Shutdown()

	// Test the different 'collect' options identified via the folders in testcases
	testcases, err := os.ReadDir("testdata/testcases")
	require.NoError(t, err)
	for _, testcase := range testcases {
		// Only handle folders
		if !testcase.IsDir() {
			continue
		}

		t.Run(testcase.Name(), func(t *testing.T) {
			// Create plugin targeting the mock and collecting the testcase's metrics
			f := &Fritzbox{
				URLs:    []string{tr064Server.Server().String()},
				Collect: []string{testcase.Name()},
				Log:     &testutil.Logger{Name: "fritzbox"},
			}

			// Verify successful Init
			require.NoError(t, f.Init())

			// Verify successfull Gather
			acc := &testutil.Accumulator{}
			require.NoError(t, acc.GatherError(f.Gather))

			// Verify collected metrics are as expected
			actualMetrics := acc.GetTelegrafMetrics()
			expectedMetricsType := telegraf.Untyped
			if len(actualMetrics) > 0 {
				expectedMetricsType = actualMetrics[0].Type()
			}
			expectedMetrics := loadExpectedMetrics(t, filepath.Join("testdata/testcases", testcase.Name(), "metrics.txt"), expectedMetricsType)
			testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func loadExpectedMetrics(t *testing.T, file string, vt telegraf.ValueType) []telegraf.Metric {
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	expectedMetrics, err := testutil.ParseMetricsFromFile(file, parser)
	require.NoError(t, err)
	for index := range expectedMetrics {
		expectedMetrics[index].SetType(vt)
	}
	return expectedMetrics
}
