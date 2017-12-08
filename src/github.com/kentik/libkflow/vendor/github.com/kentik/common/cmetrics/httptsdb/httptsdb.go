package httptsdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kentik/go-metrics"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	MIN_FOR_HOST            = 6
	MAX_SEND_TRIES          = 2
	CLIENT_RESPONSE_TIMEOUT = 5 * time.Second
	CLIENT_KEEP_ALIVE       = 60 * time.Second
	CLIENT_TLS_TIMEOUT      = 5 * time.Second
	ContentType             = "application/json"
	API_EMAIL_HEADER        = "X-CH-Auth-Email"
	API_PASSWORD_HEADER     = "X-CH-Auth-API-Token"
)

var shortHostName string = ""

type TSDBMetric struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     int64             `json:"value"`
	Tags      map[string]string `json:"tags"`
}

// OpenTSDBConfig provides a container with configuration parameters for
// the OpenTSDB exporter
type OpenTSDBConfig struct {
	Addr               string            // Network address to connect to
	Registry           metrics.Registry  // Registry to be exported
	FlushInterval      time.Duration     // Flush interval
	DurationUnit       time.Duration     // Time conversion unit for durations
	Prefix             string            // Prefix to be prepended to metric names
	Debug              bool              // write to stdout for debug
	Tags               map[string]string // add these tags to each metric writen
	Send               chan []byte       // manage # of outstanding http requests here.
	MaxHttpOutstanding int
	ProxyUrl           string
	Extra              map[string]string
	ApiEmail           *string
	ApiPassword        *string
}

// OpenTSDB is a blocking exporter function which reports metrics in r
// to a TSDB server located at addr, flushing them every d duration
// and prepending metric names with prefix.
func OpenTSDB(r metrics.Registry, d time.Duration, prefix string, addr string, maxOutstanding int) {
	OpenTSDBWithConfig(OpenTSDBConfig{
		Addr:               addr,
		Registry:           r,
		FlushInterval:      d,
		DurationUnit:       time.Nanosecond,
		Prefix:             prefix,
		Debug:              false,
		MaxHttpOutstanding: maxOutstanding,
		Send:               make(chan []byte, maxOutstanding),
		Tags:               map[string]string{},
		Extra:              nil,
	})
}

// OpenTSDBWithConfig is a blocking exporter function just like OpenTSDB,
// but it takes a OpenTSDBConfig instead.
func OpenTSDBWithConfig(c OpenTSDBConfig) {
	go c.runSend()

	for _ = range time.Tick(c.FlushInterval) {
		if err := openTSDB(&c); nil != err {
			log.Println(err)
		}
	}
}

func getShortHostname() string {
	if shortHostName == "" {
		host, _ := os.Hostname()
		strings.Replace(host, ".", "_", -1)
		shortHostName = host
	}
	return shortHostName
}

func setTags(in map[string]string, mtype string) map[string]string {
	out := make(map[string]string)

	// Copy these over as a base.
	for k, v := range in {
		out[k] = v
	}

	// Add in type, and send
	out["type"] = mtype
	return out
}

func (c *OpenTSDBConfig) runSend() {
	tr := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  false,
		Dial: (&net.Dialer{
			Timeout:   CLIENT_RESPONSE_TIMEOUT,
			KeepAlive: CLIENT_KEEP_ALIVE,
		}).Dial,
		TLSHandshakeTimeout: CLIENT_TLS_TIMEOUT,
	}

	// Add a proxy if needed.
	if c.ProxyUrl != "" {
		proxyUrl, err := url.Parse(c.ProxyUrl)
		if err != nil {
			fmt.Printf("Error setting proxy: %v\n", err)
		} else {
			tr.Proxy = http.ProxyURL(proxyUrl)
			fmt.Printf("Set outbound proxy: %s\n", c.ProxyUrl)
		}
	}

	client := &http.Client{Transport: tr, Timeout: CLIENT_RESPONSE_TIMEOUT}

	for r := range c.Send {
		if c.Debug {
			fmt.Printf("Metrics: %v", string(r))
		} else {
			for i := 0; i < MAX_SEND_TRIES; i++ {
				req, err := http.NewRequest("POST", c.Addr, bytes.NewBuffer(r))
				if err != nil {
					fmt.Printf("Error Creating Request: %v\n", err)
					continue
				}

				req.Header.Add("Content-Type", ContentType)

				if c.ApiEmail != nil && c.ApiPassword != nil {
					req.Header.Add(API_EMAIL_HEADER, *c.ApiEmail)
					req.Header.Add(API_PASSWORD_HEADER, *c.ApiPassword)
				}

				resp, err := client.Do(req)
				if err != nil {
					if i > 0 {
						fmt.Printf("Error Posting to %s: %v\n", c.Addr, err)
					} else {
						fmt.Printf("Retry Posting to %s: %v\n", c.Addr, err)
					}
					client = &http.Client{Transport: tr, Timeout: CLIENT_RESPONSE_TIMEOUT}
				} else {
					// Fire and forget
					io.Copy(ioutil.Discard, resp.Body)
					resp.Body.Close()
					break
				}
			}
		}
	}
}

