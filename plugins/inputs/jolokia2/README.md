# Jolokia2 Input Plugins

The [Jolokia](http://jolokia.org) _agent_ and _proxy_ input plugins collect JMX metrics from the Jolokia REST endpoint using the [JSON-over-HTTP protocol](https://jolokia.org/reference/html/protocol.html).

## Jolokia Agent Configuration

The `jolokia2_agent` input plugin reads JMX metrics from a [Jolokia agent](https://jolokia.org/agent/jvm.html) REST endpoint.

```toml
[[inputs.jolokia2_agent]]
  # default_field_separator = "."
  # default_field_prefix    = ""
  # default_tag_prefix      = ""

  # Add agents URLs to query
  urls = ["http://agent:8080/jolokia"]

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]
```

Specify SSL options for communicating with agents:

```toml
[[inputs.jolokia2_agent]]
  urls = ["https://agent:8080/jolokia"]
  ssl_ca   = "/var/private/ca.pem"
  ssl_cert = "/var/private/client.pem"
  ssl_key  = "/var/private/client-key.pem"
  #insecure_skip_verify = false
```

## Jolokia Proxy Configuration

To interact with JMX targets via a [Jolokia proxy](https://jolokia.org/features/proxy.html) instance, use the `jolokia2_proxy` input plugin.

```toml
[[inputs.jolokia2_proxy]]
  # default_field_separator = "."
  # default_field_prefix    = ""
  # default_tag_prefix      = ""

  url = "http://proxy:8080/jolokia"

  #default_target_username = ""
  #default_target_password = ""
  [[inputs.jolokia2_proxy.target]]
    url = "service:jmx:rmi:///jndi/rmi://targethost:9999/jmxrmi"
    # username = ""
    # password = ""

  [[inputs.jolokia2_proxy.metric]]
    name  = "jvm_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]
```

Specify SSL options for communicating with proxies:

```toml
[[inputs.jolokia2_proxy]]
  url = "https://proxy:8080/jolokia"

  ssl_ca   = "/var/private/ca.pem"
  ssl_cert = "/var/private/client.pem"
  ssl_key  = "/var/private/client-key.pem"
  #insecure_skip_verify = false
```

## Jolokia Metric Configuration

Each `metric` declaration generates a Jolokia request to fetch telemetry from a JMX MBean.

| Key            | Required | Description |
|----------------|----------|-------------|
| `mbean`        | yes      | The object name of a JMX MBean. MBean property-key values can contain a wildcard `*`, allowing you to fetch multiple MBeans with one declaration. |
| `paths`        | no       | A list of MBean attributes to read. |
| `tag_keys`     | no       | A list of MBean property-key names to convert into tags. The property-key name becomes the tag name, while the property-key value becomes the tag value. |
| `tag_prefix`   | no       | A string to prepend to the tag names produced by this `metric` declaration. |
| `field_name`   | no       | A string to set as the name of the field produced by this metric; can contain substitutions. |
| `field_prefix` | no       | A string to prepend to the field names produced by this `metric` declaration; can contain substitutions. |

Use `paths` to refine which fields to collect. The following Jolokia `metric` declaration produces the following metric: `jvm_memory HeapMemoryUsage.committed=4294967296,HeapMemoryUsage.init=4294967296,HeapMemoryUsage.max=4294967296,HeapMemoryUsage.used=1750658992,NonHeapMemoryUsage.committed=67350528,NonHeapMemoryUsage.init=2555904,NonHeapMemoryUsage.max=-1,NonHeapMemoryUsage.used=65821352,ObjectPendingFinalizationCount=0`

```toml
[[inputs.jolokia2_agent.metric]]
  name  = "jvm_memory"
  mbean = "java.lang:type=Memory"
  paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]
```

Use `*` wildcards against `mbean` property-key values to create distinct series by capturing values into `tag_keys`.

```toml
[[inputs.jolokia2_agent.metric]]
  name     = "jvm_garbage_collector"
  mbean    = "java.lang:name=*,type=GarbageCollector"
  paths    = ["CollectionTime", "CollectionCount"]
  tag_keys = ["name"]
```

Use `tag_prefix` along with `tag_keys` to add detail to tag names.

```toml
[[inputs.jolokia2_agent.metric]]
  name       = "jvm_memory_pool"
  mbean      = "java.lang:name=*,type=MemoryPool"
  paths      = ["Usage", "PeakUsage", "CollectionUsage"]
  tag_keys   = ["name"]
  tag_prefix = "pool"
```

Use substitutions to create fields and field prefixes with MBean property-keys captured by wildcards. In the following case, `$1` represents the value of the property-key `name`, and `$2` represents the value of the property-key `topic`.

```toml
[[inputs.jolokia2_agent.metric]]
  name         = "kafka_topic"
  mbean        = "kafka.server:name=*,topic=*,type=BrokerTopicMetrics"
  field_prefix = "$1"
  tag_keys     = ["topic"]
```

In this next case, a combination of tagging and field-naming allows us to coalesce many MBeans together into a single metric.

```toml
[[inputs.jolokia2_agent.metric]]
  name       = "kafka_partition"
  mbean      = "kafka.log:name=*,partition=*,topic=*,type=Log"
  field_name = "$1"
  tag_keys = ["topic", "partition"]
```
