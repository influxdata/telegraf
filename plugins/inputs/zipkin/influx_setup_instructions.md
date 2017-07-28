# How To Set Up InfluxDB For Work With Zipkin

## Steps
___Update___ InfluxDB to >= 1.3, in order to use the new tsi engine.

___Generate___ a config file with the following command:
    `influxd config > /path/for/config/file`

___Add___ the following to your config file, under the `[data]` tab:

```toml
[data]
    dir = "/Users/goller/.influxdb/data"
    index-version = "tsi1"
    wal-dir = "/Users/goller/.influxdb/wal"
    query-log-enabled = true
    cache-max-memory-size = 1073741824
    cache-snapshot-memory-size = 26214400
    cache-snapshot-write-cold-duration = "10m0s"
    compact-full-write-cold-duration = "4h0m0s"
    max-series-per-database = 1000000
    max-values-per-tag = 100000
    trace-logging-enabled = false
 ```

 ___Start___ `influxd` with your new config file:
 `$ influxd -config=/path/to/your/config/file`

___Update___ your retention policy:
```sql
ALTER RETENTION POLICY "autogen" ON "telegraf" DURATION 1d SHARD DURATION 30m
```
