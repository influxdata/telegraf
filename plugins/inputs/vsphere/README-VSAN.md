# VMware vSphere Input Plugin - vSAN extension 

vSAN resource is a special type of resource that can be collected by the plugin.
The configuration of vSAN resource is slightly different from hosts, vms and other resources.

## Prerequisites
* vSphere 5.5 and later environments are needed

* [Turn on vSAN performance service](https://docs.vmware.com/en/VMware-vSphere/6.0/com.vmware.vsphere.virtualsan.doc/GUID-02F67DC3-3D5A-48A4-A445-D2BD6AF2862C.html): When you create a vSAN cluster, the performance service is disabled. You will need to enable vSAN performance service first to monitor the performance metrics. 


## Configuration
```
[[inputs.vsphere]]
  interval = "300s"
  vcenters = ["https://<vcenter-ip>/sdk", "https://<vcenter2-ip>/sdk"]
  username = "<user>"
  password = "<pwd>"

  # Exclude all other metrics
  vm_metric_exclude = ["*"]
  datastore_metric_exclude = ["*"]
  datacenter_metric_exclude = ["*"]
  host_metric_exclude = ["*"]
  cluster_metric_exclude = ["*"]
  
  # By default all supported entity will be included
  vsan_metric_include = [
   "summary.disk-usage",
   "summary.health",
   "summary.resync",
   "performance.cluster-domclient",
   "performance.cluster-domcompmgr",
   "performance.host-domclient",
   "performance.host-domcompmgr",
   "performance.cache-disk",
   "performance.disk-group",
   "performance.capacity-disk",
   "performance.disk-group",
   "performance.virtual-machine",
   "performance.vscsi",
   "performance.virtual-disk",
   "performance.vsan-host-net",
   "performance.vsan-vnic-net",
   "performance.vsan-pnic-net",
   "performance.vsan-iscsi-host",
   "performance.vsan-iscsi-target",
   "performance.vsan-iscsi-lun",

   "performance.lsom-world-cpu",
   "performance.nic-world-cpu",
   "performance.dom-world-cpu",
   "performance.cmmds-world-cpu",
   "performance.host-cpu",
   "performance.host-domowner",
  ]
  
  # by default vsan_metric_skip_verify = false
  vsan_metric_skip_verify = true
  # vsan_metric_exclude = ["*"]
  vsan_cluster_include = ["/*/host/**"]
  
  collect_concurrency = 5
  discover_concurrency = 5
```

* Use `vsan_metric_include = [...]` to define the vSAN metrics you want to collect. 
e.g. `vsan_metric_include = ["summary.*", "performance.host-domclient", "performance.cache-disk", "performance.disk-group", "performance.capacity-disk"]`. 
To include all supported vSAN metrics, use `vsan_metric_include = [ "*" ]`
To disable all the vSAN metrics, use `vsan_metric_exclude = [ "*" ]`

* `vsan_metric_skip_verify` defines whether to skip verifying vSAN metrics against the ones from [GetSupportedEntityTypes API](https://code.vmware.com/apis/48/vsan#/doc/vim.cluster.VsanPerformanceManager.html#getSupportedEntityTypes). 
This option is given because some performance entities are not returned by the API, but we want to offer the flexibility if user really need the stats. 
When set false, anything not in supported entity list will be filtered out. 
When set true, queried metrics will be identical to vsan_metric_include and the exclusive array will not be used in this case. By default the value is false.

* `vsan_cluster_include` defines a list of inventory paths that will be used to select a portion of vSAN clusters.
vSAN metrics are only collected on cluster level. Therefore, use the same way as inventory paths for [vsphere's clusters](README.md#inventory-paths)
 

## Measurements & Fields

NOTE: vSAN performance measurements and fields may vary on the vSAN versions.
- vSAN Summary
     - OverallHealth
     - TotalCapacityB, FreeCapacityB
     - TotalBytesToSync, TotalObjectsToSync, TotalRecoveryETA
- vSAN Performance 
     - cluster-domclient
     	- iopsRead, throughputRead, latencyAvgRead, iopsWrite, throughputWrite, latencyAvgWrite, congestion, oio
     - cluster-domcompmgr	
        - iopsRead, throughputRead, latencyAvgRead, iopsWrite, throughputWrite, latencyAvgWrite, iopsRecWrite, throughputRecWrite, latencyAvgRecWrite, congestion, oio, iopsResyncRead, tputResyncRead, latAvgResyncRead
     - host-domclient
        - iopsRead, throughputRead, latencyAvgRead, readCount, iopsWrite, throughputWrite, latencyAvgWrite, writeCount, congestion, oio, clientCacheHits, clientCacheHitRate
     - host-domcompmgr
     	- iopsRead, throughputRead, latencyAvgRead, readCount, iopsWrite, throughputWrite, latencyAvgWrite, writeCount, iopsRecWrite, throughputRecWrite, latencyAvgRecWrite, recWriteCount congestion, oio, iopsResyncRead, tputResyncRead, latAvgResyncRead
     - cache-disk	
        - iopsDevRead, throughputDevRead, latencyDevRead, ioCountDevRead, iopsDevWrite, throughputDevWrite, latencyDevWrite, ioCountDevWrite, latencyDevDAvg, latencyDevGAvg
     - capacity-disk
        - iopsDevRead, throughputDevRead, latencyDevRead, ioCountDevRead, iopsDevWrite, throughputDevWrite, latencyDevWrite, ioCountDevWrite, latencyDevDAvg, latencyDevGAvg, iopsRead, latencyRead, ioCountRead, iopsWrite, latencyWrite, ioCountWrite
     - disk-group
        - iopsSched, latencySched, outstandingBytesSched, iopsSchedQueueRec, throughputSchedQueueRec,latencySchedQueueRec, iopsSchedQueueVM, throughputSchedQueueVM,latencySchedQueueVM, iopsSchedQueueMeta, throughputSchedQueueMeta,latencySchedQueueMeta, iopsDelayPctSched, latencyDelaySched, rcHitRate, wbFreePct, warEvictions, quotaEvictions, iopsRcRead, latencyRcRead, ioCountRcRead, iopsWbRead, latencyWbRead, ioCountWbRead, iopsRcWrite, latencyRcWrite, ioCountRcWrite, iopsWbWrite, latencyWbWrite, ioCountWbWrite, ssdBytesDrained, zeroBytesDrained, memCongestion, slabCongestion, ssdCongestion, iopsCongestion, logCongestion, compCongestion, iopsDirectSched, iopsRead, throughputRead, latencyAvgRead, readCount, iopsWrite, throughputWrite, latencyAvgWrite, writeCount, oioWrite, oioRecWrite, oioWriteSize, oioRecWriteSize, rcSize, wbSize, capacity, capacityUsed, capacityReserved, throughputSched, iopsResyncReadPolicy, iopsResyncReadDecom, iopsResyncReadRebalance, iopsResyncReadFixComp, iopsResyncWritePolicy, iopsResyncWriteDecom, iopsResyncWriteRebalance, iopsResyncWriteFixComp, tputResyncReadPolicy, tputResyncReadDecom, tputResyncReadRebalance, tputResyncReadFixComp, tputResyncWritePolicy, tputResyncWriteDecom, tputResyncWriteRebalance, tputResyncWriteFixComp, latResyncReadPolicy, latResyncReadDecom, latResyncReadRebalance, latResyncReadFixComp, latResyncWritePolicy, latResyncWriteDecom, latResyncWriteRebalance, latResyncWriteFixComp
     - virtual-machine	
        - iopsRead, throughputRead, latencyReadAvg, latencyReadStddev, readCount, iopsWrite, throughputWrite, latencyWriteAvg, latencyWriteStddev, writeCount
     - vscsi
     	- iopsRead, throughputRead, latencyRead, readCount, iopsWrite, throughputWrite, latencyWrite, writeCount
     - virtual-disk
     	- iopsLimit, NIOPS, NIOPSDelayed
     - vsan-host-net
     	- rxThroughput, rxPackets, rxPacketsLossRate, txThroughput, txPackets, txPacketsLossRate
     - vsan-vnic-net:
     	- rxThroughput, rxPackets, rxPacketsLossRate, txThroughput, txPackets, txPacketsLossRate
     - vsan-pnic-net
     	- rxThroughput, rxPackets, rxPacketsLossRate, txThroughput, txPackets, txPacketsLossRate
     - vsan-iscsi-host
     	- iopsRead, iopsWrite, iopsTotal, bandwidthRead, bandwidthWrite, bandwidthTotal, latencyRead, latencyWrite, latencyTotal, queueDepth
     - vsan-iscsi-target
     	- iopsRead, iopsWrite, iopsTotal, bandwidthRead, bandwidthWrite, bandwidthTotal, latencyRead, latencyWrite, latencyTotal, queueDepth
     - vsan-iscsi-lun
     	- iopsRead, iopsWrite, iopsTotal, bandwidthRead, bandwidthWrite, bandwidthTotal, latencyRead, latencyWrite, latencyTotal, queueDepth 
     	
## Tags
- all vSAN metrics
	- vcenter
	- dcname
	- clustername
	- moid (the cluster's managed object id)
-  host-domclient, host-domcompmgr
    - hostname
-  disk-group, cache-disk, capacity-disk 
    - hostname
    - deviceName
    - ssdUuid (if SSD)
- vsan-host-net
    - hostname
- vsan-pnic-net
    - pnic
- vsan-vnic-net
    - vnic
    - stackName
    
## Realtime vs. historical metrics

vSAN metrics also keep two different kinds of metrics - realtime and historical metrics.

* Realtime metrics are metrics with prefix 'summary'. These metrics are available at real-time.
* Historical metrics are metrics with prefix 'performance'. They are metrics queried from vSAN performance API, which is available at a 5-minute rollup level. 

For performance consideration, it is better to specify two instances of the plugin, one for the realtime metrics with a short collection interval and one for the historical metrics with a longer interval. For example:
```
## Realtime instance
[[inputs.vsphere]]
  interval = "30s"
  vcenters = [ "https://someaddress/sdk" ]
  username = "someuser@vsphere.local"
  password = "secret"

  insecure_skip_verify = true
  force_discover_on_init = true

  # Exclude all other metrics
  vm_metric_exclude = ["*"]
  datastore_metric_exclude = ["*"]
  datacenter_metric_exclude = ["*"]
  host_metric_exclude = ["*"]
  cluster_metric_exclude = ["*"]
  
  vsan_metric_include = [ "summary.*" ]
  vsan_metric_skip_verify = false

  collect_concurrency = 5
  discover_concurrency = 5

# Historical instance
[[inputs.vsphere]]

  interval = "300s"
  vcenters = [ "https://someaddress/sdk" ]
  username = "someuser@vsphere.local"
  password = "secret"

  insecure_skip_verify = true
  force_discover_on_init = true

  # Exclude all other metrics
  vm_metric_exclude = ["*"]
  datastore_metric_exclude = ["*"]
  datacenter_metric_exclude = ["*"]
  host_metric_exclude = ["*"]
  cluster_metric_exclude = ["*"]
  
  vsan_metric_include = [ "performance.*" ]
  vsan_metric_skip_verify = false
  
  collect_concurrency = 5
  discover_concurrency = 5
```


## Sample output
```
vsphere_vsan_performance_host-domclient,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,hostname=DC0_C0_H0,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 latencyAvg=1307,congestion=0,oio=12,iopsWrite=19,latencyMaxUnmap=0,latencyStddev=1483,latencyAvgWrite=1709,writeCount=5722,clientCacheHitRate=0,writeCongestion=0,throughputUnmap=0,throughput=313682,ioCount=8298,latencyAvgRead=415,readCount=2576,clientCacheHits=0,numOio=107657,latencyMaxRead=10780,unmapCount=0,iopsRead=8,throughputRead=36561,throughputWrite=277121,readCongestion=0,unmapCongestion=0,iops=27,iopsUnmap=0,latencyAvgUnmap=0,latencyMaxWrite=13771 1557336300000000000
vsphere_vsan_performance_cluster-domcompmgr,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXX,vcenter=localhost:8898 latencyAvgRead=865,iopsRecWrite=0,oio=40,tputResyncRead=0,latAvgResyncRead=0,throughputRead=13116039,latencyAvgWrite=659,throughputRecWrite=0,congestion=0,latencyAvgRecWrite=0,iopsResyncRead=0,iopsRead=434,iopsWrite=444,throughputWrite=12907193 1557336300000000000
vsphere_vsan_performance_cluster-domclient,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXX,vcenter=localhost:8898 iopsWrite=137,throughputWrite=6415204,latencyAvgWrite=7689,congestion=0,oio=238,iopsRead=13,throughputRead=442438,latencyAvgRead=543 1557336300000000000
vsphere_vsan_summary,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 TotalBytesToSync=0i,TotalObjectsToSync=0i,TotalRecoveryETA=0i 1557336663000000000
vsphere_vsan_summary,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 OverallHealth=1i 1557336663000000000
vsphere_vsan_summary,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 TotalCapacityB=45129177395200i,FreeCapacityB=21377606853416i 1557336663000000000
```
