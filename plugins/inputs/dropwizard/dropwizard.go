package dropwizard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Dropwizard struct {
	URLs []string `toml:"urls"`
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	Timeout internal.Duration

	// should we pass in the format or just the number of decimal places?
	FloatFieldFormat string `toml:"float_field_format"`

	// can skip any idle metric that has a count field
	SkipIdleMetrics bool `toml:"skip_idle_metrics"`
	// we don't want to store the whole metric, just the metric name and count field
	previousCountValues map[string]int64

	client *http.Client
}

func (*Dropwizard) Description() string {
	return "Read Dropwizard-formatted JSON metrics from one or more HTTP endpoints"
}

func (*Dropwizard) SampleConfig() string {
	return `
  ## Works with Dropwizard metrics endpoint out of the box

  ## Multiple URLs from which to read Dropwizard-formatted JSON
  ## Default is "http://localhost:8081/metrics".
  urls = [
    "http://localhost:8081/metrics"
  ]

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## http request & header timeout
  ## defaults to 5s if not set
  timeout = "10s"

  ## format the floating number fields on all metrics to round them off
  ## this avoids getting very small numbers like 5.647645996854652E-23
  ## defaults to "%.2f" if not set
  #float_field_format = "%.2f"

  ## skip any metric whose "count" field hasn't changed since last time the metric was pulled
  ## this applies to metric types: Counter, Histogram, Meter & Timer
  ## defaults to false if not set
  skip_idle_metrics = true

  ## exclude some built-in metrics
  # namedrop = [
  #  "jvm.classloader*",
  #  "jvm.buffers*", 
  #  "jvm.gc*",
  #  "jvm.memory.heap*",
  #  "jvm.memory.non-heap*",
  #  "jvm.memory.pools*",
  #  "jvm.threads*",
  #  "jvm.attribute.uptime",
  #  "jvm.filedescriptor",
  #  "io.dropwizard.jetty.MutableServletContextHandler*",
  #  "org.eclipse.jetty.util*" 
  # ]

  ## include only the required fields (applies to all metrics types)
  # fieldpass = [
  #  "count",
  #  "max",
  #  "p999",
  #  "m5_Rate",
  #  "value" 
  # ]
`
}

func (d *Dropwizard) Gather(acc telegraf.Accumulator) error {
	if len(d.URLs) == 0 {
		d.URLs = []string{"http://localhost:8081/metrics"}
	}

	if d.client == nil {
		tlsCfg, err := internal.GetTLSConfig(
			d.SSLCert, d.SSLKey, d.SSLCA, d.InsecureSkipVerify)
		if err != nil {
			return err
		}
		d.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: d.Timeout.Duration,
				TLSClientConfig:       tlsCfg,
			},
			Timeout: d.Timeout.Duration,
		}
	}

	var wg sync.WaitGroup
	for _, u := range d.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := d.gatherURL(acc, url); err != nil {
				acc.AddError(fmt.Errorf("[url=%s]: %s", url, err))
			}
		}(u)
	}

	wg.Wait()

	return nil
}

// Gauge values can be of different types 

type gaugeValueType int

const (
        IntType gaugeValueType = iota
        FloatType
        StringType
)

type gaugeValue struct {
	IntValue int64
	FloatValue float64
	StringValue string
	Type gaugeValueType
}

func (gv *gaugeValue) UnmarshalJSON(b []byte) error {
	jsonString := string(b)
	if intValue, err := strconv.ParseInt(jsonString, 10, 64); err == nil {
		*gv = gaugeValue{ 
			IntValue: intValue,
			Type: IntType,
		}
	} else if floatValue, err := strconv.ParseFloat(jsonString, 64); err == nil {
		*gv = gaugeValue{ 
			FloatValue: floatValue,
			Type: FloatType,
		}
	} else {
		*gv = gaugeValue{ 
			StringValue: strings.Trim(jsonString, "\""),
			Type: StringType,
		}
	}

	return nil
}

type gauge struct {
	Value gaugeValue `json:"value"`
}

type counter struct {
	Count int64 `json:"count"`
}

func (c *counter) UnmarshalJSON(b []byte) error {
	jsonString := string(b)
	if intValue, err := strconv.ParseInt(jsonString, 10, 64); err == nil {
		*c = counter{ 
			Count: intValue,
		}
	} 
	return nil
}

type histogram struct {
	Counter counter `json:"count"`
	Max    int64   `json:"max"`
	Mean   float64 `json:"mean"`
	Min    int64   `json:"min"`
	P50    float64 `json:"p50"`
	P75    float64 `json:"p75"`
	P95    float64 `json:"p95"`
	P98    float64 `json:"p98"`
	P99    float64 `json:"p99"`
	P999   float64 `json:"p999"`
	Stddev float64 `json:"stddev"`
}

type meter struct {
	Counter counter  `json:"count"`
	M15Rate  float64 `json:"m15_rate"`
	M1Rate   float64 `json:"m1_rate"`
	M5Rate   float64 `json:"m5_rate"`
	MeanRate float64 `json:"mean_rate"`
	Units    string  `json:"units"`
}

