# Netstat connections Input Plugin

This plugin collects TCP port incoming/outgoing connections state and UDP socket counts by using `lsof`, by default outgoing connections are disabled for cardinality generated to influxdb but is recomended to enable for critical hosts.

### Basic example configuration:

``` toml
# Collect TCP port incoming/outgoing connections state and UDP socket counts
[[inputs.netstat_connections]]
  remote_connections = true # Enable collect outgoing connections (default: false)
```

### Advanced example configuration (use):
In this example collect TCP connections every 20s and store in specific influxdb database
``` toml
# Create /etc/telegraf.d/netstat_connections.conf for modularity
[[inputs.netstat_connections]]
  interval = "20s" # Change default agent interval 
  remote_connections = true # Enable collect outgoing connections (default: false)
  [inputs.netstat_connections.tagdrop]
   addr = ["127.0.0.1"] #Don't collect local connections (127.0.0.1)
  [inputs.netstat_connections.tags]
    influx_routing = "netstat_connections" # Add tag for routing to special influxdb database
[[outputs.influxdb]]
  urls = ["http://127.0.0.1:8060"]
  database = "NETSTAT" #
  tagexclude = ["influx_routing"]
  tagpass = { influx_routing = ["netstat_connections"] } #Only send metrics with tag influx_routing = "netstat_connections"
  timeout = "5s"
```

Please don't forget drop metrics in /etc/telegraf/telegraf.conf (by default telegraf sends to any output influxdb databases). Exmaple:
``` toml
###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

# Configuration for influxdb server to send metrics to
[[outputs.influxdb]]
  #Drop influx_routing metrics
  tagdrop = { influx_routing = ["*"] }
```

# Measurements:

for incoming connections (connections to listen ports): netstat_incoming

for outgoing connections (connections from host to external IPs): netstat_outgoing

Supported TCP Connection states are follows.

- established
- syn_sent
- syn_recv
- fin_wait1
- fin_wait2
- time_wait
- close
- close_wait
- last_ack
- listen
- closing
- none

### TCP Connection State measurements:

Meta:
- units: counts

Measurement names:
- tcp_established
- tcp_syn_sent
- tcp_syn_recv
- tcp_fin_wait1
- tcp_fin_wait2
- tcp_time_wait
- tcp_close
- tcp_close_wait
- tcp_last_ack
- tcp_listen
- tcp_closing
- tcp_none

If there are no connection on the state, the metric is 0.

### UDP socket counts measurements:

Meta:
- units: counts

Measurement names:
- udp_socket
