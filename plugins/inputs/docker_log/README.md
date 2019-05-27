# Docker Log Input Plugin

The docker log plugin uses the Docker Engine API to get logs on running
docker containers.

The docker plugin uses the [Official Docker Client](https://github.com/moby/moby/tree/master/client)
to gather logs from the [Engine API](https://docs.docker.com/engine/api/v1.24/).
Note: This plugin works only for containers with the `local` or  `json-file` or `journald` logging driver.
### Configuration:

```toml
# Read metrics about docker containers
[[inputs.docker_log]]
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Containers to include and exclude. Collect all if empty. Globs accepted.
  container_name_include = []
  container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  # container_state_include = []
  # container_state_exclude = []

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  docker_label_include = []
  docker_label_exclude = []

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

#### Environment Configuration

When using the `"ENV"` endpoint, the connection is configured using the
[cli Docker environment variables](https://godoc.org/github.com/moby/moby/client#NewEnvClient).

### Metrics:

- docker_log
  - tags:
    - container_id
    - container_name
    - stream
  - fields:
    - message
### Example Output:

```
docker_log,com.docker.compose.config-hash=e19e13df8fd01ba2d7c1628158fca45cc91afbbe9661b2d30550547eb53a861e,com.docker.compose.container-number=1,com.docker.compose.oneoff=False,com.docker.compose.project=distribution,com.docker.compose.service=influxdb,com.docker.compose.version=1.21.2,containerId=fce475bbfa4c8380ff858d5d767f78622ca6de955b525477624c2b7896a5b8e4,containerName=aicon-influxdb,host=prash-laptop,logType=stderr log=" [httpd] 172.23.0.2 - aicon_admin [13/Apr/2019:08:35:53 +0000] \"POST /query?db=&q=SHOW+SUBSCRIPTIONS HTTP/1.1\" 200 232 \"-\" \"KapacitorInfluxDBClient\" 2661bc9c-5dc7-11e9-82f8-0242ac170007 1360\n" 1555144553541000000
docker_log,com.docker.compose.config-hash=fd91b3b096c7ab346971c681b88fe1357c609dcc6850e4ea5b1287ad28a57e5d,com.docker.compose.container-number=1,com.docker.compose.oneoff=False,com.docker.compose.project=distribution,com.docker.compose.service=kapacitor,com.docker.compose.version=1.21.2,containerId=6514d1cf6d19e7ecfedf894941f0a2ea21b8aac5e6f48e64f19dbc9bb2805a25,containerName=aicon-kapacitor,host=prash-laptop,logType=stderr log=" ts=2019-04-13T08:36:00.019Z lvl=info msg=\"http request\" service=http host=172.23.0.7 username=- start=2019-04-13T08:36:00.013835165Z method=POST uri=/write?consistency=&db=_internal&precision=ns&rp=monitor protocol=HTTP/1.1 status=204 referer=- user-agent=InfluxDBClient request-id=2a3eb481-5dc7-11e9-825b-000000000000 duration=5.814404ms\n" 1555144560024000000
```
