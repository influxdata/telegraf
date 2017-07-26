# Telegraf Plugin: Couchbase

## Configuration:

```
# Read per-node and per-bucket metrics from Couchbase
[[inputs.couchbase]]
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    http://couchbase-0.example.com/
  ##    http://admin:secret@couchbase-0.example.com:8091/
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no protocol is specifed, HTTP is used.
  ## If no port is specified, 8091 is used.
  servers = ["http://localhost:8091"]
```

## Measurements:

### couchbase_node

Tags:
- cluster: sanitized string from `servers` configuration field e.g.: `http://user:password@couchbase-0.example.com:8091/endpoint` -> `http://couchbase-0.example.com:8091/endpoint`
- hostname: Couchbase's name for the node and port, e.g., `172.16.10.187:8091`

Fields:
- memory_free (unit: bytes, example: 23181365248.0)
- memory_total (unit: bytes, example: 64424656896.0)

### couchbase_bucket

Tags:
- cluster: whatever you called it in `servers` in the configuration, e.g.: `http://couchbase-0.example.com/`)
- bucket: the name of the couchbase bucket, e.g., `blastro-df`

Fields:
- quota_percent_used (unit: percent, example: 68.85424936294555)
- ops_per_sec (unit: count, example: 5686.789686789687)
- disk_fetches (unit: count, example: 0.0)
- item_count (unit: count, example: 943239752.0)
- disk_used (unit: bytes, example: 409178772321.0)
- data_used (unit: bytes, example: 212179309111.0)
- mem_used (unit: bytes, example: 202156957464.0)


## Example output

```
$ telegraf --config telegraf.conf --input-filter couchbase --test
* Plugin: couchbase, Collection 1
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.10.187:8091 memory_free=22927384576,memory_total=64424656896 1458381183695864929
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.10.65:8091 memory_free=23520161792,memory_total=64424656896 1458381183695972112
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.13.105:8091 memory_free=23531704320,memory_total=64424656896 1458381183695995259
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.13.173:8091 memory_free=23628767232,memory_total=64424656896 1458381183696010870
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.15.120:8091 memory_free=23616692224,memory_total=64424656896 1458381183696027406
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.8.127:8091 memory_free=23431770112,memory_total=64424656896 1458381183696041040
> couchbase_node,cluster=https://couchbase-0.example.com/,hostname=172.16.8.148:8091 memory_free=23811371008,memory_total=64424656896 1458381183696059060
> couchbase_bucket,bucket=default,cluster=https://couchbase-0.example.com/ data_used=25743360,disk_fetches=0,disk_used=31744886,item_count=0,mem_used=77729224,ops_per_sec=0,quota_percent_used=10.58976636614118 1458381183696210074
> couchbase_bucket,bucket=demoncat,cluster=https://couchbase-0.example.com/ data_used=38157584951,disk_fetches=0,disk_used=62730302441,item_count=14662532,mem_used=24015304256,ops_per_sec=1207.753207753208,quota_percent_used=79.87855353525707 1458381183696242695
> couchbase_bucket,bucket=blastro-df,cluster=https://couchbase-0.example.com/ data_used=212552491622,disk_fetches=0,disk_used=413323157621,item_count=944655680,mem_used=202421103760,ops_per_sec=1692.176692176692,quota_percent_used=68.9442170551845 1458381183696272206
```
