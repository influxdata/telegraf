# VMware vSphere Input Plugin - vSAN extension 

vSAN resource is a special type of resource that can be collected by the plugin.
The configuration of vSAN resource is slightly different from hosts, vms and other resources.

## Prerequisites
* vSphere 5.5 and later environments are needed
* Clusters with vSAN enabled
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
    "performance.host-memory-slab",
    "performance.host-memory-heap",
    "performance.system-mem",
  ]  
  # by default vsan_metric_skip_verify = false
  vsan_metric_skip_verify = true
  # vsan_metric_exclude = ["*"]
  vsan_cluster_include = ["/*/host/**"]
  
  collect_concurrency = 5
  discover_concurrency = 5
  
  ## Optional SSL Config
  # ssl_ca = "/path/to/cafile"
  # ssl_cert = "/path/to/certfile"
  # ssl_key = "/path/to/keyfile"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
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

* Many vCenter environments use self-signed certificates. Be sure to update the bottom portion of the above configuration and provide proper values for all applicable SSL Config settings that apply in your vSphere environment. In some environments, setting insecure_skip_verify = true will be necessary when the SSL certificates are not available.

* To ensure consistent collection in larger vSphere environments you may need to increase concurrency for the plugin. Use the collect_concurrency setting to control concurrency. Set collect_concurrency to the number of virtual machines divided by 1500 and rounded up to the nearest integer. For example, for 1200 VMs use 1 and for 2300 VMs use 2.
 

## Measurements & Fields

NOTE: vSAN performance measurements and fields may vary on the vSAN versions.
- vSAN Summary
     - overall_health
     - total_capacity_bytes, free_capacity_bytes
     - total_bytes_to_sync, total_objects_to_sync, total_recovery_eta
