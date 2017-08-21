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

Below are enumerated the metrics taken from version 1.9 of DC/OS. See [DC/OS Metrics Reference](https://dcos.io/docs/1.9/metrics/reference//) for a description of those metrics.
Name of field has added suffix meaning unit of metric value.

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
- dcos_metrics_module
  - container_received_bytes_per_sec
  - container_throttled_bytes_per_sec
  
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

#### App metric tags
* mesos_id
* cluster_id
* container_id
* executor_id
* framework_id
* hostname"

### Example Output:
```
* Plugin: inputs.dcos, Collection 1
> dcos,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0,scope=node,cluster_url=http://m1.dcos,host=GAMGEE,hostname=192.168.65.111 in.packets.enp0s3=14873,network.out.errors.enp0s8=0,cpu.idle=98.98,network.in.errors.d-dcos=0,network.in.vtep1024=0,swap.used=0,network.out.d-dcos=0,network.out.errors.dummy0=0,network.out.errors.minuteman=0,network.in.docker0=2068,filesystem.inode.total_var_lib_docker_overlay=26214400,network.in.d-dcos=0,network.in.veth7134f20=648,network.out.dropped.veth7134f20=0,filesystem.capacity.free_=50562560000,network.in.errors.veth7134f20=0,network.out.minuteman=0,network.out.docker0=0,filesystem.capacity.total_=53660876800,memory.cached=584286208,network.out.errors.lo=0,filesystem.capacity.used_home=93216768,filesystem.capacity.total_var_lib_docker_overlay=53660876800,network.in.dropped.veth7134f20=0,network.out.dropped.lo=0,network.in.dropped.veth0a57a71=0,network.in.packets.dummy0=0,network.out.vtep1024=0,cpu.user=0.59,filesystem.inode.free_var_lib_docker_overlay=26102969,network.in.dropped.spartan=0,network.out.packets.enp0s3=14629,network.in.packets.spartan=0,network.in.errors.dummy0=0,network.in.lo=1075283285,network.in.packets.lo=509865,memory.buffers=970752,swap.free=2147479552,network.out.dropped.docker0=0,filesystem.capacity.used_var_lib_docker_overlay=3098316800,network.in.errors.spartan=0,network.in.errors.enp0s3=0,network.out.lo=1075283285,network.in.packets.vtep1024=0,network.in.errors.minuteman=0,system.uptime=133327,filesystem.inode.total_=26214400,filesystem.capacity.total_home=50432839680,network.out.dropped.d-dcos=0,filesystem.inode.used_boot=328,filesystem.inode.free_boot=523960,network.out.packets.minuteman=0,network.in.errors.docker0=0,network.out.packets.spartan=0,network.out.packets.enp0s8=3723360,network.out.veth0a57a71=648,network.in.packets.docker0=31,filesystem.capacity.used_=3098316800,network.out.packets.vtep1024=0,network.out.dropped.vtep1024=0,network.out.packets.docker0=0,filesystem.inode.free_home=24637423,filesystem.capacity.free_var_lib_docker_overlay=50562560000,network.in.packets.veth0a57a71=8,network.out.errors.veth0a57a71=0,network.out.dropped.minuteman=0,network.out.errors.spartan=0,network.in.enp0s3=1203805,network.out.packets.dummy0=0,network.in.dropped.dummy0=0,load.5min=0.04,network.in.dummy0=0,network.in.packets.minuteman=0,memory.free=4964446208,network.in.errors.veth0a57a71=0,network.in.packets.d-dcos=0,network.in.packets.enp0s8=3623456,load.1min=0,network.in.dropped.vtep1024=0,network.out.errors.vtep1024=0,cpu.cores=4,filesystem.capacity.free_home=50339622912,network.out.packets.veth7134f20=8,network.in.dropped.d-dcos=0,filesystem.capacity.total_boot=1063256064,filesystem.inode.total_boot=524288,filesystem.inode.used_home=17,filesystem.inode.used_var_lib_docker_overlay=111431,filesystem.inode.used_=111431,network.out.dropped.enp0s3=0,network.in.dropped.lo=0,network.out.errors.docker0=0,network.out.dummy0=0,network.out.packets.lo=509865,network.in.minuteman=0,filesystem.inode.free_=26102969,memory.total=6088818688,network.in.dropped.enp0s3=0,network.out.veth7134f20=648,cpu.system=0.4,network.in.errors.lo=0,network.in.packets.veth7134f20=8,swap.total=2147479552,network.out.dropped.enp0s8=0,network.out.packets.veth0a57a71=8,network.out.dropped.veth0a57a71=0,filesystem.capacity.used_boot=144031744,network.out.enp0s3=1195413,network.in.enp0s8=477720812,network.in.dropped.enp0s8=0,network.in.errors.vtep1024=0,cpu.total=0.99,cpu.wait=0,load.15min=0.05,filesystem.capacity.free_boot=919224320,network.in.dropped.docker0=0,network.out.dropped.spartan=0,network.out.errors.d-dcos=0,network.out.errors.enp0s3=0,network.out.enp0s8=564039765,process.count=218,network.in.spartan=0,network.out.packets.d-dcos=0,network.in.errors.enp0s8=0,network.in.dropped.minuteman=0,network.out.dropped.dummy0=0,filesystem.inode.total_home=24637440,network.out.spartan=0,network.in.veth0a57a71=648,network.out.errors.veth7134f20=0 1502440896000000000
> dcos,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S1,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0,hostname=192.168.65.60,scope=node,cluster_url=http://m1.dcos,host=GAMGEE network.in.errors.enp0s3=0,network.in.packets.docker0=0,process.count=179,network.in.enp0s3=1163662,network.out.dropped.vtep1024=0,network.out.errors.minuteman=0,cpu.system=0.51,network.in.packets.minuteman=0,filesystem.capacity.free_var_lib_docker_overlay=50659610624,network.in.dropped.dummy0=0,network.out.dropped.lo=0,network.out.packets.minuteman=0,network.out.dropped.docker0=0,filesystem.inode.free_=26111889,filesystem.inode.used_boot=328,network.out.dropped.d-dcos=0,network.out.lo=359641206,network.out.errors.vtep1024=0,network.in.errors.minuteman=0,filesystem.capacity.free_boot=919224320,network.in.errors.lo=0,network.out.errors.docker0=0,cpu.total=1.12,network.out.packets.d-dcos=0,network.in.dropped.lo=0,swap.used=0,network.in.packets.spartan=0,memory.free=722472960,network.out.dropped.spartan=0,cpu.idle=98.84,load.1min=0.04,network.out.packets.docker0=0,network.in.minuteman=0,network.out.docker0=0,filesystem.inode.total_=26214400,memory.buffers=970752,network.in.packets.enp0s3=14332,network.in.packets.enp0s8=3613499,network.out.errors.dummy0=0,network.in.dropped.vtep1024=0,network.in.docker0=0,filesystem.capacity.total_home=50432839680,filesystem.inode.used_var_lib_docker_overlay=102511,memory.total=1569218560,network.out.errors.enp0s3=0,network.in.dropped.enp0s8=0,load.15min=0.05,filesystem.capacity.total_boot=1063256064,network.out.dropped.enp0s3=0,network.in.errors.dummy0=0,filesystem.capacity.used_boot=144031744,filesystem.inode.free_boot=523960,network.out.enp0s3=1160102,network.out.packets.enp0s8=3696787,cpu.user=0.61,network.in.d-dcos=0,network.out.dropped.minuteman=0,network.in.errors.docker0=0,swap.total=2147479552,network.out.errors.enp0s8=0,network.in.lo=359641206,load.5min=0.05,network.out.dropped.dummy0=0,network.in.dropped.spartan=0,network.out.packets.enp0s3=14123,network.out.dummy0=0,cpu.wait=0,filesystem.inode.total_boot=524288,network.in.errors.enp0s8=0,network.out.errors.spartan=0,network.in.dropped.enp0s3=0,network.in.errors.spartan=0,network.in.errors.d-dcos=0,network.out.errors.d-dcos=0,network.in.vtep1024=0,filesystem.capacity.free_home=50339622912,network.out.packets.spartan=0,network.out.minuteman=0,network.out.d-dcos=0,memory.cached=444518400,network.in.dummy0=0,network.in.errors.vtep1024=0,filesystem.capacity.used_var_lib_docker_overlay=3001266176,filesystem.inode.free_var_lib_docker_overlay=26111889,filesystem.capacity.total_var_lib_docker_overlay=53660876800,network.in.packets.d-dcos=0,network.out.dropped.enp0s8=0,filesystem.inode.total_home=24637440,filesystem.inode.free_home=24637423,network.in.enp0s8=473115065,filesystem.capacity.total_=53660876800,filesystem.inode.used_=102511,network.in.dropped.d-dcos=0,network.out.enp0s8=523228527,network.in.dropped.docker0=0,filesystem.capacity.used_=3001266176,filesystem.capacity.used_home=93216768,network.in.dropped.minuteman=0,network.in.packets.dummy0=0,network.out.packets.vtep1024=0,swap.free=2147479552,network.in.packets.lo=502552,network.out.packets.lo=502552,network.out.vtep1024=0,system.uptime=133444,cpu.cores=2,filesystem.inode.used_home=17,network.in.spartan=0,network.out.spartan=0,network.out.errors.lo=0,filesystem.capacity.free_=50659610624,filesystem.inode.total_var_lib_docker_overlay=26214400,network.out.packets.dummy0=0,network.in.packets.vtep1024=0 1502440895000000000
> dcos,hostname=192.168.65.111,scope=container,cluster_url=http://m1.dcos,executor_name=Command\ Executor\ (Task:\ basic-0.9f578110-7b63-11e7-be7d-70b3d5800001)\ (Command:\ sh\ -c\ 'while\ [\ true...'),host=GAMGEE,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,framework_role=slave_public,framework_principal=dcos_marathon,container_id=a7bcd1ca-484b-4d36-9439-6bef12dc9613,source=basic-0.9f578110-7b63-11e7-be7d-70b3d5800001,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0,framework_name=marathon,executor_id=basic-0.9f578110-7b63-11e7-be7d-70b3d5800001,framework_id=a52c2640-d3b9-49c8-b92f-a17b2c25cd70-0001 mem.limit=44040192,net.rx.packets=0,net.rx.bytes=0,disk.limit=0,net.rx.errors=0,net.tx.bytes=0,net.tx.dropped=0,cpus.user.time=0.15,cpus.system.time=0.08,mem.total=7340032,disk.used=0,cpus.limit=0.2,cpus.throttled.time=0.407527461,net.rx.dropped=0,net.tx.packets=0,net.tx.errors=0 1502105442000000000
> dcos,source=mynginxserver.99a70d7f-7b63-11e7-be7d-70b3d5800001,framework_principal=dcos_marathon,framework_name=marathon,container_id=f942c669-6ed1-4864-ba96-20529710def5,cluster_url=http://m1.dcos,executor_name=Command\ Executor\ (Task:\ mynginxserver.99a70d7f-7b63-11e7-be7d-70b3d5800001)\ (Command:\ NO\ EXECUTABLE),scope=container,host=GAMGEE,framework_role=slave_public,hostname=192.168.65.111,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,executor_id=mynginxserver.99a70d7f-7b63-11e7-be7d-70b3d5800001,framework_id=a52c2640-d3b9-49c8-b92f-a17b2c25cd70-0001,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0 cpus.user.time=0.03,cpus.system.time=0.02,net.rx.errors=0,net.tx.bytes=0,cpus.throttled.time=0,disk.used=0,net.rx.dropped=0,net.tx.dropped=0,cpus.limit=1.1,mem.total=0,disk.limit=0,net.rx.packets=0,net.rx.bytes=0,net.tx.packets=0,mem.limit=167772160,net.tx.errors=0 1502105442000000000
> dcos,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,hostname=192.168.65.111,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0,executor_id=mynginxserver.5186b5b2-7b67-11e7-be7d-70b3d5800001,executor_name=Command\ Executor\ (Task:\ mynginxserver.5186b5b2-7b67-11e7-be7d-70b3d5800001)\ (Command:\ NO\ EXECUTABLE),framework_name=marathon,framework_id=a52c2640-d3b9-49c8-b92f-a17b2c25cd70-0001,source=mynginxserver.5186b5b2-7b67-11e7-be7d-70b3d5800001,host=GAMGEE,framework_role=slave_public,container_id=762fdbc9-1574-4ee6-ab5b-8ab4fbc1239d,framework_principal=dcos_marathon,scope=container,cluster_url=http://m1.dcos net.tx.dropped=0,cpus.throttled.time=0,net.tx.errors=0,cpus.user.time=0.03,mem.total=6660096,disk.used=0,net.rx.bytes=0,net.rx.errors=0,net.tx.packets=0,net.tx.bytes=0,cpus.limit=1.1,mem.limit=167772160,disk.limit=0,net.rx.packets=0,net.rx.dropped=0,cpus.system.time=0 1502440898000000000
> dcos,framework_principal=dcos_marathon,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0,hostname=192.168.65.111,framework_id=a52c2640-d3b9-49c8-b92f-a17b2c25cd70-0001,scope=container,executor_name=Command\ Executor\ (Task:\ mynginxserver.998cf5ce-7b63-11e7-be7d-70b3d5800001)\ (Command:\ NO\ EXECUTABLE),framework_role=slave_public,container_id=5aed0dd8-ca5e-4e79-a410-32e1d6e45cc0,cluster_url=http://m1.dcos,host=GAMGEE,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,executor_id=mynginxserver.998cf5ce-7b63-11e7-be7d-70b3d5800001,framework_name=marathon,source=mynginxserver.998cf5ce-7b63-11e7-be7d-70b3d5800001 net.tx.errors=0,cpus.system.time=0.02,mem.total=0,disk.used=0,net.rx.bytes=0,net.rx.errors=0,net.tx.bytes=0,cpus.user.time=0.03,cpus.throttled.time=0,net.rx.packets=0,cpus.limit=1.1,net.tx.packets=0,net.tx.dropped=0,mem.limit=167772160,disk.limit=0,net.rx.dropped=0 1502105442000000000
> dcos,container_id=f94cee4f-155d-430b-9d35-83441367820e,framework_principal=dcos_marathon,cluster_url=http://m1.dcos,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,framework_role=slave_public,executor_id=mynginxserver.5189c2f3-7b67-11e7-be7d-70b3d5800001,host=GAMGEE,hostname=192.168.65.111,framework_name=marathon,scope=container,executor_name=Command\ Executor\ (Task:\ mynginxserver.5189c2f3-7b67-11e7-be7d-70b3d5800001)\ (Command:\ NO\ EXECUTABLE),source=mynginxserver.5189c2f3-7b67-11e7-be7d-70b3d5800001,framework_id=a52c2640-d3b9-49c8-b92f-a17b2c25cd70-0001,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0 cpus.limit=1.1,net.rx.packets=0,net.rx.dropped=0,net.tx.packets=0,net.tx.dropped=0,net.tx.errors=0,cpus.system.time=0.01,mem.limit=167772160,disk.limit=0,disk.used=0,net.rx.errors=0,mem.total=6987776,net.rx.bytes=0,cpus.user.time=0.03,cpus.throttled.time=0,net.tx.bytes=0 1502440898000000000
> dcos,framework_id=a52c2640-d3b9-49c8-b92f-a17b2c25cd70-0001,cluster_url=http://m1.dcos,executor_name=Command\ Executor\ (Task:\ basic-0.1b75af51-7b65-11e7-be7d-70b3d5800001)\ (Command:\ sh\ -c\ 'while\ [\ true...'),host=GAMGEE,mesos_id=b0da75eb-bbe7-4ad9-80a2-582890b16a1b-S0,framework_name=marathon,cluster_id=2f4b3291-ee34-4779-b7bd-015f6594e9c0,container_id=619e0c1a-a059-4801-ae60-75f022f89df7,executor_id=basic-0.1b75af51-7b65-11e7-be7d-70b3d5800001,scope=container,source=basic-0.1b75af51-7b65-11e7-be7d-70b3d5800001,framework_role=slave_public,framework_principal=dcos_marathon,hostname=192.168.65.111 cpus.system.time=111.31,net.rx.bytes=0,net.rx.dropped=0,net.tx.errors=0,mem.total=7401472,disk.used=0,cpus.user.time=78.5,disk.limit=0,net.tx.bytes=0,net.tx.dropped=0,net.tx.packets=0,cpus.limit=0.2,cpus.throttled.time=2.715591212,mem.limit=44040192,net.rx.packets=0,net.rx.errors=0 1502440898000000000
```
