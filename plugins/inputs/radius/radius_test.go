package radius

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestRadiusLocal(t *testing.T) {
	handler := func(w radius.ResponseWriter, r *radius.Request) {
		username := rfc2865.UserName_GetString(r.Packet)
		password := rfc2865.UserPassword_GetString(r.Packet)

		var code radius.Code
		if username == "testusername" && password == "testpassword" {
			code = radius.CodeAccessAccept
		} else {
			code = radius.CodeAccessReject
		}
		if err := w.Write(r.Response(code)); err != nil {
			require.NoError(t, err, "failed writing radius server response")
		}
	}

	// Setup a connection to be able to get a random port
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer conn.Close()
	addr := conn.LocalAddr().String()
	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(`testsecret`)),
		Addr:         addr,
	}

	go func() {
		if err := server.Serve(conn); err != nil {
			if !errors.Is(err, radius.ErrServerShutdown) {
				require.NoError(t, err, "local radius server failed")
			}
		}
	}()

	plugin := &Radius{
		Servers:  []string{addr},
		Username: config.NewSecret([]byte(`testusername`)),
		Password: config.NewSecret([]byte(`testpassword`)),
		Secret:   config.NewSecret([]byte(`testsecret`)),
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	if !acc.HasMeasurement("radius") {
		t.Errorf("acc.HasMeasurement: expected radius")
	}
	require.Equal(t, true, acc.HasTag("radius", "source"))
	require.Equal(t, true, acc.HasTag("radius", "source_port"))
	require.Equal(t, true, acc.HasTag("radius", "response_code"))
	require.Equal(t, host, acc.TagValue("radius", "source"))
	require.Equal(t, port, acc.TagValue("radius", "source_port"))
	require.Equal(t, radius.CodeAccessAccept.String(), acc.TagValue("radius", "response_code"))
	require.Equal(t, true, acc.HasInt64Field("radius", "responsetime_ms"))

	if err := server.Shutdown(context.Background()); err != nil {
		require.NoError(t, err, "failed to properly shutdown local radius server")
	}
}

func TestRadiusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testdata, err := filepath.Abs("testdata/raddb/clients.conf")
	require.NoError(t, err, "determining absolute path of test-data clients.conf failed")
	testdataa, err := filepath.Abs("testdata/raddb/mods-config/files/authorize")
	require.NoError(t, err, "determining absolute path of test-data authorize failed")
	testdataaa, err := filepath.Abs("testdata/raddb/radiusd.conf")
	require.NoError(t, err, "determining absolute path of test-data radiusd.conf failed")

	container := testutil.Container{
		Image:        "freeradius/freeradius-server",
		ExposedPorts: []string{"1812/udp"},
		BindMounts: map[string]string{
			"/etc/raddb/clients.conf":                testdata,
			"/etc/raddb/mods-config/files/authorize": testdataa,
			"/etc/raddb/radiusd.conf":                testdataaa,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to process requests"),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	port := container.Ports["1812"]

	// Define the testset
	var testset = []struct {
		name               string
		testingTimeout     config.Duration
		expectedSource     string
		expectedSourcePort string
		serverToTest       string
		expectSuccess      bool
		usedPassword       string
	}{
		{
			name:               "timeout_5s",
			testingTimeout:     config.Duration(time.Second * 5),
			expectedSource:     container.Address,
			expectedSourcePort: port,
			serverToTest:       container.Address + ":" + port,
			expectSuccess:      true,
			usedPassword:       "testpassword",
		},
		{
			name:               "timeout_0s",
			testingTimeout:     config.Duration(0),
			expectedSource:     container.Address,
			expectedSourcePort: port,
			serverToTest:       container.Address + ":" + port,
			expectSuccess:      true,
			usedPassword:       "testpassword",
		},
		{
			name:               "wrong_pw",
			testingTimeout:     config.Duration(time.Second * 5),
			expectedSource:     container.Address,
			expectedSourcePort: port,
			serverToTest:       container.Address + ":" + port,
			expectSuccess:      false,
			usedPassword:       "wrongpass",
		},
		{
			name:               "unreachable",
			testingTimeout:     config.Duration(5),
			expectedSource:     "unreachable.unreachable.com",
			expectedSourcePort: "7777",
			serverToTest:       "unreachable.unreachable.com:7777",
			expectSuccess:      false,
			usedPassword:       "testpassword",
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &Radius{
				ResponseTimeout: tt.testingTimeout,
				Servers:         []string{tt.serverToTest},
				Username:        config.NewSecret([]byte(`testusername`)),
				Password:        config.NewSecret([]byte(tt.usedPassword)),
				Secret:          config.NewSecret([]byte(`testsecret`)),
				Log:             testutil.Logger{},
			}
			var acc testutil.Accumulator

			// Startup the plugin
			require.NoError(t, plugin.Init())

			// Gather
			require.NoError(t, plugin.Gather(&acc))
			require.Len(t, acc.Errors, 0)

			if !acc.HasMeasurement("radius") {
				t.Errorf("acc.HasMeasurement: expected radius")
			}
			require.Equal(t, true, acc.HasTag("radius", "source"))
			require.Equal(t, true, acc.HasTag("radius", "source_port"))
			require.Equal(t, true, acc.HasTag("radius", "response_code"))
			require.Equal(t, tt.expectedSource, acc.TagValue("radius", "source"))
			require.Equal(t, tt.expectedSourcePort, acc.TagValue("radius", "source_port"))
			require.Equal(t, true, acc.HasInt64Field("radius", "responsetime_ms"), true)
			if tt.expectSuccess {
				require.Equal(t, radius.CodeAccessAccept.String(), acc.TagValue("radius", "response_code"))
			} else {
				require.NotEqual(t, radius.CodeAccessAccept.String(), acc.TagValue("radius", "response_code"))
			}

			if tt.name == "unreachable" {
				require.Equal(t, time.Duration(tt.testingTimeout).Milliseconds(), acc.Metrics[0].Fields["responsetime_ms"])
			}
		})
	}
}
