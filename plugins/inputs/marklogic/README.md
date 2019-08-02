# Marklogic Plugin

The Marklogic Telegraf plugin gathers status metrics from one or more host in the Marklogic cluster.

### Configuration:

```toml
# Description
[[inputs.marklogic]]

## List URLs of Marklogic hosts using Management API endpoint.
# hosts = ["http://localhost:8002/manage/v2/hosts/${hostname}?view=status&format=json"]

# Using HTTP Digest Authentication.
# digest_username = "telegraf"
# digest_password = "p@ssw0rd"
```

### Measurements & Fields:

Marklogic provides one measurement named "marklogic", with the following fields:

- online
- total_cpu_stat_user
- total_cpu_stat_system
- memory_process_size
- memory_process_rss
- memory_system_total
- memory_system_free
- num_cores
- total_load
- data_dir_space
- query_read_bytes
- query_read_load
- http_server_receive_bytes
- http_server_send_bytes

### Tags:

All measurements have the following tags:

- ml_hostname (the hostname of the server address, ex. `ml1.local`)
- id (the host node unique id ex. `2592913110757471141`)

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter marklogic --test
> marklogic,host=Craigs-MacBook-Pro-2.local,id=2592913110757471141,ml_hostname=ml1.local data_dir_space=29212i,http_server_receive_bytes=0,http_server_send_bytes=0,memory_process_rss=273i,memory_process_size=702i,memory_system_free=3417i,memory_system_total=3947i,num_cores=4i,online=true,query_read_bytes=3545624i,query_read_load=0i,total_cpu_stat_system=0.63408100605011,total_cpu_stat_user=0.302343010902405,total_load=0.0847966074943543 1564731542000000000
```
