package cookie_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/cookie"
	"github.com/stretchr/testify/require"
)

const (
	reqUser                   = "testUser"
	reqPasswd                 = "testPassword"
	reqBody                   = "a body"
	authEndpoint              = "/auth"
	authEndpointWithBasicAuth = "/authWithCreds"
	authEndpointWithBody      = "/authWithBody"
)

var fakeCookie = &http.Cookie{
	Name:  "test-cookie",
	Value: "this is an auth cookie",
}

type fakeServer struct {
	*httptest.Server
	*int32
}

func newFakeServer(t *testing.T) fakeServer {
	var c int32
	return fakeServer{
		Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case authEndpoint:
				atomic.AddInt32(&c, 1)        // increment auth counter
				http.SetCookie(w, fakeCookie) // set fake cookie
			case authEndpointWithBody:
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				if !cmp.Equal([]byte(reqBody), body) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				atomic.AddInt32(&c, 1)        // increment auth counter
				http.SetCookie(w, fakeCookie) // set fake cookie
			case authEndpointWithBasicAuth:
				u, p, ok := r.BasicAuth()
				if !ok || u != reqUser || p != reqPasswd {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				atomic.AddInt32(&c, 1)        // increment auth counter
				http.SetCookie(w, fakeCookie) // set fake cookie
			default:
				// ensure cookie exists on request
				if _, err := r.Cookie(fakeCookie.Name); err != nil {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				_, _ = w.Write([]byte("good test response"))
			}
		})),
		int32: &c,
	}
}

func (s fakeServer) checkResp(t *testing.T, expCode int) {
	t.Helper()
	resp, err := s.Client().Get(s.URL + "/endpoint")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, expCode, resp.StatusCode)

	if expCode == http.StatusOK {
		require.Len(t, resp.Request.Cookies(), 1)
		require.Equal(t, "test-cookie", resp.Request.Cookies()[0].Name)
	}
}

func (s fakeServer) checkAuthCount(t *testing.T, atLeast int32) {
	t.Helper()
	require.GreaterOrEqual(t, atomic.LoadInt32(s.int32), atLeast)
}

func TestAuthConfig_Start(t *testing.T) {
	const (
		renewal      = 10 * time.Millisecond
		renewalCheck = 5 * renewal
	)
	type fields struct {
		Method   string
		Username string
		Password string
		Body     string
	}
	type args struct {
		renewal  time.Duration
		endpoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
		assert  func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer)
	}{
		{
			name:    "sets renewal default",
			wantErr: fmt.Errorf("bad response code: 403"),
			assert: func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer) {
				// default renewal set
				require.EqualValues(t, 5*time.Minute, c.Renewal)
				// should have never Cookie Authed
				srv.checkAuthCount(t, 0)
				srv.checkResp(t, http.StatusForbidden)
			},
		},
		{
			name: "success no creds, no body, default method",
			args: args{
				renewal:  renewal,
				endpoint: authEndpoint,
			},
			assert: func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer) {
				// should have Cookie Authed once
				srv.checkAuthCount(t, 1)
				// default method set
				require.Equal(t, http.MethodPost, c.Method)
				srv.checkResp(t, http.StatusOK)
				time.Sleep(renewalCheck)
				// should have Cookie Authed twice more
				srv.checkAuthCount(t, 3)
				srv.checkResp(t, http.StatusOK)
			},
		},
		{
			name: "success with creds, no body",
			fields: fields{
				Method:   http.MethodPost,
				Username: reqUser,
				Password: reqPasswd,
			},
			args: args{
				renewal:  renewal,
				endpoint: authEndpointWithBasicAuth,
			},
			assert: func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer) {
				// should have Cookie Authed once
				srv.checkAuthCount(t, 1)
				srv.checkResp(t, http.StatusOK)
				time.Sleep(renewalCheck)
				// should have Cookie Authed twice more
				srv.checkAuthCount(t, 3)
				srv.checkResp(t, http.StatusOK)
			},
		},
		{
			name: "failure with bad creds",
			fields: fields{
				Method:   http.MethodPost,
				Username: reqUser,
				Password: "a bad password",
			},
			args: args{
				renewal:  renewal,
				endpoint: authEndpointWithBasicAuth,
			},
			wantErr: fmt.Errorf("bad response code: 401"),
			assert: func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer) {
				// should have never Cookie Authed
				srv.checkAuthCount(t, 0)
				srv.checkResp(t, http.StatusForbidden)
				time.Sleep(renewalCheck)
				// should have still never Cookie Authed
				srv.checkAuthCount(t, 0)
				srv.checkResp(t, http.StatusForbidden)
			},
		},
		{
			name: "success with no creds, with good body",
			fields: fields{
				Method: http.MethodPost,
				Body:   reqBody,
			},
			args: args{
				renewal:  renewal,
				endpoint: authEndpointWithBody,
			},
			assert: func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer) {
				// should have Cookie Authed once
				srv.checkAuthCount(t, 1)
				srv.checkResp(t, http.StatusOK)
				time.Sleep(renewalCheck)
				// should have Cookie Authed twice more
				srv.checkAuthCount(t, 3)
				srv.checkResp(t, http.StatusOK)
			},
		},
		{
			name: "failure with bad body",
			fields: fields{
				Method: http.MethodPost,
				Body:   "a bad body",
			},
			args: args{
				renewal:  renewal,
				endpoint: authEndpointWithBody,
			},
			wantErr: fmt.Errorf("bad response code: 401"),
			assert: func(t *testing.T, c *cookie.CookieAuthConfig, srv fakeServer) {
				// should have never Cookie Authed
				srv.checkAuthCount(t, 0)
				srv.checkResp(t, http.StatusForbidden)
				time.Sleep(renewalCheck)
				// should have still never Cookie Authed
				srv.checkAuthCount(t, 0)
				srv.checkResp(t, http.StatusForbidden)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := newFakeServer(t)
			c := &cookie.CookieAuthConfig{
				URL:      srv.URL + tt.args.endpoint,
				Method:   tt.fields.Method,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
				Body:     tt.fields.Body,
				Renewal:  config.Duration(tt.args.renewal),
			}

			if err := c.Start(srv.Client()); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.assert != nil {
				tt.assert(t, c, srv)
			}
		})
	}
}
