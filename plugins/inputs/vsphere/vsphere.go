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
	Vcenters               []string
	Username               string
	Password               string
	GatherClusters         bool
	ClusterMetricInclude   []string
	ClusterMetricExclude   []string
	GatherHosts            bool
	HostMetricInclude      []string
	HostMetricExclude      []string
	GatherVms              bool
	VmMetricInclude        []string
	VmMetricExclude        []string
	GatherDatastores       bool
	DatastoreMetricInclude []string
	DatastoreMetricExclude []string

	ObjectsPerQuery         int32
	ObjectDiscoveryInterval internal.Duration
	Timeout                 internal.Duration

	endpoints []*Endpoint

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

# gather_vms = true # (default=true)

# Typical VM metrics (if omitted, all metrics are collected)
# vm_metric_include = [
#	"cpu.ready.summation.delta.millisecond",
#		"mem.swapinRate.average.rate.kiloBytesPerSecond",
#		"virtualDisk.numberReadAveraged.average.rate.number",
#		"virtualDisk.numberWriteAveraged.average.rate.number",
#		"virtualDisk.totalReadLatency.average.absolute.millisecond",
#		"virtualDisk.totalWriteLatency.average.absolute.millisecond",
#		"virtualDisk.readOIO.latest.absolute.number",
#		"virtualDisk.writeOIO.latest.absolute.number",
#		"net.bytesRx.average.rate.kiloBytesPerSecond",
#		"net.bytesTx.average.rate.kiloBytesPerSecond",
#		"net.droppedRx.summation.delta.number",
#		"net.droppedTx.summation.delta.number",
#		"cpu.run.summation.delta.millisecond",
#		"cpu.used.summation.delta.millisecond",
#		"mem.swapoutRate.average.rate.kiloBytesPerSecond",
#		"virtualDisk.read.average.rate.kiloBytesPerSecond",
#		"virtualDisk.write.average.rate.kiloBytesPerSecond" ]

# vm_metric_exclude []

############### Hosts ###############

# gather_hosts = true # (default=true)

## Typical host metrics (if omitted, all metrics are collected)
# host_metric_include = [
#		"cpu.ready.summation.delta.millisecond",
#		"cpu.latency.average.rate.percent",
#		"cpu.coreUtilization.average.rate.percent",
#		"mem.usage.average.absolute.percent",
#		"mem.swapinRate.average.rate.kiloBytesPerSecond",
#		"mem.state.latest.absolute.number",
#		"mem.latency.average.absolute.percent",
#		"mem.vmmemctl.average.absolute.kiloBytes",
#		"disk.read.average.rate.kiloBytesPerSecond",
#		"disk.write.average.rate.kiloBytesPerSecond",
#		"disk.numberReadAveraged.average.rate.number",
#		"disk.numberWriteAveraged.average.rate.number",
#		"disk.deviceReadLatency.average.absolute.millisecond",
#		"disk.deviceWriteLatency.average.absolute.millisecond",
#		"disk.totalReadLatency.average.absolute.millisecond",
#		"disk.totalWriteLatency.average.absolute.millisecond",
#		"storageAdapter.read.average.rate.kiloBytesPerSecond",
#		"storageAdapter.write.average.rate.kiloBytesPerSecond",
#		"storageAdapter.numberReadAveraged.average.rate.number",
#		"storageAdapter.numberWriteAveraged.average.rate.number",
#		"net.errorsRx.summation.delta.number",
#		"net.errorsTx.summation.delta.number",
#		"net.bytesRx.average.rate.kiloBytesPerSecond",
#		"net.bytesTx.average.rate.kiloBytesPerSecond",
#		"cpu.used.summation.delta.millisecond",
#		"cpu.usage.average.rate.percent",
#		"cpu.utilization.average.rate.percent",
#		"cpu.wait.summation.delta.millisecond",
#		"cpu.idle.summation.delta.millisecond",
#		"cpu.readiness.average.rate.percent",
#		"cpu.costop.summation.delta.millisecond",
#		"cpu.swapwait.summation.delta.millisecond",
#		"mem.swapoutRate.average.rate.kiloBytesPerSecond",
#		"disk.kernelReadLatency.average.absolute.millisecond",
#		"disk.kernelWriteLatency.average.absolute.millisecond" ]

# host_metric_exclude = [] # Nothing excluded by default

############### Clusters ###############

# gather_clusters = true # (default=true)

## Typical cluster metrics (if omitted, all metrics are collected)
# cluster_metric_include = [
#	  "cpu.usage.*",
#	  "cpu.usagemhz.*",
#	  "mem.usage.*",
#	  "mem.active.*" ]

# cluster_metric_exclude [] # Nothing excluded by default

############### Datastores ###############

# gather_datastore = true # (default=true)

## Typical datastore metrics (if omitted, all metrics are collected)
# datastore_metric_include = [
#   "disk.used.*",
#   "disk.provsioned.*" ]

# storage_metric_exclude = [] # Nothing excluded by default

## number of objects to retreive per query. set to 64 for vCenter 5.5 and 6.0 (default: 256)
# objects_per_query = 256

## the interval before (re)discovering objects subject to metrics collection (default: 300s)
# object_discovery_interval = "300s"

## timeout applies to any of the connection request made to vcenter
# timeout = "20s"

## Optional SSL Config
# ssl_ca = /path/to/cafile
# ssl_cert = /path/to/certfile
# ssl_key = /path/to/keyfile
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

			GatherClusters:         true,
			ClusterMetricInclude:   nil,
			ClusterMetricExclude:   nil,
			GatherHosts:            true,
			HostMetricInclude:      nil,
			HostMetricExclude:      nil,
			GatherVms:              true,
			VmMetricInclude:        nil,
			VmMetricExclude:        nil,
			GatherDatastores:       true,
			DatastoreMetricInclude: nil,
			DatastoreMetricExclude: nil,

			ObjectsPerQuery:         256,
			ObjectDiscoveryInterval: internal.Duration{Duration: time.Second * 300},
			Timeout:                 internal.Duration{Duration: time.Second * 20},
		}
	})
}
