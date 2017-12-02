package vsphere

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/vmware/govmomi/vim25/soap"
	"log"
	"sync"
	"time"
)

type VSphere struct {
	Vcenters []string

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

	endpoints               []*Endpoint
}

var sampleConfig = `
  ## List of vCenter URLs, including credentials. Note the "@" characted must be escaped as %40
  # vcenters = [ "https://administrator%40vsphere.local:VMware1!@vcenter.local/sdk" ]
`

func (v *VSphere) SampleConfig() string {
	return sampleConfig
}

func (v *VSphere) Description() string {
	return "Read metrics from VMware vCenter"
}

func (v *VSphere) checkEndpoints() {
	if v.endpoints != nil {
		return
	}

	v.endpoints = make([]*Endpoint, len(v.Vcenters))
	for i, rawUrl := range v.Vcenters {
		u, err := soap.ParseURL(rawUrl)
		if err != nil {
			log.Printf("E! Can't parse URL %s\n", rawUrl)
		}

		v.endpoints[i] = NewEndpoint(v, u)
	}
}

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

			ObjectsPerQuery:         500,
			ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
			Timeout:                 internal.Duration{Duration: time.Second * 20},
		}
	})
}
