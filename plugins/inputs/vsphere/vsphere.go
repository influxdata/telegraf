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
	Interval               time.Duration
	Vcenters               []string
	Username               string
	Password               string
	GatherClusters         bool
	ClusterInstances       bool
	ClusterMetricInclude   []string
	ClusterMetricExclude   []string
	GatherHosts            bool
	HostInstances          bool
	HostMetricInclude      []string
	HostMetricExclude      []string
	GatherVms              bool
	VMInstances            bool     `toml:"vm_instances"`
	VMMetricInclude        []string `toml:"vm_metric_include"`
	VMMetricExclude        []string `toml:"vm_metric_exclude"`
	GatherDatastores       bool
	DatastoreInstances     bool
	DatastoreMetricInclude []string
	DatastoreMetricExclude []string
	Separator              string

	ObjectsPerQuery         int
	MetricsPerQuery         int
	CollectConcurrency      int
	DiscoverConcurrency     int
	ForceDiscoverOnInit     bool
	ObjectDiscoveryInterval internal.Duration
	Timeout                 internal.Duration

	endpoints []*Endpoint
	rootCtx   context.Context
	cancel    context.CancelFunc

	// Mix in the TLS/SSL goodness from core
	tls.ClientConfig
}

var sampleConfig = `
## List of vCenter URLs to be monitored. These three lines must be uncommented
## and edited for the plugin to work.
# vcenters = [ "https://vcenter.local/sdk" ]
# username = "user@corp.local"
# password = "secret"

############### VMs ###############

# gather_vms = true ### (default=true)

## Collect VM instance metrics, such as individual cores? (default=true)
#vm_instances = true

## Typical VM metrics (if omitted, all metrics are collected)
vm_metric_include = [
		"cpu.ready.summation",
		"mem.swapinRate.average",
		"virtualDisk.numberReadAveraged.average",
		"virtualDisk.numberWriteAveraged.average",
		"virtualDisk.totalReadLatency.average",
		"virtualDisk.totalWriteLatency.average",
		"virtualDisk.readOIO.latest",
		"virtualDisk.writeOIO.latest",
		"net.bytesRx.average",
		"net.bytesTx.average",
		"net.droppedRx.summation.",
		"net.droppedTx.summation.",
		"cpu.run.summation",
		"cpu.used.summation",
		"mem.swapoutRate.average",
		"virtualDisk.read.average",
		"virtualDisk.write.average" ]

# vm_metric_exclude = [] ## Nothing is excluded by default

############### Hosts ###############

# gather_hosts = true ## (default=true)

## Collect host instance metrics, such as individual cores? (default=true)
#host_instances = true

## Typical host metrics (if omitted, all metrics are collected)
host_metric_include = [
		"cpu.ready.summation",
		"cpu.latency.average",
		"cpu.coreUtilization.average",
		"mem.usage.average",
		"mem.swapinRate.average",
		"mem.state.latest",
		"mem.latency.average",
		"mem.vmmemctl.average",
		"disk.read.average",
		"disk.write.average",
		"disk.numberReadAveraged.average",
		"disk.numberWriteAveraged.average",
		"disk.deviceReadLatency.average",
		"disk.deviceWriteLatency.average",
		"disk.totalReadLatency.average",
		"disk.totalWriteLatency.average",
		"storageAdapter.read.average",
		"storageAdapter.write.average",
		"storageAdapter.numberReadAveraged.average",
		"storageAdapter.numberWriteAveraged.average",
		"net.errorsRx.summation",
		"net.errorsTx.summation",
		"net.bytesRx.average",
		"net.bytesTx.average",
		"cpu.used.summation",
		"cpu.usage.average",
		"cpu.utilization.average",
		"cpu.wait.summation.",
		"cpu.idle.summation",
		"cpu.readiness.average",
		"cpu.costop.summation",
		"cpu.swapwait.summation",
		"mem.swapoutRate.average",
		"disk.kernelReadLatency.average",
		"disk.kernelWriteLatency.average" ]

# host_metric_exclude = [] ## Nothing excluded by default

############### Clusters ###############

# gather_clusters = true ## (default=true)

## Collect cluster instance metrics, such as individual cores? (default=true)
#cluster_instances = true

## Typical cluster metrics (if omitted, all metrics are collected)
cluster_metric_include = [
	  "cpu.usage.*",
	  "cpu.usagemhz.*",
	  "mem.usage.*",
	  "mem.active.*" ]

# cluster_metric_exclude = [] ## Nothing excluded by default

############### Datastores ###############

# gather_datastores = true ## (default=true)

## Collect datastore instance metrics, such as individual LUNs and datafiles? (default=false)
#datastore_instances = false

## Typical datastore metrics (if omitted, all metrics are collected)
##
## WARNING: It is not recommended to include all metrics. Your collection will probably time out if you do!
##
datastore_metric_include = [
	"datastore.read.average",
	"datastore.write.average",
	"datastore.throughput.usage.average",	
	"datastore.throughput.contention.average" ]

## Number of objects to retreive per query. 
# objects_per_query = 256

## Number of metrics (data points) to retrieve in each query for non-realtime metrics (cluster and datastore).
## Set to 64 for vCenter 5.5 and 6.0 (default: 256)
# metrics_per_query = 256

## the interval before (re)discovering objects subject to metrics collection (default: 300s)
# object_discovery_interval = "300s"

## timeout applies to any of the connection request made to vcenter
# timeout = "20s"

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

func (v *VSphere) checkEndpoints() error {
	if v.endpoints != nil {
		return nil
	}

	v.endpoints = make([]*Endpoint, len(v.Vcenters))
	for i, rawURL := range v.Vcenters {
		u, err := soap.ParseURL(rawURL)
		if err != nil {
			return err
		}

		v.endpoints[i] = NewEndpoint(v, u)
	}
	return nil
}

func (v *VSphere) Start(acc telegraf.Accumulator) error {
	log.Println("D! [input.vsphere]: Starting plugin")
	rootCtx, cancel := context.WithCancel(context.Background())
	v.rootCtx = rootCtx
	v.cancel = cancel
	return nil
}

func (v *VSphere) Stop() {
	log.Println("D! [input.vsphere]: Stopping plugin")
	v.cancel()

	// Wait for all endpoints to finish
	for _, ep := range v.endpoints {
		log.Printf("D! [input.vsphere]: Waiting for endpoint %s to finish", ep.URL.Host)
		ep.wg.Wait()
	}
}

// Gather is the main data collection function called by the Telegraf core. It performs all
// the data collection and writes all metrics into the Accumulator passed as an argument.
func (v *VSphere) Gather(acc telegraf.Accumulator) error {
	err := v.checkEndpoints()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, ep := range v.endpoints {
		wg.Add(1)
		go func(endpoint *Endpoint) {
			defer wg.Done()
			err := endpoint.collect(v.rootCtx, acc)
			if err == context.Canceled {

				// No need to signal errors if we were merely canceled.
				err = nil
			}
			acc.AddError(err)
		}(ep)
	}

	wg.Wait()
	return nil
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input {
		return &VSphere{
			Vcenters: []string{},

			GatherClusters:         true,
			ClusterInstances:       true,
			ClusterMetricInclude:   nil,
			ClusterMetricExclude:   nil,
			GatherHosts:            true,
			HostInstances:          true,
			HostMetricInclude:      nil,
			HostMetricExclude:      nil,
			GatherVms:              true,
			VMInstances:            true,
			VMMetricInclude:        nil,
			VMMetricExclude:        nil,
			GatherDatastores:       true,
			DatastoreInstances:     false,
			DatastoreMetricInclude: nil,
			DatastoreMetricExclude: nil,
			Separator:              ".",

			ObjectsPerQuery:         256,
			MetricsPerQuery:         256,
			CollectConcurrency:      1,
			DiscoverConcurrency:     1,
			ForceDiscoverOnInit:     true,
			ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
			Timeout:                 internal.Duration{Duration: time.Second * 20},
		}
	})
}
