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

	u.gatherWorkers(acc, &s)

	return nil
}

func (u *Uwsgi) gatherWorkers(acc telegraf.Accumulator, s *StatsServer) error {
	for _, w := range s.Workers {
		fields := map[string]interface{}{
			"requests":       w.Requests,
			"accepting":      w.Accepting,
			"delta_request":  w.DeltaRequests,
			"harakiri_count": w.HarakiriCount,
			"signals":        w.Signals,
			"signal_queue":   w.SignalQueue,
			"status":         w.Status,
			"rss":            w.RSS,
			"vsz":            w.VSZ,
			"running_time":   w.RunningTime,
			"last_spawn":     w.LastSpawn,
			"respawn_count":  w.RespawnCount,
			"tx":             w.TX,
			"avg_rt":         w.AvgRT,
		}
		tags := map[string]string{
			"id":  strconv.Itoa(w.Id),
			"url": s.Url,
			"pid": strconv.Itoa(w.Pid),
		}

		acc.AddFields("uwsgi_workers", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("uwsgi", func() telegraf.Input { return &Uwsgi{} })
}
