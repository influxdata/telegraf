# Docker container logs Input plugin

###The docker_cnt_logs plugin uses docker API to stream logs from container.

---
__The primary motivation for this input plugin is to provide the following
features that are not support by current plugins:__
1. Allow to set limits to how often and how many log data is read from the
each docker container. This primary use case here is to be protected from the situation 
when the container (from which the plugin reading logs) fall in to unlimited loop producing 
lots of log entries (stack-tracing for example). In case wy simply stream this log entries as it is, via telegraf plugin,
we will get high CPU utilization of telegraf, and in case there are several such containers on the host
we can hit the CPU limit.
2. Allows to stream logs from the particular point in time even if telegraf crashed. This is 
achieved by storing offset (unix time stamp in nanoseconds of the last read log entry)
for every container. When plugin is started it checks for the offset existence, and in case it found
it, the logs will be streamed since the offset, so no entries would be lost. Entries will come with the
original time-stamp.
3. Optimization for running under k8s for streaming logs from the containers in a POD.
Although there are some already available solutions to deliver logs from k8s containers,
this feature is of interest, because single telegraf binary can be used at the same time to
deliver metrics from the running applications and logs. No other solution needed.
Moreover, using telegraf provides great flexibility for tagging and filtering, that is
of great help. The main feature here is to shutdown telegraf when all containers from which we are streaming logs
are exited/terminated. This allows to terminate POD in a consistent way, after all logs are delivered.
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
[[inputs.docker_cnt_logs]]  
  ## Interval to gather data from docker sock.
  ## the longer the interval the fewer request is made towards docker API (less CPU utilization on dockerd).
  ## On the other hand, this increase the delay between producing logs and delivering it. Reasonable trade off
  ## should be chosen
  interval = "2000ms"
  
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

  ## Log streaming settings
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
  offset_storage_path = "/var/run/collector_offset"

  ## Shutdown telegraf if all log streaming containers stopped/killed, default - false
  ## This option make sense when telegraf started especially for streaming logs
  ## in a form of sidecar container in k8s. In case primary container exited,
  ## side-car should be terminated also.
  # shutdown_when_eof = false

  ## Settings per container (specify as many sections as needed)
  [[inputs.docker_cnt_logs.container]]
    ## Set container id (long or short from), or container name
    ## to stream logs from, this attribute is mandatory
    id = "dc23d3ea534b3a6ec3934ae21e2dd4955fdbf61106b32fa19b831a6040a7feef"

    ## Override common settings
    ## input interval (specified or inherited from agent section)
    # interval = "500ms"

    ## Initial chunk size
    initial_chunk_size = 2000 # 2K symbols

    ## Max chunk size
    max_chunk_size = 6000 # 6K symbols

    #Set additional tags that will be tagged to the stream from the current container:
    tags = [
        "tag1=value1",
        "tag2=value2"
    ]
  ##Another container to stream logs from  
  [[inputs.docker_cnt_logs.container]]
    id = "009d82030745c9994e2f5c2280571e8b9f95681793a8f7073210759c74c1ea36"
    interval = "600ms"
```

### Metrics:
* stream
  - fields:
	- value (string), the log message itself
  - tags:
    - conatainer_id
    - stream `stdin`,`stderr`,`stdout`,`interfactive`
