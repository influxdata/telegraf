package prometheus

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"io"
	"net/http"
	"sync"
	"time"
)

type Prometheus struct {
	Urls []string
}

var sampleConfig = `
  ## An array of urls to scrape metrics from.
  urls = ["http://localhost:9100/metrics"]
`

func (r *Prometheus) SampleConfig() string {
	return sampleConfig
}

func (r *Prometheus) Description() string {
	return "Read metrics from one or many prometheus clients"
}

var ErrProtocolError = errors.New("prometheus protocol error")

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *Prometheus) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	var outerr error

	for _, serv := range g.Urls {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = g.gatherURL(serv, acc)
		}(serv)
	}

	wg.Wait()

	return outerr
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

func (g *Prometheus) gatherURL(url string, acc telegraf.Accumulator) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}
	format := expfmt.ResponseFormat(resp.Header)

	decoder := expfmt.NewDecoder(resp.Body, format)

	options := &expfmt.DecodeOptions{
		Timestamp: model.Now(),
	}
	sampleDecoder := &expfmt.SampleDecoder{
		Dec:  decoder,
		Opts: options,
	}

	for {
		var samples model.Vector
		err := sampleDecoder.Decode(&samples)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error getting processing samples for %s: %s",
				url, err)
		}
		for _, sample := range samples {
			tags := make(map[string]string)
			for key, value := range sample.Metric {
				if key == model.MetricNameLabel {
					continue
				}
				tags[string(key)] = string(value)
			}
			acc.Add("prometheus_"+string(sample.Metric[model.MetricNameLabel]),
				float64(sample.Value), tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{}
	})
}
