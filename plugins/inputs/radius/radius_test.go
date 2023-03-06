package radius

import (
	"context"
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
		w.Write(r.Response(code))
	}

	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(`testsecret`)),
		Addr:         ":1813",
	}

	if err := server.ListenAndServe(); err != nil {
		require.NoError(t, err, "failed to start local radius server")
	}

	plugin := &Radius{
		Servers:  []string{"localhost:1813"},
		Username: config.NewSecret([]byte(`testusername`)),
		Password: config.NewSecret([]byte(`testpassword`)),
		Secret:   config.NewSecret([]byte(`testsecret`)),
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

	server.Shutdown(context.Background())
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
		name                 string
		testing_timeout      config.Duration
		expected_source      string
		expected_source_port string
		server_to_test       string
		expect_success       bool
		used_password        string
	}{
		{
			name:                 "timeout_5s",
			testing_timeout:      config.Duration(time.Second * 5),
			expected_source:      container.Address,
			expected_source_port: port,
			server_to_test:       container.Address + ":" + port,
			expect_success:       true,
			used_password:        "testpassword",
		},
		{
			name:                 "timeout_0s",
			testing_timeout:      config.Duration(0),
			expected_source:      container.Address,
			expected_source_port: port,
			server_to_test:       container.Address + ":" + port,
			expect_success:       true,
			used_password:        "testpassword",
		},
		{
			name:                 "wrong_pw",
			testing_timeout:      config.Duration(time.Second * 5),
			expected_source:      container.Address,
			expected_source_port: port,
			server_to_test:       container.Address + ":" + port,
			expect_success:       false,
			used_password:        "wrongpass",
		},
		{
			name:                 "unreachable",
			testing_timeout:      config.Duration(5),
			expected_source:      "unreachable.unreachable.com",
			expected_source_port: "7777",
			server_to_test:       "unreachable.unreachable.com:7777",
			expect_success:       false,
			used_password:        "testpassword",
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &Radius{
				ResponseTimeout: tt.testing_timeout,
				Servers:         []string{tt.server_to_test},
				Username:        config.NewSecret([]byte(`testusername`)),
				Password:        config.NewSecret([]byte(tt.used_password)),
				Secret:          config.NewSecret([]byte(`testsecret`)),
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
			require.Equal(t, acc.TagValue("radius", "source"), tt.expected_source)
			require.Equal(t, acc.TagValue("radius", "source_port"), tt.expected_source_port)
			require.Equal(t, acc.HasInt64Field("radius", "responsetime_ms"), true)
			if tt.expect_success {
				require.Equal(t, acc.TagValue("radius", "response_code"), radius.CodeAccessAccept.String())
			} else {
				require.NotEqual(t, acc.TagValue("radius", "response_code"), radius.CodeAccessAccept.String())
			}

			if tt.name == "unreachable" {
				require.Equal(t, time.Duration(tt.testing_timeout).Milliseconds(), acc.Metrics[0].Fields["responsetime_ms"])
			}
		})
	}
}
