# MarkLogic Plugin

The MarkLogic Telegraf plugin gathers status metrics from one or more host in the MarkLogic Cluster.

### Configuration:

```toml
[[inputs.marklogic]]
  ## Base URL of MarkLogic host for Management API endpoint.
  url = "http://localhost:8002"

  ## List of specific hostnames in a cluster to retrieve information. At least (1) required.
  # hosts = ["hostname1", "hostname2"]

  ## Using HTTP Digest Authentication. This requires 'manage-user' role privileges
  # username = "telegraf"
  # password = "p@ssw0rd"
```

### Measurement & Fields:

MarkLogic provides one measurement named "marklogic", with the following fields:

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

### Tags:

All measurements have the following tags:

- name (the hostname of the server address, ex. `ml1.local`)
- id (the host node unique id ex. `2592913110757471141`)

### Example Output:

```
$> marklogic,host=localhost,id=17879472307902936940,name=ml2.local data_dir_space=28398i,host_size=0i,http_server_receive_bytes=8059i,http_server_send_bytes=0i,memory_process_rss=303i,memory_process_size=728i,memory_size=4096i,memory_system_free=3664i,memory_system_total=3947i,ncores=4i,ncpus=1i,online=true,query_read_bytes=0i,query_read_load=0i,total_cpu_stat_idle=99.1915969848633,total_cpu_stat_iowait=0.0167552996426821,total_cpu_stat_system=0.523603975772858,total_cpu_stat_user=0.251329988241196,total_load=0,total_rate=16.1815872192383 1565067570000000000

```
