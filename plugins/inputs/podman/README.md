# Podman Input Plugin

The podman plugin uses the Podman API to gather metrics on **running**
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
podman,engine_host=ubuntu,host=xps,server_version=3.2.0 n_containers=2i,n_containers_paused=0i,n_containers_running=1i,n_containers_stopped=1i,n_cpus=8i,n_images=18i,total_mem=7966027776i 1629149800000000000
podman_container_stats,container_image=docker.io/library/nginx,container_name=nginx,container_version=latest,engine_host=ubuntu,host=xps,server_version=3.2.0 container_id="a16fe92067f111642c17e27b6c78c3d728468f162d93a02e554a3eea9326f548",cpu=0.0000000019008074018757257,mem_limit=7966027776i,mem_usage=19124224i,state="running" 1629149801000000000
```
