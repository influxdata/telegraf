# EXTR

The `extr` output data format converts metrics into JSON documents, combining those sequential metrics matching name, tags, and timestamps into a single JSON metric, combining the fields of each metric into an array of fields.

### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "extr"
  use_batch_format = true

  ## The resolution to use for the metric timestamp.  Must be a duration string
  ## such as "1ns", "1us", "1ms", "10ms", "1s".  Durations are truncated to
  ## the power of 10 less than the specified units.
  json_timestamp_units = "1s"


[[outputs.http]]
  url = "http://10.139.101.72:9443/telegraf/rest/v1"
  method = "POST"

  data_format = "extr"
  flush_interval = "2s"

  [outputs.http.headers]
  Content-Type = "application/json; charset=utf-8"

```

### Examples:

The following Telegraf batched metrics
   
```text
StatsCpu,node=NODE1  cpu=0,min=20,max=30,avg=25,interval=1,samplePeriod=10 1556813561098000000
StatsCpu,node=NODE1  cpu=1,min=31,max=42,avg=76,interval=1,samplePeriod=10 1556813561098000000
StatsCpu,node=NODE1  cpu=2,min=22,max=52,avg=11,interval=1,samplePeriod=10 1556813561098000000
EventInterfaceStatus,node=NODE2  ifIndex="1001",port="1:1",adminStatus=1,operStatus=1 1557813561098000000
EventInterfaceStatus,node=NODE2  ifIndex="1002",port="1:2",adminStatus=0,operStatus=0 1557813561098000000
```

will serialize into the following extr JSON ouput
   
```json
[{
   "fields": [
      {"avg":25,"cpu":0,"interval":1,"max":30,"min":20,"samplePeriod":10},
      {"avg":76,"cpu":1,"interval":1,"max":42,"min":31,"samplePeriod":10},
      {"avg":11,"cpu":2,"interval":1,"max":52,"min":22,"samplePeriod":10}
   ],
   "name":"StatsCpu",
   "tags":{"node":"NODE1"},
   "timestamp":1556813561
},
{
   "fields":[
      {"adminStatus":1,"ifIndex":"1001","operStatus":1,"port":"1:1"},
      {"adminStatus":0,"ifIndex":"1002","operStatus":0,"port":"1:2"}],
   "name":"EventInterfaceStatus",
   "tags":{"node":"NODE2"},
   "timestamp":1557899561}]
```
