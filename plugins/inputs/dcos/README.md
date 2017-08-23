# DC/OS Input Plugin

This input plugin gathers metrics from a DC/OS cluster.
For more information, please check the [DC/OS Metrics](https://dcos.io/docs/1.9/metrics/) page.

### Configuration:

```toml
# Input plugin for gathering DC/OS agent metrics
[[inputs.dcos]]
  # Base URL of DC/OS cluster, e.g. http://dcos.example.com
  cluster_url=""
  # Authentication token, obtained by running: dcos config show core.dcos_acs_token
  auth_token=""
  # List of  DC/OS agent hostnames from which the metrics should be gathers. Leave empty for all.
  agents = []
  # DC/OS agent node file system mount for which related metrics should be gathered. Leave empty for all.
  file_system_mounts = []
  # DC/OS agent node network interface names for which related metrics should be gathered. Leave empty for all.
  network_interfaces = []
  # HTTP Response timeout, value must be more than a second
  #client_timeout = 4s
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
* mesos_id
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
> dcos_load,host=GAMGEE,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,hostname=10.0.1.178,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com 1min_count=0,5min_count=0,15min_count=0 1503487657000000000
> dcos_filesystem,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,hostname=10.0.1.178,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,path=/,host=GAMGEE inode_free_count=1431256,capacity_total_bytes=5843333120,capacity_used_bytes=1738280960,capacity_free_bytes=3822092288,inode_total_count=1498496,inode_used_count=67240 1503487657000000000
> dcos_memory,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,hostname=10.0.1.178,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,host=GAMGEE total_bytes=15776616448,free_bytes=13291368448,buffers_bytes=71159808,cached_bytes=1875877888 1503487657000000000
> dcos_swap,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,hostname=10.0.1.178,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,host=GAMGEE total_bytes=0,free_bytes=0,used_bytes=0 1503487657000000000
> dcos_network,hostname=10.0.1.178,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,interface=eth0,host=GAMGEE,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5 in_packets_count=745823,out_packets_count=452194,in_dropped_count=0,out_dropped_count=0,in_errors_count=0,out_errors_count=0,in_bytes=670494225,out_bytes=51747081 1503487657000000000
> dcos_process,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,host=GAMGEE,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,hostname=10.0.1.178 count=188 1503487657000000000
> dcos_system,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,hostname=10.0.1.178,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,host=GAMGEE uptime_count=10068 1503487657000000000
> dcos_cpu,hostname=10.0.1.178,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S5,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,scope=node,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,host=GAMGEE user_percent=0.17,system_percent=0.13,idle_percent=99.58,wait_percent=0.05,cores_count=4,total_percent=0.30000000000000004 1503487657000000000
> dcos_cpus,framework_principal=cassandra-principal,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S3,framework_id=84259197-1dda-44b1-934b-63fdf2ab1d37-0002,hostname=10.0.3.245,scope=container,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,executor_name=node-2_executor,executor_id=node-2_executor__171b3056-0eda-4eb5-ba55-65b4a5f82c84,host=GAMGEE,framework_name=cassandra,framework_role=cassandra-role,container_id=2ded88f3-ecce-4dad-9718-c472a65dde27,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5 user_time_seconds=270.73,system_time_seconds=158.24,limit_count=1,throttled_time_seconds=697.690003783 1503487657000000000
> dcos_mem,source=oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3,executor_id=oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3,HAPROXY_0_VHOST=oinker.acme.org,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,framework_principal=dcos_marathon,executor_name=Command\ Executor\ (Task:\ oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3)\ (Command:\ sh\ -c\ 'export\ OINKE...'),host=GAMGEE,container_id=aa3e66b2-b77e-42c9-9e5f-86e6b1cfd1fe,hostname=10.0.3.245,HAPROXY_GROUP=external,framework_name=marathon,framework_id=84259197-1dda-44b1-934b-63fdf2ab1d37-0001,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S3,framework_role=slave_public,scope=container total_bytes=310349824,limit_bytes=167772160 1503487657000000000
> dcos_disk,framework_name=marathon,framework_principal=dcos_marathon,scope=container,source=oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3,HAPROXY_GROUP=external,framework_role=slave_public,host=GAMGEE,container_id=aa3e66b2-b77e-42c9-9e5f-86e6b1cfd1fe,executor_id=oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3,HAPROXY_0_VHOST=oinker.acme.org,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,framework_id=84259197-1dda-44b1-934b-63fdf2ab1d37-0001,hostname=10.0.3.245,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S3,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,executor_name=Command\ Executor\ (Task:\ oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3)\ (Command:\ sh\ -c\ 'export\ OINKE...') limit_bytes=0,used_bytes=0 1503487657000000000
> dcos_net,hostname=10.0.3.245,HAPROXY_GROUP=external,framework_name=marathon,framework_role=slave_public,framework_principal=dcos_marathon,host=GAMGEE,executor_id=oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5,container_id=aa3e66b2-b77e-42c9-9e5f-86e6b1cfd1fe,HAPROXY_0_VHOST=oinker.acme.org,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S3,framework_id=84259197-1dda-44b1-934b-63fdf2ab1d37-0001,scope=container,executor_name=Command\ Executor\ (Task:\ oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3)\ (Command:\ sh\ -c\ 'export\ OINKE...'),source=oinker.9bbc8ad4-87e3-11e7-86f2-b2ce0138a0e3 tx_bytes=0,tx_errors_count=0,tx_dropped_count=0,rx_packets_count=0,rx_bytes=0,rx_errors_count=0,rx_dropped_count=0,tx_packets_count=0 1503487657000000000
> dcos_app,cluster_url=http://dcos-elasticloadba-ivcirgsfmba4-1406858042.us-west-1.elb.amazonaws.com,hostname=10.0.3.245,mesos_id=84259197-1dda-44b1-934b-63fdf2ab1d37-S3,container_id=32b42c71-3b6c-48fd-9db3-e8d5d20f1cb2,scope=app,host=GAMGEE,executor_id=cassandra.f9bbb8fe-87e1-11e7-86f2-b2ce0138a0e3,framework_id=84259197-1dda-44b1-934b-63fdf2ab1d37-0001,cluster_id=5e3d4c26-6bc0-4ad3-8142-d104376ed2b5 dcos.metrics.module.container_received_bytes_per_sec=0,dcos.metrics.module.container_throttled_bytes_per_sec=0 1503487657000000000
```
