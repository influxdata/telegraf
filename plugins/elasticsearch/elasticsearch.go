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
	Host       string            `json:"host"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
	Indices    interface{}       `json:"indices"`
	Os         interface{}       `json:"os"`
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

	for id, n := range esRes.Nodes {
		tags := map[string]string{
			"node_id":      id,
			"node_host":    n.Host,
			"node_name":    n.Name,
			"cluster_name": esRes.ClusterName,
		}

		for k, v := range n.Attributes {
			tags["node_attribute_"+k] = v
		}

		if err := e.parseInterface(acc, "indices", tags, n.Indices); err != nil {
			return err
		}
		if err := e.parseInterface(acc, "os", tags, n.Os); err != nil {
			return err
		}
	}

	return nil
}

func (e *Elasticsearch) parseInterface(acc plugins.Accumulator, prefix string, tags map[string]string, v interface{}) error {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			if err := e.parseInterface(acc, prefix+"_"+k, tags, v); err != nil {
				return err
			}
		}
	case float64:
		acc.Add(prefix, t, tags)
	case bool, string, []interface{}:
		// ignored types
		return nil
	default:
		return fmt.Errorf("elasticsearch: got unexpected type %T with value %v (%s)", t, t, prefix)
	}
	return nil
}

func init() {
	plugins.Add("elasticsearch", func() plugins.Plugin {
		return NewElasticsearch()
	})
}
