# VMware vSphere Input Plugin

The VMware vSphere plugin uses the vSphere API to gather metrics from multiple vCenter servers.
 
* Clusters
* Hosts
* VMs
* Data stores

## Configuration

NOTE: To disable collection of a specific resource type, simply exclude all metrics using the XX_metric_exclude. 
For example, to disable collection of VMs, add this:
```vm_metric_exclude = [ "*" ]```

```
# Read metrics from one or many vCenters
[[inputs.vsphere]]
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
  # cluster_instances = true ## true by default

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
  # timeout = "20s"

  ## Optional SSL Config
  # ssl_ca = "/path/to/cafile"
  # ssl_cert = "/path/to/certfile"
  # ssl_key = "/path/to/keyfile"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
``` 

### Objects and Metrics Per Query

Default settings for vCenter 6.5 and above is 256. Prior versions of vCenter have this set to 64. A vCenter administrator
can change this setting, which should be reflected in this plugin. See this [VMware KB article](https://kb.vmware.com/s/article/2107096)
for more information.

### Collection and Discovery concurrency

On large vCenter setups it may be prudent to have multiple concurrent go routines collect performance metrics
in order to avoid potential errors for time elapsed during a collection cycle. This should never be greater than 8,
though the default of 1 (no concurrency) should be sufficient for most configurations. 

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

For a detailed list of commonly available metrics, please refer to [METRICS.MD](METRICS.MD)
	
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

## Sample output

```
vsphere_vm_cpu,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 run_summation=2608i,ready_summation=129i,usage_average=5.01,used_summation=2134i,demand_average=326i 1535660299000000000
vsphere_vm_net,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 bytesRx_average=321i,bytesTx_average=335i 1535660299000000000
vsphere_vm_virtualDisk,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 write_average=144i,read_average=4i 1535660299000000000
vsphere_vm_net,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 bytesRx_average=242i,bytesTx_average=308i 1535660299000000000
vsphere_vm_virtualDisk,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 write_average=232i,read_average=4i 1535660299000000000
vsphere_vm_cpu,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 usage_average=5.49,used_summation=1804i,demand_average=308i,run_summation=2001i,ready_summation=120i 1535660299000000000
vsphere_vm_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 usage_average=4.19,used_summation=2108i,demand_average=285i,run_summation=1793i,ready_summation=93i 1535660299000000000
vsphere_vm_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 bytesRx_average=272i,bytesTx_average=419i 1535660299000000000
vsphere_vm_virtualDisk,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 write_average=229i,read_average=4i 1535660299000000000
vsphere_vm_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 run_summation=2277i,ready_summation=118i,usage_average=4.67,used_summation=2546i,demand_average=289i 1535660299000000000
vsphere_vm_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 bytesRx_average=243i,bytesTx_average=296i 1535660299000000000
vsphere_vm_virtualDisk,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 write_average=158i,read_average=4i 1535660299000000000
vsphere_host_net,esxhostname=DC0_H0,host=host.example.com,interface=vmnic0,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 usage_average=1042i,bytesTx_average=753i,bytesRx_average=660i 1535660299000000000
vsphere_host_cpu,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 utilization_average=10.46,usage_average=22.4,readiness_average=0.4,costop_summation=2i,coreUtilization_average=19.61,wait_summation=5148518i,idle_summation=58581i,latency_average=0.6,ready_summation=13370i,used_summation=19219i 1535660299000000000
vsphere_host_cpu,cpu=0,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 coreUtilization_average=25.6,utilization_average=11.58,used_summation=24306i,usage_average=24.26,idle_summation=86688i 1535660299000000000
vsphere_host_cpu,cpu=1,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 coreUtilization_average=12.29,utilization_average=8.32,used_summation=31312i,usage_average=22.47,idle_summation=94934i 1535660299000000000
vsphere_host_disk,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 read_average=331i,write_average=2800i 1535660299000000000
vsphere_host_disk,disk=/var/folders/rf/txwdm4pj409f70wnkdlp7sz80000gq/T/govcsim-DC0-LocalDS_0-367088371@folder-5,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 write_average=2701i,read_average=258i 1535660299000000000
vsphere_host_mem,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 usage_average=93.27 1535660299000000000
vsphere_host_net,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 bytesTx_average=650i,usage_average=1414i,bytesRx_average=569i 1535660299000000000
vsphere_host_cpu,clustername=DC0_C0,cpu=1,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 utilization_average=12.6,used_summation=25775i,usage_average=24.44,idle_summation=68886i,coreUtilization_average=17.59 1535660299000000000
vsphere_host_disk,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 read_average=340i,write_average=2340i 1535660299000000000
vsphere_host_disk,clustername=DC0_C0,disk=/var/folders/rf/txwdm4pj409f70wnkdlp7sz80000gq/T/govcsim-DC0-LocalDS_0-367088371@folder-5,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 write_average=2277i,read_average=282i 1535660299000000000
vsphere_host_mem,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=104.78 1535660299000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 bytesTx_average=463i,usage_average=1131i,bytesRx_average=719i 1535660299000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,interface=vmnic0,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=1668i,bytesTx_average=838i,bytesRx_average=921i 1535660299000000000
vsphere_host_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 used_summation=28952i,utilization_average=11.36,idle_summation=93261i,latency_average=0.46,ready_summation=12837i,usage_average=21.56,readiness_average=0.39,costop_summation=2i,coreUtilization_average=27.19,wait_summation=3820829i 1535660299000000000
vsphere_host_cpu,clustername=DC0_C0,cpu=0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 coreUtilization_average=24.12,utilization_average=13.83,used_summation=22462i,usage_average=24.69,idle_summation=96993i 1535660299000000000
internal_vsphere,host=host.example.com,os=Mac,vcenter=localhost:8989 connect_ns=4727607i,discover_ns=65389011i,discovered_objects=8i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=datastore,vcenter=localhost:8989 gather_duration_ns=296223i,gather_count=0i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=vm,vcenter=192.168.1.151 gather_duration_ns=136050i,gather_count=0i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=host,vcenter=localhost:8989 gather_count=62i,gather_duration_ns=8788033i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=host,vcenter=192.168.1.151 gather_count=0i,gather_duration_ns=162002i 1535660309000000000
internal_gather,host=host.example.com,input=vsphere,os=Mac gather_time_ns=17483653i,metrics_gathered=28i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,vcenter=192.168.1.151 connect_ns=0i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=vm,vcenter=localhost:8989 gather_duration_ns=7291897i,gather_count=36i 1535660309000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=datastore,vcenter=192.168.1.151 gather_duration_ns=958474i,gather_count=0i 1535660309000000000
vsphere_vm_cpu,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 usage_average=8.82,used_summation=3192i,demand_average=283i,run_summation=2419i,ready_summation=115i 1535660319000000000
vsphere_vm_net,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 bytesRx_average=277i,bytesTx_average=343i 1535660319000000000
vsphere_vm_virtualDisk,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 read_average=1i,write_average=741i 1535660319000000000
vsphere_vm_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 bytesRx_average=386i,bytesTx_average=369i 1535660319000000000
vsphere_vm_virtualDisk,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 write_average=814i,read_average=1i 1535660319000000000
vsphere_vm_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 run_summation=1778i,ready_summation=111i,usage_average=7.54,used_summation=2339i,demand_average=297i 1535660319000000000
vsphere_vm_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 usage_average=6.98,used_summation=2125i,demand_average=211i,run_summation=2990i,ready_summation=141i 1535660319000000000
vsphere_vm_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 bytesRx_average=357i,bytesTx_average=268i 1535660319000000000
vsphere_vm_virtualDisk,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 write_average=528i,read_average=1i 1535660319000000000
vsphere_vm_cpu,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 used_summation=2374i,demand_average=195i,run_summation=3454i,ready_summation=110i,usage_average=7.34 1535660319000000000
vsphere_vm_net,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 bytesRx_average=308i,bytesTx_average=246i 1535660319000000000
vsphere_vm_virtualDisk,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 write_average=1178i,read_average=1i 1535660319000000000
vsphere_host_net,esxhostname=DC0_H0,host=host.example.com,interface=vmnic0,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 bytesRx_average=773i,usage_average=1521i,bytesTx_average=890i 1535660319000000000
vsphere_host_cpu,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 wait_summation=3421258i,idle_summation=67994i,latency_average=0.36,usage_average=29.86,readiness_average=0.37,used_summation=25244i,costop_summation=2i,coreUtilization_average=21.94,utilization_average=17.19,ready_summation=15897i 1535660319000000000
vsphere_host_cpu,cpu=0,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 utilization_average=11.32,used_summation=19333i,usage_average=14.29,idle_summation=92708i,coreUtilization_average=27.68 1535660319000000000
vsphere_host_cpu,cpu=1,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 used_summation=28596i,usage_average=25.32,idle_summation=79553i,coreUtilization_average=28.01,utilization_average=11.33 1535660319000000000
vsphere_host_disk,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 read_average=86i,write_average=1659i 1535660319000000000
vsphere_host_disk,disk=/var/folders/rf/txwdm4pj409f70wnkdlp7sz80000gq/T/govcsim-DC0-LocalDS_0-367088371@folder-5,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 write_average=1997i,read_average=58i 1535660319000000000
vsphere_host_mem,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 usage_average=68.45 1535660319000000000
vsphere_host_net,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 bytesTx_average=679i,usage_average=2286i,bytesRx_average=719i 1535660319000000000
vsphere_host_cpu,clustername=DC0_C0,cpu=1,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 utilization_average=10.52,used_summation=21693i,usage_average=23.09,idle_summation=84590i,coreUtilization_average=29.92 1535660319000000000
vsphere_host_disk,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 read_average=113i,write_average=1236i 1535660319000000000
vsphere_host_disk,clustername=DC0_C0,disk=/var/folders/rf/txwdm4pj409f70wnkdlp7sz80000gq/T/govcsim-DC0-LocalDS_0-367088371@folder-5,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 write_average=1708i,read_average=110i 1535660319000000000
vsphere_host_mem,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=111.46 1535660319000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 bytesTx_average=998i,usage_average=2000i,bytesRx_average=881i 1535660319000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,interface=vmnic0,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=1683i,bytesTx_average=675i,bytesRx_average=1078i 1535660319000000000
vsphere_host_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 used_summation=28531i,wait_summation=3139129i,utilization_average=9.99,idle_summation=98579i,latency_average=0.51,costop_summation=2i,coreUtilization_average=14.35,ready_summation=16121i,usage_average=34.19,readiness_average=0.4 1535660319000000000
vsphere_host_cpu,clustername=DC0_C0,cpu=0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 utilization_average=12.2,used_summation=22750i,usage_average=18.84,idle_summation=99539i,coreUtilization_average=23.05 1535660319000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=host,vcenter=localhost:8989 gather_duration_ns=7076543i,gather_count=62i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=host,vcenter=192.168.1.151 gather_duration_ns=4051303i,gather_count=0i 1535660339000000000
internal_gather,host=host.example.com,input=vsphere,os=Mac metrics_gathered=56i,gather_time_ns=13555029i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,vcenter=192.168.1.151 connect_ns=0i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=vm,vcenter=localhost:8989 gather_duration_ns=6335467i,gather_count=36i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=datastore,vcenter=192.168.1.151 gather_duration_ns=958474i,gather_count=0i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,vcenter=localhost:8989 discover_ns=65389011i,discovered_objects=8i,connect_ns=4727607i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=datastore,vcenter=localhost:8989 gather_duration_ns=296223i,gather_count=0i 1535660339000000000
internal_vsphere,host=host.example.com,os=Mac,resourcetype=vm,vcenter=192.168.1.151 gather_count=0i,gather_duration_ns=1540920i 1535660339000000000
vsphere_vm_virtualDisk,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 write_average=302i,read_average=11i 1535660339000000000
vsphere_vm_cpu,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 usage_average=5.58,used_summation=2941i,demand_average=298i,run_summation=3255i,ready_summation=96i 1535660339000000000
vsphere_vm_net,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-35,os=Mac,source=DC0_H0_VM0,vcenter=localhost:8989,vmname=DC0_H0_VM0 bytesRx_average=155i,bytesTx_average=241i 1535660339000000000
vsphere_vm_cpu,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 usage_average=10.3,used_summation=3053i,demand_average=346i,run_summation=3289i,ready_summation=122i 1535660339000000000
vsphere_vm_net,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 bytesRx_average=215i,bytesTx_average=275i 1535660339000000000
vsphere_vm_virtualDisk,esxhostname=DC0_H0,guest=other,host=host.example.com,moid=vm-38,os=Mac,source=DC0_H0_VM1,vcenter=localhost:8989,vmname=DC0_H0_VM1 write_average=252i,read_average=14i 1535660339000000000
vsphere_vm_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 usage_average=8,used_summation=2183i,demand_average=354i,run_summation=3542i,ready_summation=128i 1535660339000000000
vsphere_vm_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 bytesRx_average=178i,bytesTx_average=200i 1535660339000000000
vsphere_vm_virtualDisk,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-41,os=Mac,source=DC0_C0_RP0_VM0,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM0 write_average=283i,read_average=12i 1535660339000000000
vsphere_vm_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 demand_average=328i,run_summation=3481i,ready_summation=122i,usage_average=7.95,used_summation=2167i 1535660339000000000
vsphere_vm_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 bytesTx_average=282i,bytesRx_average=196i 1535660339000000000
vsphere_vm_virtualDisk,clustername=DC0_C0,esxhostname=DC0_C0_H0,guest=other,host=host.example.com,moid=vm-44,os=Mac,source=DC0_C0_RP0_VM1,vcenter=localhost:8989,vmname=DC0_C0_RP0_VM1 write_average=321i,read_average=13i 1535660339000000000
vsphere_host_disk,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 read_average=39i,write_average=2635i 1535660339000000000
vsphere_host_disk,disk=/var/folders/rf/txwdm4pj409f70wnkdlp7sz80000gq/T/govcsim-DC0-LocalDS_0-367088371@folder-5,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 write_average=2635i,read_average=30i 1535660339000000000
vsphere_host_mem,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 usage_average=98.5 1535660339000000000
vsphere_host_net,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 usage_average=1887i,bytesRx_average=662i,bytesTx_average=251i 1535660339000000000
vsphere_host_net,esxhostname=DC0_H0,host=host.example.com,interface=vmnic0,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 usage_average=1481i,bytesTx_average=899i,bytesRx_average=992i 1535660339000000000
vsphere_host_cpu,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 used_summation=50405i,costop_summation=2i,utilization_average=17.32,latency_average=0.61,ready_summation=14843i,usage_average=27.94,coreUtilization_average=32.12,wait_summation=3058787i,idle_summation=56600i,readiness_average=0.36 1535660339000000000
vsphere_host_cpu,cpu=0,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 coreUtilization_average=37.61,utilization_average=17.05,used_summation=38013i,usage_average=32.66,idle_summation=89575i 1535660339000000000
vsphere_host_cpu,cpu=1,esxhostname=DC0_H0,host=host.example.com,moid=host-19,os=Mac,source=DC0_H0,vcenter=localhost:8989 coreUtilization_average=25.92,utilization_average=18.72,used_summation=39790i,usage_average=40.42,idle_summation=69457i 1535660339000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,interface=vmnic0,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=1246i,bytesTx_average=673i,bytesRx_average=781i 1535660339000000000
vsphere_host_cpu,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 coreUtilization_average=33.8,idle_summation=77121i,ready_summation=15857i,readiness_average=0.39,used_summation=29554i,costop_summation=2i,wait_summation=4338417i,utilization_average=17.87,latency_average=0.44,usage_average=28.78 1535660339000000000
vsphere_host_cpu,clustername=DC0_C0,cpu=0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 idle_summation=86610i,coreUtilization_average=34.36,utilization_average=19.03,used_summation=28766i,usage_average=23.72 1535660339000000000
vsphere_host_cpu,clustername=DC0_C0,cpu=1,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 coreUtilization_average=33.15,utilization_average=16.8,used_summation=44282i,usage_average=30.08,idle_summation=93490i 1535660339000000000
vsphere_host_disk,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 read_average=56i,write_average=1672i 1535660339000000000
vsphere_host_disk,clustername=DC0_C0,disk=/var/folders/rf/txwdm4pj409f70wnkdlp7sz80000gq/T/govcsim-DC0-LocalDS_0-367088371@folder-5,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 write_average=2110i,read_average=48i 1535660339000000000
vsphere_host_mem,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=116.21 1535660339000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 bytesRx_average=726i,bytesTx_average=643i,usage_average=1504i 1535660339000000000
vsphere_host_mem,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 usage_average=116.21 1535660339000000000
vsphere_host_net,clustername=DC0_C0,esxhostname=DC0_C0_H0,host=host.example.com,moid=host-30,os=Mac,source=DC0_C0_H0,vcenter=localhost:8989 bytesRx_average=726i,bytesTx_average=643i,usage_average=1504i 1535660339000000000
```
