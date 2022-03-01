# Memcached Input Plugin

This plugin gathers statistics data from a Memcached server.

## Configuration

```toml
# Read metrics from one or many memcached servers.
[[inputs.memcached]]
  # An array of address to gather stats about. Specify an ip on hostname
  # with optional port. ie localhost, 10.0.0.1:11211, etc.
  servers = ["localhost:11211"]
  # An array of unix memcached sockets to gather stats about.
  # unix_sockets = ["/var/run/memcached.sock"]

  ## Optional TLS Config
  # enable_tls = true
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## If false, skip chain & host verification
  # insecure_skip_verify = true
```

## Measurements & Fields

The fields from this plugin are gathered in the *memcached* measurement.

Fields:

* accepting_conns - Whether or not server is accepting conns
* auth_cmds - Number of authentication commands handled, success or failure
* auth_errors - Number of failed authentications
* bytes - Current number of bytes used to store items
* bytes_read - Total number of bytes read by this server from network
* bytes_written - Total number of bytes sent by this server to network
* cas_badval - Number of CAS reqs for which a key was found, but the CAS value did not match
* cas_hits - Number of successful CAS reqs
* cas_misses - Number of CAS reqs against missing keys
* cmd_flush - Cumulative number of flush reqs
* cmd_get - Cumulative number of retrieval reqs
* cmd_set - Cumulative number of storage reqs
* cmd_touch - Cumulative number of touch reqs
* conn_yields - Number of times any connection yielded to another due to hitting the -R limit
* connection_structures - Number of connection structures allocated by the server
* curr_connections - Number of open connections
* curr_items - Current number of items stored
* decr_hits - Number of successful decr reqs
* decr_misses - Number of decr reqs against missing keys
* delete_hits - Number of deletion reqs resulting in an item being removed
* delete_misses - umber of deletions reqs for missing keys
* evicted_active - Items evicted from LRU that had been hit recently but did not jump to top of LRU
* evicted_unfetched - Items evicted from LRU that were never touched by get/incr/append/etc
* evictions - Number of valid items removed from cache to free memory for new items
* expired_unfetched - Items pulled from LRU that were never touched by get/incr/append/etc before expiring
* get_expired - Number of items that have been requested but had already expired
* get_flushed - Number of items that have been requested but have been flushed via flush_all
* get_hits - Number of keys that have been requested and found present
* get_misses - Number of items that have been requested and not found
* hash_bytes - Bytes currently used by hash tables
* hash_is_expanding - Indicates if the hash table is being grown to a new size
* hash_power_level - Current size multiplier for hash table
* incr_hits - Number of successful incr reqs
* incr_misses - Number of incr reqs against missing keys
* limit_maxbytes - Number of bytes this server is allowed to use for storage
* listen_disabled_num - Number of times server has stopped accepting new connections (maxconns)
* max_connections - Max number of simultaneous connections
* reclaimed - Number of times an entry was stored using memory from an expired entry
* rejected_connections - Conns rejected in maxconns_fast mode
* store_no_memory - Number of rejected storage requests caused by exhaustion of the memory limit when evictions are disabled
* store_too_large - Number of rejected storage requests caused by attempting to write a value larger than the item size limit
* threads - Number of worker threads requested
* total_connections - Total number of connections opened since the server started running
* total_items - Total number of items stored since the server started
* touch_hits - Number of keys that have been touched with a new expiration time
* touch_misses - Number of items that have been touched and not found
* uptime - Number of secs since the server started

Description of gathered fields taken from [here](https://github.com/memcached/memcached/blob/master/doc/protocol.txt).

## Tags

* Memcached measurements have the following tags:
  * server (the host name from which metrics are gathered)

## Sample Queries

You can use the following query to get the average get hit and miss ratio, as well as the total average size of cached items, number of cached items and average connection counts per server.

```sql
SELECT mean(get_hits) / mean(cmd_get) as get_ratio, mean(get_misses) / mean(cmd_get) as get_misses_ratio, mean(bytes), mean(curr_items), mean(curr_connections) FROM memcached WHERE time > now() - 1h GROUP BY server
```

## Example Output

```shell
$ ./telegraf --config telegraf.conf --input-filter memcached --test
memcached,server=localhost:11211 accepting_conns=1i,auth_cmds=0i,auth_errors=0i,bytes=0i,bytes_read=7i,bytes_written=0i,cas_badval=0i,cas_hits=0i,cas_misses=0i,cmd_flush=0i,cmd_get=0i,cmd_set=0i,cmd_touch=0i,conn_yields=0i,connection_structures=3i,curr_connections=2i,curr_items=0i,decr_hits=0i,decr_misses=0i,delete_hits=0i,delete_misses=0i,evicted_active=0i,evicted_unfetched=0i,evictions=0i,expired_unfetched=0i,get_expired=0i,get_flushed=0i,get_hits=0i,get_misses=0i,hash_bytes=524288i,hash_is_expanding=0i,hash_power_level=16i,incr_hits=0i,incr_misses=0i,limit_maxbytes=67108864i,listen_disabled_num=0i,max_connections=1024i,reclaimed=0i,rejected_connections=0i,store_no_memory=0i,store_too_large=0i,threads=4i,total_connections=3i,total_items=0i,touch_hits=0i,touch_misses=0i,uptime=3i 1644771989000000000
```
