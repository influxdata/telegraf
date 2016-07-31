// Package uwsgi implements a telegraf plugin for
// collecting uwsgi stats from the uwsgi stats server.
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

// uWSGI Server struct
type Uwsgi struct {
	Servers []string `toml:"server"`
}

// Errors is a list of errors accumulated during an interval.
type Errors []error

func (errs Errors) Error() string {
	s := ""
	for _, err := range errs {
		if s == "" {
			s = err.Error()
		} else {
			s = s + ". " + err.Error()
		}
	}
	return s
}

// NestedError wraps an error returned from deeper in the code.
type NestedError struct {
	// Err is the error from where the NestedError was constructed.
	Err error
	// NestedError is the error that was passed back from the called function.
	NestedErr error
}

// Error returns a concatenated string of all the nested errors.
func (ne NestedError) Error() string {
	return ne.Err.Error() + ": " + ne.NestedErr.Error()
}

// Errorf is a convenience function for constructing a NestedError.
func Errorf(err error, msg string, format ...interface{}) error {
	return NestedError{
		NestedErr: err,
		Err:       fmt.Errorf(msg, format...),
	}
}

// Description returns the plugin description
func (u *Uwsgi) Description() string {
	return "Read uWSGI metrics."
}

// SampleConfig returns the sample configuration
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

// Gather collect data from uWSGI Server
func (u *Uwsgi) Gather(acc telegraf.Accumulator) error {
	var errs Errors
	for _, s := range u.Servers {
		n, err := url.Parse(s)
		if err != nil {
			return fmt.Errorf("Could not parse uWSGI Stats Server url '%s': %s", s, err)
		}
		if err := u.gatherServer(acc, n); err != nil {
			errs = append(errs, Errorf(err, "server %s", s))
		}

	}
	return errs
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
			return Errorf(err, "Could not connect to uWSGI Stats Server '%s'", url.String())
		}
		r = resp.Body
	default:
		return fmt.Errorf("'%s' is not a valid URL", url.String())
	}

	if err != nil {
		return Errorf(err, "Could not connect to uWSGI Stats Server '%s'", url.String())
	}
	defer r.Close()

	var s StatsServer
	s.Url = url.String()

	dec := json.NewDecoder(r)
	if err := dec.Decode(&s); err != nil {
		return Errorf(err, "Could not decode json payload from server '%s'", url.String())
	}

	u.gatherStatServer(acc, &s)

	return err
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
