package ctrlx_datalayer

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/boschrexroth/ctrlx-datalayer-golang/pkg/token"
	"github.com/influxdata/telegraf/config"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const path = "/automation/api/v2/events"

var multiEntries = false
var mux sync.Mutex

func setMultiEntries(m bool) {
	mux.Lock()
	defer mux.Unlock()
	multiEntries = m
}

func getMultiEntries() bool {
	mux.Lock()
	defer mux.Unlock()
	return multiEntries
}

func TestCtrlXCreateSubscriptionBasic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("201 created"))
		require.NoError(t, err)
	}))
	defer server.Close()

	subs := make([]Subscription, 0)
	subs = append(subs, Subscription{
		index: 0,
		Nodes: []Node{
			{Name: "counter", Address: "plc/app/Application/sym/PLC_PRG/counter"},
			{Name: "counterReverse", Address: "plc/app/Application/sym/PLC_PRG/counterReverse"},
		},
	},
	)
	s := &CtrlXDataLayer{
		connection:   &http.Client{},
		url:          server.URL,
		Username:     config.NewSecret([]byte("user")),
		Password:     config.NewSecret([]byte("password")),
		tokenManager: token.TokenManager{Url: server.URL, Username: "user", Password: "password", Connection: &http.Client{}},
		Subscription: subs,
		Log:          testutil.Logger{},
	}

	subID, err := s.createSubscription(&subs[0])

	require.NoError(t, err)
	require.NotEmpty(t, subID)
}

func TestCtrlXCreateSubscriptionDriven(t *testing.T) {
	var tests = []struct {
		res       string
		status    int
		wantError bool
	}{
		{res: "{\"status\":200}", status: 200, wantError: false},
		{res: "{\"status\":401}", status: 401, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.res, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				_, err := w.Write([]byte(test.res))
				require.NoError(t, err)
			}))
			defer server.Close()
			subs := make([]Subscription, 0)
			subs = append(subs, Subscription{
				Nodes: []Node{
					{Name: "counter", Address: "plc/app/Application/sym/PLC_PRG/counter"},
					{Name: "counterReverse", Address: "plc/app/Application/sym/PLC_PRG/counterReverse"},
				},
			},
			)
			s := &CtrlXDataLayer{
				connection:   &http.Client{},
				url:          server.URL,
				Username:     config.NewSecret([]byte("user")),
				Password:     config.NewSecret([]byte("password")),
				Subscription: subs,
				tokenManager: token.TokenManager{Url: server.URL, Username: "user", Password: "password", Connection: &http.Client{}},
				Log:          testutil.Logger{},
			}
			subID, err := s.createSubscription(&subs[0])

			if test.wantError {
				require.EqualError(t, err, "failed to create sse subscription 0, status: 401 Unauthorized")
				require.Empty(t, subID)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, subID)
			}
		})
	}
}

func newServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	// Handle request to fetch token
	mux.HandleFunc("/identity-manager/api/v2/auth/token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("{\"access_token\": \"eyJhbGciOiJIU.xxx.xxx\", \"token_type\":\"Bearer\"}"))
		require.NoError(t, err)
	}))
	// Handle request to validate token
	mux.HandleFunc("/identity-manager/api/v2/auth/token/validity", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("{\"valid\": \"true\"}"))
		require.NoError(t, err)
	}))
	// Handle request to create subscription
	mux.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("201 created"))
		require.NoError(t, err)
	}))
	// Handle request to fetch sse data
	mux.HandleFunc(path+"/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("event: update\n"))
			require.NoError(t, err)
			_, err = w.Write([]byte("id: 12345\n"))
			require.NoError(t, err)
			if getMultiEntries() {
				data := "data: {\n"
				_, err = w.Write([]byte(data))
				require.NoError(t, err)
				data = "data: \"node\":\"plc/app/Application/sym/PLC_PRG/counter\", \"timestamp\":132669450604571037,\"type\":\"double\",\"value\":44.0\n"
				_, err = w.Write([]byte(data))
				require.NoError(t, err)
				data = "data: }\n"
				_, err = w.Write([]byte(data))
				require.NoError(t, err)
			} else {
				data := "data: {\"node\":\"plc/app/Application/sym/PLC_PRG/counter\", \"timestamp\":132669450604571037,\"type\":\"double\",\"value\":43.0}\n"
				_, err = w.Write([]byte(data))
				require.NoError(t, err)
			}
			_, err = w.Write([]byte("\n"))
			require.NoError(t, err)
		}
	}))
	return httptest.NewServer(mux)
}

func cleanup(server *httptest.Server) {
	server.CloseClientConnections()
	server.Close()
}

func initRunner(t *testing.T) (*CtrlXDataLayer, *httptest.Server) {
	server := newServer(t)

	subs := make([]Subscription, 0)
	subs = append(subs, Subscription{
		Measurement: "ctrlx",
		Nodes: []Node{
			{Name: "counter", Address: "plc/app/Application/sym/PLC_PRG/counter"},
		},
	},
	)
	s := &CtrlXDataLayer{
		connection: &http.Client{},
		url:        server.URL,
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("password")),
		HTTPClientConfig: httpconfig.HTTPClientConfig{
			ClientConfig: tls.ClientConfig{
				InsecureSkipVerify: true,
			},
		},
		Subscription: subs,
		tokenManager: token.TokenManager{Url: server.URL, Username: "user", Password: "password", Connection: &http.Client{}},
		Log:          testutil.Logger{},
	}
	return s, server
}

func TestCtrlXMetricsField(t *testing.T) {
	const measurement = "ctrlx"
	const fieldName = "counter"

	s, server := initRunner(t)
	defer cleanup(server)

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(s.Start))
	require.Eventually(t, func() bool {
		if v, found := acc.FloatField(measurement, fieldName); found {
			require.Equal(t, 43.0, v)
			return true
		}
		return false
	}, time.Second*10, time.Second)
}

func TestCtrlXMetricsMulti(t *testing.T) {
	const measurement = "ctrlx"
	const fieldName = "counter"

	setMultiEntries(true)
	s, server := initRunner(t)
	defer cleanup(server)

	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(s.Start))
	require.Eventually(t, func() bool {
		if v, found := acc.FloatField(measurement, fieldName); found {
			require.Equal(t, 44.0, v)
			return true
		}
		return false
	}, time.Second*10, time.Second)

	setMultiEntries(false)
}

func TestCtrlXCreateSseClient(t *testing.T) {
	sub := Subscription{
		Measurement: "ctrlx",
		Nodes: []Node{
			{Name: "counter", Address: "plc/app/Application/sym/PLC_PRG/counter"},
			{Name: "counterReverse", Address: "plc/app/Application/sym/PLC_PRG/counterReverse"},
		},
	}
	s, server := initRunner(t)
	defer cleanup(server)
	client, err := s.createSubscriptionAndSseClient(&sub)
	require.NoError(t, err)
	require.NotEmpty(t, client)
}

func TestConvertTimestamp2UnixTime(t *testing.T) {
	expected := time.Date(2022, 02, 14, 14, 22, 38, 333552400, time.UTC)
	actual := convertTimestamp2UnixTime(132893221583335524)
	require.EqualValues(t, expected.UnixNano(), actual.UnixNano())
}
