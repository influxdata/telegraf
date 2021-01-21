package aurora

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type (
	TestHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
	CheckFunc       func(t *testing.T, err error, acc *testutil.Accumulator)
)

func TestAurora(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name         string
		plugin       *Aurora
		schedulers   []string
		roles        []string
		leaderhealth TestHandlerFunc
		varsjson     TestHandlerFunc
		check        CheckFunc
	}{
		{
			name: "minimal",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body := `{
					"variable_scrape_events": 2958,
					"variable_scrape_events_per_sec": 1.0,
					"variable_scrape_micros_per_event": 1484.0,
					"variable_scrape_micros_total": 4401084,
					"variable_scrape_micros_total_per_sec": 1485.0
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.Equal(t, 1, len(acc.Metrics))
				acc.AssertContainsTaggedFields(t,
					"aurora",
					map[string]interface{}{
						"variable_scrape_events":               int64(2958),
						"variable_scrape_events_per_sec":       1.0,
						"variable_scrape_micros_per_event":     1484.0,
						"variable_scrape_micros_total":         int64(4401084),
						"variable_scrape_micros_total_per_sec": 1485.0,
					},
					map[string]string{
						"scheduler": u.String(),
						"role":      "leader",
					},
				)
			},
		},
		{
			name:  "disabled role",
			roles: []string{"leader"},
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.NoError(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
		{
			name: "no metrics available",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.NoError(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
		{
			name: "string metrics skipped",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body := `{
					"foo": "bar"
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.NoError(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
		{
			name: "float64 unparseable",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				// too large
				body := `{
					"foo": 1e309
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.Error(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
		{
			name: "int64 unparseable",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				// too large
				body := `{
					"foo": 9223372036854775808
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.Error(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
		{
			name: "bad json",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body := `{]`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(body))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.Error(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
		{
			name: "wrong status code",
			leaderhealth: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			varsjson: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body := `{
					"value": 42
				}`
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(body))
			},
			check: func(t *testing.T, err error, acc *testutil.Accumulator) {
				require.NoError(t, err)
				require.Error(t, acc.FirstError())
				require.Equal(t, 0, len(acc.Metrics))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/leaderhealth":
					tt.leaderhealth(t, w, r)
				case "/vars.json":
					tt.varsjson(t, w, r)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			})

			var acc testutil.Accumulator
			plugin := &Aurora{}
			plugin.Schedulers = []string{u.String()}
			plugin.Roles = tt.roles
			err := plugin.Gather(&acc)
			tt.check(t, err, &acc)
		})
	}
}

func TestBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name: "no auth",
		},
		{
			name:     "basic auth",
			username: "username",
			password: "pa$$word",
		},
		{
			name:     "username only",
			username: "username",
		},
		{
			name:     "password only",
			password: "pa$$word",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, password, _ := r.BasicAuth()
				require.Equal(t, tt.username, username)
				require.Equal(t, tt.password, password)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
			})

			var acc testutil.Accumulator
			plugin := &Aurora{}
			plugin.Schedulers = []string{u.String()}
			plugin.Username = tt.username
			plugin.Password = tt.password
			err := plugin.Gather(&acc)
			require.NoError(t, err)
		})
	}
}