/**
Write out additional tags
*/
func openTSDB(c *OpenTSDBConfig) error {

	shortHostnameBase := getShortHostname()
	now := time.Now().Unix()
	sendBody := make([]TSDBMetric, 0)
	du := float64(c.DurationUnit)

	c.Registry.Each(func(baseName string, i interface{}) {

		pts := strings.Split(baseName, ".")
		name := pts[0]
		tags := make(map[string]string)

		// Copy these over as a base.
		for k, v := range c.Tags {
			tags[k] = v
		}

		// And add the rest here.
		if len(pts) > MIN_FOR_HOST {
			pLen := len(pts)
			tags["host"] = shortHostnameBase
			tags["cid"] = strings.Join(pts[pLen-4:pLen-3], "_")
			tags["did"] = strings.Join(pts[pLen-4:pLen-1], "_")
			tags["sid"] = strings.Join(pts[pLen-1:], "_")
		} else {
			tags["host"] = shortHostnameBase

			if len(pts) >= 4 {
				tags["cid"] = strings.Join(pts[1:2], "_")
				tags["did"] = strings.Join(pts[3:4], "_")
			}
		}

		if c.Extra != nil {
			for k, v := range c.Extra {
				tags[k] = v
			}
		}

		switch metric := i.(type) {
		case metrics.Counter:
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: metric.Count(), Tags: setTags(tags, "count")})
		case metrics.Gauge:
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: metric.Value(), Tags: setTags(tags, "value")})
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: h.Count(), Tags: setTags(tags, "count")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: h.Min(), Tags: setTags(tags, "min")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: h.Max(), Tags: setTags(tags, "max")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(h.Mean()), Tags: setTags(tags, "mean")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(h.StdDev()), Tags: setTags(tags, "std-dev")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[0]), Tags: setTags(tags, "50-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[1]), Tags: setTags(tags, "75-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[2]), Tags: setTags(tags, "95-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[3]), Tags: setTags(tags, "99-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[4]), Tags: setTags(tags, "999-percentile")})
			metric.Clear()
		case metrics.Meter:
			m := metric.Snapshot()
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: m.Count(), Tags: setTags(tags, "count")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(m.Rate1()), Tags: setTags(tags, "one-minute")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(m.Rate5()), Tags: setTags(tags, "five-minute")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(m.Rate15()), Tags: setTags(tags, "fifteen-minute")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(m.RateMean()), Tags: setTags(tags, "mean")})
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: t.Count(), Tags: setTags(tags, "count")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: t.Min() / int64(du), Tags: setTags(tags, "min")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: t.Max() / int64(du), Tags: setTags(tags, "max")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(t.Mean() / du), Tags: setTags(tags, "mean")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(t.StdDev()), Tags: setTags(tags, "std-dev")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[0] / du), Tags: setTags(tags, "50-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[1] / du), Tags: setTags(tags, "75-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[2] / du), Tags: setTags(tags, "95-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[3] / du), Tags: setTags(tags, "99-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(ps[4] / du), Tags: setTags(tags, "999-percentile")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(t.Rate1()), Tags: setTags(tags, "one-minute")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(t.Rate5()), Tags: setTags(tags, "five-minute")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(t.Rate15()), Tags: setTags(tags, "fifteen-minute")})
			sendBody = append(sendBody, TSDBMetric{Metric: c.Prefix + "." + name, Timestamp: now, Value: int64(t.RateMean()), Tags: setTags(tags, "mean-rate")})
			metric.Clear()
		}
	})

	if len(sendBody) > 0 {
		if ebytes, err := json.Marshal(sendBody); err != nil {
			fmt.Printf("Error encoding json: %v\n", err)
		} else {
			if len(c.Send) < c.MaxHttpOutstanding {
				c.Send <- ebytes
			} else {
				fmt.Printf("Dropping flow: Q at %d\n", len(c.Send))
			}
		}
	}

	return nil
}
