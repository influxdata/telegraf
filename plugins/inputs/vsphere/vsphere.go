package vsphere

import (
	"context"
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
	Vcenters                []string
	Username                string
	Password                string
	DatacenterInstances     bool
	DatacenterMetricInclude []string
	DatacenterMetricExclude []string
	DatacenterInclude       []string
	ClusterInstances        bool
	ClusterMetricInclude    []string
	ClusterMetricExclude    []string
	ClusterInclude          []string
	HostInstances           bool
	HostMetricInclude       []string
	HostMetricExclude       []string
	HostInclude             []string
	VMInstances             bool     `toml:"vm_instances"`
	VMMetricInclude         []string `toml:"vm_metric_include"`
	VMMetricExclude         []string `toml:"vm_metric_exclude"`
	VMInclude               []string `toml:"vm_include"`
	DatastoreInstances      bool
	DatastoreMetricInclude  []string
	DatastoreMetricExclude  []string
	DatastoreInclude        []string
	Separator               string
	UseIntSamples           bool

	MaxQueryObjects         int
	MaxQueryMetrics         int
	CollectConcurrency      int
	DiscoverConcurrency     int
	ForceDiscoverOnInit     bool
	ObjectDiscoveryInterval internal.Duration
	Timeout                 internal.Duration

	endpoints []*Endpoint
	cancel    context.CancelFunc

	// Mix in the TLS/SSL goodness from core
	tls.ClientConfig
}

