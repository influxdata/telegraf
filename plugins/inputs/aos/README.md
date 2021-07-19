# Apstra AOS Input Plugin
The [AOS](https://www.apstra.com/products/aos-overview/) input plugin parses
protobuf messages sent by the AOS server and formats them as time series 
database metrics, e.g., Influxdb. These messages contain telemetry data
that AOS generates and consist of performance monitoring (perfmon), alerts and events.

The plugin configures the telegraf instance as a protobuf streaming endpoint to
receive either perfmon, alert or event messages, or any combination thereof.
The message data is also augmented with information collected from the AOS server
via the REST API.

## Configuration
```
[[inputs.aos]]
  # TCP Port to listen for incoming sessions from the AOS Server.
  port = 7777

  # Address of the server running Telegraf, it needs to be reachable from AOS.
  address = "192.168.59.1"

  # Interval to refresh content from the AOS server (in sec).
  # This is no longer needed as of AOS 3.2.
  # refresh_interval = 30

  # Streaming Type Can be "perfmon", "alerts" or "events".
  streaming_type = [ "perfmon", "alerts" ]

  # Define parameters to configure the AOS Server using the REST API.
  aos_server = "192.168.59.250"
  aos_port = 443
  aos_login = "admin"
  aos_password = "admin"
  aos_protocol = "https"

```

## Metrics
Data streamed from AOS is of three different types: perfmon, alerts, events. 
### perfmon metrics
The fields included in the interface_counters metric are:  
* tx_bytes
* tx_unicast_packets
* tx_broadcast_packets
* tx_multicast_packets
* tx_error_packets
* tx_discard_packets
* tx_bps
* tx_unicast_pps
* tx_broadcast_pps
* tx_multicast_pps
* tx_error_pps
* tx_discard_pps
* rx_bytes
* rx_unicast_packets
* rx_broadcast_packets
* rx_multicast_packets
* rx_error_packets
* rx_discard_packets
* rx_bps
* rx_unicast_pps
* rx_broadcast_pps
* rx_multicast_pps
* rx_error_pps
* rx_discard_pps
* alignment_errors
* fcs_errors
* symbol_errors
* runts
* giants
* delta_seconds

The fields included in the system_info metric are:  
* cpu_idle
* cpu_system
* cpu_user
* memory_total
* memory_used

The fields included the process_info metric are:  
* cpu_system
* cpu_user
* memory_used

The fields included in the file_info metric are:
* file_size

The fields included in the probe_message metric are:
* int64 
* float
* string
* EvpnType3RouteEvent
* EvpnType5RouteEvent
* InterfaceCountersUtilization
* SystemInterfaceUtilization
* ActiveFloodlistEvent

### alerts
Alert metrics include a value of 0 or 1.  
1 - indicates that the alert is active.  
0 - indicates that the alert has cleared.  
AOS provides the following alert metrics:
* config_deviation_alert
* streaming_alert
* cable_peer_mismatch_alert
* bgp_neighbor_mismatch_alert
* interface_link_status_mismatch_alert
* hostname_alert
* route_alert
* liveness_alert
* deployment_alert
* blueprint_rendering_alert
* counters_alert
* mac_alert
* arp_alert
* headroom_alert
* lag_alert
* mlag_alert
* probe_alert
* config_mismatch_alert
* extensible_alert

### events
Event metrics include:
* device_state
* streaming
* cable_peer
* bgp_neighbor
* link_status
* traffic
* mac_state
* arp_state
* lag_state
* mlag_state
* extensible_event
* route_state

### message_loss
When using AOS in sequenced mode, each streamed message will include a sequence number, which allows the plugin to generate loss metrics for each type:
* message_loss_perfmon
* message_loss_alert
* message_loss_event

## Metric Buffer Limit Recommendations
Receiving a high rate of telemetry messages can cause metric buffer drops if the metric_buffer_limit is too low. Testing with both Prometheus and InfluxDB as output the plugin, we've found that setting the metric_buffer_limit to 35000 eliminates buffer drops with an input stream of 2000 perfmon messages/sec.

## Support 
This plugin supports AOS up to version 4.0.0.
