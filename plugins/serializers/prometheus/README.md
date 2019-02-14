# Prometheus

The `prmetheus` serializer translates the Telegraf metric format to the [prometheus format](https://prometheus.io/docs/concepts/data_model/).

### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "prometheus"
```

### Example

If we take the following InfluxDB Line Protocol:

```
disk,device=vda3,fstype=ext4,host=localhost,mode=rw,path=/var/lib/docker/overlay used_percent=48.903132176204835
disk,device=vda3,fstype=ext4,host=localhost,mode=rw,path=/var/lib/docker/plugins used_percent=48.903132176204835
disk,device=vda1,fstype=ext4,host=localhost,mode=rw,path=/boot used_percent=11.304095488859774
disk,device=vda3,fstype=ext4,host=localhost,mode=rw,path=/ used=95105536000i,used_percent=48.903132176204835

```

after serializing in Prometheus, the result would be:

```
# HELP disk_used_percent Telegraf collected metric
# TYPE disk_used_percent gauge
disk_used_percent{mode="rw",path="/var/lib/docker/overlay",device="vda3",fstype="ext4",host="localhost"} 48.87792147200521
disk_used_percent{device="vda3",fstype="ext4",host="localhost",mode="rw",path="/"} 48.87790462274593
disk_used_percent{fstype="ext4",host="localhost",mode="rw",path="/boot",device="vda1"} 11.304095488859774
disk_used_percent{mode="rw",path="/var/lib/docker/plugins",device="vda3",fstype="ext4",host="localhost"} 48.87791725969039
```

### Fields and Tags with spaces
When a field key or tag key/value have spaces, spaces will be replaced with `_`.
