package tacacs

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/nwaples/tacplus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

type testRequestHandler map[string]struct {
	password string
	args     []string
}

func (t testRequestHandler) HandleAuthenStart(_ context.Context, a *tacplus.AuthenStart, s *tacplus.ServerSession) *tacplus.AuthenReply {
	user := a.User
	for user == "" {
		c, err := s.GetUser(context.Background(), "Username:")
		if err != nil || c.Abort {
			return nil
		}
		user = c.Message
	}
	pass := ""
	for pass == "" {
		c, err := s.GetPass(context.Background(), "Password:")
		if err != nil || c.Abort {
			return nil
		}
		pass = c.Message
	}
	if u, ok := t[user]; ok && u.password == pass {
		return &tacplus.AuthenReply{Status: tacplus.AuthenStatusPass}
	}
	return &tacplus.AuthenReply{Status: tacplus.AuthenStatusFail}
}

func (t testRequestHandler) HandleAuthorRequest(_ context.Context, a *tacplus.AuthorRequest, _ *tacplus.ServerSession) *tacplus.AuthorResponse {
	if u, ok := t[a.User]; ok {
		return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusPassAdd, Arg: u.args}
	}
	return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}
}

func (t testRequestHandler) HandleAcctRequest(_ context.Context, _ *tacplus.AcctRequest, _ *tacplus.ServerSession) *tacplus.AcctReply {
	return &tacplus.AcctReply{Status: tacplus.AcctStatusSuccess}
}

func TestTacacsLocal(t *testing.T) {
	testHandler := tacplus.ServerConnHandler{
		Handler: &testRequestHandler{
			"testusername": {
				password: "testpassword",
			},
		},
		ConnConfig: tacplus.ConnConfig{
			Secret: []byte(`testsecret`),
			Mux:    true,
		},
	}
	l, err := net.Listen("tcp", "localhost:1049") // Use port 1049 instead of 49, so we are above 1023 for unix tests
	require.NoError(t, err, "local net listen failed to start listening on port 1049")

	srv := &tacplus.Server{
		ServeConn: func(nc net.Conn) {
			testHandler.Serve(nc)
		},
	}

	go func() {
		err = srv.Serve(l)
		require.NoError(t, err, "local srv.Serve failed to start serving on port 1049")
	}()

	plugin := &Tacacs{
		Servers:         []string{"localhost:1049"},
		Username:        config.NewSecret([]byte(`testusername`)),
		Password:        config.NewSecret([]byte(`testpassword`)),
		Secret:          config.NewSecret([]byte(`testsecret`)),
		RequestAddr:     "127.0.0.1",
		ResponseTimeout: config.Duration(0),
		Log:             testutil.Logger{},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.Errors, 0)
	if !acc.HasMeasurement("tacacs") {
		t.Errorf("acc.HasMeasurement: expected tacacs")
	}
	require.Equal(t, true, acc.HasTag("tacacs", "source"))
	require.Equal(t, "localhost:1049", acc.TagValue("tacacs", "source"))
	require.Equal(t, true, acc.HasInt64Field("tacacs", "responsetime_ms"))
	require.Equal(t, true, acc.HasTag("tacacs", "response_code"))
	require.Equal(t, strconv.FormatUint(uint64(tacplus.AuthenStatusPass), 10), acc.TagValue("tacacs", "response_code"))

	plugin = &Tacacs{
		Servers:         []string{"localhost:1049"},
		Username:        config.NewSecret([]byte(`testusername`)),
		Password:        config.NewSecret([]byte(`WRONGPASSWORD`)),
		Secret:          config.NewSecret([]byte(`testsecret`)),
		RequestAddr:     "127.0.0.1",
		ResponseTimeout: config.Duration(time.Second * 5),
		Log:             testutil.Logger{},
	}
	var acc2 testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Gather(&acc2))
	require.Len(t, acc2.Errors, 0)
	require.Equal(t, true, acc2.HasTag("tacacs", "response_code"))
	require.Equal(t, strconv.FormatUint(uint64(tacplus.AuthenStatusFail), 10), acc2.TagValue("tacacs", "response_code"))
	require.Equal(t, true, acc.HasTag("tacacs", "source"))
	require.Equal(t, "localhost:1049", acc.TagValue("tacacs", "source"))
	require.Equal(t, true, acc2.HasInt64Field("tacacs", "responsetime_ms"))
	require.Equal(t, time.Duration(plugin.ResponseTimeout).Milliseconds(), acc2.Metrics[0].Fields["responsetime_ms"])

	plugin = &Tacacs{
		Servers:         []string{"localhost:1049"},
		Username:        config.NewSecret([]byte(`testusername`)),
		Password:        config.NewSecret([]byte(`testpassword`)),
		Secret:          config.NewSecret([]byte(`WRONGSECRET`)),
		RequestAddr:     "127.0.0.1",
		ResponseTimeout: config.Duration(time.Second * 5),
		Log:             testutil.Logger{},
	}
	var acc3 testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Gather(&acc3))
	require.Len(t, acc3.Errors, 1)
	require.ErrorContains(t, acc3.Errors[0], "error on new tacacs authentication start request to localhost:1049 : bad secret or packet")
	require.Equal(t, false, acc3.HasTag("tacacs", "source"))
	require.Equal(t, false, acc3.HasInt64Field("tacacs", "responsetime_ms"))

	plugin = &Tacacs{
		Servers:         []string{"localhost:9999"},
		Username:        config.NewSecret([]byte(`testusername`)),
		Password:        config.NewSecret([]byte(`testpassword`)),
		Secret:          config.NewSecret([]byte(`testsecret`)),
		RequestAddr:     "127.0.0.1",
		ResponseTimeout: config.Duration(time.Nanosecond * 1),
		Log:             testutil.Logger{},
	}
	var acc4 testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Gather(&acc4))
	require.Len(t, acc4.Errors, 1)
	require.ErrorContains(t, acc4.Errors[0], "error on new tacacs authentication start request to localhost:9999 : dial tcp")
	require.Equal(t, false, acc4.HasTag("tacacs", "source"))
	require.Equal(t, false, acc4.HasInt64Field("tacacs", "responsetime_ms"))
}

func TestTacacsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "dchidell/docker-tacacs",
		ExposedPorts: []string{"49/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Starting server..."),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	port := container.Ports["49"]

	// Define the testset
	var testset = []struct {
		name           string
		testingTimeout config.Duration
		serverToTest   string
		expectSuccess  bool
		usedPassword   string
	}{
		{
			name:           "timeout_3s",
			testingTimeout: config.Duration(time.Second * 3),
			serverToTest:   container.Address + ":" + port,
			expectSuccess:  true,
			usedPassword:   "cisco",
		},
		{
			name:           "timeout_0s",
			testingTimeout: config.Duration(0),
			serverToTest:   container.Address + ":" + port,
			expectSuccess:  true,
			usedPassword:   "cisco",
		},
		{
			name:           "wrong_pw",
			testingTimeout: config.Duration(time.Second * 5),
			serverToTest:   container.Address + ":" + port,
			expectSuccess:  false,
			usedPassword:   "wrongpass",
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &Tacacs{
				ResponseTimeout: tt.testingTimeout,
				Servers:         []string{tt.serverToTest},
				Username:        config.NewSecret([]byte(`iosadmin`)),
				Password:        config.NewSecret([]byte(tt.usedPassword)),
				Secret:          config.NewSecret([]byte(`ciscotacacskey`)),
				RequestAddr:     "127.0.0.1",
				Log:             testutil.Logger{},
			}
			var acc testutil.Accumulator

			// Startup the plugin
			require.NoError(t, plugin.Init())

			// Gather
			require.NoError(t, plugin.Gather(&acc))

			if tt.expectSuccess {
				if len(acc.Errors) > 0 {
					t.Errorf("error occured in test where should be none, error was: %s", acc.Errors[0].Error())
				}
				if !acc.HasMeasurement("tacacs") {
					t.Errorf("acc.HasMeasurement: expected tacacs")
				}
				require.Equal(t, true, acc.HasTag("tacacs", "source"))
				require.Equal(t, tt.serverToTest, acc.TagValue("tacacs", "source"))
				require.Equal(t, true, acc.HasInt64Field("tacacs", "responsetime_ms"), true)
				require.Len(t, acc.Errors, 0)
			} else {
				require.Equal(t, false, acc.HasTag("tacacs", "source"))
				require.Equal(t, false, acc.HasInt64Field("tacacs", "responsetime_ms"), true)
				require.Len(t, acc.Errors, 1)
			}

			if tt.name == "wrong_pw" {
				require.NoError(t, plugin.Init())
				require.NoError(t, plugin.Gather(&acc))
				require.Len(t, acc.Errors, 0)
				require.Equal(t, true, acc.HasTag("tacacs", "response_code"))
				require.Equal(t, strconv.FormatUint(uint64(tacplus.AuthenStatusFail), 10), acc.TagValue("tacacs", "response_code"))
				require.Equal(t, true, acc.HasTag("tacacs", "source"))
				require.Equal(t, tt.serverToTest, acc.TagValue("tacacs", "source"))
				require.Equal(t, true, acc.HasInt64Field("tacacs", "responsetime_ms"))
				require.Equal(t, time.Duration(plugin.ResponseTimeout).Milliseconds(), acc.Metrics[0].Fields["responsetime_ms"])
			}
		})
	}
}
