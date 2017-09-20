# DC/OS Input Plugin

This input plugin gathers metrics from a DC/OS cluster.
For more information, please check the [DC/OS Metrics](https://dcos.io/docs/1.9/metrics/) page.

### Configuration:

```toml
# Input plugin for gathering DC/OS agent metrics
[[inputs.dcos]]
  # Hostname or ip address of DC/OS master for access from within DC/OS cluster
  master_hostname=""
  # Public URL of DC/OS cluster, e.g. http://dcos.example.com. Use of access from outside of the DC/OS cluster. master_hostname has higher priority, if set
  #cluster_url=""
  # Authentication token, obtained by running: dcos config show core.dcos_acs_token. Leave empty for no authentication.
  # Warning: authentication token is valid only 5 days in DC/OS 1.10.
  #auth_token=""
  # List of  DC/OS agent hostnames from which the metrics should be gathered. Leave empty for all.
  agent_include = []
  # DC/OS agent node file system mount for which related metrics should be gathered. Leave empty for all.
  path_include = []
  # DC/OS agent node network interface names for which related metrics should be gathered. Leave empty for all.
  interface_include = []
  # HTTP Response timeout, value must be more than a second
  #client_timeout = 30s
  # Set of default allowed tags. See readme.md for more tag keys.
  taginclude = ["cluster_url","path","interface","hostname","container_id","mesos_id","framework_name"]
  # Port number of Mesos component on DC/OS master for access from within DC/OS cluster
  #master_port = 5050
  # Port number of DC/OS metrics component on DC/OS agents. Must be the same on all agents
  #metrics_port = 61001
```

### Measurements & Fields

Below are enumerated the metrics taken from version 1.9 of DC/OS. For a description of those metrics, please see the [DC/OS Metrics Reference](https://dcos.io/docs/1.9/metrics/reference//).
Each field name has an additional suffix following the final underscore to indicate the unit of metric value.

#### Node (Agent) metric fields
- dcos_system
  - uptime_count
- dcos_cpu
  - cores_count
  - total_percent
  - user_percent
  - system_percent
  - idle_percent
  - wait_percent
- dcos_load
  - 1min_count
  - 5min_count
  - 15min_count
- dcos_filesystem
  - capacity_total_bytes
  - capacity_used_bytes
  - capacity_free_bytes
  - inode_total_count
  - inode_used_count
  - inode_free_count
- dcos_memory
  - total_bytes
  - free_bytes
  - buffers_bytes
  - cached_bytes
- dcos_swap
  - total_bytes
  - free_bytes
  - used_bytes
- dcos_network
  - in_bytes
  - out_bytes
  - in_packets_count
  - out_packets_count
  - in_dropped_count
  - out_dropped_count
  - in_errors_count
  - out_errors_count
- dcos_process
  - count

#### Container metric fields
- dcos_cpus
  - user_time_seconds
  - system_time_seconds
  - limit_count
  - throttled_time_seconds
- dcos_mem
  - total_bytes
  - limit_bytes
- dcos_disk
  - limit_bytes
  - used_bytes
- dcos_net
  - rx_packets_count
  - rx_bytes
  - rx_errors_count
  - rx_dropped_count
  - tx_packets_count
  - tx_bytes
  - tx_errors_count
  - tx_dropped_count

#### App metric fields
- dcos_app
  - _app specific metrics_ 
  
The set of DC/OS app metrics depends on the metrics exposed by an application running inside the container.

### Tags
#### Node (Agent) metric tags
* cluster_id
* cluster_url
* hostname
* interface - _only for dcos_network metric_
* mesos_id
* path - _only for dcos_filesystem metric_
* scope

#### Container metric tags
* cluster_id
* cluster_url
* container_id
* executor_id
* executor_name
* framework_id
* framework_name
* framework_principal
* framework_role
* hostname
* mesos_id
* scope
* source
* _additional container specific labels_

#### App metric tags
* mesos_id
* cluster_id
* cluster_url
* container_id
* executor_id
* framework_id
* hostname

### Example Output:
```
* Plugin: inputs.dcos, Collection 1
> dcos_swap,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,cluster_url=http://m1.dcos,hostname=192.168.65.60 total_bytes=2147479552,free_bytes=2147282944,used_bytes=196608 1505917759000000000
> dcos_system,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,cluster_url=http://m1.dcos,hostname=192.168.65.60 uptime_count=137478 1505917759000000000
> dcos_cpu,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,cluster_url=http://m1.dcos,hostname=192.168.65.60 user_percent=1.32,total_percent=2.3,idle_percent=97.61,system_percent=0.98,cores_count=2,wait_percent=0 1505917759000000000
> dcos_filesystem,hostname=192.168.65.60,path=/,cluster_url=http://m1.dcos,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3 inode_used_count=127502,inode_total_count=26214400,inode_free_count=26086898,capacity_total_bytes=53660876800,capacity_free_bytes=50554769408,capacity_used_bytes=3106107392 1505917759000000000
> dcos_memory,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,hostname=192.168.65.60,cluster_url=http://m1.dcos total_bytes=1569218560,buffers_bytes=733184,free_bytes=82296832,cached_bytes=835018752 1505917759000000000
> dcos_process,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,cluster_url=http://m1.dcos,hostname=192.168.65.60 count=194 1505917759000000000
> dcos_load,cluster_url=http://m1.dcos,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,hostname=192.168.65.60 1min_count=0.1,5min_count=0.14,15min_count=0.1 1505917759000000000
> dcos_cpus,cluster_url=http://m1.dcos,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,framework_name=marathon,container_id=8fe8b8d1-9549-4f1a-b96c-9ac6a13b78d1,hostname=192.168.65.60 throttled_time_seconds=1.931479049,user_time_seconds=112.1,system_time_seconds=324.08,limit_count=2.1 1505917759000000000
> dcos_disk,framework_name=marathon,container_id=8fe8b8d1-9549-4f1a-b96c-9ac6a13b78d1,cluster_url=http://m1.dcos,hostname=192.168.65.60,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3 used_bytes=0,limit_bytes=0 1505917759000000000
> dcos_net,container_id=8fe8b8d1-9549-4f1a-b96c-9ac6a13b78d1,cluster_url=http://m1.dcos,framework_name=marathon,hostname=192.168.65.60,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3 rx_dropped_count=0,rx_bytes=0,tx_bytes=0,tx_packets_count=0,rx_errors_count=0,tx_dropped_count=0,rx_packets_count=0,tx_errors_count=0 1505917759000000000
> dcos_mem,container_id=8fe8b8d1-9549-4f1a-b96c-9ac6a13b78d1,cluster_url=http://m1.dcos,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S3,framework_name=marathon,hostname=192.168.65.60 limit_bytes=301989888,total_bytes=0 1505917759000000000
> dcos_system,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S2,hostname=192.168.65.111,cluster_url=http://m1.dcos uptime_count=137755 1505917761000000000
> dcos_swap,hostname=192.168.65.111,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S2,cluster_url=http://m1.dcos used_bytes=94208,free_bytes=2147385344,total_bytes=2147479552 1505917761000000000
> dcos_memory,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S2,cluster_url=http://m1.dcos,hostname=192.168.65.111 total_bytes=6088818688,buffers_bytes=339968,free_bytes=217067520,cached_bytes=3629289472 1505917761000000000
> dcos_filesystem,path=/,cluster_url=http://m1.dcos,mesos_id=16a563a0-1560-4e1f-b886-9f3e487b85a6-S2,hostname=192.168.65.111 capacity_used_bytes=25282076672,capacity_free_bytes=28378800128,inode_free_count=26033073,inode_used_count=181327,inode_total_count=26214400,capacity_total_bytes=53660876800 1505917761000000000
....
```
