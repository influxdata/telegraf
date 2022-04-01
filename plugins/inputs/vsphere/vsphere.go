package vsphere

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/vmware/govmomi/vim25/soap"
)

// VSphere is the top level type for the vSphere input plugin. It contains all the configuration
// and a list of connected vSphere endpoints
type VSphere struct {
	Vcenters                []string
	Username                string
	Password                string
	DatacenterInstances     bool
	DatacenterMetricInclude []string
	DatacenterMetricExclude []string
	DatacenterInclude       []string
	DatacenterExclude       []string
	ClusterInstances        bool
	ClusterMetricInclude    []string
	ClusterMetricExclude    []string
	ClusterInclude          []string
	ClusterExclude          []string
	HostInstances           bool
	HostMetricInclude       []string
	HostMetricExclude       []string
	HostInclude             []string
	HostExclude             []string
	VMInstances             bool     `toml:"vm_instances"`
	VMMetricInclude         []string `toml:"vm_metric_include"`
	VMMetricExclude         []string `toml:"vm_metric_exclude"`
	VMInclude               []string `toml:"vm_include"`
	VMExclude               []string `toml:"vm_exclude"`
	DatastoreInstances      bool
	DatastoreMetricInclude  []string
	DatastoreMetricExclude  []string
	DatastoreInclude        []string
	DatastoreExclude        []string
	Separator               string
	CustomAttributeInclude  []string
	CustomAttributeExclude  []string
	UseIntSamples           bool
	IPAddresses             []string
	MetricLookback          int

	MaxQueryObjects         int
	MaxQueryMetrics         int
	CollectConcurrency      int
	DiscoverConcurrency     int
	ForceDiscoverOnInit     bool `toml:"force_discover_on_init" deprecated:"1.14.0;option is ignored"`
	ObjectDiscoveryInterval config.Duration
	Timeout                 config.Duration
	HistoricalInterval      config.Duration

	endpoints []*Endpoint
	cancel    context.CancelFunc

	// Mix in the TLS/SSL goodness from core
	tls.ClientConfig

	Log telegraf.Logger
}

// Start is called from telegraf core when a plugin is started and allows it to
// perform initialization tasks.
func (v *VSphere) Start(_ telegraf.Accumulator) error {
	v.Log.Info("Starting plugin")
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	// Create endpoints, one for each vCenter we're monitoring
	v.endpoints = make([]*Endpoint, len(v.Vcenters))
	for i, rawURL := range v.Vcenters {
		u, err := soap.ParseURL(rawURL)
		if err != nil {
			return err
		}
		ep, err := NewEndpoint(ctx, v, u, v.Log)
		if err != nil {
			return err
		}
		v.endpoints[i] = ep
	}
	return nil
}

// Stop is called from telegraf core when a plugin is stopped and allows it to
// perform shutdown tasks.
func (v *VSphere) Stop() {
	v.Log.Info("Stopping plugin")
	v.cancel()

	// Wait for all endpoints to finish. No need to wait for
	// Gather() to finish here, since it Stop() will only be called
	// after the last Gather() has finished. We do, however, need to
	// wait for any discovery to complete by trying to grab the
	// "busy" mutex.
	for _, ep := range v.endpoints {
		v.Log.Debugf("Waiting for endpoint %q to finish", ep.URL.Host)
		func() {
			ep.busy.Lock() // Wait until discovery is finished
			defer ep.busy.Unlock()
			ep.Close()
		}()
	}
}

// Gather is the main data collection function called by the Telegraf core. It performs all
// the data collection and writes all metrics into the Accumulator passed as an argument.
func (v *VSphere) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, ep := range v.endpoints {
		wg.Add(1)
		go func(endpoint *Endpoint) {
			defer wg.Done()
			err := endpoint.Collect(context.Background(), acc)
			if err == context.Canceled {
				// No need to signal errors if we were merely canceled.
				err = nil
			}
			if err != nil {
				acc.AddError(err)
			}
		}(ep)
	}

	wg.Wait()
	return nil
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input {
		return &VSphere{
			Vcenters: []string{},

			DatacenterInstances:     false,
			DatacenterMetricInclude: nil,
			DatacenterMetricExclude: nil,
			DatacenterInclude:       []string{"/*"},
			ClusterInstances:        false,
			ClusterMetricInclude:    nil,
			ClusterMetricExclude:    nil,
			ClusterInclude:          []string{"/*/host/**"},
			HostInstances:           true,
			HostMetricInclude:       nil,
			HostMetricExclude:       nil,
			HostInclude:             []string{"/*/host/**"},
			VMInstances:             true,
			VMMetricInclude:         nil,
			VMMetricExclude:         nil,
			VMInclude:               []string{"/*/vm/**"},
			DatastoreInstances:      false,
			DatastoreMetricInclude:  nil,
			DatastoreMetricExclude:  nil,
			DatastoreInclude:        []string{"/*/datastore/**"},
			Separator:               "_",
			CustomAttributeInclude:  []string{},
			CustomAttributeExclude:  []string{"*"},
			UseIntSamples:           true,
			IPAddresses:             []string{},

			MaxQueryObjects:         256,
			MaxQueryMetrics:         256,
			CollectConcurrency:      1,
			DiscoverConcurrency:     1,
			MetricLookback:          3,
			ForceDiscoverOnInit:     true,
			ObjectDiscoveryInterval: config.Duration(time.Second * 300),
			Timeout:                 config.Duration(time.Second * 60),
			HistoricalInterval:      config.Duration(time.Second * 300),
		}
	})
}
