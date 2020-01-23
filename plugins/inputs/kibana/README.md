# Kibana input plugin

The [kibana](https://www.elastic.co/) plugin queries Kibana status API to
obtain the health status of Kibana and some useful metrics.

This plugin has been tested and works on Kibana 6.x versions.

### Configuration

```toml
[[inputs.kibana]]
  ## specify a list of one or more Kibana servers
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

### Status mappings

When reporting health (green/yellow/red), additional field `status_code`
is reported. Field contains mapping from status:string to status_code:int
with following rules:

- `green` - 1
- `yellow` - 2
- `red` - 3
- `unknown` - 0

### Measurements & Fields

- kibana
  - status_code: integer (1, 2, 3, 0)
  - heap_total_bytes: integer
  - heap_max_bytes: integer
  - heap_used_bytes: integer
  - uptime_ms: integer
  - response_time_avg_ms: float
  - response_time_max_ms: integer
  - concurrent_connections: integer
  - requests_per_sec: float

### Tags

- name (Kibana reported name)
- source (Kibana server hostname or IP)
- status (Kibana health: green, yellow, red)
- version (Kibana version)

### Example Output
```
kibana,host=myhost,name=my-kibana,source=localhost:5601,status=green,version=6.5.4 concurrent_connections=8i,heap_max_bytes=447778816i,heap_total_bytes=447778816i,heap_used_bytes=380603352i,requests_per_sec=1,response_time_avg_ms=57.6,response_time_max_ms=220i,status_code=1i,uptime_ms=6717489805i 1534864502000000000
```
