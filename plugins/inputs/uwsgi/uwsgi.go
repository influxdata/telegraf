package uwsgi

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net/http"
	"strconv"
	"time"
)

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

type Uwsgi struct {
	URLs []string `toml:"urls"`
}

func (u *Uwsgi) Description() string {
	return "Read uWSGI metrics."
}

func (u *Uwsgi) SampleConfig() string {
	return `
    ### List with urls of uWSGI Stats servers
    urls = []
`
}

func (u *Uwsgi) Gather(acc telegraf.Accumulator) error {
	for _, url := range u.URLs {
		err := u.gatherURL(acc, url)
		if err != nil {
			return err
		}

	}
	return nil
}

func (u *Uwsgi) gatherURL(acc telegraf.Accumulator, url string) error {
	resp, err := client.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var s StatsServer
	s.Url = url

	dec := json.NewDecoder(resp.Body)
	dec.Decode(&s)

	if err != nil {
		return err
	}

	u.gatherStatServer(acc, &s)
	u.gatherWorkers(acc, &s)
	u.gatherApps(acc, &s)
	u.gatherCores(acc, &s)

	return nil
}

func (u *Uwsgi) gatherStatServer(acc telegraf.Accumulator, s *StatsServer) error {
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

	return nil

}

func (u *Uwsgi) gatherWorkers(acc telegraf.Accumulator, s *StatsServer) error {
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

	return nil
}

func (u *Uwsgi) gatherApps(acc telegraf.Accumulator, s *StatsServer) error {
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

	return nil
}

func (u *Uwsgi) gatherCores(acc telegraf.Accumulator, s *StatsServer) error {
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

	return nil
}

func init() {
	inputs.Add("uwsgi", func() telegraf.Input { return &Uwsgi{} })
}
