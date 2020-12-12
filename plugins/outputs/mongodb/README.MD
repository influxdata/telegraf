# MongoDB Output Plugin

This plugin writes to a MongoDB collection.

```toml
[[outputs.mqtt]]
## URI of MongoDB servers. Requried.
uri = "mongodb://admin:123456@localhost:27017"

## mongodb database. Requried.
db  = "testing"

## mongodb collection. Requried.
collection = "numbers"

## timeout for connections, default: 10s
timeout = 10

## Max size for connection pool, default: 10
max_pool_size = 10

## Min size for connection pool, default: 5
min_pool_size = 5

## Max idle time for connections, default: 30m
max_conn_idle_time = 30

## white list fields can insert to collection. If empty, all fields can be inserted.
fields = ["total", "count"]

## Data format to output.
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
data_format = "json"
```

