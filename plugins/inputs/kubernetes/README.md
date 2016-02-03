# Kubernetes Input Plugin
The kubernetes plugin uses the docker remote API to gather metrics on running
docker containers. You can read Docker's documentation for their remote API
[here](https://docs.docker.com/engine/reference/api/docker_remote_api_v1.20/#get-container-stats-based-on-resource-usage)

It then decorates those metrics with the kubernetes labels (and docker labels).

The kubernetes plugin uses the excellent
[fsouza go-dockerclient](https://github.com/fsouza/go-dockerclient) library to
gather stats. Documentation for the library can be found
[here](https://godoc.org/github.com/fsouza/go-dockerclient) and documentation
for the stat structure can be found
[here](https://godoc.org/github.com/fsouza/go-dockerclient#Stats)

### Configuration:

```
# Read metrics about docker containers
[[inputs.kubernetes]]
  # Docker Endpoint
  #   To use TCP, set endpoint = "tcp://[ip]:[port]"
  #   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"
  # Only collect metrics for these containers, collect all if empty
  container_names = []
```

### Measurements & Fields:

Please see the [docker input plugin](../docker/README.md) for detailed list of measurements.
