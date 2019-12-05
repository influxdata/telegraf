# Breaker processor Plugin

This processor could drop metrics based on the value of another selected (flag) metric.

The parameters ``name`` y ``field`` are used to select with metric will be used to define the status of the breaker.

``value_enable`` y ``value_disable`` define which values from the control metric will be used to enable or disable the breaker.

Breaker is "active" means it will block all metrics passing through it.

``enabled`` is the state of the breaker at startup.

The metric used se the "flag" will pass also.

Normally this processor will be used with some filter to select which metrics should be affected.

### Configuration:

```toml
[[processors.breaker]]
  # By default is false, meaning it will block
  enabled = false

  # Select which metric should be used to define the state of the breaker
  name = "flag_metric"
  field = "value"

  # Which values of the selected metric will be used to enable or disable the breaker
  # These values will be transformed to string to be compared with metrics values (also converted to strings)
  value_enable = "foo"
  value_disable = "bar"
```

Example configuration dropping metrics if host does not have configured the IP resolved by a domain.

The idea is two servers in a active-pasive cluster configuration, where only the active one should send metrics.
```toml
[[inputs.exec]]
  # This command resolve the domain of the cluster VIP (simulated with that .xip.io domain) and check if it configured
  # in the host.
  # It returns "active" if the IP is configured, "pasive" otherwise.
  commands = [
    '''/bin/sh -c "ip -4 -o a | egrep \"[^0-9]$(host 127.0.0.1.xip.io | head -1 | cut -d \  -f 4)[^0-9]\" >& /dev/null && echo active || echo pasive"'''
  ]
  timeout = "5s"
  data_format = "value"
  data_type = "string"
  name_override = "node_state"

# This generate a metric with the hostname of the active host.
# Is an example of the metrics that should pass, or not, the breaker.
# The idea is that both servers, active and passive, send the same metric
[[inputs.exec]]
  commands = [ "hostname" ]
  timeout = "5s"
  data_format = "value"
  data_type = "string"
  name_override = "active_node"
  [inputs.exec.tags]
    host = "virtualHost"

# B
[[processors.breaker]]
  # By default drop metrics, to avoid the race condition at the beggining between
  # the flag and the rest of the metrics
  enabled = true

  # Here we select with metric should be used as the flag
  name = "node_state"
  field = "value"
  # And which value of the flag metric will enable/disable the breaker
  value_enable = "pasive"
  value_disable = "active"
```
