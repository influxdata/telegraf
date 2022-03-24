package uwsgi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs/uwsgi"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	js := `
{
    "version":"2.0.12",
    "listen_queue":0,
    "listen_queue_errors":0,
    "signal_queue":0,
    "load":0,
    "pid":28372,
    "uid":1000,
    "gid":1000,
    "cwd":"/opt/uwsgi",
    "locks":[
	{
	    "user 0":0
	},
	{
	    "signal":0
	},
	{
	    "filemon":0
	},
	{
	    "timer":0
	},
	{
	    "rbtimer":0
	},
	{
	    "cron":0
	},
	{
	    "rpc":0
	},
	{
	    "snmp":0
	}
    ],
    "sockets":[
	{
	    "name":"127.0.0.1:47430",
	    "proto":"uwsgi",
	    "queue":0,
	    "max_queue":100,
	    "shared":0,
	    "can_offload":0
	}
    ],
    "workers":[
	{
	    "id":1,
	    "pid":28375,
	    "accepting":1,
	    "requests":0,
	    "delta_requests":0,
	    "exceptions":0,
	    "harakiri_count":0,
	    "signals":0,
	    "signal_queue":0,
	    "status":"idle",
	    "rss":0,
	    "vsz":0,
	    "running_time":0,
	    "last_spawn":1459942782,
	    "respawn_count":1,
	    "tx":0,
	    "avg_rt":0,
	    "apps":[
		{
		    "id":0,
		    "modifier1":0,
		    "mountpoint":"",
		    "startup_time":0,
		    "requests":0,
		    "exceptions":0,
		    "chdir":""
		}
	    ],
	    "cores":[
		{
		    "id":0,
		    "requests":0,
		    "static_requests":0,
		    "routed_requests":0,
		    "offloaded_requests":0,
		    "write_errors":0,
		    "read_errors":0,
		    "in_request":0,
		    "vars":[

		    ]
		}
	    ]
	}
    ]
}
`

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(js))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer fakeServer.Close()

	plugin := &uwsgi.Uwsgi{
		Servers: []string{fakeServer.URL + "/"},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Equal(t, 0, len(acc.Errors))
}

func TestInvalidJSON(t *testing.T) {
	js := `
{
    "version":"2.0.12",
    "listen_queue":0,
    "listen_queue_errors":0,
    "signal_queue":0,
    "load":0,
    "pid:28372
    "uid":10
}
`

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(js))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer fakeServer.Close()

	plugin := &uwsgi.Uwsgi{
		Servers: []string{fakeServer.URL + "/"},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Equal(t, 1, len(acc.Errors))
}

func TestHttpError(t *testing.T) {
	plugin := &uwsgi.Uwsgi{
		Servers: []string{"http://novalidurladress/"},
		Timeout: config.Duration(10 * time.Millisecond),
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Equal(t, 1, len(acc.Errors))
}

func TestTcpError(t *testing.T) {
	plugin := &uwsgi.Uwsgi{
		Servers: []string{"tcp://novalidtcpadress/"},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Equal(t, 1, len(acc.Errors))
}

func TestUnixSocketError(t *testing.T) {
	plugin := &uwsgi.Uwsgi{
		Servers: []string{"unix:///novalidunixsocket"},
	}
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Equal(t, 1, len(acc.Errors))
}
