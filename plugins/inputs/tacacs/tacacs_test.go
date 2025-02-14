package tacacs

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nwaples/tacplus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
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

func (testRequestHandler) HandleAcctRequest(context.Context, *tacplus.AcctRequest, *tacplus.ServerSession) *tacplus.AcctReply {
	return &tacplus.AcctReply{Status: tacplus.AcctStatusSuccess}
}

func TestTacacsInit(t *testing.T) {
	var testset = []struct {
		name           string
		testingTimeout config.Duration
		serversToTest  []string
		usedUsername   config.Secret
		usedPassword   config.Secret
		usedSecret     config.Secret
		requestAddr    string
		errContains    string
	}{
		{
			name:           "empty_creds",
			testingTimeout: config.Duration(time.Second * 5),
			serversToTest:  []string{"foo.bar:80"},
			usedUsername:   config.NewSecret([]byte(``)),
			usedPassword:   config.NewSecret([]byte(`testpassword`)),
			usedSecret:     config.NewSecret([]byte(`testsecret`)),
			errContains:    "empty credentials were provided (username, password or secret)",
		},
		{
			name:           "wrong_reqaddress",
			testingTimeout: config.Duration(time.Second * 5),
			serversToTest:  []string{"foo.bar:80"},
			usedUsername:   config.NewSecret([]byte(`testusername`)),
			usedPassword:   config.NewSecret([]byte(`testpassword`)),
			usedSecret:     config.NewSecret([]byte(`testsecret`)),
			requestAddr:    "257.257.257.257",
			errContains:    "invalid ip address provided for request_ip",
		},
		{
			name:           "no_reqaddress_no_servers",
			testingTimeout: config.Duration(time.Second * 5),
			usedUsername:   config.NewSecret([]byte(`testusername`)),
			usedPassword:   config.NewSecret([]byte(`testpassword`)),
			usedSecret:     config.NewSecret([]byte(`testsecret`)),
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Tacacs{
				ResponseTimeout: tt.testingTimeout,
				Servers:         tt.serversToTest,
				Username:        tt.usedUsername,
				Password:        tt.usedPassword,
				Secret:          tt.usedSecret,
				RequestAddr:     tt.requestAddr,
				Log:             testutil.Logger{},
			}

			err := plugin.Init()
			if tt.errContains == "" {
				require.NoError(t, err)
				if tt.requestAddr == "" {
					require.Equal(t, "127.0.0.1", plugin.RequestAddr)
				}
				if len(tt.serversToTest) == 0 {
					require.Equal(t, []string{"127.0.0.1:49"}, plugin.Servers)
				}
			} else {
				require.ErrorContains(t, err, tt.errContains)
			}
		})
	}
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
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "local net listen failed to start listening")

	srvLocal := l.Addr().String()

	srv := &tacplus.Server{
		ServeConn: func(nc net.Conn) {
			testHandler.Serve(nc)
		},
	}

	go func() {
		if err := srv.Serve(l); err != nil {
			t.Logf("local srv.Serve failed to start serving on %s", srvLocal)
		}
	}()

	var testset = []struct {
		name           string
		testingTimeout config.Duration
		serverToTest   []string
		usedUsername   config.Secret
		usedPassword   config.Secret
		usedSecret     config.Secret
		requestAddr    string
		errContains    string
		reqRespStatus  string
	}{
		{
			name:           "success_timeout_0s",
			testingTimeout: config.Duration(0),
			serverToTest:   []string{srvLocal},
			usedUsername:   config.NewSecret([]byte(`testusername`)),
			usedPassword:   config.NewSecret([]byte(`testpassword`)),
			usedSecret:     config.NewSecret([]byte(`testsecret`)),
			requestAddr:    "127.0.0.1",
			reqRespStatus:  "AuthenStatusPass",
		},
		{
			name:           "wrongpw",
			testingTimeout: config.Duration(time.Second * 5),
			serverToTest:   []string{srvLocal},
			usedUsername:   config.NewSecret([]byte(`testusername`)),
			usedPassword:   config.NewSecret([]byte(`WRONGPASSWORD`)),
			usedSecret:     config.NewSecret([]byte(`testsecret`)),
			requestAddr:    "127.0.0.1",
			reqRespStatus:  "AuthenStatusFail",
		},
		{
			name:           "wrongsecret",
			testingTimeout: config.Duration(time.Second * 5),
			serverToTest:   []string{srvLocal},
			usedUsername:   config.NewSecret([]byte(`testusername`)),
			usedPassword:   config.NewSecret([]byte(`testpassword`)),
			usedSecret:     config.NewSecret([]byte(`WRONGSECRET`)),
			requestAddr:    "127.0.0.1",
			errContains:    "error on new tacacs authentication start request to " + srvLocal + " : bad secret or packet",
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Tacacs{
				ResponseTimeout: tt.testingTimeout,
				Servers:         tt.serverToTest,
				Username:        tt.usedUsername,
				Password:        tt.usedPassword,
				Secret:          tt.usedSecret,
				RequestAddr:     tt.requestAddr,
				Log:             testutil.Logger{},
			}

			var acc testutil.Accumulator

			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Gather(&acc))

			if tt.errContains == "" {
				require.Empty(t, acc.Errors)
				expected := []telegraf.Metric{
					metric.New(
						"tacacs",
						map[string]string{"source": srvLocal},
						map[string]interface{}{
							"responsetime_ms": int64(0),
							"response_status": tt.reqRespStatus,
						},
						time.Unix(0, 0),
					),
				}
				options := []cmp.Option{
					testutil.IgnoreTime(),
					testutil.IgnoreFields("responsetime_ms"),
				}
				testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), options...)
			} else {
				require.Len(t, acc.Errors, 1)
				require.ErrorContains(t, acc.FirstError(), tt.errContains)
				require.Empty(t, acc.GetTelegrafMetrics())
			}
		})
	}
}