- vSAN Performance 
     - cluster-domclient
     	- iops_read, throughput_read, latency_avg_read, iops_write, throughput_write, latency_avg_write, congestion, oio
     - cluster-domcompmgr	
        - iops_read, throughput_read, latency_avg_read, iops_write, throughput_write, latency_avg_write, iops_rec_write, throughput_rec_write, latency_avg_rec_write, congestion, oio, iops_resync_read, tput_resync_read, lat_avg_resyncread
     - host-domclient
        - iops_read, throughput_read, latency_avg_read, read_count, iops_write, throughput_write, latency_avg_write, write_count, congestion, oio, client_cache_hits, client_cache_hit_rate
     - host-domcompmgr
     	- iops_read, throughput_read, latency_avg_read, read_count, iops_write, throughput_write, latency_avg_write, write_count, iops_rec_write, throughput_rec_write, latency_avg_rec_write, rec_write_count congestion, oio, iops_resync_read, tput_resync_read, lat_avg_resync_read
     - cache-disk	
        - iops_dev_read, throughput_dev_read, latency_dev_read, io_count_dev_read, iops_dev_write, throughput_dev_write, latency_dev_write, io_count_dev_write, latency_dev_d_avg, latency_dev_g_avg
     - capacity-disk
        - iops_dev_read, throughput_dev_read, latency_dev_read, io_count_dev_read, iops_dev_write, throughput_dev_write, latency_dev_write, io_count_dev_write, latency_dev_d_avg, latency_dev_g_avg, iops_read, latency_read, io_count_read, iops_write, latency_write, io_count_write
     - disk-group
        - iops_sched, latency_sched, outstanding_bytes_sched, iops_sched_queue_rec, throughput_sched_queue_rec,latency_sched_queue_rec, iops_sched_queue_vm, throughput_sched_queue_vm,latency_sched_queue_vm, iops_sched_queue_meta, throughput_sched_queue_meta,latency_sched_queue_meta, iops_delay_pct_sched, latency_delay_sched, rc_hit_rate, wb_free_pct, war_evictions, quota_evictions, iops_rc_read, latency_rc_read, io_count_rc_read, iops_wb_read, latency_wb_read, io_count_wb_read, iops_rc_write, latency_rc_write, io_count_rc_write, iops_wb_write, latency_wb_write, io_count_wb_write, ssd_bytes_drained, zero_bytes_drained, mem_congestion, slab_congestion, ssd_congestion, iops_congestion, log_congestion, comp_congestion, iops_direct_sched, iops_read, throughput_read, latency_avg_read, read_count, iops_write, throughput_write, latency_avg_write, write_count, oio_write, oio_rec_write, oio_write_size, oio_rec_write_size, rc_size, wb_size, capacity, capacity_used, capacity_reserved, throughput_sched, iops_resync_read_policy, iops_resync_read_decom, iops_resync_read_rebalance, iops_resync_read_fix_comp, iops_resync_write_policy, iops_resync_write_decom, iops_resync_write_rebalance, iops_resync_write_fix_comp, tput_resync_read_policy, tput_resync_read_decom, tput_resync_read_rebalance, tput_resync_read_fix_comp, tput_resync_write_policy, tput_resync_write_decom, tput_resync_write_rebalance, tput_resync_write_fix_comp, lat_resync_read_policy, lat_resync_read_decom, lat_resync_read_rebalance, lat_resync_read_fix_comp, lat_resync_write_policy, lat_resync_write_decom, lat_resync_write_rebalance, lat_resync_write_fix_comp
     - virtual-machine	
        - iops_read, throughput_read, latency_read_avg, latency_read_stddev, read_count, iops_write, throughput_write, latency_write_avg, latency_write_stddev, write_count
     - vscsi
     	- iops_read, throughput_read, latency_read, read_count, iops_write, throughput_write, latency_write, write_count
     - virtual-disk
     	- iops_limit, niops, niops_delayed
     - vsan-host-net
     	- rx_throughput, rx_packets, rx_packets_loss_rate, tx_throughput, tx_packets, tx_packets_loss_rate
     - vsan-vnic-net:
     	- rx_throughput, rx_packets, rx_packets_loss_rate, tx_throughput, tx_packets, tx_packets_loss_rate 
     - vsan-pnic-net
     	- rx_throughput, rx_packets, rx_packets_loss_rate, tx_throughput, tx_packets, tx_packets_loss_rate
     - vsan-iscsi-host
     	- iops_read, iops_write, iops_total, bandwidth_read, bandwidth_write, bandwidth_total, latency_read, latency_write, latency_total, queue_depth
     - vsan-iscsi-target
     	- iops_read, iops_write, iops_total, bandwidth_read, bandwidth_write, bandwidth_total, latency_read, latency_write, latency_total, queue_depth
     - vsan-iscsi-lun
     	- iops_read, iops_write, iops_total, bandwidth_read, bandwidth_write, bandwidth_total, latency_read, latency_write, latency_total, queue_depth
     	
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
vsphere_vsan_performance_hostdomclient,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,hostname=DC0_C0_H0,moid=domain-c8,source=Example-VSAN,vcenter=localhost:8898 iops_read=7,write_congestion=0,unmap_congestion=0,read_count=2199,iops=8,latency_max_write=8964,latency_avg_unmap=0,latency_avg_write=1883,write_count=364,num_oio=12623,throughput=564127,client_cache_hits=0,latency_max_read=17821,latency_max_unmap=0,read_congestion=0,latency_avg=1154,congestion=0,throughput_read=554721,latency_avg_read=1033,throughput_write=9406,client_cache_hit_rate=0,iops_unmap=0,throughput_unmap=0,latency_stddev=1315,io_count=2563,oio=4,iops_write=1,unmap_count=0 1578955200000000000
vsphere_vsan_performance_clusterdomcompmgr,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXX,vcenter=localhost:8898 latency_avg_rec_write=0,latency_avg_write=9886,congestion=0,iops_resync_read=0,lat_avg_resync_read=0,iops_read=289,latency_avg_read=1184,throughput_write=50137368,iops_rec_write=0,throughput_rec_write=0,tput_resync_read=0,throughput_read=9043654,iops_write=1272,oio=97 1578954900000000000
vsphere_vsan_performance_clusterdomclient,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXX,vcenter=localhost:8898 latency_avg_write=1011,congestion=0,oio=26,iops_read=6,throughput_read=489093,latency_avg_read=1085,iops_write=43,throughput_write=435142 1578955200000000000
vsphere_vsan_summary,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 total_bytes_to_sync=0i,total_objects_to_sync=0i,total_recovery_eta=0i 1578955489000000000
vsphere_vsan_summary,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 overall_health=1i 1578955489000000000
vsphere_vsan_summary,clustername=Example-VSAN,dcname=Example-DC,host=host.example.com,moid=domain-c7,source=Example-VSAN,vcenter=localhost:8898 free_capacity_byte=11022535578757i,total_capacity_byte=14102625779712i 1578955488000000000
```