var sampleConfig = `
  ## List of vCenter URLs to be monitored. These three lines must be uncommented
  ## and edited for the plugin to work.
  vcenters = [ "https://vcenter.local/sdk" ]
  username = "user@corp.local"
  password = "secret"

  ## VMs
  ## Typical VM metrics (if omitted or empty, all metrics are collected)
  vm_metric_include = [
    "cpu.demand.average",
    "cpu.idle.summation",
    "cpu.latency.average",
    "cpu.readiness.average",
    "cpu.ready.summation",
    "cpu.run.summation",
    "cpu.usagemhz.average",
    "cpu.used.summation",
    "cpu.wait.summation",
    "mem.active.average",
    "mem.granted.average",
    "mem.latency.average",
    "mem.swapin.average",
    "mem.swapinRate.average",
    "mem.swapout.average",
    "mem.swapoutRate.average",
    "mem.usage.average",
    "mem.vmmemctl.average",
    "net.bytesRx.average",
    "net.bytesTx.average",
    "net.droppedRx.summation",
    "net.droppedTx.summation",
    "net.usage.average",
    "power.power.average",    
    "virtualDisk.numberReadAveraged.average",
    "virtualDisk.numberWriteAveraged.average",
    "virtualDisk.read.average",
    "virtualDisk.readOIO.latest",
    "virtualDisk.throughput.usage.average",
    "virtualDisk.totalReadLatency.average",
    "virtualDisk.totalWriteLatency.average",
    "virtualDisk.write.average",
    "virtualDisk.writeOIO.latest",
    "sys.uptime.latest",
  ]
  # vm_metric_exclude = [] ## Nothing is excluded by default
  # vm_instances = true ## true by default

  ## Hosts 
  ## Typical host metrics (if omitted or empty, all metrics are collected)
  host_metric_include = [
    "cpu.coreUtilization.average",
    "cpu.costop.summation",
    "cpu.demand.average",
    "cpu.idle.summation",
    "cpu.latency.average",
    "cpu.readiness.average",
    "cpu.ready.summation",
    "cpu.swapwait.summation",
    "cpu.usage.average",
    "cpu.usagemhz.average",
    "cpu.used.summation",
    "cpu.utilization.average",
    "cpu.wait.summation",
    "disk.deviceReadLatency.average",
    "disk.deviceWriteLatency.average",
    "disk.kernelReadLatency.average",
    "disk.kernelWriteLatency.average",
    "disk.numberReadAveraged.average",
    "disk.numberWriteAveraged.average",
    "disk.read.average",
    "disk.totalReadLatency.average",
    "disk.totalWriteLatency.average",
    "disk.write.average",
    "mem.active.average",
    "mem.latency.average",
    "mem.state.latest",
    "mem.swapin.average",
    "mem.swapinRate.average",
    "mem.swapout.average",
    "mem.swapoutRate.average",
    "mem.totalCapacity.average",
    "mem.usage.average",
    "mem.vmmemctl.average",
    "net.bytesRx.average",
    "net.bytesTx.average",
    "net.droppedRx.summation",
    "net.droppedTx.summation",
    "net.errorsRx.summation",
    "net.errorsTx.summation",
    "net.usage.average",
    "power.power.average",
    "storageAdapter.numberReadAveraged.average",
    "storageAdapter.numberWriteAveraged.average",
    "storageAdapter.read.average",
    "storageAdapter.write.average",
    "sys.uptime.latest",
  ]
  # host_metric_exclude = [] ## Nothing excluded by default
  # host_instances = true ## true by default

  ## Clusters 
  # cluster_metric_include = [] ## if omitted or empty, all metrics are collected
  # cluster_metric_exclude = [] ## Nothing excluded by default
  # cluster_instances = false ## false by default

  ## Datastores 
  # datastore_metric_include = [] ## if omitted or empty, all metrics are collected
  # datastore_metric_exclude = [] ## Nothing excluded by default
  # datastore_instances = false ## false by default for Datastores only

  ## Datacenters
  datacenter_metric_include = [] ## if omitted or empty, all metrics are collected
  datacenter_metric_exclude = [ "*" ] ## Datacenters are not collected by default.
  # datacenter_instances = false ## false by default for Datastores only

  ## Plugin Settings  
  ## separator character to use for measurement and field names (default: "_")
  # separator = "_"

  ## number of objects to retreive per query for realtime resources (vms and hosts)
  ## set to 64 for vCenter 5.5 and 6.0 (default: 256)
  # max_query_objects = 256

  ## number of metrics to retreive per query for non-realtime resources (clusters and datastores)
  ## set to 64 for vCenter 5.5 and 6.0 (default: 256)
  # max_query_metrics = 256

  ## number of go routines to use for collection and discovery of objects and metrics
  # collect_concurrency = 1
  # discover_concurrency = 1

  ## whether or not to force discovery of new objects on initial gather call before collecting metrics
  ## when true for large environments this may cause errors for time elapsed while collecting metrics
  ## when false (default) the first collection cycle may result in no or limited metrics while objects are discovered
  # force_discover_on_init = false

  ## the interval before (re)discovering objects subject to metrics collection (default: 300s)
  # object_discovery_interval = "300s"

  ## timeout applies to any of the api request made to vcenter
  # timeout = "60s"

  ## When set to true, all samples are sent as integers. This makes the output data types backwards compatible
  ## with Telegraf 1.9 or lower. Normally all samples from vCenter, with the exception of percentages, are 
  ## integer values, but under some conditions, some averaging takes place internally in the plugin. Setting this
  ## flag to "false" will send values as floats to preserve the full precision when averaging takes place.
  # use_int_samples = true

  ## Optional SSL Config
  # ssl_ca = "/path/to/cafile"
  # ssl_cert = "/path/to/certfile"
  # ssl_key = "/path/to/keyfile"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
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

// Start is called from telegraf core when a plugin is started and allows it to
// perform initialization tasks.
func (v *VSphere) Start(acc telegraf.Accumulator) error {
	log.Println("D! [inputs.vsphere]: Starting plugin")
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	// Create endpoints, one for each vCenter we're monitoring
	v.endpoints = make([]*Endpoint, len(v.Vcenters))
	for i, rawURL := range v.Vcenters {
		u, err := soap.ParseURL(rawURL)
		if err != nil {
			return err
		}
		ep, err := NewEndpoint(ctx, v, u)
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
	log.Println("D! [inputs.vsphere]: Stopping plugin")
	v.cancel()

	// Wait for all endpoints to finish. No need to wait for
	// Gather() to finish here, since it Stop() will only be called
	// after the last Gather() has finished. We do, however, need to
	// wait for any discovery to complete by trying to grab the
	// "busy" mutex.
	for _, ep := range v.endpoints {
		log.Printf("D! [inputs.vsphere]: Waiting for endpoint %s to finish", ep.URL.Host)
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
			UseIntSamples:           true,

			MaxQueryObjects:         256,
			MaxQueryMetrics:         256,
			CollectConcurrency:      1,
			DiscoverConcurrency:     1,
			ForceDiscoverOnInit:     false,
			ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
			Timeout:                 internal.Duration{Duration: time.Second * 60},
		}
	})
}
