# Arista LANZ Consumer Input Plugin

This plugin provides a consumer for use with Arista Networks’ Latency Analyzer
(LANZ)

Metrics are read from a stream of data via TCP through port 50001 on the
switches management IP. The data is in Protobuffers format. For more information
on Arista LANZ

- <https://www.arista.com/en/um-eos/eos-latency-analyzer-lanz>

This plugin uses Arista's sdk.

- <https://github.com/aristanetworks/goarista>

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics off Arista LANZ, via socket
[[inputs.lanz]]
  ## URL to Arista LANZ endpoint
  servers = [
    "tcp://switch1.int.example.com:50001",
    "tcp://switch2.int.example.com:50001",
  ]
```

You will need to configure LANZ and enable streaming LANZ data.

- <https://www.arista.com/en/um-eos/eos-section-44-3-configuring-lanz>
- <https://www.arista.com/en/um-eos/eos-section-44-3-configuring-lanz#ww1149292>

## Metrics

For more details on the metrics see
<https://github.com/aristanetworks/goarista/blob/master/lanz/proto/lanz.proto>

- lanz_congestion_record:
  - tags:
    - intf_name
    - switch_id
    - port_id
    - entry_type
    - traffic_class
    - fabric_peer_intf_name
    - source
    - port
  - fields:
    - timestamp        (integer)
    - queue_size       (integer)
    - time_of_max_qlen (integer)
    - tx_latency       (integer)
    - q_drop_count     (integer)

- lanz_global_buffer_usage_record
  - tags:
    - entry_type
    - source
    - port
  - fields:
    - timestamp   (integer)
    - buffer_size (integer)
    - duration    (integer)

## Sample Queries

Get the max tx_latency for the last hour for all interfaces on all switches.

```sql
SELECT max("tx_latency") AS "max_tx_latency" FROM "congestion_record" WHERE time > now() - 1h GROUP BY time(10s), "hostname", "intf_name"
```

Get the max tx_latency for the last hour for all interfaces on all switches.

```sql
SELECT max("queue_size") AS "max_queue_size" FROM "congestion_record" WHERE time > now() - 1h GROUP BY time(10s), "hostname", "intf_name"
```

Get the max buffer_size for over the last hour for all switches.

```sql
SELECT max("buffer_size") AS "max_buffer_size" FROM "global_buffer_usage_record" WHERE time > now() - 1h GROUP BY time(10s), "hostname"
```

## Example Output

```text
lanz_global_buffer_usage_record,entry_type=2,host=telegraf.int.example.com,port=50001,source=switch01.int.example.com timestamp=158334105824919i,buffer_size=505i,duration=0i 1583341058300643815
lanz_congestion_record,entry_type=2,host=telegraf.int.example.com,intf_name=Ethernet36,port=50001,port_id=61,source=switch01.int.example.com,switch_id=0,traffic_class=1 time_of_max_qlen=0i,tx_latency=564480i,q_drop_count=0i,timestamp=158334105824919i,queue_size=225i 1583341058300636045
lanz_global_buffer_usage_record,entry_type=2,host=telegraf.int.example.com,port=50001,source=switch01.int.example.com timestamp=158334105824919i,buffer_size=589i,duration=0i 1583341058300457464
lanz_congestion_record,entry_type=1,host=telegraf.int.example.com,intf_name=Ethernet36,port=50001,port_id=61,source=switch01.int.example.com,switch_id=0,traffic_class=1 q_drop_count=0i,timestamp=158334105824919i,queue_size=232i,time_of_max_qlen=0i,tx_latency=584640i 1583341058300450302
```
