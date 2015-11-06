# Telegraf plugin: Jolokia

#### Plugin arguments:
- **context** string: Context root used of jolokia url
- **servers** []Server: List of servers
  + **name** string: Server's logical name
  + **host** string: Server's ip address or hostname
  + **port** string: Server's listening port
- **metrics** []Metric
  + **name** string: Name of the measure
  + **jmx** string: Jmx path that identifies mbeans attributes
  + **pass** []string: Attributes to retain when collecting values
  + **drop** []string: Attributes to drop when collecting values

#### Description

The Jolokia plugin collects JVM metrics exposed as MBean's attributes through jolokia REST endpoint. All metrics
are collected for each server configured.

See: https://jolokia.org/

# Measurements:
Jolokia plugin produces one measure for each metric configured, adding Server's `name`, `host` and `port` as tags.

Given a configuration like:

```ini
[jolokia]

[[jolokia.servers]]
  name = "as-service-1"
  host = "127.0.0.1"
  port = "8080"

[[jolokia.servers]]
  name = "as-service-2"
  host = "127.0.0.1"
  port = "8180"

[[jolokia.metrics]]
  name = "heap_memory_usage"
  jmx  = "/java.lang:type=Memory/HeapMemoryUsage"
  pass = ["used", "max"]
```

The collected metrics will be:

```
jolokia_heap_memory_usage name=as-service-1,host=127.0.0.1,port=8080 used=xxx,max=yyy
jolokia_heap_memory_usage name=as-service-2,host=127.0.0.1,port=8180 used=vvv,max=zzz
```
