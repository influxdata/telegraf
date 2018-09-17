# Memcached Input Plugin

This plugin gathers statistics data from a Memcached server.

### Configuration:

```toml
# Read metrics from one or many memcached servers.
[[inputs.memcached]]
  # An array of address to gather stats about. Specify an ip on hostname
  # with optional port. ie localhost, 10.0.0.1:11211, etc.
  servers = ["localhost:11211"]
  # An array of unix memcached sockets to gather stats about.
  # unix_sockets = ["/var/run/memcached.sock"]
```

### Measurements & Fields:

The fields from this plugin are gathered in the *memcached* measurement.

Fields:

* get_hits - Number of keys that have been requested and found present
* get_misses - Number of items that have been requested and not found
* evictions - Number of valid items removed from cache to free memory for new items
* limit_maxbytes - Number of bytes this server is allowed to use for storage
* bytes - Current number of bytes used to store items
* uptime - Number of secs since the server started
* curr_items - Current number of items stored
* total_items - Total number of items stored since the server started
* curr_connections - Number of open connections
* total_connections - Total number of connections opened since the server started running
* connection_structures - Number of connection structures allocated by the server
* cmd_get - Cumulative number of retrieval reqs
* cmd_set - Cumulative number of storage reqs
* delete_hits - Number of deletion reqs resulting in an item being removed
* delete_misses - umber of deletions reqs for missing keys
* incr_hits - Number of successful incr reqs
* incr_misses - Number of incr reqs against missing keys
* decr_hits - Number of successful decr reqs
* decr_misses - Number of decr reqs against missing keys
* cas_hits - Number of successful CAS reqs
* cas_misses - Number of CAS reqs against missing keys
* bytes_read - Total number of bytes read by this server from network
* bytes_written - Total number of bytes sent by this server to network
* threads - Number of worker threads requested
* conn_yields - Number of times any connection yielded to another due to hitting the -R limit

Description of gathered fields taken from [here](https://github.com/memcached/memcached/blob/master/doc/protocol.txt).

### Tags:

* Memcached measurements have the following tags:
    - server (the host name from which metrics are gathered)

### Sample Queries:

You can use the following query to get the average get hit and miss ratio, as well as the total average size of cached items, number of cached items and average connection counts per server.

```
SELECT mean(get_hits) / mean(cmd_get) as get_ratio, mean(get_misses) / mean(cmd_get) as get_misses_ratio, mean(bytes), mean(curr_items), mean(curr_connections) FROM memcached WHERE time > now() - 1h GROUP BY server
```

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter memcached --test
memcached,server=localhost:11211 get_hits=1,get_misses=2,evictions=0,limit_maxbytes=0,bytes=10,uptime=3600,curr_items=2,total_items=2,curr_connections=1,total_connections=2,connection_structures=1,cmd_get=2,cmd_set=1,delete_hits=0,delete_misses=0,incr_hits=0,incr_misses=0,decr_hits=0,decr_misses=0,cas_hits=0,cas_misses=0,bytes_read=10,bytes_written=10,threads=1,conn_yields=0 1453831884664956455
```
