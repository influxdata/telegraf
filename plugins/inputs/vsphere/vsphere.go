//go:generate ../../../tools/readme_config_includer/generator
package vsphere

import (
	"context"
	_ "embed"
	"errors"
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/soap"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// VSphere is the top level type for the vSphere input plugin. It contains all the configuration
// and a list of connected vSphere endpoints
type VSphere struct {
	Vcenters                    []string
	Username                    config.Secret `toml:"username"`
	Password                    config.Secret `toml:"password"`
	DatacenterInstances         bool
	DatacenterMetricInclude     []string
	DatacenterMetricExclude     []string
	DatacenterInclude           []string
	DatacenterExclude           []string
	ClusterInstances            bool
	ClusterMetricInclude        []string
	ClusterMetricExclude        []string
	ClusterInclude              []string
	ClusterExclude              []string
	ResourcePoolInstances       bool
	ResourcePoolMetricInclude   []string
	ResourcePoolMetricExclude   []string
	ResourcePoolInclude         []string
	ResourcePoolExclude         []string
	HostInstances               bool
	HostMetricInclude           []string
	HostMetricExclude           []string
	HostInclude                 []string
	HostExclude                 []string
	VMInstances                 bool     `toml:"vm_instances"`
	VMMetricInclude             []string `toml:"vm_metric_include"`
	VMMetricExclude             []string `toml:"vm_metric_exclude"`
	VMInclude                   []string `toml:"vm_include"`
	VMExclude                   []string `toml:"vm_exclude"`
	DatastoreInstances          bool
	DatastoreMetricInclude      []string
	DatastoreMetricExclude      []string
	DatastoreInclude            []string
	DatastoreExclude            []string
	VSANMetricInclude           []string `toml:"vsan_metric_include"`
	VSANMetricExclude           []string `toml:"vsan_metric_exclude"`
	VSANMetricSkipVerify        bool     `toml:"vsan_metric_skip_verify"`
	VSANClusterInclude          []string `toml:"vsan_cluster_include"`
	Separator                   string
	CustomAttributeInclude      []string
	CustomAttributeExclude      []string
	UseIntSamples               bool
	IPAddresses                 []string
	MetricLookback              int
	DisconnectedServersBehavior string
	MaxQueryObjects             int
	MaxQueryMetrics             int
	CollectConcurrency          int
	DiscoverConcurrency         int
	ForceDiscoverOnInit         bool `toml:"force_discover_on_init" deprecated:"1.14.0;option is ignored"`
	ObjectDiscoveryInterval     config.Duration
	Timeout                     config.Duration
	HistoricalInterval          config.Duration

	endpoints []*Endpoint
	cancel    context.CancelFunc

	// Mix in the TLS/SSL goodness from core
	tls.ClientConfig

	Log telegraf.Logger
}

func (*VSphere) SampleConfig() string {
	return sampleConfig
}

// Start is called from telegraf core when a plugin is started and allows it to
// perform initialization tasks.
func (v *VSphere) Start(_ telegraf.Accumulator) error {
	v.Log.Info("Starting plugin")
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	// Create endpoints, one for each vCenter we're monitoring
	v.endpoints = make([]*Endpoint, 0, len(v.Vcenters))
	for _, rawURL := range v.Vcenters {
		u, err := soap.ParseURL(rawURL)
		if err != nil {
			return err
		}
		ep, err := NewEndpoint(ctx, v, u, v.Log)
		if err != nil {
			return err
		}
		v.endpoints = append(v.endpoints, ep)
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
			if errors.Is(err, context.Canceled) {
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
			Vcenters:                    []string{},
			DatacenterInstances:         false,
			DatacenterMetricInclude:     nil,
			DatacenterMetricExclude:     nil,
			DatacenterInclude:           []string{"/*"},
			ClusterInstances:            false,
			ClusterMetricInclude:        nil,
			ClusterMetricExclude:        nil,
			ClusterInclude:              []string{"/*/host/**"},
			HostInstances:               true,
			HostMetricInclude:           nil,
			HostMetricExclude:           nil,
			HostInclude:                 []string{"/*/host/**"},
			ResourcePoolInstances:       false,
			ResourcePoolMetricInclude:   nil,
			ResourcePoolMetricExclude:   nil,
			ResourcePoolInclude:         []string{"/*/host/**"},
			VMInstances:                 true,
			VMMetricInclude:             nil,
			VMMetricExclude:             nil,
			VMInclude:                   []string{"/*/vm/**"},
			DatastoreInstances:          false,
			DatastoreMetricInclude:      nil,
			DatastoreMetricExclude:      nil,
			DatastoreInclude:            []string{"/*/datastore/**"},
			VSANMetricInclude:           nil,
			VSANMetricExclude:           []string{"*"},
			VSANMetricSkipVerify:        false,
			VSANClusterInclude:          []string{"/*/host/**"},
			Separator:                   "_",
			CustomAttributeInclude:      []string{},
			CustomAttributeExclude:      []string{"*"},
			UseIntSamples:               true,
			IPAddresses:                 []string{},
			MaxQueryObjects:             256,
			MaxQueryMetrics:             256,
			CollectConcurrency:          1,
			DiscoverConcurrency:         1,
			MetricLookback:              3,
			ForceDiscoverOnInit:         true,
			ObjectDiscoveryInterval:     config.Duration(time.Second * 300),
			Timeout:                     config.Duration(time.Second * 60),
			HistoricalInterval:          config.Duration(time.Second * 300),
			DisconnectedServersBehavior: "error",
		}
	})
}
