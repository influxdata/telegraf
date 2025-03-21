# Data Plane Development Kit (DPDK) Input Plugin

This plugin collects metrics exposed by applications built with the
[Data Plane Development Kit][dpdk] which is an extensive set of open
source libraries designed for accelerating packet processing workloads.

> [!NOTE]
> Since DPDK will most likely run with root privileges, the telemetry socket
> exposed by DPDK will also require root access. Please adjust permissions
> accordingly!

Refer to the [Telemetry User Guide][user_guide] for details and examples on how
to use DPDK in your application.

> [!IMPORTANT]
> This plugin uses the `v2` interface to read telemetry > data from applications
> and required DPDK version `v20.05` or higher. Some metrics might require later
> versions.
> The recommended version, especially in conjunction with the `in_memory`
> option is `DPDK 21.11.2` or higher.

⭐ Telegraf v1.19.0
🏷️ applications, network
💻 linux

[dpdk]: https://www.dpdk.org
[user_guide]: https://doc.dpdk.org/guides/howto/telemetry.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Reads metrics from DPDK applications using v2 telemetry interface.
# This plugin ONLY supports Linux
[[inputs.dpdk]]
  ## Path to DPDK telemetry socket. This shall point to v2 version of DPDK
  ## telemetry interface.
  # socket_path = "/var/run/dpdk/rte/dpdk_telemetry.v2"

  ## Duration that defines how long the connected socket client will wait for
  ## a response before terminating connection.
  ## This includes both writing to and reading from socket. Since it's local
  ## socket access to a fast packet processing application, the timeout should
  ## be sufficient for most users.
  ## Setting the value to 0 disables the timeout (not recommended)
  # socket_access_timeout = "200ms"

  ## Enables telemetry data collection for selected device types.
  ## Adding "ethdev" enables collection of telemetry from DPDK NICs (stats, xstats, link_status, info).
  ## Adding "rawdev" enables collection of telemetry from DPDK Raw Devices (xstats).
  # device_types = ["ethdev"]

  ## List of custom, application-specific telemetry commands to query
  ## The list of available commands depend on the application deployed.
  ## Applications can register their own commands via telemetry library API
  ## https://doc.dpdk.org/guides/prog_guide/telemetry_lib.html#registering-commands
  ## For L3 Forwarding with Power Management Sample Application this could be:
  ##   additional_commands = ["/l3fwd-power/stats"]
  # additional_commands = []

  ## List of plugin options.
  ## Supported options:
  ##  - "in_memory" option enables reading for multiple sockets when a dpdk application is running with --in-memory option.
  ##    When option is enabled plugin will try to find additional socket paths related to provided socket_path.
  ##    Details: https://doc.dpdk.org/guides/howto/telemetry.html#connecting-to-different-dpdk-processes
  # plugin_options = ["in_memory"]

  ## Specifies plugin behavior regarding unreachable socket (which might not have been initialized yet).
  ## Available choices:
  ##   - error: Telegraf will return an error during the startup and gather phases if socket is unreachable
  ##   - ignore: Telegraf will ignore error regarding unreachable socket on both startup and gather
  # unreachable_socket_behavior = "error"

  ## List of metadata fields which will be added to every metric produced by the plugin.
  ## Supported options:
  ##  - "pid" - exposes PID of DPDK process. Example: pid=2179660i
  ##  - "version" - exposes version of DPDK. Example: version="DPDK 21.11.2"
  # metadata_fields = ["pid", "version"]

  ## Allows turning off collecting data for individual "ethdev" commands.
  ## Remove "/ethdev/link_status" from list to gather link status metrics.
  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]

  ## When running multiple instances of the plugin it's recommended to add a
  ## unique tag to each instance to identify metrics exposed by an instance
  ## of DPDK application. This is useful when multiple DPDK apps run on a
  ## single host.
  ##  [inputs.dpdk.tags]
  ##    dpdk_instance = "my-fwd-app"
