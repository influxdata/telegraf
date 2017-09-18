# Tile38 Input Plugin

[Tile38](http://tile38.com/) input plugin gathers metrics from any Tile38 server instance.

### Configuration:

```toml
[[inputs.tile38]]
## specify servers via a url matching:
## [:password]@address[:port]
## e.g.
##   localhost:9851
##   :password@192.168.100.100
##
## If no servers are specified, then localhost is used as the host.
## If no port is specified, 9851 is used
  servers = ["localhost:9851"]

## If true, collect stats for all keys. Default: false.
#  keys_stats = true
```

### Measurements & Fields:

There are two measurements: tile38_server and tile38_stats. 
 - tile_server is the results gathers from [SERVER](http://tile38.com/commands/server/) tile38 command.
 - tile38_stats is the results gathers from [STATS](http://tile38.com/commands/stats/) tile38 command. By default tile38_stats is not collect, unless change config option *keys_stats = true*.

- tile38_server
  - aof_size (int, number)
  - avg_item_size (int, number)
  - heap_released (int, number)
  - heap_size (int, number)
  - http_transport (int, 0 | 1)
  - in_memory_size (int, number)
  - max_heap_size (int, number)
  - mem_alloc (int, number)
  - num_collections (int, number)
  - num_hooks (int, number)
  - num_objects (int, number)
  - num_points (int, number)
  - num_strings (int, number)
  - pid (int, number)
  - pointer_size (int, number)
  - read_only (int, 0 | 1)

- tile38_stats
  - in_memory_size (int, number)
  - num_objects (int, number)
  - num_points (int, number)
  - num_strings (int, number)

###  Tags

- All measurements have the following tags:
  - server
  - port
  - id

- The has an additional *key* tag
  - key

### Example Output:

Configuration:
```
[[inputs.tile38]]
## specify servers via a url matching:
## [:password]@address[:port]
## e.g.
##   localhost:9851
##   :password@192.168.100.100
##
## If no servers are specified, then localhost is used as the host.
## If no port is specified, 9851 is used
  servers = ["localhost:9851"]

## If true, collect stats for all keys. Default: false.
  keys_stats = true
```
Run with
```
$ telegraf --config /etc/telegraf/telegraf.conf --input-filter tile38 --test
```

tile38_server:
```
* Plugin: inputs.tile38, Collection 1
> tile38_server,server=localhost,port=9851,id=955dd310a736a9cae6b93a2f8d14747c,host=host heap_size=18292408i,in_memory_size=15585622i,num_hooks=0i,pid=8712i,read_only=0i,mem_alloc=18292408i,avg_item_size=29i,heap_released=10862592i,max_heap_size=0i,pointer_size=8i,aof_size=20398395i,http_transport=1i,num_collections=9i,num_objects=2799i,num_points=625088i,num_strings=1i 1505458700000000000
```

tile38_key:
```
> tile38_key,id=955dd310a736a9cae6b93a2f8d14747c,server=localhost,port=9851,key=max,host=host in_memory_size=443552i,num_objects=14i,num_points=18358i,num_strings=0i 1505458700000000000
> tile38_key,id=955dd310a736a9cae6b93a2f8d14747c,server=localhost,port=9851,key=mid,host=host num_points=254759i,num_strings=0i,in_memory_size=6217173i,num_objects=492i 1505458700000000000
```
