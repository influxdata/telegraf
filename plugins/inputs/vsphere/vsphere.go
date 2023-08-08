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
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// VSphere is the top level type for the vSphere input plugin. It contains all the configuration
// and a list of connected vSphere endpoints
type VSphere struct {
	Vcenters                    []string        `toml:"vcenters"`
	Username                    config.Secret   `toml:"username"`
	Password                    config.Secret   `toml:"password"`
	DatacenterInstances         bool            `toml:"datacenter_instances"`
	DatacenterMetricInclude     []string        `toml:"datacenter_metric_include"`
	DatacenterMetricExclude     []string        `toml:"datacenter_metric_exclude"`
	DatacenterInclude           []string        `toml:"datacenter_include"`
	DatacenterExclude           []string        `toml:"datacenter_exclude"`
	ClusterInstances            bool            `toml:"cluster_instances"`
	ClusterMetricInclude        []string        `toml:"cluster_metric_include"`
	ClusterMetricExclude        []string        `toml:"cluster_metric_exclude"`
	ClusterInclude              []string        `toml:"cluster_include"`
	ClusterExclude              []string        `toml:"cluster_exclude"`
	ResourcePoolInstances       bool            `toml:"resource_pool_instances"`
	ResourcePoolMetricInclude   []string        `toml:"resource_pool_metric_include"`
	ResourcePoolMetricExclude   []string        `toml:"resource_pool_metric_exclude"`
	ResourcePoolInclude         []string        `toml:"resource_pool_include"`
	ResourcePoolExclude         []string        `toml:"resource_pool_exclude"`
	HostInstances               bool            `toml:"host_instances"`
	HostMetricInclude           []string        `toml:"host_metric_include"`
	HostMetricExclude           []string        `toml:"host_metric_exclude"`
	HostInclude                 []string        `toml:"host_include"`
	HostExclude                 []string        `toml:"host_exclude"`
	VMInstances                 bool            `toml:"vm_instances"`
	VMMetricInclude             []string        `toml:"vm_metric_include"`
	VMMetricExclude             []string        `toml:"vm_metric_exclude"`
	VMInclude                   []string        `toml:"vm_include"`
	VMExclude                   []string        `toml:"vm_exclude"`
	DatastoreInstances          bool            `toml:"datastore_instances"`
	DatastoreMetricInclude      []string        `toml:"datastore_metric_include"`
	DatastoreMetricExclude      []string        `toml:"datastore_metric_exclude"`
	DatastoreInclude            []string        `toml:"datastore_include"`
	DatastoreExclude            []string        `toml:"datastore_exclude"`
	VSANMetricInclude           []string        `toml:"vsan_metric_include"`
	VSANMetricExclude           []string        `toml:"vsan_metric_exclude"`
	VSANMetricSkipVerify        bool            `toml:"vsan_metric_skip_verify"`
	VSANClusterInclude          []string        `toml:"vsan_cluster_include"`
	Separator                   string          `toml:"separator"`
	CustomAttributeInclude      []string        `toml:"custom_attribute_include"`
	CustomAttributeExclude      []string        `toml:"custom_attribute_exclude"`
	UseIntSamples               bool            `toml:"use_int_samples"`
	IPAddresses                 []string        `toml:"ip_addresses"`
	MetricLookback              int             `toml:"metric_lookback"`
	DisconnectedServersBehavior string          `toml:"disconnected_servers_behavior"`
	MaxQueryObjects             int             `toml:"max_query_objects"`
	MaxQueryMetrics             int             `toml:"max_query_metrics"`
	CollectConcurrency          int             `toml:"collect_concurrency"`
	DiscoverConcurrency         int             `toml:"discover_concurrency"`
	ForceDiscoverOnInit         bool            `toml:"force_discover_on_init" deprecated:"1.14.0;option is ignored"`
	ObjectDiscoveryInterval     config.Duration `toml:"object_discovery_interval"`
	Timeout                     config.Duration `toml:"timeout"`
	HistoricalInterval          config.Duration `toml:"historical_interval"`
	Log                         telegraf.Logger `toml:"-"`

	tls.ClientConfig // Mix in the TLS/SSL goodness from core
	proxy.HTTPProxy

	endpoints []*Endpoint
	cancel    context.CancelFunc
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
			DatacenterInclude:           []string{"/*"},
			ClusterInclude:              []string{"/*/host/**"},
			HostInstances:               true,
			HostInclude:                 []string{"/*/host/**"},
			ResourcePoolInclude:         []string{"/*/host/**"},
			VMInstances:                 true,
			VMInclude:                   []string{"/*/vm/**"},
			DatastoreInclude:            []string{"/*/datastore/**"},
			VSANMetricExclude:           []string{"*"},
			VSANClusterInclude:          []string{"/*/host/**"},
			Separator:                   "_",
			CustomAttributeExclude:      []string{"*"},
			UseIntSamples:               true,
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
			HTTPProxy:                   proxy.HTTPProxy{UseSystemProxy: true},
		}
	})
}
