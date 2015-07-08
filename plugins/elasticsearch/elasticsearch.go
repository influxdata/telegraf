package elasticsearch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/influxdb/telegraf/plugins"
)

const indicesStatsPath = "/_nodes/stats/indices"
const indicesStatsPathLocal = "/_nodes/_local/stats/indices"

type node struct {
	Host    string                            `json:"host"`
	Indices map[string]map[string]interface{} `json:"indices"`
}

const sampleConfig = `
# specify a list of one or more Elasticsearch servers
servers = ["http://localhost:9200"]
#
# set local to false when you want to read the indices stats from all nodes
# within the cluster
local = true
`

// Elasticsearch is a plugin to read stats from one or many Elasticsearch
// servers.
type Elasticsearch struct {
	Local   bool
	Servers []string
	client  *http.Client
}

// NewElasticsearch return a new instance of Elasticsearch
func NewElasticsearch() *Elasticsearch {
	return &Elasticsearch{client: http.DefaultClient}
}

// SampleConfig returns sample configuration for this plugin.
func (e *Elasticsearch) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (e *Elasticsearch) Description() string {
	return "Read indices stats from one or more Elasticsearch servers or clusters"
}

// Gather reads the stats from Elasticsearch and writes it to the
// Accumulator.
func (e *Elasticsearch) Gather(acc plugins.Accumulator) error {
	for _, serv := range e.Servers {
		var url string
		if e.Local {
			url = serv + indicesStatsPathLocal
		} else {
			url = serv + indicesStatsPath
		}
		if err := e.gatherUrl(url, acc); err != nil {
			return err
		}
	}
	return nil
}

func (e *Elasticsearch) gatherUrl(url string, acc plugins.Accumulator) error {
	r, err := e.client.Get(url)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("elasticsearch: API responded with status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}
	d := json.NewDecoder(r.Body)
	esRes := &struct {
		ClusterName string           `json:"cluster_name"`
		Nodes       map[string]*node `json:"nodes"`
	}{}
	if err = d.Decode(esRes); err != nil {
		return err
	}

	for _, n := range esRes.Nodes {
		tags := map[string]string{
			"node_host":    n.Host,
			"cluster_name": esRes.ClusterName,
		}

		for group, stats := range n.Indices {
			for statName, value := range stats {
				floatVal, ok := value.(float64)
				if !ok {
					// there are a couple of values that we can't cast to float,
					// this is fine :-)
					continue
				}
				acc.Add(fmt.Sprintf("indices_%s_%s", group, statName), int(floatVal), tags)
			}
		}
	}

	return nil
}

func init() {
	plugins.Add("elasticsearch", func() plugins.Plugin {
		return NewElasticsearch()
	})
}
