# Dropwizard Input Plugin

The dropwizard plugin gathers metrics from a [Dropwizard](http://www.dropwizard.io/) web application via HTTP.

It also supports gathering metrics from a non-Dropwizard web application that uses the [Dropwizard Metrics](http://metrics.dropwizard.io) library and exposes the library's AdminServlet or MetricsServlet endpoint.

The plugin expects that the web application will serve a JSON representation of all registered metrics via HTTP.

This plugin is an alternate to using Dropwizard Metrics Reporter implementations like the ones from [iZettle](https://github.com/iZettle/dropwizard-metrics-influxdb) and [kickstarter](https://github.com/kickstarter/dropwizard-influxdb-reporter). The features in this plugin are inspired by the above mentioned Reporter libraries.
The main differences with this plugin compared to the Reporters are:

- you can change metrics collection configuration or certain functionality without changing or restarting the Dropwizard application. This gives Operations greater flexibility to manage the infrastructure without affecting the application.
- metrics will be pulled from the Dropwizard application rather than it pushing metrics via the Reporters. This reduces the responsibilities of the web application.


### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage dropwizard`.

```toml
# Read Dropwizard-formatted JSON metrics from one or more HTTP endpoints
[[inputs.dropwizard]]
  ## Works with Dropwizard metrics endpoint out of the box

  ## Multiple URLs from which to read Dropwizard-formatted JSON
  ## Default is "http://localhost:8081/metrics".
  urls = [
    "http://localhost:8081/metrics"
  ]

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## http request & header timeout
  ## defaults to 5s if not set
  timeout = "10s"

  ## format the floating number fields on all metrics to round them off
  ## this avoids getting very small numbers like 5.647645996854652E-23
  ## defaults to "%.2f" if not set
  #float_field_format = "%.2f"

  ## skip any metric whose "count" field hasn't changed since last time the metric was pulled
  ## this applies to metric types: Counter, Histogram, Meter & Timer
  ## defaults to false if not set
  skip_idle_metrics = true

  ## exclude some built-in metrics
  # namedrop = [
  #  "jvm.classloader*",
  #  "jvm.buffers*",
  #  "jvm.gc*",
  #  "jvm.memory.heap*",
  #  "jvm.memory.non-heap*",
  #  "jvm.memory.pools*",
  #  "jvm.threads*",
  #  "jvm.attribute.uptime",
  #  "jvm.filedescriptor",
  #  "io.dropwizard.jetty.MutableServletContextHandler*",
  #  "org.eclipse.jetty.util*"
  # ]

  ## include only the required fields (applies to all metrics types)
  # fieldpass = [
  #  "count",
  #  "max",
  #  "p999",
  #  "m5_Rate",
  #  "value"
  # ]
```

### Metrics:

The [Dropwizard Metrics](http://metrics.dropwizard.io) library supports 5 metric types and each have a fixed number of fields shown in brackets:

- Gauge (value)
- Counter (count)
- Histogram (count, max, mean, min, p50, p75, p95, p98, p99, p999, stddev)
- Meter (count, m15_rate, m1_rate, m5_rate, mean_rate)
- Timer (count, max, mean, min, p50, p75, p95, p98, p99, p999, stddev, m15_rate, m1_rate, m5_rate, mean_rate)

The metrics will include any that you have added in your application plus some built-in ones (e.g. JVM, jetty and logback). 
You can omit some of the built-in ones by using Telegraf's ```namedrop``` configuration that is available on all input plugins. An example is included in the sample config.
Any metrics that are non-numeric will be dropped.

### Sample Queries:

To plot the memory used by the Dropwizard application:
```
SELECT mean("value") AS "mean_value" FROM "telegraf"."autogen"."jvm.memory.total.used" WHERE time > now() - 5m GROUP BY time(15s) FILL(null)
```

### Example Output:

```
./telegraf --input-filter dropwizard --test
2017/11/24 07:23:32 I! Using config file: /Users/dropwizard/.telegraf/telegraf.conf
* Plugin: inputs.dropwizard, Collection 1
> jvm.memory.total.max,host=dropwizard value=1908932607i 1511468612000000000
> jvm.memory.total.used,host=dropwizard value=62960264i 1511468612000000000
> jvm.memory.total.committed,host=dropwizard value=179109888i 1511468612000000000
> jvm.memory.total.init,host=dropwizard value=136773632i 1511468612000000000
> ch.qos.logback.core.Appender.all,host=dropwizard count=16i,m5_rate=0 1511468612000000000
> ch.qos.logback.core.Appender.debug,host=dropwizard count=0i,m5_rate=0 1511468612000000000
> ch.qos.logback.core.Appender.error,host=dropwizard count=0i,m5_rate=0 1511468612000000000
> ch.qos.logback.core.Appender.info,host=dropwizard count=16i,m5_rate=0 1511468612000000000
> ch.qos.logback.core.Appender.trace,host=dropwizard count=0i,m5_rate=0 1511468612000000000
> ch.qos.logback.core.Appender.warn,host=dropwizard m5_rate=0,count=0i 1511468612000000000
> org.eclipse.jetty.server.HttpConnectionFactory.9000.connections,host=dropwizard p999=40.16,count=1i,max=40.16,m5_rate=0 1511468612000000000
> org.eclipse.jetty.server.HttpConnectionFactory.9001.connections,host=dropwizard count=6i,m5_rate=0,p999=50.01,max=235.67 1511468612000000000
```

### TODO:

Currently this plugin does the basics of pulling the metrics from Dropwizard JSON/HTTP endpoint and you can use Telegraf's built-in features to determine which metrics and fields get sent to your output. 
It would be nice to have some additional features like the following:

- Group single field metrics (i.e. gauges and counters) with a common prefix into 1 measurement with multiple fields. For example, they are 4 jvm gauge metrics with the common prefix "jvm.memory.total". Instead of the 4 single field measurements, it would create 1 measurement with 4 fields. This would help improve the resulting influxdb schema.

- Pull other information from Dropwizard's AdminServlet like Healthchecks

- Per-metric tags, these could be derived using a naming convention like "measurement.name,tag1=value1,tag2=value2"

- Metric name to Measurement name mapping (i.e. renaming). For example, could support mapping "jvm.memory.total" metrics to "jvm_memory" through configuration