# VMware vSphere Input Plugin

The VMware vSphere plugin uses the vSphere API to gather metrics from multiple vCenter servers.
 
* Clusters
* Hosts
* VMs
* Data stores

## Configuration

```
# Read metrics from one or many vCenters
[[inputs.vsphere]]
  ## List of vCenter URLs to be monitored. 
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
```

### Metrics Selection

vSphere metrics are available in 4 categories: Clusters, Hosts, VMs, and Datastores. Each category can be enabled
independently, and for each category a list with wildcards can be applied to filter metrics. If the &lt;category&gt;_metrics 
property is not specified all metrics will be included for the category. Metrics can be specified with standard glob
style wildcards (*, ?). You can specify metrics to be negated, by using an exclamation mark (!) as the first character. 
Metrics will be checked against every item in the list in order, meaning you can include multiple metrics with the first 
item, and negate selectively within that list on a subsequent item. Negating metrics that have not been whitelisted in 
a prior item of the list, has no effect.

A configuration to include all metrics under the host category that start with `cpu.*` but will not include
`cpu.idle.summation`.

```
  host_metrics = ["cpu.*", "!cpu.idle.summation"]

``` 

### Objects Per Query

Default settings for vCenter 6.5 and above is 256. Prior versions of vCenter have this set to 64. A vCenter administrator
can change this setting, which should be reflected in this plugin. See this [VMware KB article](https://kb.vmware.com/s/article/2107096)
for more information.

## Measurements &amp; Fields

- Cluster Stats
	- Cluster services: CPU, memory, failover
	- CPU: total, usage
	- Memory: consumed, total, vmmemctl
	- VM operations: # changes, clone, create, deploy, destroy, power, reboot, reconfigure, register, reset, shutdown, standby, vmotion
- Host Stats:
	- CPU: total, usage, cost, mhz
	- Datastore: iops, latency, read/write bytes, # reads/writes
	- Disk: commands, latency, kernel reads/writes, # reads/writes, queues
	- Memory: total, usage, active, latency, swap, shared, vmmemctl
	- Network: broadcast, bytes, dropped, errors, multicast, packets, usage
	- Power: energy, usage, capacity
	- Res CPU: active, max, running
	- Storage Adapter: commands, latency, # reads/writes
	- Storage Path: commands, latency, # reads/writes
	- System Resources: cpu active, cpu max, cpu running, cpu usage, mem allocated, mem consumed, mem shared, swap
	- System: uptime
	- Flash Module: active VMDKs 
- VM Stats:
	- CPU: demand, usage, readiness, cost, mhz
	- Datastore: latency, # reads/writes
	- Disk: commands, latency, # reads/writes, provisioned, usage
	- Memory: granted, usage, active, swap, vmmemctl
	- Network: broadcast, bytes, dropped, multicast, packets, usage
	- Power: energy, usage
	- Res CPU: active, max, running
	- System: operating system uptime, uptime
	- Virtual Disk: seeks, # reads/writes, latency, load 
- Datastore stats:
	- Disk: Capacity, provisioned, used  
	
## Tags
- all metrics
	- vcenter (vcenter url)
- all host metrics
	- cluster (vcenter cluster)
- all vm metrics
	- cluster (vcenter cluster)
	- esxhost (name of ESXi host)
	- guest (guest operating system id)
- cpu stats for Host and VM
	- cpu (cpu core - not all CPU fields will have this tag)
- datastore stats for Host and VM
	- datastore (id of datastore)
- disk stats for Host and VM
	- disk (name of disk)
- disk.used.capacity for Datastore
	- disk (type of disk)
- net stats for Host and VM
	- interface (name of network interface)
- storageAdapter stats for Host
	- adapter (name of storage adapter)
- storagePath stats for Host 
	- path (id of storage path)
- sys.resource* stats for Host
	- resource (resource type)
- vflashModule stats for Host
	- module (name of flash module)
- virtualDisk stats for VM
	- disk (name of virtual disk)