type timer struct {
	Counter	      counter `json:"count"`
	Max           float64 `json:"max"`
	Mean          float64 `json:"mean"`
	Min           float64 `json:"min"`
	P50           float64 `json:"p50"`
	P75           float64 `json:"p75"`
	P95           float64 `json:"p95"`
	P98           float64 `json:"p98"`
	P99           float64 `json:"p99"`
	P999          float64 `json:"p999"`
	Stddev        float64 `json:"stddev"`
	M15Rate       float64 `json:"m15_rate"`
	M1Rate        float64 `json:"m1_rate"`
	M5Rate        float64 `json:"m5_rate"`
	MeanRate      float64 `json:"mean_rate"`
	DurationUnits string  `json:"duration_units"`
	RateUnits     string  `json:"rate_units"`
}

type metrics struct {
	Version    string               `json:"version"`
	Gauges     map[string]gauge     `json:"gauges"`
	Counters   map[string]counter   `json:"counters"`
	Histograms map[string]histogram `json:"histograms"`
	Meters     map[string]meter     `json:"meters"`
	Timers     map[string]timer     `json:"timers"`
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (d *Dropwizard) gatherURL(
	acc telegraf.Accumulator,
	url string,
) error {
	now := time.Now()

	resp, err := d.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	metrics, err := d.DecodeJSONMetrics(resp.Body)
	if err != nil {
		return err
	}

	// tags can be specified in config either globally or per input
	// through built-in functionality
	var tags map[string]string = nil

	if d.SkipIdleMetrics && d.previousCountValues == nil {
		d.previousCountValues = make(map[string]int64)
	}

	for name, g := range metrics.Gauges {
		if g.Value.Type == IntType {
			acc.AddGauge(name,
				map[string]interface{}{ "value": g.Value.IntValue },
				tags,
				now)
		} else if g.Value.Type == FloatType {
			acc.AddGauge(name,
				map[string]interface{}{ "value": d.FormatFloat(g.Value.FloatValue) },
				tags,
				now)
		}
	}

	for name, c := range metrics.Counters {
		if d.canSkipMetric(name, &c) {
			continue
		}

		acc.AddCounter(name,
			map[string]interface{}{ "count": c.Count },
			tags,
			now)
	}

	for name, h := range metrics.Histograms {
		if d.canSkipMetric(name, &h.Counter) {
			continue
		}

		acc.AddHistogram(name,
			map[string]interface{}{ 
				"count": h.Counter.Count,
				"max": h.Max,
				"mean": d.FormatFloat(h.Mean),
				"min": h.Min,
				"p50": d.FormatFloat(h.P50),
				"p75": d.FormatFloat(h.P75),
				"p95": d.FormatFloat(h.P95),
				"p98": d.FormatFloat(h.P98),
				"p99": d.FormatFloat(h.P99),
				"p999": d.FormatFloat(h.P999),
				"stddev": h.Stddev,
			},
			tags,
			now)
	}

	//TODO what to do with the Units?
	for name, m := range metrics.Meters {
		if d.canSkipMetric(name, &m.Counter) {
			continue
		}

		acc.AddHistogram(name,
			map[string]interface{}{ 
				"count": m.Counter.Count,
				"m15_rate": d.FormatFloat(m.M15Rate),
				"m1_rate": d.FormatFloat(m.M1Rate),
				"m5_rate": d.FormatFloat(m.M5Rate),
				"mean_rate": d.FormatFloat(m.MeanRate),
			},
			tags,
			now)
	}

	//TODO what to do with duration and rate units?
	for name, t := range metrics.Timers {
		if d.canSkipMetric(name, &t.Counter) {
			continue
		}

		acc.AddFields(name,
			map[string]interface{}{ 
				"count": t.Counter.Count,
				"max": d.FormatFloat(t.Max),
				"mean": d.FormatFloat(t.Mean),
				"min": d.FormatFloat(t.Min),
				"p50": d.FormatFloat(t.P50),
				"p75": d.FormatFloat(t.P75),
				"p95": d.FormatFloat(t.P95),
				"p98": d.FormatFloat(t.P98),
				"p99": d.FormatFloat(t.P99),
				"p999": d.FormatFloat(t.P999),
				"stddev": d.FormatFloat(t.Stddev),
				"m15_rate": d.FormatFloat(t.M15Rate),
				"m1_rate": d.FormatFloat(t.M1Rate),
				"m5_rate": d.FormatFloat(t.M5Rate),
				"mean_rate": d.FormatFloat(t.MeanRate),
			},
			tags,
			now)
	}

	return nil
}

func init() {
	inputs.Add("dropwizard", func() telegraf.Input {
		return &Dropwizard{
			Timeout: internal.Duration{Duration: time.Second * 5},
			FloatFieldFormat: "%.2f",
		}
	})
}

func (*Dropwizard) DecodeJSONMetrics(r io.Reader) (metrics, error) {
	var decodedMetrics metrics
	err := json.NewDecoder(r).Decode(&decodedMetrics)
	if err != nil {
		return decodedMetrics, err
	}
	return decodedMetrics, nil
}

func (d *Dropwizard) FormatFloat(f float64) float64 {
	if d.FloatFieldFormat == "" {
		return f
	}
	floatValue, err := strconv.ParseFloat(fmt.Sprintf(d.FloatFieldFormat, f), 64)
	if err != nil {
		return f
	}
	return floatValue
}

func (d *Dropwizard) canSkipMetric(name string, c *counter) bool {
	if d.SkipIdleMetrics {
		if val, ok := d.previousCountValues[name]; ok {
			if val == c.Count {
				return true
			}
		} 
		d.previousCountValues[name] = c.Count
	}

	return false
}