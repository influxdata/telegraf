package radius

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
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

	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(`testsecret`)),
		Addr:         ":1813",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, radius.ErrServerShutdown) {
				require.NoError(t, err, "local radius server failed")
			}
		}
	}()

	plugin := &Radius{
		Servers:  []string{"localhost:1813"},
		Username: config.NewSecret([]byte(`testusername`)),
		Password: config.NewSecret([]byte(`testpassword`)),
		Secret:   config.NewSecret([]byte(`testsecret`)),
		Log:      testutil.Logger{},
	}
	var acc testutil.Accumulator

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.Errors, 0)
	if !acc.HasMeasurement("radius") {
		t.Errorf("acc.HasMeasurement: expected radius")
	}
	require.Equal(t, acc.HasTag("radius", "source"), true)
	require.Equal(t, acc.HasTag("radius", "source_port"), true)
	require.Equal(t, acc.HasTag("radius", "response_code"), true)
	require.Equal(t, acc.TagValue("radius", "source"), "localhost")
	require.Equal(t, acc.TagValue("radius", "source_port"), "1813")
	require.Equal(t, acc.TagValue("radius", "response_code"), radius.CodeAccessAccept.String())
	require.Equal(t, acc.HasInt64Field("radius", "responsetime_ms"), true)

	if err := server.Shutdown(context.Background()); err != nil {
		require.NoError(t, err, "failed to properly shutdown local radius server")
	}
}

func TestRadiusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	port := "1812"

	testdata, err := filepath.Abs("testdata/raddb")
	require.NoError(t, err, "determining absolute path of test-data failed")

	container := testutil.Container{
		Image:        "freeradius/freeradius-server",
		ExposedPorts: []string{port},
		BindMounts: map[string]string{
			"/etc/raddb": testdata,
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(port)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

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
			require.Equal(t, acc.HasTag("radius", "source"), true)
			require.Equal(t, acc.HasTag("radius", "source_port"), true)
			require.Equal(t, acc.HasTag("radius", "response_code"), true)
			require.Equal(t, acc.TagValue("radius", "source"), tt.expectedSource)
			require.Equal(t, acc.TagValue("radius", "source_port"), tt.expectedSourcePort)
			require.Equal(t, acc.HasInt64Field("radius", "responsetime_ms"), true)
			if tt.expectSuccess {
				require.Equal(t, acc.TagValue("radius", "response_code"), radius.CodeAccessAccept.String())
			} else {
				require.NotEqual(t, acc.TagValue("radius", "response_code"), radius.CodeAccessAccept.String())
			}

			if tt.name == "unreachable" {
				require.Equal(t, time.Duration(tt.testingTimeout).Milliseconds(), acc.Metrics[0].Fields["responsetime_ms"])
			}
		})
	}
}
