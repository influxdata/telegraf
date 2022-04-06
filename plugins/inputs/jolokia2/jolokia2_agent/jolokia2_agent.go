package jolokia2_agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2/common"
)

type JolokiaAgent struct {
	DefaultFieldPrefix    string
	DefaultFieldSeparator string
	DefaultTagPrefix      string

	URLs            []string `toml:"urls"`
	Username        string
	Password        string
	ResponseTimeout config.Duration `toml:"response_timeout"`

	tls.ClientConfig

	Metrics  []common.MetricConfig `toml:"metric"`
	gatherer *common.Gatherer
	clients  []*common.Client
}

func (ja *JolokiaAgent) Gather(acc telegraf.Accumulator) error {
	if ja.gatherer == nil {
		ja.gatherer = common.NewGatherer(ja.createMetrics())
	}

	// Initialize clients once
	if ja.clients == nil {
		ja.clients = make([]*common.Client, 0, len(ja.URLs))
		for _, url := range ja.URLs {
			client, err := ja.createClient(url)
			if err != nil {
				acc.AddError(fmt.Errorf("unable to create client for %s: %v", url, err))
				continue
			}
			ja.clients = append(ja.clients, client)
		}
	}

	var wg sync.WaitGroup

	for _, client := range ja.clients {
		wg.Add(1)
		go func(client *common.Client) {
			defer wg.Done()

			err := ja.gatherer.Gather(client, acc)
			if err != nil {
				acc.AddError(fmt.Errorf("unable to gather metrics for %s: %v", client.URL, err))
			}
		}(client)
	}

	wg.Wait()

	return nil
}

func (ja *JolokiaAgent) createMetrics() []common.Metric {
	var metrics []common.Metric

	for _, metricConfig := range ja.Metrics {
		metrics = append(metrics, common.NewMetric(metricConfig,
			ja.DefaultFieldPrefix, ja.DefaultFieldSeparator, ja.DefaultTagPrefix))
	}

	return metrics
}

func (ja *JolokiaAgent) createClient(url string) (*common.Client, error) {
	return common.NewClient(url, &common.ClientConfig{
		Username:        ja.Username,
		Password:        ja.Password,
		ResponseTimeout: time.Duration(ja.ResponseTimeout),
		ClientConfig:    ja.ClientConfig,
	})
}
