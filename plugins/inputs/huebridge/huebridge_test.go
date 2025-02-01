package huebridge

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-hue/mock"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConfig(t *testing.T) {
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfig("testdata/conf/huebridge.conf"))
	require.Len(t, conf.Inputs, 1)
	h, ok := conf.Inputs[0].Input.(*HueBridge)
	require.True(t, ok)

	require.NoError(t, h.Init())

	require.Len(t, h.Bridges, 4)
	require.NotEmpty(t, h.RemoteClientId)
	require.NotEmpty(t, h.RemoteClientSecret)
	require.NotEmpty(t, h.RemoteCallbackUrl)
	require.NotEmpty(t, h.RemoteTokenDir)
	require.Len(t, h.RoomAssignments, 2)
	require.Equal(t, config.Duration(60*time.Second), h.Timeout)
	require.Equal(t, "secret", h.TLSKeyPwd)
	require.True(t, h.InsecureSkipVerify)
}

func TestInitSuccess(t *testing.T) {
	h := &HueBridge{
		Bridges: []string{
			"address://12345678:secret@localhost/",
			"cloud://12345678:secret@localhost/discovery/",
			"mdns://12345678:secret@/",
			"remote://12345678:secret@localhost/",
		},
		RemoteClientConfig: RemoteClientConfig{
			RemoteClientId:     mock.MockClientId,
			RemoteClientSecret: mock.MockClientSecret,
			RemoteTokenDir:     ".",
		},
		ClientConfig: tls.ClientConfig{
			InsecureSkipVerify: true,
		},
		Timeout: config.Duration(10 * time.Second),
		Log:     &testutil.Logger{Name: "huebridge"},
	}

	require.NoError(t, h.Init())

	require.Len(t, h.configuredBridges, 4)
}

func TestInitIgnoreInvalidUrls(t *testing.T) {
	// The following URLs must all be ignored during Init
	h := &HueBridge{
		Bridges: []string{
			"invalid://12345678:secret@invalid-scheme.net/",
			"address://12345678@missing-password.net/",
			"cloud://12345678@missing-password.net/",
			"mdns://12345678@missing-password.net/",
			"remote://12345678@missing-password.net/",
			"remote://12345678:secret@missing-remote-config.net/",
		},
		Timeout: config.Duration(10 * time.Second),
		Log:     &testutil.Logger{Name: "huebridge"},
	}

	require.NoError(t, h.Init())

	require.Len(t, h.configuredBridges, 0)
}

func TestGatherLocal(t *testing.T) {
	bridgeMock := mock.Start()
	require.NotNil(t, bridgeMock)
	defer bridgeMock.Shutdown()

	h := &HueBridge{
		Bridges: []string{
			fmt.Sprintf("address://%s:%s@%s/", mock.MockBridgeId, mock.MockBridgeUsername, bridgeMock.Server().Host),
		},
		RoomAssignments: map[string]string{"Name#7": "Name#15"},
		Timeout:         config.Duration(10 * time.Second),
		Log:             &testutil.Logger{Name: "huebridge"},
	}

	require.NoError(t, h.Init())

	acc := &testutil.Accumulator{}

	require.NoError(t, acc.GatherError(h.Gather))
	testMetric(t, acc, "testdata/metrics/huebridge.txt", telegraf.Gauge)
}

func testMetric(t *testing.T, acc *testutil.Accumulator, file string, vt telegraf.ValueType) {
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	expectedMetrics, err := testutil.ParseMetricsFromFile(file, parser)
	require.NoError(t, err)
	for index := range expectedMetrics {
		expectedMetrics[index].SetType(vt)
	}
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}
