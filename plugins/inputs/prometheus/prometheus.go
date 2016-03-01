package prometheus

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
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
	collectDate := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	// Headers
	headers := make(map[string]string)
	for key, value := range headers {
		headers[key] = value
	}

	// Prepare Prometheus parser config
	promparser := PrometheusParser{
		PromFormat: headers,
	}

	metrics, err := promparser.Parse(body)
	if err != nil {
		return fmt.Errorf("error getting processing samples for %s: %s",
			url, err)
	}
	// Add (or not) collected metrics
	for _, metric := range metrics {
		tags := metric.Tags()
		tags["url"] = url
		acc.AddFields(metric.Name(), metric.Fields(), tags, collectDate)
	}

	return nil
}

func init() {
	inputs.Add("prometheus", func() telegraf.Input {
		return &Prometheus{}
	})
}
