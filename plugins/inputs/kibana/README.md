# Kibana Input Plugin

The `kibana` plugin queries the [Kibana][] API to obtain the service status.

- Telegraf minimum version: 1.8
- Kibana minimum tested version: 6.0

[Kibana]: https://www.elastic.co/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read status information from one or more Kibana servers
[[inputs.kibana]]
  ## Specify a list of one or more Kibana servers
  servers = ["http://localhost:5601"]

  ## Timeout for HTTP requests
  timeout = "5s"

  ## HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

- kibana
  - tags:
    - name (Kibana reported name)
    - source (Kibana server hostname or IP)
    - status (Kibana health: green, yellow, red)
    - version (Kibana version)
  - fields:
    - status_code (integer, green=1 yellow=2 red=3 unknown=0)
    - heap_total_bytes (integer)
    - heap_max_bytes (integer; deprecated in 1.13.3: use `heap_total_bytes` field)
    - heap_used_bytes (integer)
    - heap_size_limit (integer)
    - uptime_ms (integer)
    - response_time_avg_ms (float)
    - response_time_max_ms (integer)
    - concurrent_connections (integer)
    - requests_per_sec (float)

## Example Output

```text
kibana,host=myhost,name=my-kibana,source=localhost:5601,status=green,version=6.5.4 concurrent_connections=8i,heap_max_bytes=447778816i,heap_total_bytes=447778816i,heap_used_bytes=380603352i,requests_per_sec=1,response_time_avg_ms=57.6,response_time_max_ms=220i,status_code=1i,uptime_ms=6717489805i 1534864502000000000
```

## Run example environment

Requires the following tools:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

From the root of this project execute the following script:
`./plugins/inputs/kibana/test_environment/run_test_env.sh`

This will build the latest Telegraf and then start up Kibana and Elasticsearch,
Telegraf will begin monitoring Kibana's status and write its results to the file
`/tmp/metrics.out` in the Telegraf container.

Then you can attach to the telegraf container to inspect the file
`/tmp/metrics.out` to see if the status is being reported.

The Visual Studio Code [Remote - Containers][remote] extension provides an easy
user interface to attach to the running container.

[remote]: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers
