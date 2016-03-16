# Telegraf Plugin: Couchbase

### Configuration:

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

### Per-node measurements

Meta:
- units: bytes
- tags: `cluster`, `hostname`

Measurement names:
- memory_free
- memory_total

### Per-bucket measurements

Meta:
- units: varies
- tags: `cluster`, `bucket`

Measurement names:
- quotaPercentUsed (unit: percent)
- opsPerSec (unit: count)
- diskFetches (unit: count)
- itemCount (unit: count)
- diskUsed (unit: bytes)
- dataUsed (unit: bytes)
- memUsed (unit: bytes)

