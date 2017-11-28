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
  #response_timeout = "30s"
  # Port number of Mesos component on DC/OS master for access from within DC/OS cluster
  #master_port = 5050
  # Port number of DC/OS metrics component on DC/OS agents. Must be the same on all agents
  #metrics_port = 61001
  # TLS/SSL configuration for cluster_url
  #ssl_ca = "/etc/telegraf/ca.pem"
  #ssl_cert = "/etc/telegraf/cert.cer"
  #ssl_key = "/etc/telegraf/key.key"
  #insecure_skip_verify = false
```

### Measurements & Fields

Below are enumerated the metrics taken from version 1.10 of DC/OS. For a description of those metrics, please see the [DC/OS Metrics Reference](https://dcos.io/docs/1.10/metrics/reference/).
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
* framework_id
* hostname
* mesos_id
* package_name
* scope

#### App metric tags
* cluster_id
* cluster_url
* container_id
* executor_id
* framework_id
* hostname
* mesos_id

### Example Output:
```
* Plugin: inputs.dcos, Collection 1
> dcos_load,scope=node,cluster_url=https://m1.dcos,host=GAMGEE,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60 1min_count=0.02,15min_count=0.05,5min_count=0.05 1507644146000000000
> dcos_filesystem,cluster_url=https://m1.dcos,path=/,host=GAMGEE,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60,scope=node inode_total_count=26214400,inode_free_count=26064405,inode_used_count=149995,capacity_total_bytes=53660876800,capacity_used_bytes=3394183168,capacity_free_bytes=50266693632 1507644146000000000
> dcos_swap,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60,scope=node,cluster_url=https://m1.dcos,host=GAMGEE free_bytes=2147045376,total_bytes=2147479552,used_bytes=434176 1507644146000000000
> dcos_memory,scope=node,cluster_url=https://m1.dcos,host=GAMGEE,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1 total_bytes=1569218560,free_bytes=92704768,buffers_bytes=471040,cached_bytes=783945728 1507644146000000000
> dcos_cpu,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60,scope=node,cluster_url=https://m1.dcos,host=GAMGEE cores_count=2,idle_percent=98.08,wait_percent=0,user_percent=1.12,total_percent=1.86,system_percent=0.74 1507644146000000000
> dcos_process,scope=node,cluster_url=https://m1.dcos,host=GAMGEE,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60 count=193 1507644146000000000
> dcos_system,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.60,scope=node,cluster_url=https://m1.dcos,host=GAMGEE uptime_count=50059 1507644146000000000
> dcos_cpus,hostname=192.168.65.60,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,executor_id=marathon-lb.ca05c4b2-adb9-11e7-aa1f-70b3d5800001,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_url=https://m1.dcos,container_id=6ff2c2a3-6669-4a03-bf79-ee783884e39a,framework_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-0001,package_name=marathon-lb,scope=container,host=GAMGEE user_time_seconds=7.44,throttled_time_seconds=0,limit_count=2.1,system_time_seconds=16.05 1507644146000000000
> dcos_mem,hostname=192.168.65.60,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,scope=container,container_id=6ff2c2a3-6669-4a03-bf79-ee783884e39a,framework_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-0001,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,cluster_url=https://m1.dcos,host=GAMGEE,executor_id=marathon-lb.ca05c4b2-adb9-11e7-aa1f-70b3d5800001,package_name=marathon-lb limit_bytes=301989888,total_bytes=0 1507644146000000000
> dcos_net,package_name=marathon-lb,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,host=GAMGEE,scope=container,cluster_url=https://m1.dcos,framework_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-0001,hostname=192.168.65.60,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,container_id=6ff2c2a3-6669-4a03-bf79-ee783884e39a,executor_id=marathon-lb.ca05c4b2-adb9-11e7-aa1f-70b3d5800001 tx_bytes=0,rx_bytes=0,tx_dropped_count=0,tx_packets_count=0,tx_errors_count=0,rx_dropped_count=0,rx_packets_count=0,rx_errors_count=0 1507644146000000000
> dcos_disk,executor_id=marathon-lb.ca05c4b2-adb9-11e7-aa1f-70b3d5800001,cluster_url=https://m1.dcos,host=GAMGEE,framework_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-0001,hostname=192.168.65.60,package_name=marathon-lb,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S1,scope=container,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,container_id=6ff2c2a3-6669-4a03-bf79-ee783884e39a limit_bytes=0,used_bytes=0 1507644146000000000
> dcos_swap,hostname=192.168.65.111,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,scope=node,cluster_url=https://m1.dcos,host=GAMGEE used_bytes=0,total_bytes=2147479552,free_bytes=2147479552 1507644147000000000
> dcos_filesystem,host=GAMGEE,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.111,scope=node,cluster_url=https://m1.dcos,path=/ capacity_free_bytes=49171406848,capacity_used_bytes=4489469952,inode_total_count=26214400,capacity_total_bytes=53660876800,inode_free_count=26082535,inode_used_count=131865 1507644147000000000
> dcos_cpu,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.111,scope=node,cluster_url=https://m1.dcos,host=GAMGEE system_percent=0.72,total_percent=3.4799999999999995,idle_percent=96.39,wait_percent=0,user_percent=2.76,cores_count=4 1507644147000000000
> dcos_system,hostname=192.168.65.111,scope=node,cluster_url=https://m1.dcos,host=GAMGEE,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87 uptime_count=50160 1507644147000000000
> dcos_load,host=GAMGEE,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.111,scope=node,cluster_url=https://m1.dcos 1min_count=0.35,15min_count=0.22,5min_count=0.3 1507644147000000000
> dcos_process,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,scope=node,cluster_url=https://m1.dcos,host=GAMGEE,hostname=192.168.65.111 count=208 1507644147000000000
> dcos_memory,mesos_id=ce86d3aa-65e2-49fa-a0b7-dc933ad82fe2-S0,cluster_id=0d6f9827-190a-402c-91e3-5dd9c1288a87,hostname=192.168.65.111,scope=node,cluster_url=https://m1.dcos,host=GAMGEE buffers_bytes=970752,total_bytes=6088818688,free_bytes=225067008,cached_bytes=3123318784 1507644147000000000
  ...
```
