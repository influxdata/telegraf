package influxdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	maxErrorResponseBodyLength = 1024
)

type APIError struct {
	StatusCode  int
	Reason      string
	Description string `json:"error"`
}

func (e *APIError) Error() string {
	if e.Description != "" {
		return e.Reason + ": " + e.Description
	}
	return e.Reason
}

type InfluxDB struct {
	URLs     []string        `toml:"urls"`
	Username string          `toml:"username"`
	Password string          `toml:"password"`
	Timeout  config.Duration `toml:"timeout"`
	tls.ClientConfig

	client *http.Client
}

func (i *InfluxDB) Gather(acc telegraf.Accumulator) error {
	if len(i.URLs) == 0 {
		i.URLs = []string{"http://localhost:8086/debug/vars"}
	}

	if i.client == nil {
		tlsCfg, err := i.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		i.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(i.Timeout),
				TLSClientConfig:       tlsCfg,
			},
			Timeout: time.Duration(i.Timeout),
		}
	}

	var wg sync.WaitGroup
	for _, u := range i.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := i.gatherURL(acc, url); err != nil {
				acc.AddError(err)
			}
		}(u)
	}

	wg.Wait()

	return nil
}

type point struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

type memstats struct {
	Alloc         int64      `json:"Alloc"`
	TotalAlloc    int64      `json:"TotalAlloc"`
	Sys           int64      `json:"Sys"`
	Lookups       int64      `json:"Lookups"`
	Mallocs       int64      `json:"Mallocs"`
	Frees         int64      `json:"Frees"`
	HeapAlloc     int64      `json:"HeapAlloc"`
	HeapSys       int64      `json:"HeapSys"`
	HeapIdle      int64      `json:"HeapIdle"`
	HeapInuse     int64      `json:"HeapInuse"`
	HeapReleased  int64      `json:"HeapReleased"`
	HeapObjects   int64      `json:"HeapObjects"`
	StackInuse    int64      `json:"StackInuse"`
	StackSys      int64      `json:"StackSys"`
	MSpanInuse    int64      `json:"MSpanInuse"`
	MSpanSys      int64      `json:"MSpanSys"`
	MCacheInuse   int64      `json:"MCacheInuse"`
	MCacheSys     int64      `json:"MCacheSys"`
	BuckHashSys   int64      `json:"BuckHashSys"`
	GCSys         int64      `json:"GCSys"`
	OtherSys      int64      `json:"OtherSys"`
	NextGC        int64      `json:"NextGC"`
	LastGC        int64      `json:"LastGC"`
	PauseTotalNs  int64      `json:"PauseTotalNs"`
	PauseNs       [256]int64 `json:"PauseNs"`
	NumGC         int64      `json:"NumGC"`
	GCCPUFraction float64    `json:"GCCPUFraction"`
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (i *InfluxDB) gatherURL(
	acc telegraf.Accumulator,
	url string,
) error {
	shardCounter := 0
	now := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	req.Header.Set("User-Agent", "Telegraf/"+internal.Version())

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readResponseError(resp)
	}

	// It would be nice to be able to decode into a map[string]point, but
	// we'll get a decoder error like:
	// `json: cannot unmarshal array into Go value of type influxdb.point`
	// if any of the values aren't objects.
	// To avoid that error, we decode by hand.
	dec := json.NewDecoder(resp.Body)

	// Parse beginning of object
	if t, err := dec.Token(); err != nil {
		return err
	} else if t != json.Delim('{') {
		return errors.New("document root must be a JSON object")
	}

	// Loop through rest of object
	for {
		// Nothing left in this object, we're done
		if !dec.More() {
			break
		}

		// Read in a string key. We don't do anything with the top-level keys,
		// so it's discarded.
		key, err := dec.Token()
		if err != nil {
			return err
		}

		if keyStr, ok := key.(string); ok {
			if keyStr == "memstats" {
				var m memstats
				if err := dec.Decode(&m); err != nil {
					continue
				}
				acc.AddFields("influxdb_memstats",
					map[string]interface{}{
						"alloc":           m.Alloc,
						"total_alloc":     m.TotalAlloc,
						"sys":             m.Sys,
						"lookups":         m.Lookups,
						"mallocs":         m.Mallocs,
						"frees":           m.Frees,
						"heap_alloc":      m.HeapAlloc,
						"heap_sys":        m.HeapSys,
						"heap_idle":       m.HeapIdle,
						"heap_inuse":      m.HeapInuse,
						"heap_released":   m.HeapReleased,
						"heap_objects":    m.HeapObjects,
						"stack_inuse":     m.StackInuse,
						"stack_sys":       m.StackSys,
						"mspan_inuse":     m.MSpanInuse,
						"mspan_sys":       m.MSpanSys,
						"mcache_inuse":    m.MCacheInuse,
						"mcache_sys":      m.MCacheSys,
						"buck_hash_sys":   m.BuckHashSys,
						"gc_sys":          m.GCSys,
						"other_sys":       m.OtherSys,
						"next_gc":         m.NextGC,
						"last_gc":         m.LastGC,
						"pause_total_ns":  m.PauseTotalNs,
						"pause_ns":        m.PauseNs[(m.NumGC+255)%256],
						"num_gc":          m.NumGC,
						"gc_cpu_fraction": m.GCCPUFraction,
					},
					map[string]string{
						"url": url,
					})
			}
		}

		// Attempt to parse a whole object into a point.
		// It might be a non-object, like a string or array.
		// If we fail to decode it into a point, ignore it and move on.
		var p point
		if err := dec.Decode(&p); err != nil {
			continue
		}

		if p.Tags == nil {
			p.Tags = make(map[string]string)
		}

		// If the object was a point, but was not fully initialized,
		// ignore it and move on.
		if p.Name == "" || p.Values == nil || len(p.Values) == 0 {
			continue
		}

		if p.Name == "shard" {
			shardCounter++
		}

		// Add a tag to indicate the source of the data.
		p.Tags["url"] = url

		acc.AddFields(
			"influxdb_"+p.Name,
			p.Values,
			p.Tags,
			now,
		)
	}

	acc.AddFields("influxdb",
		map[string]interface{}{
			"n_shards": shardCounter,
		},
		nil,
		now,
	)

	return nil
}

func readResponseError(resp *http.Response) error {
	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Reason:     resp.Status,
	}

	var buf bytes.Buffer
	r := io.LimitReader(resp.Body, maxErrorResponseBodyLength)
	_, err := buf.ReadFrom(r)
	if err != nil {
		return apiError
	}

	err = json.Unmarshal(buf.Bytes(), apiError)
	if err != nil {
		return apiError
	}

	return apiError
}

func init() {
	inputs.Add("influxdb", func() telegraf.Input {
		return &InfluxDB{
			Timeout: config.Duration(time.Second * 5),
		}
	})
}
