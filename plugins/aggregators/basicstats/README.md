# BasicStats Aggregator Plugin

The BasicStats aggregator plugin give us count,max,min,mean,s2(variance), stdev for a set of values,
emitting the aggregate every `period` seconds.

### Configuration:

```toml
# Keep the aggregate basicstats of each metric passing through.
[[aggregators.basicstats]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
```

### Measurements & Fields:

- measurement1
    - field1_count
    - field1_max
    - field1_min
    - field1_mean
    - field1_s2 (variance)
    - field1_stdev (standard deviation)

### Tags:

No tags are applied by this aggregator.

### Example Output:

```
$ telegraf --config telegraf.conf --quiet
system,host=tars load1=1 1475583980000000000
system,host=tars load1=1 1475583990000000000
system,host=tars load1_count=2,load1_max=1,load1_min=1,load1_mean=1,load1_s2=0,load1_stdev=0 1475584010000000000
system,host=tars load1=1 1475584020000000000
system,host=tars load1=3 1475584030000000000
system,host=tars load1_count=2,load1_max=3,load1_min=1,load1_mean=2,load1_s2=2,load1_stdev=1.414162 1475584010000000000
```
