# Kibana input plugin

The [kibana](https://www.elastic.co/) plugin queries Kibana status API to
obtain the health status of Kibana and some useful metrics.

### Configuration

```toml
[[inputs.kibana]]
   ## specify a list of one or more Kibana servers
  # you can add username and password to your url to use basic authentication:
  # servers = ["http://user:pass@localhost:5601"]
  servers = ["http://localhost:5601"]

  ## Timeout for HTTP requests
  http_timeout = "5s"

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
  - status: string (green, yellow, red)
  - status_code: integer (1, 2, 3, 0)
  - heap_max_bytes: integer
  - heap_used_bytes: integer
  - uptime_ms: integer
  - response_time_avg_ms: integer
  - response_time_max_ms: integer
  - concurrent_connections: integer

### Tags

- name (Kibana reported name)
- uuid (Kibana reported UUID)
- version (Kibana version)
- server (Kibana server hostname or IP)

### Example Output

kibana,host=myhost,name=my-kibana,server=localhost:5601,uuid=00000000-0000-0000-0000-000000000000,version=6.3.2 concurrent_connections=0i,heap_max_bytes=136478720i,heap_used_bytes=119231088i,response_time_avg_ms=0i,response_time_max_ms=0i,status="green",status_code=1i,uptime_ms=2187428019i 1534864502000000000
