# Prometheus remote write

Converts prometheus remote write samples directly into Telegraf metrics. It can be used with [http_listener_v2](/plugins/inputs/http_listener_v2). There are no additional configuration options for Prometheus Remote Write Samples.

### Configuration

```toml
[[inputs.http_listener_v2]]
  ## Address and port to host HTTP listener on
  service_address = ":1234"

  ## Path to listen to.
  path = "/recieve"

  ## Data format to consume.
  data_format = "prometheusremotewrite"
```

### Example

**Example Input**
```
prompb.WriteRequest{
		Timeseries: []*prompb.TimeSeries{
			{
				Labels: []*prompb.Label{
					{Name: "__name__", Value: "go_gc_duration_seconds"},
					{Name: "instance", Value: "localhost:9090"},
					{Name: "job", Value: "prometheus"},
					{Name: "quantile", Value: "0.99"},
				},
				Samples: []prompb.Sample{
					{Value: 4.63, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
		},
	}

```

**Example Output**
```
prometheusremotewrite,instance=localhost:9090,job=prometheus,quantile=0.99 go_gc_duration_seconds=4.63 1614889298859000000
```
