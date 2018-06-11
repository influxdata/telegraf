# Libvirt Plugin

As of now, this plugin parses the dump xml to get the Domain devices (disks) metrics and interfaces
so that calls can be made to gather metrics.

# TODO

- write tests

### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
# Description
[[inputs.libvirtd]]
  lib_virt_sock = "/var/run/libvirt/libvirt-sock"
  
  ## Network Devices is a WIP
  network_devices = []
```

### Metrics:
- libvirt_vmdata
  - tags:
    - host (optional description)
    - id
    - name
  - fields:
    - alignment_faults
    - branch_instructions
    - branch_misses
    - bus_cycles
    - cache_misses
    - cache_references
    - cmt
    - context_switches
    - cpu_clock
    - cpu_count
    - cpu_cycles
    - cpu_migrations
    - cpu_time
    - emulation_faults
    - instructions
    - max_memory
    - mbml
    - mbmt
    - mem_used_percent
    - page_faults
    - page_faults_maj
    - page_faults_min
    - ref_cpu_cycles
    - stalled_cycles_backend
    - stalled_cycles_frontend
    - state
    - task_clock
    - total_memory
    - vcpu_time

- libvirtd_disk
  - tags:
    - id
    - device
    - name
    - host
  - fields:
    - allocation
    - capacity
    - flush_operations
    - flush_total_times
    - physical
    - rd_bytes
    - rd_operations
    - rd_total_times
    - wr_bytes
    - wr_operations
    - wr_total_times



### Example Output:

```
> livbirtd_disk,device=hdc,name=r-2824-QA,ID=̡,host=njcloudhost.dev.ena.net wr_bytes="&{3 0}",rd_operations="&{3 68}",
flush_operations="&{3 0}",wr_operations="&{3 0}",rd_bytes="&{3 450836}",wr_total_times="&{3 0}",
rd_total_times="&{3 97202811}",flush_total_times="&{3 0}",allocation=78528512i,capacity=78526464i,physical=78526464i 
1513349120000000000


>libvirtd_vmdata,name=r-2828-QA,id=̢,host=njcloudhost.dev.ena.net branch_misses="&{6 0}",ref_cpu_cycles="&{6 0}",
mem_used_percent=1i,max_memory=262144i,cmt="&{6 0}",cache_references="&{6 0}",stalled_cycles_frontend="&{6 0}",
cpu_count=1i,mbmt="&{6 0}",cpu_cycles="&{6 0}",cache_misses="&{6 0}",cpu_clock="&{6 0}",task_clock="&{6 0}",
bus_cycles="&{6 0}",vcpu_time="&{4 77308414402}",mbml="&{6 0}",instructions="&{6 0}",branch_instructions="&{6 0}",
cpu_migrations="&{6 0}",page_faults_min="&{6 0}",context_switches="&{6 0}",emulation_faults="&{6 0}",state=1i,
total_memory=262144i,stalled_cycles_backend="&{6 0}",page_faults="&{6 0}",page_faults_maj="&{6 0}",
alignment_faults="&{6 0}",cpu_time=511790000000i 1513349119000000000
```
