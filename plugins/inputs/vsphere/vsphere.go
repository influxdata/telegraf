package vsphere

import (
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/vmware/govmomi/vim25/soap"
)

// VSphere is the top level type for the vSphere input plugin. It contains all the configuration
// and a list of connected vSphere endpoints
type VSphere struct {
	Vcenters         []string
	Username         string
	Password         string
	GatherClusters   bool
	ClusterMetrics   []string
	GatherHosts      bool
	HostMetrics      []string
	GatherVms        bool
	VmMetrics        []string
	GatherDatastores bool
	DatastoreMetrics []string

	ObjectsPerQuery         int32
	ObjectDiscoveryInterval internal.Duration
	Timeout                 internal.Duration

	endpoints []*Endpoint

	// Mix in the TLS/SSL goodness from core
	tls.ClientConfig
}

var sampleConfig = `
  ## List of vCenter URLs, including credentials. Note the "@" characted must be escaped as %40
  # vcenters = [ "https://administrator%40vsphere.local:VMware1!@vcenter.local/sdk" ]
`

// SampleConfig returns a set of default configuration to be used as a boilerplate when setting up
// Telegraf.
func (v *VSphere) SampleConfig() string {
	return sampleConfig
}

// Description returns a short textual description of the plugin
func (v *VSphere) Description() string {
	return "Read metrics from VMware vCenter"
}

func (v *VSphere) checkEndpoints() {
	if v.endpoints != nil {
		return
	}

	v.endpoints = make([]*Endpoint, len(v.Vcenters))
	for i, rawURL := range v.Vcenters {
		u, err := soap.ParseURL(rawURL)
		if err != nil {
			log.Printf("E! Can't parse URL %s\n", rawURL)
		}

		v.endpoints[i] = NewEndpoint(v, u)
	}
}

// Gather is the main data collection function called by the Telegraf core. It performs all
// the data collection and writes all metrics into the Accumulator passed as an argument.
func (v *VSphere) Gather(acc telegraf.Accumulator) error {

	v.checkEndpoints()

	var wg sync.WaitGroup

	for _, ep := range v.endpoints {
		wg.Add(1)
		go func(endpoint *Endpoint) {
			defer wg.Done()
			acc.AddError(endpoint.collect(acc))
		}(ep)
	}

	wg.Wait()

	return nil
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input {
		return &VSphere{
			Vcenters: []string{},

			GatherClusters:   true,
			ClusterMetrics:   nil,
			GatherHosts:      true,
			HostMetrics:      nil,
			GatherVms:        true,
			VmMetrics:        nil,
			GatherDatastores: true,
			DatastoreMetrics: nil,

			ObjectsPerQuery:         256,
			ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
			Timeout:                 internal.Duration{Duration: time.Second * 20},
		}
	})
}
