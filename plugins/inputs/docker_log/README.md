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

  ## Only collect metrics for these containers. Values will be appended to
  ## container_name_include.
  ## Deprecated (1.4.0), use container_name_include
  container_names = []

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
  - fields:
    - log
### Example Output:

```
docker_log,containerId=4325333a47ab42c78b8bf5cb01d5b0972321f857a4b9e116856b4f0459047077,host=prash-laptop log=" root@4325333a47ab:/# ls -l\r\n" 1530162134000000000
docker_log,containerId=4325333a47ab42c78b8bf5cb01d5b0972321f857a4b9e116856b4f0459047077,host=prash-laptop log=" total 64\r\n drwxr-xr-x   2 root root 4096 May 26 00:45 bin\r\n drwxr-xr-x   2 root root 4096 Apr 24 08:34 boot\r\n drwxr-xr-x   5 root root  360 Jun 28 05:01 dev\r\n drwxr-xr-x   1 root root 4096 Jun 28 05:01 etc\r\n drwxr-xr-x   2 root root 4096 Apr 24 08:34 home\r\n drwxr-xr-x   8 root root 4096 May 26 00:44 lib\r\n drwxr-xr-x   2 root root 4096 May 26 00:44 lib64\r\n drwxr-xr-x   2 root root 4096 May 26 00:44 media\r\n drwxr-xr-x   2 root root 4096 May 26 00:44 mnt\r\n" 1530162134000000000
```
