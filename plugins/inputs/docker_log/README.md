# Docker Log Input Plugin

The docker log plugin uses the Docker Engine API to get logs on running
docker containers.

The docker plugin uses the [Official Docker Client](https://github.com/moby/moby/tree/master/client)
to gather logs from the [Engine API](https://docs.docker.com/engine/api/v1.24/).
Note: This plugin works only for containers with the `local` or  `json-file` or `journald` logging driver.

### Configuration:

```toml
[[inputs.docker_log]]
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  # endpoint = "unix:///var/run/docker.sock"

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
  ## When empty only containers in the "running" state will be captured.
  # container_state_include = []
  # container_state_exclude = []

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  # docker_label_include = []
  # docker_label_exclude = []

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
```
