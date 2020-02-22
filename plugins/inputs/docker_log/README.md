# Docker container logs Input plugin

###The docker_log plugin uses docker API to stream logs from container.

---
__The primary motivation for this input plugin is to provide the following
features:__
1. Throttling for gathering container logs by limiting amount of log data to be read from the
each docker container and interval of log's reading. This primary use case here is to be protected from the situation 
when the container (from which the plugin reading logs) fall in to unlimited loop producing 
lots of log entries (stack-tracing for example). This can cause high CPU utilisation of telegraf, and in case there are several such containers on the host we can hit the CPU limit.
2. Allows to stream logs from the particular point in time even if telegraf crashed. This is achieved by storing offset (unix time stamp in nanoseconds of the last read log entry) for every container. When plugin is started it checks for the offset existence, and in case it found it, the logs will be streamed since the offset, so no entries would be lost. Entries will come with the original time-stamp.
3. Use original docker timestamps - to get the precise picture.
4. Optimisation for running under k8s for streaming logs from the containers in a POD.
Although there are some already available solutions to deliver logs from k8s containers,
this feature is of interest, because single telegraf binary can be used at the same time to
deliver metrics from the running applications and logs. No other solution needed.
Moreover, using telegraf provides great flexibility for tagging and filtering, that is of great help. 
The main feature here is to execute command when all containers from which we are streaming logs are exited/terminated.
This allows to terminate POD in a consistent way, after all logs are delivered.
---

To be able to use it, docker socket should be provided for runtime,
and docker container logging driver should be set to `json-file` or `journald`.

To query API, the possible oldest version used - 1.21 (https://docs-stage.docker.com/engine/api/v1.21/), 
to support as much of variety of docker versions as possible. 
To stream logs the following API endpoint is used `GET /containers/(id or name)/logs`

API version vs Docker version compatibility matrix: https://docs.docker.com/develop/sdk/
(see `API version matrix` chapter)

### Configuration:

```toml
[[inputs.docker_log]]  
  ## Docker Endpoint
  ##  To use unix, set endpoint = "unix:///var/run/docker.sock" (/var/run/docker.sock is default mount path)
  ##  To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##  To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
  ## When true, container logs are read from the beginning; otherwise
  ## reading begins at the end of the log.
  # from_beginning = false

  ## Timeout for Docker API calls.
  # timeout = "5s"

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  # container_name_include = []
  # container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty all states will be captured.
  ## Valid values are: "created", "restarting", "running", "removing", "paused", "exited", "dead"
  # container_state_include = []
  # container_state_exclude = []

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  # docker_label_include = []
  # docker_label_exclude = []
 
 
  ## Log streaming settings:
  ## Interval to gather data from docker sock.
  ## the longer the interval the fewer request is made towards docker API (less CPU utilization on dockerd).
  ## On the other hand, this increase the delay between producing logs and delivering it. Reasonable trade off
  ## should be chosen. Default value is 2000 ms.
  # log_gather_interval = "2000ms"

  ## Set the source tag for the metrics to the container ID hostname, eg first 12 chars
  source_tag = false

  ## Set initial chunk size (length of []byte buffer to read from docker socket)
  ## If not set, default value of 'defaultInitialChunkSize = 1000' will be used
  # initial_chunk_size = 1000 # 1K symbols (half of 80x25 screen)

  ## Max chunk size (length of []byte buffer to read from docker socket)
  ## Buffer can grow in capacity adjusting to volume of data received from docker sock
  ## to the maximum volume limited by this parameter. The bigger buffer is set
  ## the more data potentially it can read during 1 API call to docker.
  ## And all of this data will be processed before sending, that increase CPU utilization.
  ## This parameter should be set carefully.
  # max_chunk_size = 5000 # 5K symbols

  ## Offset flush interval. How often the offset pointer (see below) in the
  ## log stream is flashed to file.Offset pointer represents the unix time stamp
  ## in nano seconds for the last message read from log stream (default - 3 sec)
  # offset_flush = "3s"

  ## Offset storage path (mandatory), make sure the user on behalf 
  ## of which the telegraf is started has appropriate rights to read and write to chosen path.
  ## default value is "/var/run/telegraf/docker_log_offset"
  offset_storage_path = "/var/run/telegraf/docker_log_offset"
  
  ## Command to be run when all static containers (see section below) are processed.
  ## 'Processed' in this context mean that logs are delivered and container is not in a running state 
  #[inputs.docker_log.when_static_container_processed]
  #  execute_cmd=["s6-svc", "-d", "/services/service/run"]

  ## Optional static (means containers are not dinamycally discovered) containers configuration (specify as many sections as needed).
  ## The section below is mutually exclusive with the
  ## 'container_name_include' & 'container_name_exclude' options!
  ## The section below used to configure input for delivering logs from specific containers with
  ## individual settings for throttling. Primary use case is to define this config for containers in a k8s POD
  ## in which the telegraf is running as a separate container. This section used to be paired with 'when_static_container_processed'
  ## section, as it provides ability to finalized telegraf container in POD when the target static containers
  ## are exited.
  ## 
  #[[inputs.docker_log.container]]
  ## Set container id (long or short from, mutually exclusive with container name)
  #  id = "dc23d3ea534b3a6ec3934ae21e2dd4955fdbf61106b32fa19b831a6040a7feef"
  ## Set container name (mutually exclusive with container id)
  #  name = "quirky_fermi"

  ## Overriding common settings:
  # log_gather_interval = "500ms"

  ## Initial chunk size
  #  initial_chunk_size = 2000 # 2K symbols

  ## Max chunk size
  #  max_chunk_size = 6000 # 6K symbols

  #Set additional tags that will be tagged to the stream from the current container:
  # tags = [
  #      "tag1=value1",
  #      "tag2=value2"
  #  ]
  ##Another static container to stream logs from  
  #[[inputs.docker_log.container]]
  #  id = "009d82030745"
  #  interval = "600ms"
```

#### Environment Configuration

When using the `"ENV"` endpoint, the connection is configured using the
[CLI Docker environment variables][env]

[env]: https://godoc.org/github.com/moby/moby/client#NewEnvClient

### source tag

Selecting the containers can be tricky if you have many containers with the same name.
To alleviate this issue you can set the below value to `true`

```toml
source_tag = true
```
This will cause all data points to have the `source` tag be set to the first 12 characters of the container id. The first 12 characters is the common hostname for containers that have no explicit hostname set, as defined by docker.

### Metrics
- docker_log
  - fields:
    - container_id
    - message
  - tags (custom tags if specified by static container configuration is not listed):
    - container_image
    - container_version
    - container_name
    - stream (stdout, stderr, or tty)
    - source