```

This plugin offers multiple configuration options, please review examples below
for additional usage information.

### Example: Minimal Configuration for NIC metrics

This configuration allows getting metrics for all devices reported via
`/ethdev/list` command:

* `/ethdev/info` - device information: name, MAC address, buffers size, etc
                   (since `DPDK 21.11`)
* `/ethdev/stats` - basic device statistics (since `DPDK 20.11`)
* `/ethdev/xstats` - extended device statistics
* `/ethdev/link_status` - up/down link status

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]
```

Since this configuration will query `/ethdev/link_status` it's recommended to
increase timeout to `socket_access_timeout = "10s"`.

The [plugin collecting interval](../../../docs/CONFIGURATION.md#input-plugins)
should be adjusted accordingly (e.g. `interval = "30s"`).

### Example: Excluding NIC link status from being collected

Checking link status depending on underlying implementation may take more time
to complete. This configuration can be used to exclude this telemetry command
to allow faster response for metrics.

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]
```

A separate plugin instance with higher timeout settings can be used to get
`/ethdev/link_status` independently.  Consult [Independent NIC link status
configuration](#example-independent-nic-link-status-configuration) and [Getting
metrics from multiple DPDK instances running on same
host](#example-getting-metrics-from-multiple-dpdk-instances-on-same-host)
examples for further details.

### Example: Independent NIC link status configuration

This configuration allows getting `/ethdev/link_status` using separate
configuration, with higher timeout.

```toml
[[inputs.dpdk]]
  interval = "30s"
  socket_access_timeout = "10s"
  device_types = ["ethdev"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/info", "/ethdev/stats", "/ethdev/xstats"]
```

### Example: Getting application-specific metrics

This configuration allows reading custom metrics exposed by
applications. Example telemetry command obtained from
[L3 Forwarding with Power Management Sample Application][sample].

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]
  additional_commands = ["/l3fwd-power/stats"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]
```

Command entries specified in `additional_commands` should match DPDK command
format:

* Command entry format: either `command` or `command,params` for commands that
  expect parameters, where comma (`,`) separates command from params.
* Command entry length (command with params) should be `< 1024` characters.
* Command length (without params) should be `< 56` characters.
* Commands have to start with `/`.

Providing invalid commands will prevent the plugin from starting. Additional
commands allow duplicates, but they will be removed during execution, so each
command will be executed only once during each metric gathering interval.

[sample]: https://doc.dpdk.org/guides/sample_app_ug/l3_forward_power_man.html

### Example: Getting metrics from multiple DPDK instances on same host

This configuration allows getting metrics from two separate applications
exposing their telemetry interfaces via separate sockets. For each plugin
instance a unique tag `[inputs.dpdk.tags]` allows distinguishing between them.

```toml
# Instance #1 - L3 Forwarding with Power Management Application
[[inputs.dpdk]]
  socket_path = "/var/run/dpdk/rte/l3fwd-power_telemetry.v2"
  device_types = ["ethdev"]
  additional_commands = ["/l3fwd-power/stats"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]

  [inputs.dpdk.tags]
    dpdk_instance = "l3fwd-power"

# Instance #2 - L2 Forwarding with Intel Cache Allocation Technology (CAT)
# Application
[[inputs.dpdk]]
  socket_path = "/var/run/dpdk/rte/l2fwd-cat_telemetry.v2"
  device_types = ["ethdev"]

[inputs.dpdk.ethdev]
  exclude_commands = ["/ethdev/link_status"]

  [inputs.dpdk.tags]
    dpdk_instance = "l2fwd-cat"
```

This utilizes Telegraf's standard capability of [adding custom
tags](../../../docs/CONFIGURATION.md#input-plugins) to input plugin's
measurements.

## Metrics

The DPDK socket accepts `command,params` requests and returns metric data in
JSON format. All metrics from DPDK socket become flattened using [Telegraf's
JSON Flattener](../../parsers/json/README.md) and exposed as fields.  If DPDK
response contains no information (is empty or is null) then such response will
be discarded.

> **NOTE:** Since DPDK allows registering custom metrics in its telemetry
> framework the JSON response from DPDK may contain various sets of metrics.
> While metrics from `/ethdev/stats` should be most stable, the `/ethdev/xstats`
> may contain driver-specific metrics (depending on DPDK application
> configuration). The application-specific commands like `/l3fwd-power/stats`
> can return their own specific set of metrics.

## Example Output

The output consists of plugin name (`dpdk`), and a set of tags that identify
querying hierarchy:

```text
dpdk,host=dpdk-host,dpdk_instance=l3fwd-power,command=/ethdev/stats,params=0 [fields] [timestamp]
```

| Tag | Description |
|-----|-------------|
| `host` | hostname of the machine (consult [Telegraf Agent configuration](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#agent) for additional details) |
| `dpdk_instance` | custom tag from `[inputs.dpdk.tags]` (optional) |
| `command` | executed command (without params) |
| `params` | command parameter, e.g. for `/ethdev/stats` it is the ID of NIC as exposed by `/ethdev/list`. For DPDK app that uses 2 NICs the metrics will output e.g. `params=0`, `params=1`. |

When running plugin configuration below...

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]
  additional_commands = ["/l3fwd-power/stats"]
  metadata_fields = []
  [inputs.dpdk.tags]
    dpdk_instance = "l3fwd-power"
```

...expected output for `dpdk` plugin instance running on host named
`host=dpdk-host`:

```text
dpdk,command=/ethdev/info,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 all_multicast=0,dev_configured=1,dev_flags=74,dev_started=1,ethdev_rss_hf=0,lro=0,mac_addr="E4:3D:1A:DD:13:31",mtu=1500,name="0000:ca:00.1",nb_rx_queues=1,nb_tx_queues=1,numa_node=1,port_id=0,promiscuous=1,rx_mbuf_alloc_fail=0,rx_mbuf_size_min=2176,rx_offloads=0,rxq_state_0=1,scattered_rx=0,state=1,tx_offloads=65536,txq_state_0=1 1659017414000000000
dpdk,command=/ethdev/stats,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 q_opackets_0=0,q_ipackets_5=0,q_errors_11=0,ierrors=0,q_obytes_5=0,q_obytes_10=0,q_opackets_10=0,q_ipackets_4=0,q_ipackets_7=0,q_ipackets_15=0,q_ibytes_5=0,q_ibytes_6=0,q_ibytes_9=0,obytes=0,q_opackets_1=0,q_opackets_11=0,q_obytes_7=0,q_errors_5=0,q_errors_10=0,q_ibytes_4=0,q_obytes_6=0,q_errors_1=0,q_opackets_5=0,q_errors_3=0,q_errors_12=0,q_ipackets_11=0,q_ipackets_12=0,q_obytes_14=0,q_opackets_15=0,q_obytes_2=0,q_errors_8=0,q_opackets_12=0,q_errors_0=0,q_errors_9=0,q_opackets_14=0,q_ibytes_3=0,q_ibytes_15=0,q_ipackets_13=0,q_ipackets_14=0,q_obytes_3=0,q_errors_13=0,q_opackets_3=0,q_ibytes_0=7092,q_ibytes_2=0,q_ibytes_8=0,q_ipackets_8=0,q_ipackets_10=0,q_obytes_4=0,q_ibytes_10=0,q_ibytes_13=0,q_ibytes_1=0,q_ibytes_12=0,opackets=0,q_obytes_1=0,q_errors_15=0,q_opackets_2=0,oerrors=0,rx_nombuf=0,q_opackets_8=0,q_ibytes_11=0,q_ipackets_3=0,q_obytes_0=0,q_obytes_12=0,q_obytes_11=0,q_obytes_13=0,q_errors_6=0,q_ipackets_1=0,q_ipackets_6=0,q_ipackets_9=0,q_obytes_15=0,q_opackets_7=0,q_ibytes_14=0,ipackets=98,q_ipackets_2=0,q_opackets_6=0,q_ibytes_7=0,imissed=0,q_opackets_4=0,q_opackets_9=0,q_obytes_8=0,q_obytes_9=0,q_errors_4=0,q_errors_14=0,q_opackets_13=0,ibytes=7092,q_ipackets_0=98,q_errors_2=0,q_errors_7=0 1606310780000000000
dpdk,command=/ethdev/stats,dpdk_instance=l3fwd-power,host=dpdk-host,params=1 q_opackets_0=0,q_ipackets_5=0,q_errors_11=0,ierrors=0,q_obytes_5=0,q_obytes_10=0,q_opackets_10=0,q_ipackets_4=0,q_ipackets_7=0,q_ipackets_15=0,q_ibytes_5=0,q_ibytes_6=0,q_ibytes_9=0,obytes=0,q_opackets_1=0,q_opackets_11=0,q_obytes_7=0,q_errors_5=0,q_errors_10=0,q_ibytes_4=0,q_obytes_6=0,q_errors_1=0,q_opackets_5=0,q_errors_3=0,q_errors_12=0,q_ipackets_11=0,q_ipackets_12=0,q_obytes_14=0,q_opackets_15=0,q_obytes_2=0,q_errors_8=0,q_opackets_12=0,q_errors_0=0,q_errors_9=0,q_opackets_14=0,q_ibytes_3=0,q_ibytes_15=0,q_ipackets_13=0,q_ipackets_14=0,q_obytes_3=0,q_errors_13=0,q_opackets_3=0,q_ibytes_0=7092,q_ibytes_2=0,q_ibytes_8=0,q_ipackets_8=0,q_ipackets_10=0,q_obytes_4=0,q_ibytes_10=0,q_ibytes_13=0,q_ibytes_1=0,q_ibytes_12=0,opackets=0,q_obytes_1=0,q_errors_15=0,q_opackets_2=0,oerrors=0,rx_nombuf=0,q_opackets_8=0,q_ibytes_11=0,q_ipackets_3=0,q_obytes_0=0,q_obytes_12=0,q_obytes_11=0,q_obytes_13=0,q_errors_6=0,q_ipackets_1=0,q_ipackets_6=0,q_ipackets_9=0,q_obytes_15=0,q_opackets_7=0,q_ibytes_14=0,ipackets=98,q_ipackets_2=0,q_opackets_6=0,q_ibytes_7=0,imissed=0,q_opackets_4=0,q_opackets_9=0,q_obytes_8=0,q_obytes_9=0,q_errors_4=0,q_errors_14=0,q_opackets_13=0,ibytes=7092,q_ipackets_0=98,q_errors_2=0,q_errors_7=0 1606310780000000000
dpdk,command=/ethdev/xstats,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 out_octets_encrypted=0,rx_fcoe_mbuf_allocation_errors=0,tx_q1packets=0,rx_priority0_xoff_packets=0,rx_priority7_xoff_packets=0,rx_errors=0,mac_remote_errors=0,in_pkts_invalid=0,tx_priority3_xoff_packets=0,tx_errors=0,rx_fcoe_bytes=0,rx_flow_control_xon_packets=0,rx_priority4_xoff_packets=0,tx_priority2_xoff_packets=0,rx_illegal_byte_errors=0,rx_xoff_packets=0,rx_management_packets=0,rx_priority7_dropped=0,rx_priority4_dropped=0,in_pkts_unchecked=0,rx_error_bytes=0,rx_size_256_to_511_packets=0,tx_priority4_xoff_packets=0,rx_priority6_xon_packets=0,tx_priority4_xon_to_xoff_packets=0,in_pkts_delayed=0,rx_priority0_mbuf_allocation_errors=0,out_octets_protected=0,tx_priority7_xon_to_xoff_packets=0,tx_priority1_xon_to_xoff_packets=0,rx_fcoe_no_direct_data_placement_ext_buff=0,tx_priority6_xon_to_xoff_packets=0,flow_director_filter_add_errors=0,rx_total_packets=99,rx_crc_errors=0,flow_director_filter_remove_errors=0,rx_missed_errors=0,tx_size_64_packets=0,rx_priority3_dropped=0,flow_director_matched_filters=0,tx_priority2_xon_to_xoff_packets=0,rx_priority1_xon_packets=0,rx_size_65_to_127_packets=99,rx_fragment_errors=0,in_pkts_notusingsa=0,rx_q0bytes=7162,rx_fcoe_dropped=0,rx_priority1_dropped=0,rx_fcoe_packets=0,rx_priority5_xoff_packets=0,out_pkts_protected=0,tx_total_packets=0,rx_priority2_dropped=0,in_pkts_late=0,tx_q1bytes=0,in_pkts_badtag=0,rx_multicast_packets=99,rx_priority6_xoff_packets=0,tx_flow_control_xoff_packets=0,rx_flow_control_xoff_packets=0,rx_priority0_xon_packets=0,in_pkts_untagged=0,tx_fcoe_packets=0,rx_priority7_mbuf_allocation_errors=0,tx_priority0_xon_to_xoff_packets=0,tx_priority5_xon_to_xoff_packets=0,tx_flow_control_xon_packets=0,tx_q0packets=0,tx_xoff_packets=0,rx_size_512_to_1023_packets=0,rx_priority3_xon_packets=0,rx_q0errors=0,rx_oversize_errors=0,tx_priority4_xon_packets=0,tx_priority5_xoff_packets=0,rx_priority5_xon_packets=0,rx_total_missed_packets=0,rx_priority4_mbuf_allocation_errors=0,tx_priority1_xon_packets=0,tx_management_packets=0,rx_priority5_mbuf_allocation_errors=0,rx_fcoe_no_direct_data_placement=0,rx_undersize_errors=0,tx_priority1_xoff_packets=0,rx_q0packets=99,tx_q2packets=0,tx_priority6_xon_packets=0,rx_good_packets=99,tx_priority5_xon_packets=0,tx_size_256_to_511_packets=0,rx_priority6_dropped=0,rx_broadcast_packets=0,tx_size_512_to_1023_packets=0,tx_priority3_xon_to_xoff_packets=0,in_pkts_unknownsci=0,in_octets_validated=0,tx_priority6_xoff_packets=0,tx_priority7_xoff_packets=0,rx_jabber_errors=0,tx_priority7_xon_packets=0,tx_priority0_xon_packets=0,in_pkts_unusedsa=0,tx_priority0_xoff_packets=0,mac_local_errors=33,rx_total_bytes=7162,in_pkts_notvalid=0,rx_length_errors=0,in_octets_decrypted=0,rx_size_128_to_255_packets=0,rx_good_bytes=7162,tx_size_65_to_127_packets=0,rx_mac_short_packet_dropped=0,tx_size_1024_to_max_packets=0,rx_priority2_mbuf_allocation_errors=0,flow_director_added_filters=0,tx_multicast_packets=0,rx_fcoe_crc_errors=0,rx_priority1_xoff_packets=0,flow_director_missed_filters=0,rx_xon_packets=0,tx_size_128_to_255_packets=0,out_pkts_encrypted=0,rx_priority4_xon_packets=0,rx_priority0_dropped=0,rx_size_1024_to_max_packets=0,tx_good_bytes=0,rx_management_dropped=0,rx_mbuf_allocation_errors=0,tx_xon_packets=0,rx_priority3_xoff_packets=0,tx_good_packets=0,tx_fcoe_bytes=0,rx_priority6_mbuf_allocation_errors=0,rx_priority2_xon_packets=0,tx_broadcast_packets=0,tx_q2bytes=0,rx_priority7_xon_packets=0,out_pkts_untagged=0,rx_priority2_xoff_packets=0,rx_priority1_mbuf_allocation_errors=0,tx_q0bytes=0,rx_size_64_packets=0,rx_priority5_dropped=0,tx_priority2_xon_packets=0,in_pkts_nosci=0,flow_director_removed_filters=0,in_pkts_ok=0,rx_l3_l4_xsum_error=0,rx_priority3_mbuf_allocation_errors=0,tx_priority3_xon_packets=0 1606310780000000000
dpdk,command=/ethdev/xstats,dpdk_instance=l3fwd-power,host=dpdk-host,params=1 tx_priority5_xoff_packets=0,in_pkts_unknownsci=0,tx_q0packets=0,tx_total_packets=0,rx_crc_errors=0,rx_priority4_xoff_packets=0,rx_priority5_dropped=0,tx_size_65_to_127_packets=0,rx_good_packets=98,tx_priority6_xoff_packets=0,tx_fcoe_bytes=0,out_octets_protected=0,out_pkts_encrypted=0,rx_priority1_xon_packets=0,tx_size_128_to_255_packets=0,rx_flow_control_xoff_packets=0,rx_priority7_xoff_packets=0,tx_priority0_xon_to_xoff_packets=0,rx_broadcast_packets=0,tx_priority1_xon_packets=0,rx_xon_packets=0,rx_fragment_errors=0,tx_flow_control_xoff_packets=0,tx_q0bytes=0,out_pkts_untagged=0,rx_priority4_xon_packets=0,tx_priority5_xon_packets=0,rx_priority1_xoff_packets=0,rx_good_bytes=7092,rx_priority4_mbuf_allocation_errors=0,in_octets_decrypted=0,tx_priority2_xon_to_xoff_packets=0,rx_priority3_dropped=0,tx_multicast_packets=0,mac_local_errors=33,in_pkts_ok=0,rx_illegal_byte_errors=0,rx_xoff_packets=0,rx_q0errors=0,flow_director_added_filters=0,rx_size_256_to_511_packets=0,rx_priority3_xon_packets=0,rx_l3_l4_xsum_error=0,rx_priority6_dropped=0,in_pkts_notvalid=0,rx_size_64_packets=0,tx_management_packets=0,rx_length_errors=0,tx_priority7_xon_to_xoff_packets=0,rx_mbuf_allocation_errors=0,rx_missed_errors=0,rx_priority1_mbuf_allocation_errors=0,rx_fcoe_no_direct_data_placement=0,tx_priority3_xoff_packets=0,in_pkts_delayed=0,tx_errors=0,rx_size_512_to_1023_packets=0,tx_priority4_xon_packets=0,rx_q0bytes=7092,in_pkts_unchecked=0,tx_size_512_to_1023_packets=0,rx_fcoe_packets=0,in_pkts_nosci=0,rx_priority6_mbuf_allocation_errors=0,rx_priority1_dropped=0,tx_q2packets=0,rx_priority7_dropped=0,tx_size_1024_to_max_packets=0,rx_management_packets=0,rx_multicast_packets=98,rx_total_bytes=7092,mac_remote_errors=0,tx_priority3_xon_packets=0,rx_priority2_mbuf_allocation_errors=0,rx_priority5_mbuf_allocation_errors=0,tx_q2bytes=0,rx_size_128_to_255_packets=0,in_pkts_badtag=0,out_pkts_protected=0,rx_management_dropped=0,rx_fcoe_bytes=0,flow_director_removed_filters=0,tx_priority2_xoff_packets=0,rx_fcoe_crc_errors=0,rx_priority0_mbuf_allocation_errors=0,rx_priority0_xon_packets=0,rx_fcoe_dropped=0,tx_priority1_xon_to_xoff_packets=0,rx_size_65_to_127_packets=98,rx_q0packets=98,tx_priority0_xoff_packets=0,rx_priority6_xon_packets=0,rx_total_packets=98,rx_undersize_errors=0,flow_director_missed_filters=0,rx_jabber_errors=0,in_pkts_invalid=0,in_pkts_late=0,rx_priority5_xon_packets=0,tx_priority4_xoff_packets=0,out_octets_encrypted=0,tx_q1packets=0,rx_priority5_xoff_packets=0,rx_priority6_xoff_packets=0,rx_errors=0,in_octets_validated=0,rx_priority3_xoff_packets=0,tx_priority4_xon_to_xoff_packets=0,tx_priority5_xon_to_xoff_packets=0,tx_flow_control_xon_packets=0,rx_priority0_dropped=0,flow_director_filter_add_errors=0,tx_q1bytes=0,tx_priority6_xon_to_xoff_packets=0,flow_director_matched_filters=0,tx_priority2_xon_packets=0,rx_fcoe_mbuf_allocation_errors=0,rx_priority2_xoff_packets=0,tx_priority7_xoff_packets=0,rx_priority0_xoff_packets=0,rx_oversize_errors=0,in_pkts_notusingsa=0,tx_size_64_packets=0,rx_size_1024_to_max_packets=0,tx_priority6_xon_packets=0,rx_priority2_dropped=0,rx_priority4_dropped=0,rx_priority7_mbuf_allocation_errors=0,rx_flow_control_xon_packets=0,tx_good_bytes=0,tx_priority3_xon_to_xoff_packets=0,rx_total_missed_packets=0,rx_error_bytes=0,tx_priority7_xon_packets=0,rx_mac_short_packet_dropped=0,tx_priority1_xoff_packets=0,tx_good_packets=0,tx_broadcast_packets=0,tx_xon_packets=0,in_pkts_unusedsa=0,rx_priority2_xon_packets=0,in_pkts_untagged=0,tx_fcoe_packets=0,flow_director_filter_remove_errors=0,rx_priority3_mbuf_allocation_errors=0,tx_priority0_xon_packets=0,rx_priority7_xon_packets=0,rx_fcoe_no_direct_data_placement_ext_buff=0,tx_xoff_packets=0,tx_size_256_to_511_packets=0 1606310780000000000
dpdk,command=/ethdev/link_status,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 status="UP",link_status=1,speed=10000,duplex="full-duplex" 1606310780000000000
dpdk,command=/ethdev/link_status,dpdk_instance=l3fwd-power,host=dpdk-host,params=1 status="UP",link_status=1,speed=10000,duplex="full-duplex" 1606310780000000000
dpdk,command=/l3fwd-power/stats,dpdk_instance=l3fwd-power,host=dpdk-host empty_poll=49506395979901,full_poll=0,busy_percent=0 1606310780000000000
```

When running plugin configuration below...

```toml
[[inputs.dpdk]]
  interval = "30s"
  socket_access_timeout = "10s"
  device_types = ["ethdev"]
  metadata_fields = ["version", "pid"]
  plugin_options = ["in_memory"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/info", "/ethdev/stats", "/ethdev/xstats"]
```

Expected output for `dpdk` plugin instance running with `link_status` command
and all metadata fields enabled, additionally `link_status` field will be
exposed to represent string value of `status` field (`DOWN`=0,`UP`=1):

```text
dpdk,command=/ethdev/link_status,host=dpdk-host,params=0 pid=100988i,version="DPDK 21.11.2",status="DOWN",link_status=0i 1660295749000000000
dpdk,command=/ethdev/link_status,host=dpdk-host,params=0 pid=2401624i,version="DPDK 21.11.2",status="UP",link_status=1i 1660295749000000000
```
