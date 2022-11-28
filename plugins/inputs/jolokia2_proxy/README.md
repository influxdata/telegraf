# Jolokia2 Proxy Input Plugin

The `jolokia2_proxy` input plugin reads JMX metrics from one or more _targets_
by interacting with a [Jolokia proxy](https://jolokia.org/features/proxy.html)
REST endpoint.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Read JMX metrics from a Jolokia REST proxy endpoint
[[inputs.jolokia2_proxy]]
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."

  ## Proxy agent
  url = "http://localhost:8080/jolokia"
  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional origin URL to include as a header in the request. Some endpoints
  ## may reject an empty origin.
  # origin = ""

  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Add proxy targets to query
  # default_target_username = ""
  # default_target_password = ""
  [[inputs.jolokia2_proxy.target]]
    url = "service:jmx:rmi:///jndi/rmi://targethost:9999/jmxrmi"
    # username = ""
    # password = ""

  ## Add metrics to read
  [[inputs.jolokia2_proxy.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]
```

Optionally, specify TLS options for communicating with proxies:

```toml
[[inputs.jolokia2_proxy]]
  url = "https://proxy:8080/jolokia"

  tls_ca   = "/var/private/ca.pem"
  tls_cert = "/var/private/client.pem"
  tls_key  = "/var/private/client-key.pem"
  #insecure_skip_verify = false

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

### Metric Configuration

Please see
[Jolokia agent documentation](../jolokia2_agent/README.md#metric-configuration).

## Metrics

The metrics depend on the definition(s) in the `inputs.jolokia2_proxy.metric`
section(s).

## Example Output

```text
jvm_memory_pool,pool_name=Compressed\ Class\ Space PeakUsage.max=1073741824,PeakUsage.committed=3145728,PeakUsage.init=0,Usage.committed=3145728,Usage.init=0,PeakUsage.used=3017976,Usage.max=1073741824,Usage.used=3017976 1503764025000000000
jvm_memory_pool,pool_name=Code\ Cache PeakUsage.init=2555904,PeakUsage.committed=6291456,Usage.committed=6291456,PeakUsage.used=6202752,PeakUsage.max=251658240,Usage.used=6210368,Usage.max=251658240,Usage.init=2555904 1503764025000000000
jvm_memory_pool,pool_name=G1\ Eden\ Space CollectionUsage.max=-1,PeakUsage.committed=56623104,PeakUsage.init=56623104,PeakUsage.used=53477376,Usage.max=-1,Usage.committed=49283072,Usage.used=19922944,CollectionUsage.committed=49283072,CollectionUsage.init=56623104,CollectionUsage.used=0,PeakUsage.max=-1,Usage.init=56623104 1503764025000000000
jvm_memory_pool,pool_name=G1\ Old\ Gen CollectionUsage.max=1073741824,CollectionUsage.committed=0,PeakUsage.max=1073741824,PeakUsage.committed=1017118720,PeakUsage.init=1017118720,PeakUsage.used=137032208,Usage.max=1073741824,CollectionUsage.init=1017118720,Usage.committed=1017118720,Usage.init=1017118720,Usage.used=134708752,CollectionUsage.used=0 1503764025000000000
jvm_memory_pool,pool_name=G1\ Survivor\ Space Usage.max=-1,Usage.init=0,CollectionUsage.max=-1,CollectionUsage.committed=7340032,CollectionUsage.used=7340032,PeakUsage.committed=7340032,Usage.committed=7340032,Usage.used=7340032,CollectionUsage.init=0,PeakUsage.max=-1,PeakUsage.init=0,PeakUsage.used=7340032 1503764025000000000
jvm_memory_pool,pool_name=Metaspace PeakUsage.init=0,PeakUsage.used=21852224,PeakUsage.max=-1,Usage.max=-1,Usage.committed=22282240,Usage.init=0,Usage.used=21852224,PeakUsage.committed=22282240 1503764025000000000
```
