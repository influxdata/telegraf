# Docker Log Input Plugin

The docker log plugin uses the Docker Engine API to get logs on running
docker containers.

The docker plugin uses the [Official Docker Client](https://github.com/moby/moby/tree/master/client)
to gather logs from the [Engine API](https://docs.docker.com/engine/api/v1.24/).

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

```


### Metrics:

- docker_log
  - tags:
    - containerId
    - containerName
  - fields:
    - log
### Example Output:

```
docker_log,containerId=168c940a98b4317de15e336140bf6caae009c1ea948226d7fac84c839ccf6e6d,containerName=loving_leavitt,host=prash-laptop log=" root@168c940a98b4:/# ls\r\n" 1538210547000000000
docker_log,containerId=168c940a98b4317de15e336140bf6caae009c1ea948226d7fac84c839ccf6e6d,containerName=loving_leavitt,host=prash-laptop log=" bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var\r\n" 1538210547000000000
docker_log,containerId=168c940a98b4317de15e336140bf6caae009c1ea948226d7fac84c839ccf6e6d,containerName=loving_leavitt,host=prash-laptop log=" root@168c940a98b4:/# pwd\r\n /\r\n" 1538210552000000000
```
