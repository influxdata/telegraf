# VMware vSphere Input Plugin

The VMware vSphere plugin uses the vSphere API to gather metrics from multiple vCenter servers.
 
* Clusters
* Hosts
* VMs
* Data stores

## Configuration

NOTE: To disable collection av a specific resource type, simply exclude all metrics using the XX_metric_exclude. For example, to disable collection of VMs, add this:
```vm_metric_exclude = [ "*" ]```

```
# Read metrics from one or many vCenters
[[inputs.vsphere]]
  # List of vCenter URLs to be monitored. These three lines must be uncommented
## and edited for the plugin to work.
# vcenters = [ "https://vcenter.local/sdk" ]
# username = "user@corp.local"
# password = "secret"

############### VMs ###############

## Typical VM metrics (if omitted, all metrics are collected)
# vm_metric_include = [
#		"cpu.ready.summation",
#		"mem.swapinRate.average",
#		"virtualDisk.numberReadAveraged.average",
#		"virtualDisk.numberWriteAveraged.average",
#		"virtualDisk.totalReadLatency.average",
#		"virtualDisk.totalWriteLatency.average",
#		"virtualDisk.readOIO.latest",
#		"virtualDisk.writeOIO.latest",
#		"net.bytesRx.average",
#		"net.bytesTx.average",
#		"net.droppedRx.summation.",
#		"net.droppedTx.summation.",
#		"cpu.run.summation",
#		"cpu.used.summation",
#		"mem.swapoutRate.average",
#		"virtualDisk.read.average",
#		"virtualDisk.write.average" ]

# vm_metric_exclude = [] ## Nothing is excluded by default

############### Hosts ###############

## Typical host metrics (if omitted, all metrics are collected)
# host_metric_include = [
#		"cpu.ready.summation",
#		"cpu.latency.average",
#		"cpu.coreUtilization.average",
#		"mem.usage.average",
#		"mem.swapinRate.average",
#		"mem.state.latest",
#		"mem.latency.average",
#		"mem.vmmemctl.average",
#		"disk.read.average",
#		"disk.write.average",
#		"disk.numberReadAveraged.average",
#		"disk.numberWriteAveraged.average",
#		"disk.deviceReadLatency.average",
#		"disk.deviceWriteLatency.average",
#		"disk.totalReadLatency.average",
#		"disk.totalWriteLatency.average",
#		"storageAdapter.read.average",
#		"storageAdapter.write.average",
#		"storageAdapter.numberReadAveraged.average",
#		"storageAdapter.numberWriteAveraged.average",
#		"net.errorsRx.summation",
#		"net.errorsTx.summation",
#		"net.bytesRx.average",
#		"net.bytesTx.average",
#		"cpu.used.summation",
#		"cpu.usage.average",
#		"cpu.utilization.average",
#		"cpu.wait.summation.",
#		"cpu.idle.summation",
#		"cpu.readiness.average",
#		"cpu.costop.summation",
#		"cpu.swapwait.summation",
#		"mem.swapoutRate.average",
#		"disk.kernelReadLatency.average",
#		"disk.kernelWriteLatency.average" ]

# host_metric_exclude = [] ## Nothing excluded by default

############### Clusters ###############

## Typical cluster metrics (if omitted, all metrics are collected)
#cluster_metric_include = [
#	  "cpu.usage.*",
#	  "cpu.usagemhz.*",
#	  "mem.usage.*",
#	  "mem.active.*" ]

# cluster_metric_exclude = [] ## Nothing excluded by default

############### Datastores ###############

## Typical datastore metrics (if omitted, all metrics are collected)
##
## WARNING: It is not recommended to include all metrics. Your collection will probably time out if you do!
##

datastore_metric_include = [
	"datastore.read.average",
	"datastore.write.average",
	"datastore.throughput.usage.average",	
	"datastore.throughput.contention.average" ]

## number of objects to retreive per query. set to 64 for vCenter 5.5 and 6.0 (default: 256)
# objects_per_query = 256

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
