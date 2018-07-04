# BasicStats Aggregator Plugin

The BasicStats aggregator plugin give us count,max,min,mean,sum,s2(variance), stdev for a set of values,
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

  ## BasicStats Arguments:

  ## Configures which basic stats to push as fields
  stats = ["count","min","max","mean","stdev","s2","sum"]
```

- stats
    - If not specified, then `count`, `min`, `max`, `mean`, `stdev`, and `s2` are aggregated and pushed as fields.  `sum` is not aggregated by default to maintain backwards compatibility.
    - If empty array, no stats are aggregated

### Measurements & Fields:

- measurement1
    - field1_count
    - field1_max
    - field1_min
    - field1_mean
    - field1_sum
    - field1_s2 (variance)
    - field1_stdev (standard deviation)

### Tags:

No tags are applied by this aggregator.

### Example Output:

```
$ telegraf --config telegraf.conf --quiet
system,host=tars load1=1 1475583980000000000
system,host=tars load1=1 1475583990000000000
system,host=tars load1_count=2,load1_max=1,load1_min=1,load1_mean=1,load1_sum=2,load1_s2=0,load1_stdev=0 1475584010000000000
system,host=tars load1=1 1475584020000000000
system,host=tars load1=3 1475584030000000000
system,host=tars load1_count=2,load1_max=3,load1_min=1,load1_mean=2,load1_sum=4,load1_s2=2,load1_stdev=1.414162 1475584010000000000
```


# CoStats Aggregator Extension

The CoStats aggregator extension gives the ability to add optional covariance and correlation stats e.g. covariance and correlation, every `period` seconds. Predetermined metric name and specific fields need to be selected for pairing and calculating the correlation and covariance.

### Configuration:

```toml
  ## Keep the aggregate costats for metric passing through (pair1 for costat).
  ##[[aggregators.basicstats.costat]]

  ### Metric Info Arguments: The Metric Info about the pair of metrics for which the covariance and correlation stats are required.
  ###[[aggregators.basicstats.costat.metrics]]
  #### Configures 1st metric name and field for covariance and correlation
  ####name = "measurementName1"
  ####field = "fieldName1"
  ###[[aggregators.basicstats.costat.metrics]]
  #### Configures 2nd metric name and field for covariance and correlation
  ####name = "measurementName2"
  ####field = "fieldName2"

  ## Keep the aggregate costats for metric passing through (another pair for costat).
  ##[[aggregators.basicstats.costat]]

  ### Metric Info Arguments: The Metric Info about the pair of metrics for which the covariance and correlation stats are required.
  ###[[aggregators.basicstats.costat.metrics]]
  #### Configures 1st metric name and field for covariance and correlation
  ####name = "measurementName3"
  ####field = "fieldName3"
  ###[[aggregators.basicstats.costat.metrics]]
  #### Configures 2nd metric name and field for covariance and correlation
  ####name = "measurementName4"
  ####field = "fieldName4"
```

### Measurements & Fields:

covariance
correlation

Note, any tags as applicable are appended uniquely to the covariance and correlation measurements to identify the unique pair of fields.

E.g.
covariance[measurementName1/fieldName1/<tag1_1>:<val1_1>/<tag1_2>:<val1_2>/...][measurementName2/fieldName2/<tag2_1>:<val2_1>/<tag2_2>:<val2_2>/...]
correlation[measurementName1/fieldName1/<tag1_1>:<val1_1>/<tag1_2>:<val1_2>/...][measurementName2/fieldName2/<tag2_1>:<val2_1>/<tag2_2>:<val2_2>/...]
covariance[measurementName2/fieldName2/<tag2_1>:<val2_1>/<tag2_2>:<val2_2>/...][measurementName1/fieldName1/<tag1_1>:<val1_1>/<tag1_2>:<val1_2>/...]
correlation[measurementName2/fieldName2/<tag2_1>:<val2_1>/<tag2_2>:<val2_2>/...][measurementName1/fieldName1/<tag1_1>:<val1_1>/<tag1_2>:<val1_2>/...]

covariance[measurementName3/fieldName3/<tag3_1>:<val3_1>/<tag3_2>:<val3_2>/...][measurementName4/fieldName4/<tag4_1>:<val4_1>/<tag4_2>:<val4_2>/...]
correlation[measurementName3/fieldName3/<tag3_1>:<val3_1>/<tag3_2>:<val3_2>/...][measurementName4/fieldName4/<tag4_1>:<val4_1>/<tag4_2>:<val4_2>/...]
covariance[measurementName4/fieldName4/<tag4_1>:<val4_1>/<tag4_2>:<val4_2>/...][measurementName3/fieldName3/<tag3_1>:<val3_1>/<tag3_2>:<val3_2>/...]
correlation[measurementName4/fieldName4/<tag4_1>:<val4_1>/<tag4_2>:<val4_2>/...][measurementName3/fieldName3/<tag3_1>:<val3_1>/<tag3_2>:<val3_2>/...]


### Tags:

No new tags are applied by this aggregator presently.

### Example Output:

Example configuration:
```
[[aggregators.basicstats.costat]]
 [[aggregators.basicstats.costat.metrics]]
   name = "cpu"
   field = "usage_system"
 [[aggregators.basicstats.costat.metrics]]
   name = "cpu"
   field = "usage_user"

```

```
$ telegraf --config telegraf.conf --input-filter cpu:mem --quiet

covariance[cpu/usage_system/cpu:cpu-total/host:hostName][cpu/usage_user/cpu:cpu-total/host:hostName]=-0.02373428406780692,
correlation[cpu/usage_user/cpu:cpu-total/host:hostName][cpu/usage_system/cpu:cpu-total/host:hostName]=1.0714185814230373,
covariance[cpu/usage_user/cpu:cpu-total/host:hostName][cpu/usage_system/cpu:cpu-total/host:hostName]=0.1361977595801846,
correlation[cpu/usage_system/cpu:cpu-total/host:hostName][cpu/usage_user/cpu:cpu-total/host:hostName]=-0.18670904018835857
```
