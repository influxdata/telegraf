package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

// remoteConfigServer serves a configuration revision with a matching
// Last-Modified header and can fail HEAD requests on demand.
type remoteConfigServer struct {
	sync.Mutex
	modified   string
	failHeads  int
	headsCount int
}

func (s *remoteConfigServer) set(modified string) {
	s.Lock()
	defer s.Unlock()
	s.modified = modified
}

func (s *remoteConfigServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	if r.Method == http.MethodHead {
		s.headsCount++
		if s.failHeads > 0 {
			s.failHeads--
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	w.Header().Set("Last-Modified", s.modified)
	w.Header().Set("Content-Type", "text/plain")
	if r.Method == http.MethodGet {
		_, _ = w.Write([]byte("[agent]\n"))
	}
}

func (s *remoteConfigServer) heads() int {
	s.Lock()
	defer s.Unlock()
	return s.headsCount
}

func TestWatchRemoteConfigs(t *testing.T) {
	tests := []struct {
		name string
		// modified is the revision served when the config is loaded
		modified string
		// update is the revision served after the config was loaded
		update    string
		failHeads int
		reload    bool
	}{
		{
			name:     "unchanged",
			modified: "Mon, 26 Oct 2026 12:00:00 GMT",
			update:   "Mon, 26 Oct 2026 12:00:00 GMT",
		},
		{
			name:     "changed after load before first check",
			modified: "Mon, 26 Oct 2026 12:00:00 GMT",
			update:   "Mon, 26 Oct 2026 12:05:00 GMT",
			reload:   true,
		},
		{
			name:      "changed after failed check",
			modified:  "Mon, 26 Oct 2026 12:00:00 GMT",
			update:    "Mon, 26 Oct 2026 12:05:00 GMT",
			failHeads: 1,
			reload:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &remoteConfigServer{modified: tt.modified, failHeads: tt.failHeads}
			ts := httptest.NewServer(srv)
			defer ts.Close()

			_, remote, err := config.LoadConfigFile(ts.URL)
			require.NoError(t, err)
			require.True(t, remote)
			require.Equal(t, tt.modified, config.RemoteConfigLastModified(ts.URL))

			srv.set(tt.update)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			signals := make(chan os.Signal, 1)
			done := make(chan struct{})
			go func() {
				defer close(done)
				(&Telegraf{}).watchRemoteConfigs(ctx, signals, 10*time.Millisecond, []string{ts.URL})
			}()

			if tt.reload {
				select {
				case sig := <-signals:
					require.Equal(t, syscall.SIGHUP, sig)
				case <-time.After(5 * time.Second):
					t.Fatal("remote config change was not detected")
				}
			} else {
				require.Eventually(t, func() bool { return srv.heads() >= 5 }, 5*time.Second, 10*time.Millisecond)
				require.Empty(t, signals)
			}
			cancel()
			<-done
		})
	}
}
