# MarkLogic Plugin

The MarkLogic Telegraf plugin gathers health status metrics from one or more host.

### Configuration:

```toml
[[inputs.marklogic]]
  ## Base URL of the MarkLogic HTTP Server.
  url = "http://localhost:8002"

  ## List of specific hostnames to retrieve information. At least (1) required.
  # hosts = ["hostname1", "hostname2"]

  ## Using HTTP Digest Authentication. Management API requires 'manage-user' role privileges
  # username = "myuser"
  # password = "mypassword"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Metrics

- marklogic
  - tags:
    - source (the hostname of the server address, ex. `ml1.local`)
    - id (the host node unique id ex. `2592913110757471141`)
  - fields:
    - online
    - total_load
    - total_rate
    - ncpus
    - ncores
    - total_cpu_stat_user
    - total_cpu_stat_system
    - total_cpu_stat_idle
    - total_cpu_stat_iowait
    - memory_process_size
    - memory_process_rss
    - memory_system_total
    - memory_system_free
    - memory_size
    - host_size
    - data_dir_space
    - query_read_bytes
    - query_read_load
    - http_server_receive_bytes
    - http_server_send_bytes

### Example Output:

```
$> marklogic,host=localhost,id=17879472307902936940,source=ml1.local data_dir_space=28394i,host_size=0i,http_server_receive_bytes=0i,http_server_send_bytes=0i,memory_process_rss=187i,memory_process_size=622i,memory_size=4096i,memory_system_free=3772i,memory_system_total=3947i,ncores=4i,ncpus=1i,online=true,query_read_bytes=0i,query_read_load=0i,total_cpu_stat_idle=97.9717025756836,total_cpu_stat_iowait=0.0168676991015673,total_cpu_stat_system=0.695792019367218,total_cpu_stat_user=1.29881000518799,total_load=0,total_rate=33.8406105041504 1566024482000000000

```
