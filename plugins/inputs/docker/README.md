# Docker Input Plugin

The docker plugin uses the Docker Engine API to gather metrics on running
docker containers.

The docker plugin uses the [Official Docker Client](https://github.com/moby/moby/tree/master/client)
to gather stats from the [Engine API](https://docs.docker.com/engine/api/v1.20/).
[Library Documentation](https://godoc.org/github.com/moby/moby/client)

### Configuration:

```
# Read metrics about docker containers
[[inputs.docker]]
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Only collect metrics for these containers. Values will be appended to
  ## container_name_include.
  ## Deprecated (1.4.0), use container_name_include
  container_names = []

  ## Containers to include and exclude. Collect all if empty. Globs accepted.
  container_name_include = []
  container_name_exclude = []

  ## Timeout for docker list, info, and stats commands
  timeout = "5s"

  ## Whether to report for each container per-device blkio (8:0, 8:1...) and
  ## network (eth0, eth1, ...) stats or not
  perdevice = true

  ## Whether to report for each container total blkio and network stats or not
  total = false

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  docker_label_include = []
  docker_label_exclude = []

  ## Which environment variables should we use as a tag
  tag_env = ["JAVA_HOME", "HEAP_SIZE"]

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

#### Environment Configuration

When using the `"ENV"` endpoint, the connection is configured using the
[cli Docker environment variables](https://godoc.org/github.com/moby/moby/client#NewEnvClient).

### Measurements & Fields:

Every effort was made to preserve the names based on the JSON response from the
docker API.

Note that the docker_container_cpu metric may appear multiple times per collection,
based on the availability of per-cpu stats on your system.

- docker_container_mem
    - total_pgmafault
    - cache
    - mapped_file
    - total_inactive_file
    - pgpgout
    - rss
    - total_mapped_file
    - writeback
    - unevictable
    - pgpgin
    - total_unevictable
    - pgmajfault
    - total_rss
    - total_rss_huge
    - total_writeback
    - total_inactive_anon
    - rss_huge
    - hierarchical_memory_limit
    - total_pgfault
    - total_active_file
    - active_anon
    - total_active_anon
    - total_pgpgout
    - total_cache
    - inactive_anon
    - active_file
    - pgfault
    - inactive_file
    - total_pgpgin
    - max_usage
    - usage
    - failcnt
    - limit
    - container_id
- docker_container_cpu
    - throttling_periods
    - throttling_throttled_periods
    - throttling_throttled_time
    - usage_in_kernelmode
    - usage_in_usermode
    - usage_system
    - usage_total
    - usage_percent
    - container_id
- docker_container_net
    - rx_dropped
    - rx_bytes
    - rx_errors
    - tx_packets
    - tx_dropped
    - rx_packets
    - tx_errors
    - tx_bytes
    - container_id
- docker_container_blkio
    - io_service_bytes_recursive_async
    - io_service_bytes_recursive_read
    - io_service_bytes_recursive_sync
    - io_service_bytes_recursive_total
    - io_service_bytes_recursive_write
    - io_serviced_recursive_async
    - io_serviced_recursive_read
    - io_serviced_recursive_sync
    - io_serviced_recursive_total
    - io_serviced_recursive_write
    - container_id
- docker_
    - n_used_file_descriptors
    - n_cpus
    - n_containers
    - n_containers_running
    - n_containers_stopped
    - n_containers_paused
    - n_images
    - n_goroutines
    - n_listener_events
    - memory_total
    - pool_blocksize
- docker_data
    - available
    - total
    - used
- docker_metadata
    - available
    - total
    - used


### Tags:
#### Docker Engine tags
- docker (memory_total)
    - unit=bytes
    - engine_host
- docker (pool_blocksize)
    - unit=bytes
    - engine_host
- docker_data
    - unit=bytes
    - engine_host
- docker_metadata
    - unit=bytes
    - engine_host

#### Docker Container tags
- Tags on all containers:
    - engine_host
    - container_image
    - container_name
    - container_version
- docker_container_mem specific:
- docker_container_cpu specific:
    - cpu
- docker_container_net specific:
    - network
- docker_container_blkio specific:
    - device

### Example Output:

```
% ./telegraf --config ~/ws/telegraf.conf --input-filter docker --test
* Plugin: docker, Collection 1
> docker n_cpus=8i 1456926671065383978
> docker n_used_file_descriptors=15i 1456926671065383978
> docker n_containers=7i 1456926671065383978
> docker n_containers_running=7i 1456926671065383978
> docker n_containers_stopped=3i 1456926671065383978
> docker n_containers_paused=0i 1456926671065383978
> docker n_images=152i 1456926671065383978
> docker n_goroutines=36i 1456926671065383978
> docker n_listener_events=0i 1456926671065383978
> docker,unit=bytes memory_total=18935443456i 1456926671065383978
> docker,unit=bytes pool_blocksize=65540i 1456926671065383978
> docker_data,unit=bytes available=24340000000i,total=107400000000i,used=14820000000i 1456926671065383978
> docker_metadata,unit=bytes available=2126999999i,total=2146999999i,used=20420000i 145692667106538
> docker_container_mem,
container_image=spotify/kafka,container_name=kafka \
active_anon=52568064i,active_file=6926336i,cache=12038144i,fail_count=0i,\
hierarchical_memory_limit=9223372036854771712i,inactive_anon=52707328i,\
inactive_file=5111808i,limit=1044578304i,mapped_file=10301440i,\
max_usage=140656640i,pgfault=63762i,pgmajfault=2837i,pgpgin=73355i,\
pgpgout=45736i,rss=105275392i,rss_huge=4194304i,total_active_anon=52568064i,\
total_active_file=6926336i,total_cache=12038144i,total_inactive_anon=52707328i,\
total_inactive_file=5111808i,total_mapped_file=10301440i,total_pgfault=63762i,\
total_pgmafault=0i,total_pgpgin=73355i,total_pgpgout=45736i,\
total_rss=105275392i,total_rss_huge=4194304i,total_unevictable=0i,\
total_writeback=0i,unevictable=0i,usage=117440512i,writeback=0i 1453409536840126713
> docker_container_cpu,
container_image=spotify/kafka,container_name=kafka,cpu=cpu-total \
throttling_periods=0i,throttling_throttled_periods=0i,\
throttling_throttled_time=0i,usage_in_kernelmode=440000000i,\
usage_in_usermode=2290000000i,usage_system=84795360000000i,\
usage_total=6628208865i 1453409536840126713
> docker_container_cpu,
container_image=spotify/kafka,container_name=kafka,cpu=cpu0 \
usage_total=6628208865i 1453409536840126713
> docker_container_net,\
container_image=spotify/kafka,container_name=kafka,network=eth0 \
rx_bytes=7468i,rx_dropped=0i,rx_errors=0i,rx_packets=94i,tx_bytes=946i,\
tx_dropped=0i,tx_errors=0i,tx_packets=13i 1453409536840126713
> docker_container_blkio,
container_image=spotify/kafka,container_name=kafka,device=8:0 \
io_service_bytes_recursive_async=80216064i,io_service_bytes_recursive_read=79925248i,\
io_service_bytes_recursive_sync=77824i,io_service_bytes_recursive_total=80293888i,\
io_service_bytes_recursive_write=368640i,io_serviced_recursive_async=6562i,\
io_serviced_recursive_read=6492i,io_serviced_recursive_sync=37i,\
io_serviced_recursive_total=6599i,io_serviced_recursive_write=107i 1453409536840126713
```