func TestTacacsLocalTimeout(t *testing.T) {
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
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "local net listen failed to start listening")

	srvLocal := l.Addr().String()

	srv := &tacplus.Server{
		ServeConn: func(nc net.Conn) {
			testHandler.Serve(nc)
		},
	}

	go func() {
		if err := srv.Serve(l); err != nil {
			t.Logf("local srv.Serve failed to start serving on %s", srvLocal)
		}
	}()

	// Initialize the plugin
	plugin := &Tacacs{
		ResponseTimeout: config.Duration(time.Microsecond),
		Servers:         []string{"unreachable.test:49"},
		Username:        config.NewSecret([]byte(`testusername`)),
		Password:        config.NewSecret([]byte(`testpassword`)),
		Secret:          config.NewSecret([]byte(`testsecret`)),
		RequestAddr:     "127.0.0.1",
		Log:             &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Try to connect, this will return a metric with the timeout...
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"tacacs",
			map[string]string{"source": "unreachable.test:49"},
			map[string]interface{}{
				"response_status": string("Timeout"),
				"responsetime_ms": int64(0),
			},
			time.Unix(0, 0),
		),
	}

	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.IgnoreFields("responsetime_ms"),
	}

	require.Empty(t, acc.Errors)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), options...)
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
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	port := container.Ports["49"]

	// Define the testset
	var testset = []struct {
		name           string
		testingTimeout config.Duration
		usedPassword   string
		reqRespStatus  string
	}{
		{
			name:           "timeout_3s",
			testingTimeout: config.Duration(time.Second * 3),
			usedPassword:   "cisco",
			reqRespStatus:  "AuthenStatusPass",
		},
		{
			name:           "timeout_0s",
			testingTimeout: config.Duration(0),
			usedPassword:   "cisco",
			reqRespStatus:  "AuthenStatusPass",
		},
		{
			name:           "wrong_pw",
			testingTimeout: config.Duration(time.Second * 5),
			usedPassword:   "wrongpass",
			reqRespStatus:  "AuthenStatusFail",
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &Tacacs{
				ResponseTimeout: tt.testingTimeout,
				Servers:         []string{container.Address + ":" + port},
				Username:        config.NewSecret([]byte(`iosadmin`)),
				Password:        config.NewSecret([]byte(tt.usedPassword)),
				Secret:          config.NewSecret([]byte(`ciscotacacskey`)),
				RequestAddr:     "127.0.0.1",
				Log:             testutil.Logger{},
			}
			var acc testutil.Accumulator

			// Startup the plugin & Gather
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Gather(&acc))

			require.NoError(t, acc.FirstError())
			expected := []telegraf.Metric{
				metric.New(
					"tacacs",
					map[string]string{"source": container.Address + ":" + port},
					map[string]interface{}{
						"responsetime_ms": int64(0),
						"response_status": tt.reqRespStatus,
					},
					time.Unix(0, 0),
				),
			}
			options := []cmp.Option{
				testutil.IgnoreTime(),
				testutil.IgnoreFields("responsetime_ms"),
			}
			testutil.RequireMetricsStructureEqual(t, expected, acc.GetTelegrafMetrics(), options...)
		})
	}
}
