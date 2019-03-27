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
  ## List of vCenter URLs to be monitored.
  vcenters = ["https://vcenter.local/sdk"]

  ## vCenter Username
  username = "user@corp.local"

  ## vCenter Password
  password = "secret"

  ## Inventory paths for the virtual machines to gather metrics on.  If empty,
  ## all VMs are included.
  ##   ex: vm_include = ["/*/vm/**"]
  # vm_include = []

  ## Virtual machine metrics to gather.
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
  vm_metric_exclude = []

  ## If true, gather virtual machine instance metrics.
  # vm_instances = true

  ## Inventory path of the hosts to gather metrics on.  If empty, all hosts
  ## are included.
  ##   ex: host_include = [ "/*/host/**"]
  # host_include = []

  ## Host metrics to gather.
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
  host_metric_exclude = []

  ## If true, gather host instance metrics.
  # host_instances = true

  ## Inventory paths for the clusters to gather metrics on.  If empty,
  ## all clusters are included.
  ##   ex: cluster_include = ["/*/cluster/**"]
  # cluster_include = []

  ## Cluster metrics to gather.
  cluster_metric_include = []
  cluster_metric_exclude = []

  ## If true, gather cluster instance metrics.
  # cluster_instances = false

  ## Inventory paths for the datastores to gather metrics on.  If empty,
  ## all datastores are included.
  ##   ex: datastore_include = ["/*/datastore/**"]
  # datastore_include = []

  ## Datastore metrics to gather.
  datastore_metric_include = []
  datastore_metric_exclude = []

  ## If true, gather datastore instance metrics.
  # datastore_instances = false

  ## Inventory paths for the datacenter to gather metrics on.  If empty,
  ## all datacenter are included.
  ##   ex: datacenter_include = ["/*/datacenter/**"]
  # datacenter_include = []

  ## Datacenter metrics to gather.
  datacenter_metric_include = []
  datacenter_metric_exclude = ["*"]

  ## If true, gather datacenter instance metrics.
  # datacenter_instances = false

  ## Separator character to use for measurement and field names
  # separator = "_"

  ## Number of objects to retreive per query for realtime resources (vms and
  ## hosts).  Set to 64 for vCenter 5.5 and 6.0.
  # max_query_objects = 256

  ## Number of metrics to retreive per query for non-realtime resources
  ## (clusters and datastores) Set to 64 for vCenter 5.5 and 6.0.
  # max_query_metrics = 256

  ## Number of goroutines to use for collection and discovery of objects and
  ## metrics.
  # collect_concurrency = 1
  # discover_concurrency = 1

  ## Force discovery of new objects on initial gather call before collecting
  ## metrics.  When true, in large environments this may cause errors for time
  ## elapsed while collecting metrics.  When false, the first collection cycle
  ## may result in no or limited metrics while objects are discovered.
  # force_discover_on_init = false

  ## The interval before (re)discovering objects subject to metrics collection.
  # object_discovery_interval = "300s"

  ## Timeout applies to any API request made to vCenter.
  # timeout = "60s"

  ## When set to true, all samples are sent as integers. This makes the output
  ## data types backwards compatible with Telegraf 1.9 or lower. Normally all
  ## samples from vCenter, with the exception of percentages, are integer
  ## values, but under some conditions, some averaging takes place internally in
  ## the plugin. Setting this flag to "false" will send values as floats to
  ## preserve the full precision when averaging takes place.
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
