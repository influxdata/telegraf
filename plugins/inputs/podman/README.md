# Podman Input Plugin

The podman plugin uses the Podman API to gather metrics on running
podman containers.

The podman plugin uses the [Official libpod Client](https://pkg.go.dev/github.com/containers/podman/v3@v3.2.3/pkg/bindings)
to gather stats from the [Engine API](https://pkg.go.dev/github.com/containers/podman/v3@v3.2.3/pkg/api/server).

### Configuration:

```toml
# Read metrics about docker containers
[[inputs.podman]]
  ## Podman Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  endpoint = "unix:///run/podman/podman.sock"

  ## Containers to include and exclude. Collect all if empty. 
  container_name_include = []
  container_name_exclude = []

  ## Timeout for podman list, info, and stats commands
  timeout = "5s"

```

Podman creates the default API socket for root environments in `/run/podman/podman.sock`. For non root environments, Podman will create the socket in a user's runntime directory. The socket path can be easily located with the `XDG_RUNTIME_DIR` environment variable:

```Go
	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
        socket := "unix:" + sock_dir + "/podman/podman.sock"
```

### Metrics:

- podman
  - tags:
    - engine_host
    - server_version
  + fields:
    - n_cpus
    - n_containers
    - n_containers_running
    - n_containers_stopped
    - n_containers_paused
    - n_images
    - total_mem

- podman_container_stats
  - tags:
    - engine_host
    - server_version
    - container_name
    - container_image
    - container_version
    - pod_name
  + fields:
    - container_id
    - state
    - cpu
    - mem_usage
    - mem_limit
	
### Example Output:

```

```
