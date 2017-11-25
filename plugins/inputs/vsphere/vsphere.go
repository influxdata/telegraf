package vsphere

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/vmware/govmomi/vim25/soap"
	"log"
	"net/url"
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

	VmSamplingPeriod        internal.Duration
	HostSamplingPeriod      internal.Duration
	ClusterSamplingPeriod   internal.Duration
	DatastoreSamplingPeriod internal.Duration

	endpoints []Endpoint
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

func (v *VSphere) vSphereInit() {
	if v.endpoints != nil {
		return
	}

	var wg sync.WaitGroup

	v.endpoints = make([]Endpoint, len(v.Vcenters))
	for i, rawUrl := range v.Vcenters {
		u, err := soap.ParseURL(rawUrl)
		if err != nil {
			log.Printf("E! Can't parse URL %s\n", rawUrl)
		}

		wg.Add(1)
		go func(url *url.URL, j int) {
			defer wg.Done()
			v.endpoints[j] = NewEndpoint(v, url)
		}(u, i)
	}

	wg.Wait()
}

func (v *VSphere) Gather(acc telegraf.Accumulator) error {

	v.vSphereInit()

	start := time.Now()

	var wg sync.WaitGroup

	for _, ep := range v.endpoints {
		wg.Add(1)
		go func(endpoint Endpoint) {
			defer wg.Done()
			acc.AddError(endpoint.collect(acc))
		}(ep)
	}

	wg.Wait()

	// Add gauge to show how long it took to gather all the metrics on this cycle
	//
	acc.AddGauge("vsphere", map[string]interface{}{"gather.duration": time.Now().Sub(start).Seconds()}, nil, time.Now())

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

			ClusterSamplingPeriod:   internal.Duration{Duration: time.Second * 300},
			HostSamplingPeriod:      internal.Duration{Duration: time.Second * 20},
			VmSamplingPeriod:        internal.Duration{Duration: time.Second * 20},
			DatastoreSamplingPeriod: internal.Duration{Duration: time.Second * 300},
		}
	})
}
