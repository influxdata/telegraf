# Jolokia2 Proxy input plugin

The `jolokia2_proxy` input plugin reads JMX metrics from one or more _targets_ by interacting with a [Jolokia proxy](https://jolokia.org/features/proxy.html) REST endpoint.

## Configuration

```toml
# Read JMX metrics from a Jolokia REST agent endpoint
[[inputs.jolokia2_agent]]
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."

  # Add agents URLs to query
  urls = ["http://localhost:8080/jolokia"]
  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Add metrics to read
  [[inputs.jolokia2_agent.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]
```
