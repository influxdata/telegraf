# F5 Load Balancer Input Plugin

The `f5_load_balancer` plugin gathers metrics from an F5 Load Balancer's iControl Rest API. Versions 15.1+ are supported, although some functionality might work on older versions. Version specific documentation can be found [here](https://clouddocs.f5.com/api/icontrol-rest/).
### Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage f5_load_balancer`.

```toml
[[inputs.f5_load_balancer]]
  ## F5 Load Balancer Username
  username = "" # required
  ## F5 Load Balancer Password
  password = "" # required
  ## F5 Load Balancer User Interface Endpoint
  url = "https://f5.example.com/" # required
  ## Metrics to collect from the F5
  collectors = ["node","virtual","pool","net_interface"]
```

### Metrics

- node
  - tags:
    - name
  - fields:
    - node_current_sessions
    - node_serverside_bits_in
    - node_serverside_bits_out
    - node_serverside_current_connections
    - node_serverside_packets_in
    - node_serverside_packets_out
    - node_serverside_total_connections
    - node_total_requests

+ virtual
  - tags:
    - name
  - fields:
    - virtual_clientside_bits_in
    - virtual_clientside_bits_out
    - virtual_clientside_current_connections
    - virtual_clientside_packets_in
    - virtual_clientside_packets_out
    - virtual_total_requests
    - virtual_one_minute_avg_usage
    - virtual_available

- pool
  - tags:
    - name
  - fields:
    - pool_active_member_count
    - pool_current_sessions
    - pool_serverside_bits_in
    - pool_serverside_bits_out
    - pool_serverside_current_connections
    - pool_serverside_packets_in
    - pool_serverside_packets_out
    - pool_serverside_total_connections
    - pool_total_requests
    - pool_available

- net_interface
  - tags:
    - name
  - fields:
    - net_interface_counter_bits_in
    - net_interface_counter_bits_out
    - net_interface_counter_packets_in
    - net_interface_counter_packets_out
    - net_interface_status