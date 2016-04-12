# Telegraf plugin: Jolokia

#### Configuration

```toml
[[inputs.jolokia]]
  ## This is the context root used to compose the jolokia url
  context = "/jolokia/read"

  ## List of servers exposing jolokia read service
  [[inputs.jolokia.servers]]
    name = "stable"
    host = "192.168.103.2"
    port = "8180"
    # username = "myuser"
    # password = "mypassword"

  ## List of metrics collected on above servers
  ## Each metric consists in a name, a jmx path and either
  ## a pass or drop slice attribute.
  ## This collect all heap memory usage metrics.
  [[inputs.jolokia.metrics]]
    name = "heap_memory_usage"
    jmx  = "/java.lang:type=Memory/HeapMemoryUsage"
    
  ## This collect thread counts metrics.
  [[inputs.jolokia.metrics]]
    name = "thread_count"
    jmx  = "/java.lang:type=Threading/TotalStartedThreadCount,ThreadCount,DaemonThreadCount,PeakThreadCount"
 
  ## This collect number of class loaded/unloaded counts metrics.
  [[inputs.jolokia.metrics]]
    name = "class_count"
    jmx  = "/java.lang:type=ClassLoading/LoadedClassCount,UnloadedClassCount,TotalLoadedClassCount"
```

#### Description

The Jolokia plugin collects JVM metrics exposed as MBean's attributes through jolokia REST endpoint. All metrics
are collected for each server configured.

See: https://jolokia.org/

# Measurements:
Jolokia plugin produces one measure for each metric configured, adding Server's `name`, `host` and `port` as tags.
