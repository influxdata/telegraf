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
- memory_free (example: 23181365248.0)
- memory_total (example: 64424656896.0)

### Per-bucket measurements

Meta:
- units: varies
- tags: `cluster`, `bucket`

Measurement names:
- quotaPercentUsed (unit: percent, example: 68.85424936294555)
- opsPerSec (unit: count, example: 5686.789686789687)
- diskFetches (unit: count, example: 0.0)
- itemCount (unit: count, example: 943239752.0)
- diskUsed (unit: bytes, example: 409178772321.0)
- dataUsed (unit: bytes, example: 212179309111.0)
- memUsed (unit: bytes, example: 202156957464.0)

