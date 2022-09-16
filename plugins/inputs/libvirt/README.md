# Libvirt Input Plugin

The `libvirt` plugin collects statistics about virtualized
guests on a system by using virtualization libvirt API,
created by RedHat's Emerging Technology group.  
Metrics are gathered directly from the hypervisor on a host
system, which means that Telegraf doesn't have to be installed
and configured on a guest system.

## Prerequisites

For proper operation of the libvirt plugin,
it is required that the host system has:

- enabled virtualization options for host CPU
- libvirtd and its dependencies installed and running
- qemu hypervisor installed and running
- at least one virtual machine for statistics monitoring

Useful links:

- [libvirt](https://libvirt.org/)
- [qemu](https://www.qemu.org/)

## Configuration

```toml
# The libvirt plugin collects statistics from virtualized guests using virtualization libvirt API.
[[inputs.libvirt]]
     ## Domain names from which libvirt gather statistics.
     ## By default (empty or missing array) the plugin gather statistics from each domain registered in the host system.
     # domains = []

     ## Libvirt connection URI with hypervisor.
     ## The plugin supports multiple transport protocols and approaches which are configurable via the URI.
     ## The general URI form: driver[+transport]://[username@][hostname][:port]/[path][?extraparameters]
     ## Supported transport protocols: ssh, tcp, tls, unix
     ## URI examples for each type of transport protocol:
     ## 1. SSH:  qemu+ssh://<USER@IP_OR_HOSTNAME>/system?keyfile=/<PATH_TO_PRIVATE_KEY>&known_hosts=/<PATH_TO_known_hosts>
     ## 2. TCP:  qemu+tcp://<IP_OR_HOSTNAME>/system
     ## 3. TLS:  qemu+tls://<HOSTNAME>/system?pkipath=/certs_dir/<COMMON_LOCATION_OF_CACERT_AND_SERVER_CLIENT_CERTS>
     ## 4. UNIX: qemu+unix:///system?socket=/<PATH_TO_libvirt-sock>
     ## Default URI is qemu:///system
     # libvirt_uri = "qemu:///system"

     ## Statistics groups for which libvirt plugin will gather statistics.
     ## Supported statistics groups: state, cpu_total, balloon, vcpu, interface, block, perf, iothread, memory, dirtyrate
     ## Empty array means no metrics for statistics groups will be exposed by the plugin.
     ## By default the plugin will gather all available statistics.
     # statistics_groups = ["state", "cpu_total", "balloon", "vcpu", "interface", "block", "perf", "iothread", "memory", "dirtyrate"]

     ## A list containing additional statistics to be exposed by libvirt plugin.
     ## Supported additional statistics: vcpu_mapping
     ## By default (empty or missing array) the plugin will not collect additional statistics.
     # additional_statistics = []
```

Useful links:

- [Libvirt URI docs](https://libvirt.org/uri.html)
- [TLS setup for libvirt](https://wiki.libvirt.org/page/TLSSetup)

In cases when one or more of the following occur:

- the global Telegraf variable `interval` is set to a low value (e.g. 1s),
- a significant number of VMs are monitored,
- the medium connecting the plugin to the hypervisor is inefficient,  

It is possible that following warning in the logs appears:
`Collection took longer than expected`.

For that case, `interval` should be set inside plugin configuration.
Its value should be adjusted to plugin's runtime environment.

Example:

```toml
[[inputs.libvirt]]
  interval = "30s"
```

### Example configuration

```toml
[[inputs.libvirt]]
  domain_names = ["ubuntu_20"]
  libvirt_uri = "qemu:///system"
  libvirt_metrics = ["state", "interface"]
  additional_statistics = ["vcpu_mapping"]
```

## Metrics

Below the table containing a list of all metrics
supported by the libvirt plugin is presented.  
The metrics are divided into the following groups of statistics:

- state
- cpu_total
- balloon
- vcpu
- net
- perf
- block
- iothread
- memory
- dirtyrate
- vcpu_mapping - additional statistics

Statistics groups from the plugin corresponds to the grouping of
metrics directly read from libvirtd using the `virsh domstats` command.  
More details about metrics can be found at the links below:

- [Domain statistics](https://libvirt.org/manpages/virsh.html#domstats)
- [Performance monitoring events](https://libvirt.org/formatdomain.html#performance-monitoring-events)

| **Statistics group** | **Metric name** | **Exposed Telegraf field** | **Description** |
|:---|:---|:---|:---|
| **state** | state.state | state | state of the VM, returned as number from virDomainState enum |
||state.reason | reason | reason for entering given state, returned as int from virDomain*Reason enum corresponding to given state |
| **cpu_total** | cpu.time | time | total cpu time spent for this domain in nanoseconds |
|| cpu.user | user | user cpu time spent in nanoseconds |
|| cpu.system | system | system cpu time spent in nanoseconds |
|| cpu.haltpoll.success.time | haltpoll_success_time | cpu halt polling success time spent in nanoseconds |
|| cpu.haltpoll.fail.time | haltpoll_fail_time | cpu halt polling fail time spent in nanoseconds |
|| cpu.cache.monitor.count |count | the number of cache monitors for this domain |
|| cpu.cache.monitor.\<num\>.name | name | the name of cache monitor \<num\>, not available for kernels from 4.14 upwards |
|| cpu.cache.monitor.\<num\>.vcpus| vcpus |vcpu list of cache monitor \<num\>, not available for kernels from 4.14 upwards |
|| cpu.cache.monitor.\<num\>.bank.count | bank_count | the number of cache banks in cache monitor \<num\>, not available for kernels from 4.14 upwards |
|| cpu.cache.monitor.\<num\>.bank.\<index\>.id | id|host allocated cache id for bank \<index\> in cache monitor \<num\>, not available for kernels from 4.14 upwards |
|| cpu.cache.monitor.\<num\>.bank.\<index\>.bytes | bytes | the number of bytes of last level cache that the domain is using on cache bank \<index\>, not available for kernels from 4.14 upwards|
| **balloon** | balloon.current | current | the memory in KiB currently used |
|| balloon.maximum | maximum | the maximum memory in KiB allowed |
|| balloon.swap_in | swap_in | the amount of data read from swap space (in KiB) |
|| balloon.swap_out | swap_out | the amount of memory written out to swap space (in KiB) |
|| balloon.major_fault | major_fault | the number of page faults when disk IO was required |
|| balloon.minor_fault | minor_fault | the number of other page faults |
|| balloon.unused | unused | the amount of memory left unused by the system (in KiB) |
|| balloon.available | available | the amount of usable memory as seen by the domain (in KiB) |
|| balloon.rss | rss | Resident Set Size of running domain's process (in KiB) |
|| balloon.usable | usable | the amount of memory which can be reclaimed by balloon without causing host swapping (in KiB) |
|| balloon.last-update | last_update | timestamp of the last update of statistics (in seconds) |
|| balloon.disk_caches | disk_caches | the amount of memory that can be reclaimed without additional I/O, typically disk (in KiB) |
|| balloon.hugetlb_pgalloc | hugetlb_pgalloc | the number of successful huge page allocations from inside the domain via virtio balloon |
|| balloon.hugetlb_pgfail | hugetlb_pgfail | the number of failed huge page allocations from inside the domain via virtio balloon |
| **vcpu** | vcpu.current | current | yes current number of online virtual CPUs |
|| vcpu.maximum | maximum | maximum number of online virtual CPUs |
|| vcpu.\<num\>.state | state | state of the virtual CPU \<num\>, as number from virVcpuState enum |
|| vcpu.\<num\>.time | time | virtual cpu time spent by virtual CPU \<num\> (in microseconds) |
|| vcpu.\<num\>.wait | wait | virtual cpu time spent by virtual CPU \<num\> waiting on I/O (in microseconds) |
|| vcpu.\<num\>.halted | halted | virtual CPU \<num\> is halted: yes or no (may indicate the processor is idle or even disabled, depending on the architecture) |
|| vcpu.\<num\>.halted | halted_i | virtual CPU \<num\> is halted: 1 (for "yes") or 0 (for other values) (may indicate the processor is idle or even disabled, depending on the architecture) |
|| vcpu.\<num\>.delay | delay | time the vCPU \<num\> thread was enqueued by the host scheduler, but was waiting in the queue instead of running. Exposed to the VM as a steal time. |
|| --- | cpu_id | Information about mapping vcpu_id to cpu_id (id of physical cpu). Should only be exposed when statistics_group contains vcpu and additional_statistics contains vcpu_mapping (in config) |
| **interface** | net.count | count | number of network interfaces on this domain |
|| net.\<num\>.name | name | name of the interface  \<num\> |
|| net.\<num\>.rx.bytes | rx_bytes | number of bytes received |
|| net.\<num\>.rx.pkts | rx_pkts | number of packets received |
|| net.\<num\>.rx.errs | rx_errs | number of receive errors |
|| net.\<num\>.rx.drop | rx_drop | number of receive packets dropped |
|| net.\<num\>.tx.bytes | tx_bytes | number of bytes transmitted |
|| net.\<num\>.tx.pkts | tx_pkts | number of packets transmitted |
|| net.\<num\>.tx.errs | tx_errs | number of transmission errors |
|| net.\<num\>.tx.drop | tx_drop | number of transmit packets dropped |
| **perf** | perf.cmt | cmt | the cache usage in Byte currently used, not available for kernels from 4.14 upwards |
|| perf.mbmt | mbmt | total system bandwidth from one level of cache, not available for kernels from 4.14 upwards |
|| perf.mbml | mbml | bandwidth of memory traffic for a memory controller, not available for kernels from 4.14 upwards |
|| perf.cpu_cycles | cpu_cycles | the count of cpu cycles (total/elapsed) |
|| perf.instructions | instructions |  the count of instructions |
|| perf.cache_references | cache_references | the count of cache hits |
|| perf.cache_misses | cache_misses | the count of caches misses |
|| perf.branch_instructions | branch_instructions | the count of branch instructions |
|| perf.branch_misses | branch_misses | the count of branch misses |
|| perf.bus_cycles | bus_cycles | the count of bus cycles |
|| perf.stalled_cycles_frontend | stalled_cycles_frontend | the count of stalled frontend cpu cycles |
|| perf.stalled_cycles_backend | stalled_cycles_backend | the count of stalled backend cpu cycles |
|| perf.ref_cpu_cycles | ref_cpu_cycles | the count of ref cpu cycles |
|| perf.cpu_clock | cpu_clock | the count of cpu clock time |
|| perf.task_clock | task_clock | the count of task clock time |
|| perf.page_faults | page_faults | the count of page faults |
|| perf.context_switches | context_switches | the count of context switches |
|| perf.cpu_migrations | cpu_migrations | the count of cpu migrations |
|| perf.page_faults_min | page_faults_min | the count of minor page faults |
|| perf.page_faults_maj | page_faults_maj | the count of major page faults |
|| perf.alignment_faults | alignment_faults | the count of alignment faults |

|| perf.emulation_faults | emulation_faults | the count of emulation faults |
| **block** | block.count | count | number of block devices being listed |
|| block.\<num\>.name | name | name of the target of the block device  \<num\> (the same name for multiple entries if --backing is present) |
|| block.\<num\>.backingIndex | backingIndex | when --backing is present, matches up with the \<backingStore\> index listed in domain XML for backing files |
|| block.\<num\>.path | path | file source of block device  \<num\>, if it is a local file or block device |
|| block.\<num\>.rd.reqs | rd_reqs | number of read requests |
|| block.\<num\>.rd.bytes | rd_bytes | number of read bytes |
|| block.\<num\>.rd.times | rd_times | total time (ns) spent on reads |
|| block.\<num\>.wr.reqs | wr_reqs | number of write requests |
|| block.\<num\>.wr.bytes | wr_bytes | number of written bytes |
|| block.\<num\>.wr.times | wr_times | total time (ns) spent on writes |
|| block.\<num\>.fl.reqs | fl_reqs | total flush requests |
|| block.\<num\>.fl.times | fl_times | total time (ns) spent on cache flushing |
|| block.\<num\>.errors | errors | Xen only: the 'oo_req' value |
|| block.\<num\>.allocation | allocation | offset of highest written sector in bytes |
|| block.\<num\>.capacity | capacity | logical size of source file in bytes |
|| block.\<num\>.physical | physical | physical size of source file in bytes |
|| block.\<num\>.threshold | threshold | threshold (in bytes) for delivering the VIR_DOMAIN_EVENT_ID_BLOCK_THRESHOLD event. See domblkthreshold |
| **iothread** | iothread.count | count | maximum number of IOThreads in the subsequent list as unsigned int. Each IOThread in the list will will use it's iothread_id value as the \<id\>. There may be fewer \<id\> entries than the iothread.count value if the polling values are not supported |
|| iothread.\<id\>.poll-max-ns | poll_max_ns | maximum polling time in nanoseconds used by the \<id\> IOThread. A value of 0 (zero) indicates polling is disabled |
|| iothread.\<id\>.poll-grow | poll_grow | polling time grow value. A value of 0 (zero) growth is managed by the hypervisor |
|| iothread.\<id\>.poll-shrink | poll_shrink | polling time shrink value. A value of (zero) indicates shrink is managed by hypervisor |
| **memory** | memory.bandwidth.monitor.count | count | the number of memory bandwidth monitors for this domain, not available for kernels from 4.14 upwards |
|| memory.bandwidth.monitor.\<num\>.name | name | the name of monitor  \<num\>, not available for kernels from 4.14 upwards |
|| memory.bandwidth.monitor.\<num\>.vcpus | vcpus | the vcpu list of monitor  \<num\>, not available for kernels from 4.14 upwards |
|| memory.bandwidth.monitor.\<num\>.node.count | node_count | the number of memory controller in monitor \<num\>, not available for kernels from 4.14 upwards |
|| memory.bandwidth.monitor.\<num\>.node.\<index\>.id | id | host allocated memory controller id for controller \<index\> of monitor \<num\>, not available for kernels from 4.14 upwards |
|| memory.bandwidth.monitor.\<num\>.node.\<index\>.bytes.local | bytes_local | the accumulative bytes consumed by \@vcpus that passing through the memory controller in the same processor that the scheduled host CPU belongs to, not available for kernels from 4.14 upwards |
|| memory.bandwidth.monitor.\<num\>.node.\<index\>.bytes.total | bytes_total | the total bytes consumed by \@vcpus that passing through all memory controllers, either local or remote controller, not available for kernels from 4.14 upwards |
| **dirtyrate** | dirtyrate.calc_status | calc_status | the status of last memory dirty rate calculation, returned as number from virDomainDirtyRateStatus enum |
|| dirtyrate.calc_start_time | calc_start_time the | start time of last memory dirty rate calculation |
|| dirtyrate.calc_period | calc_period | the period of last memory dirty rate calculation |
|| dirtyrate.megabytes_per_second | megabytes_per_second | the calculated memory dirty rate in MiB/s |
|| dirtyrate.calc_mode | calc_mode | the calculation mode used last measurement (page-sampling/dirty-bitmap/dirty-ring) |
|| dirtyrate.vcpu.\<num\>.megabytes_per_second | megabytes_per_second | the calculated memory dirty rate for a virtual cpu in MiB/s |

### Additional statistics

| **Statistics group**           | **Exposed Telegraf tag**      | **Exposed Telegraf field**      |**Description**         |
|:-------------------------------|:-------------------------:|:-------------------------:|:-------------------------|
| **vcpu_mapping** | vcpu_id | --- | ID of Virtual CPU |
|| --- | cpu_id | Comma separated list (exposed as a string) of Physical CPU IDs |

## Example Output

```shell
libvirt_cpu_affinity,domain_name=U22,host=localhost,vcpu_id=0 cpu_id="1,2,3" 1662383707000000000
libvirt_cpu_affinity,domain_name=U22,host=localhost,vcpu_id=1 cpu_id="1,2,3,4,5,6,7,8,9,10" 1662383707000000000
libvirt_balloon,domain_name=U22,host=localhost current=4194304i,maximum=4194304i,swap_in=0i,swap_out=0i,major_fault=0i,minor_fault=0i,unused=3928628i,available=4018480i,rss=1036012i,usable=3808724i,last_update=1654611373i,disk_caches=68820i,hugetlb_pgalloc=0i,hugetlb_pgfail=0i 1662383709000000000
libvirt_vcpu_total,domain_name=U22,host=localhost maximum=2i,current=2i 1662383709000000000
libvirt_vcpu,domain_name=U22,host=localhost,vcpu_id=0 state=1i,time=17943740000000i,wait=0i,halted="no",halted_i=0i,delay=14246609424i,cpu_id=1i 1662383709000000000
libvirt_vcpu,domain_name=U22,host=localhost,vcpu_id=1 state=1i,time=18288400000000i,wait=0i,halted="yes",halted_i=1i,delay=12902231142i,cpu_id=3i 1662383709000000000
libvirt_net_total,domain_name=U22,host=localhost count=1i 1662383709000000000
libvirt_net,domain_name=U22,host=localhost,interface_id=0 name="vnet0",rx_bytes=110i,rx_pkts=1i,rx_errs=0i,rx_drop=31007i,tx_bytes=0i,tx_pkts=0i,tx_errs=0i,tx_drop=0i 1662383709000000000
libvirt_block_total,domain_name=U22,host=localhost count=1i 1662383709000000000
libvirt_block,domain_name=U22,host=localhost,block_id=0 rd=17337818234i,path=name="vda",backingIndex=1i,path="/tmp/ubuntu_image.img",rd_reqs=11354i,rd_bytes=330314752i,rd_times=6240559566i,wr_reqs=52440i,wr_bytes=1183828480i,wr_times=21887150375i,fl_reqs=32250i,fl_times=23158998353i,errors=0i,allocation=770048000i,capacity=2361393152i,physical=770052096i,threshold=2147483648i
libvirt_perf,domain_name=U22,host=localhost cmt=19087360i,mbmt=77168640i,mbml=67788800i,cpu_cycles=29858995122i,instructions=0i,cache_references=3053301695i,cache_misses=609441024i,branch_instructions=2623890194i,branch_misses=103707961i,bus_cycles=188105628i,stalled_cycles_frontend=0i,stalled_cycles_backend=0i,ref_cpu_cycles=30766094039i,cpu_clock=25166642695i,task_clock=25263578917i,page_faults=2670i,context_switches=294284i,cpu_migrations=17949i,page_faults_min=2670i,page_faults_maj=0i,alignment_faults=0i,emulation_faults=0i 1662383709000000000
libvirt_dirtyrate,domain_name=U22,host=localhost calc_status=2i,calc_start_time=348414i,calc_period=1i,dirtyrate.megabytes_per_second=4i,calc_mode="dirty-ring" 1662383709000000000
libvirt_dirtyrate_vcpu,domain_name=U22,host=localhost,vcpu_id=0 megabytes_per_second=2i 1662383709000000000
libvirt_dirtyrate_vcpu,domain_name=U22,host=localhost,vcpu_id=1 megabytes_per_second=2i 1662383709000000000
libvirt_state,domain_name=U22,host=localhost state=1i,reason=5i 1662383709000000000
libvirt_cpu,domain_name=U22,host=localhost time=67419144867000i,user=63886161852000i,system=3532983015000i,haltpoll_success_time=516907915i,haltpoll_fail_time=2727253643i 1662383709000000000
libvirt_cpu_cache_monitor_total,domain_name=U22,host=localhost count=1i 1662383709000000000
libvirt_cpu_cache_monitor,domain_name=U22,host=localhost,cache_monitor_id=0 name="any_name_vcpus_0-3",vcpus="0-3",bank_count=1i 1662383709000000000
libvirt_cpu_cache_monitor_bank,domain_name=U22,host=localhost,cache_monitor_id=0,bank_index=0 id=0i,bytes=5406720i 1662383709000000000
libvirt_iothread_total,domain_name=U22,host=localhost count=1i 1662383709000000000
libvirt_iothread,domain_name=U22,host=localhost,iothread_id=0 poll_max_ns=32768i,poll_grow=0i,poll_shrink=0i 1662383709000000000
libvirt_memory_bandwidth_monitor_total,domain_name=U22,host=localhost count=2i 1662383709000000000
libvirt_memory_bandwidth_monitor,domain_name=U22,host=localhost,memory_bandwidth_monitor_id=0 name="any_name_vcpus_0-4",vcpus="0-4",node_count=2i 1662383709000000000
libvirt_memory_bandwidth_monitor,domain_name=U22,host=localhost,memory_bandwidth_monitor_id=1 name="vcpus_7",vcpus="7",node_count=2i 1662383709000000000
libvirt_memory_bandwidth_monitor_node,domain_name=U22,host=localhost,memory_bandwidth_monitor_id=0,controller_index=0 id=0i,bytes_total=10208067584i,bytes_local=4807114752i 1662383709000000000
libvirt_memory_bandwidth_monitor_node,domain_name=U22,host=localhost,memory_bandwidth_monitor_id=0,controller_index=1 id=1i,bytes_total=8693735424i,bytes_local=5850161152i 1662383709000000000
libvirt_memory_bandwidth_monitor_node,domain_name=U22,host=localhost,memory_bandwidth_monitor_id=1,controller_index=0 id=0i,bytes_total=853811200i,bytes_local=290701312i 1662383709000000000
libvirt_memory_bandwidth_monitor_node,domain_name=U22,host=localhost,memory_bandwidth_monitor_id=1,controller_index=1 id=1i,bytes_total=406044672i,bytes_local=229425152i 1662383709000000000
```
