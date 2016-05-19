package uwsgi

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var timeout = 5 * time.Second

type Uwsgi struct {
	Servers []string `toml:"server"`
}

func (u *Uwsgi) Description() string {
	return "Read uWSGI metrics."
}

func (u *Uwsgi) SampleConfig() string {
	return `
    ## List with urls of uWSGI Stats servers. Url must match pattern:
    ## scheme://address[:port]
    ##
    ## For example:
    ## servers = ["tcp://localhost:5050", "http://localhost:1717", "unix:///tmp/statsock"]
    servers = []
`
}

func (u *Uwsgi) Gather(acc telegraf.Accumulator) error {
	for _, s := range u.Servers {
		n, err := url.Parse(s)
		if err != nil {
			return fmt.Errorf("Could not parse uWSGI Stats Server url '%s': %s", s, err)
		}

		u.gatherServer(acc, n)

	}
	return nil
}

func (u *Uwsgi) gatherServer(acc telegraf.Accumulator, url *url.URL) error {
	var err error
	var r io.ReadCloser

	switch url.Scheme {
	case "unix":
		r, err = net.DialTimeout(url.Scheme, url.Path, timeout)
	case "tcp":
		r, err = net.DialTimeout(url.Scheme, url.Host, timeout)
	case "http":
		resp, err := http.Get(url.String())
		if err != nil {
			return fmt.Errorf("Could not connect to uWSGI Stats Server '%s': %s", url.String(), err)
		}
		r = resp.Body
	default:
		return fmt.Errorf("'%s' is not a valid URL", url.String())
	}

	if err != nil {
		return fmt.Errorf("Could not connect to uWSGI Stats Server '%s': %s", url.String(), err)
	}
	defer r.Close()

	var s StatsServer
	s.Url = url.String()

	dec := json.NewDecoder(r)
	dec.Decode(&s)

	u.gatherStatServer(acc, &s)

	return nil
}

func (u *Uwsgi) gatherStatServer(acc telegraf.Accumulator, s *StatsServer) {
	fields := map[string]interface{}{
		"listen_queue":        s.ListenQueue,
		"listen_queue_errors": s.ListenQueueErrors,
		"signal_queue":        s.SignalQueue,
		"load":                s.Load,
	}

	tags := map[string]string{
		"url":     s.Url,
		"pid":     strconv.Itoa(s.Pid),
		"uid":     strconv.Itoa(s.Uid),
		"gid":     strconv.Itoa(s.Gid),
		"version": s.Version,
		"cwd":     s.Cwd,
	}
	acc.AddFields("uwsgi_overview", fields, tags)

	u.gatherWorkers(acc, s)
	u.gatherApps(acc, s)
	u.gatherCores(acc, s)
}

func (u *Uwsgi) gatherWorkers(acc telegraf.Accumulator, s *StatsServer) {
	for _, w := range s.Workers {
		fields := map[string]interface{}{
			"requests":       w.Requests,
			"accepting":      w.Accepting,
			"delta_request":  w.DeltaRequests,
			"exceptions":     w.Exceptions,
			"harakiri_count": w.HarakiriCount,
			"signals":        w.Signals,
			"signal_queue":   w.SignalQueue,
			"status":         w.Status,
			"rss":            w.Rss,
			"vsz":            w.Vsz,
			"running_time":   w.RunningTime,
			"last_spawn":     w.LastSpawn,
			"respawn_count":  w.RespawnCount,
			"tx":             w.Tx,
			"avg_rt":         w.AvgRt,
		}
		tags := map[string]string{
			"worker_id": strconv.Itoa(w.WorkerId),
			"url":       s.Url,
			"pid":       strconv.Itoa(w.Pid),
		}

		acc.AddFields("uwsgi_workers", fields, tags)
	}
}

func (u *Uwsgi) gatherApps(acc telegraf.Accumulator, s *StatsServer) {
	for _, w := range s.Workers {
		for _, a := range w.Apps {
			fields := map[string]interface{}{
				"modifier1":    a.Modifier1,
				"requests":     a.Requests,
				"startup_time": a.StartupTime,
				"exceptions":   a.Exceptions,
			}
			tags := map[string]string{
				"app_id":     strconv.Itoa(a.AppId),
				"worker_id":  strconv.Itoa(w.WorkerId),
				"mountpoint": a.MountPoint,
				"chdir":      a.Chdir,
			}
			acc.AddFields("uwsgi_apps", fields, tags)
		}
	}
}

func (u *Uwsgi) gatherCores(acc telegraf.Accumulator, s *StatsServer) {
	for _, w := range s.Workers {
		for _, c := range w.Cores {
			fields := map[string]interface{}{
				"requests":           c.Requests,
				"static_requests":    c.StaticRequests,
				"routed_requests":    c.RoutedRequests,
				"offloaded_requests": c.OffloadedRequests,
				"write_errors":       c.WriteErrors,
				"read_errors":        c.ReadErrors,
				"in_request":         c.InRequest,
			}
			tags := map[string]string{
				"core_id":   strconv.Itoa(c.CoreId),
				"worker_id": strconv.Itoa(w.WorkerId),
			}
			acc.AddFields("uwsgi_cores", fields, tags)
		}

	}
}

func init() {
	inputs.Add("uwsgi", func() telegraf.Input { return &Uwsgi{} })
}
