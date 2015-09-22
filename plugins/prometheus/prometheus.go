package prometheus

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/koksan83/telegraf/plugins"
	"github.com/prometheus/client_golang/extraction"
	"github.com/prometheus/client_golang/model"
)

type Prometheus struct {
	Urls []string
}

var sampleConfig = `
	# An array of urls to scrape metrics from.
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
func (g *Prometheus) Gather(acc plugins.Accumulator) error {
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

func (g *Prometheus) gatherURL(url string, acc plugins.Accumulator) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}
	processor, err := extraction.ProcessorForRequestHeader(resp.Header)
	if err != nil {
		return fmt.Errorf("error getting extractor for %s: %s", url, err)
	}

	ingestor := &Ingester{
		acc: acc,
	}

	options := &extraction.ProcessOptions{
		Timestamp: model.TimestampFromTime(time.Now()),
	}

	err = processor.ProcessSingle(resp.Body, ingestor, options)
	if err != nil {
		return fmt.Errorf("error getting processing samples for %s: %s", url, err)
	}
	return nil
}

type Ingester struct {
	acc plugins.Accumulator
}

// Ingest implements an extraction.Ingester.
func (i *Ingester) Ingest(samples model.Samples) error {
	for _, sample := range samples {
		tags := map[string]string{}
		for key, value := range sample.Metric {
			if key == model.MetricNameLabel {
				continue
			}
			tags[string(key)] = string(value)
		}
		i.acc.Add(string(sample.Metric[model.MetricNameLabel]), float64(sample.Value), tags)
	}
	return nil
}

func init() {
	plugins.Add("prometheus", func() plugins.Plugin {
		return &Prometheus{}
	})
}
