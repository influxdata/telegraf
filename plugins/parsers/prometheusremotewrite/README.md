# Prometheus remote write

Converts prometheus remote write metrics directly into Telegraf metrics. It can be used in [http_listener_v2](/plugins/inputs/http_listener_v2).


### Configuration

```toml
[[inputs.http_listener_v2]]
  ## Address and port to host HTTP listener on
  service_address = ":1234"

  ## Path to listen to.
  # path = "/recieve"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "prometheusremotewrite"
```

### Metrics

A Prometheus metric is created for each integer, float, boolean or unsigned
field.  Boolean values are converted to *1.0* for true and *0.0* for false.

The Prometheus metric names are produced by joining the measurement name with
the field key.  In the special case where the measurement name is `prometheus`
it is not included in the final metric name.

Prometheus labels are produced for each tag.

**Note:** String fields are ignored and do not produce Prometheus metrics.

### Example

**Example Input**
```
{[
    labels:<name:"__name__" value:"go_gc_duration_seconds" > labels:<name:"instance" value:"localhost:9090" > labels:<name:"job" value:"prometheus" > labels:<name:"quantile" value:"0" > samples:<value:4.262e-05 timestamp:1614889298859 >  
    labels:<name:"__name__" value:"go_gc_duration_seconds" > labels:<name:"instance" value:"localhost:9090" > labels:<name:"job" value:"prometheus" > labels:<name:"quantile" value:"0.25" > samples:<value:4.6339e-05 timestamp:1614889298859 >  
]}`

```

**Example Output**
```
prometheusremotewrite,instance=localhost:9090,job=prometheus,quantile=0 go_gc_duration_seconds=0.00004262 1614889298859000000
prometheusremotewrite,instance=localhost:9090,job=prometheus,quantile=0.25 go_gc_duration_seconds=0.000046339 1614889298859000000
```
