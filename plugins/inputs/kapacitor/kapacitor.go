package kapacitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Kapacitor struct {
	URLs []string `toml:"urls"`

	Timeout internal.Duration

	client *http.Client
}

func (*Kapacitor) Description() string {
	return "Read Kapacitor-formatted JSON metrics from one or more HTTP endpoints"
}

func (*Kapacitor) SampleConfig() string {
	return `
  ## Multiple URLs from which to read Kapacitor-formatted JSON
  ## Default is "http://localhost:9092/kapacitor/v1/debug/vars".
  urls = [
    "http://localhost:9092/kapacitor/v1/debug/vars"
  ]

  ## http request & header timeout
  timeout = "5s"
`
}

func (k *Kapacitor) Gather(acc telegraf.Accumulator) error {
	if len(k.URLs) == 0 {
		k.URLs = []string{"http://localhost:9092/kapacitor/v1/debug/vars"}
	}

	if k.client == nil {
		k.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: k.Timeout.Duration,
			},
			Timeout: k.Timeout.Duration,
		}
	}

	errorChannel := make(chan error, len(k.URLs))

	var wg sync.WaitGroup
	for _, u := range k.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := k.gatherURL(acc, url); err != nil {
				errorChannel <- fmt.Errorf("[url=%s]: %s", url, err)
			}
		}(u)
	}

	wg.Wait()
	close(errorChannel)

	// If there weren't any errors, we can return nil now.
	if len(errorChannel) == 0 {
		return nil
	}

	// There were errors, so join them all together as one big error.
	errorStrings := make([]string, 0, len(errorChannel))
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	return errors.New(strings.Join(errorStrings, "\n"))
}

type object struct {
	Name   string                 `json:"name"`
	Values map[string]interface{} `json:"values"`
	Tags   map[string]string      `json:"tags"`
}

type memstats struct {
	Alloc         int64   `json:"Alloc"`
	TotalAlloc    int64   `json:"TotalAlloc"`
	Sys           int64   `json:"Sys"`
	Lookups       int64   `json:"Lookups"`
	Mallocs       int64   `json:"Mallocs"`
	Frees         int64   `json:"Frees"`
	HeapAlloc     int64   `json:"HeapAlloc"`
	HeapSys       int64   `json:"HeapSys"`
	HeapIdle      int64   `json:"HeapIdle"`
	HeapInuse     int64   `json:"HeapInuse"`
	HeapReleased  int64   `json:"HeapReleased"`
	HeapObjects   int64   `json:"HeapObjects"`
	StackInuse    int64   `json:"StackInuse"`
	StackSys      int64   `json:"StackSys"`
	MSpanInuse    int64   `json:"MSpanInuse"`
	MSpanSys      int64   `json:"MSpanSys"`
	MCacheInuse   int64   `json:"MCacheInuse"`
	MCacheSys     int64   `json:"MCacheSys"`
	BuckHashSys   int64   `json:"BuckHashSys"`
	GCSys         int64   `json:"GCSys"`
	OtherSys      int64   `json:"OtherSys"`
	NextGC        int64   `json:"NextGC"`
	LastGC        int64   `json:"LastGC"`
	PauseTotalNs  int64   `json:"PauseTotalNs"`
	NumGC         int64   `json:"NumGC"`
	GCCPUFraction float64 `json:"GCCPUFraction"`
}

type stats struct {
	CmdLine          []string           `json:"cmdline"`
	ClusterID        string             `json:"cluster_id"`
	Host             string             `json:"host"`
	Kapacitor        *map[string]object `json:"kapacitor"`
	MemStats         *memstats          `json:"memstats"`
	NumEnabledTasks  int                `json:"num_enabled_tasks"`
	NumSubscriptions int                `json:"num_subscriptions"`
	NumTasks         int                `json:"num_tasks"`
	Product          string             `json:"product"`
	ServerID         string             `json:"server_id"`
	Version          string             `json:"version"`
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (k *Kapacitor) gatherURL(
	acc telegraf.Accumulator,
	url string,
) error {
	now := time.Now()

	resp, err := k.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	var s stats
	err = dec.Decode(&s)
	if err != nil {
		return err
	}

	acc.AddFields("kapacitor_memstats",
		map[string]interface{}{
			"alloc":           s.MemStats.Alloc,
			"total_alloc":     s.MemStats.TotalAlloc,
			"sys":             s.MemStats.Sys,
			"lookups":         s.MemStats.Lookups,
			"mallocs":         s.MemStats.Mallocs,
			"frees":           s.MemStats.Frees,
			"heap_alloc":      s.MemStats.HeapAlloc,
			"heap_sys":        s.MemStats.HeapSys,
			"heap_idle":       s.MemStats.HeapIdle,
			"heap_inuse":      s.MemStats.HeapInuse,
			"heap_released":   s.MemStats.HeapReleased,
			"heap_objects":    s.MemStats.HeapObjects,
			"stack_inuse":     s.MemStats.StackInuse,
			"stack_sys":       s.MemStats.StackSys,
			"mspan_inuse":     s.MemStats.MSpanInuse,
			"mspan_sys":       s.MemStats.MSpanSys,
			"mcache_inuse":    s.MemStats.MCacheInuse,
			"mcache_sys":      s.MemStats.MCacheSys,
			"buck_hash_sys":   s.MemStats.BuckHashSys,
			"gc_sys":          s.MemStats.GCSys,
			"other_sys":       s.MemStats.OtherSys,
			"next_gc":         s.MemStats.NextGC,
			"last_gc":         s.MemStats.LastGC,
			"pause_total_ns":  s.MemStats.PauseTotalNs,
			"num_gc":          s.MemStats.NumGC,
			"gcc_pu_fraction": s.MemStats.GCCPUFraction,
		},
		map[string]string{
			"url":         url,
			"kap_version": s.Version,
		},
		now)

	acc.AddFields("kapacitor",
		map[string]interface{}{
			"num_enabled_tasks": s.NumEnabledTasks,
			"num_subscriptions": s.NumSubscriptions,
			"num_tasks":         s.NumTasks,
		},
		map[string]string{
			"url":         url,
			"kap_version": s.Version,
		},
		now)

	for _, obj := range *s.Kapacitor {

		// Strip out high-cardinality or duplicative tags
		excludeTags := []string{"host", "cluster_id", "server_id"}
		for _, key := range excludeTags {
			if _, ok := obj.Tags[key]; ok {
				delete(obj.Tags, key)
			}
		}

		// Convert time-related string field to int
		if _, ok := obj.Values["avg_exec_time_ns"]; ok {
			d, err := time.ParseDuration(obj.Values["avg_exec_time_ns"].(string))
			if err != nil {
				continue
			}
			obj.Values["avg_exec_time_ns"] = d.Nanoseconds()
		}

		acc.AddFields(
			"kapacitor_"+obj.Name,
			obj.Values,
			obj.Tags,
			now,
		)
	}

	return nil
}

func init() {
	inputs.Add("kapacitor", func() telegraf.Input {
		return &Kapacitor{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
