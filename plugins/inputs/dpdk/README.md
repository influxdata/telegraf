# Data Plane Development Kit (DPDK) Input Plugin

The `dpdk` plugin collects metrics exposed by applications built with [Data Plane Development Kit](https://www.dpdk.org/)
which is an extensive set of open source libraries designed for accelerating packet processing workloads.

DPDK provides APIs that enable exposing various statistics from the devices used by DPDK applications and enable exposing
KPI metrics directly from applications. Device statistics include e.g. common statistics available across NICs, like:
received and sent packets, received and sent bytes etc. In addition to this generic statistics, an extended statistics API
is available that allows providing more detailed, driver-specific metrics that are not available as generic statistics.

[DPDK Release 20.05](https://doc.dpdk.org/guides/rel_notes/release_20_05.html) introduced updated telemetry interface
that enables DPDK libraries and applications to provide their telemetry. This is referred to as `v2` version of this
socket-based telemetry interface. This release enabled e.g. reading driver-specific extended stats (`/ethdev/xstats`)
via this new interface.

[DPDK Release 20.11](https://doc.dpdk.org/guides/rel_notes/release_20_11.html) introduced reading via `v2` interface
common statistics (`/ethdev/stats`) in addition to existing (`/ethdev/xstats`).

The example usage of `v2` telemetry interface can be found in [Telemetry User Guide](https://doc.dpdk.org/guides/howto/telemetry.html).
A variety of [DPDK Sample Applications](https://doc.dpdk.org/guides/sample_app_ug/index.html) is also available for users
to discover and test the capabilities of DPDK libraries and to explore the exposed metrics.

> **DPDK Version Info:** This plugin uses this `v2` interface to read telemetry data from applications build with
> `DPDK version >= 20.05`. The default configuration include reading common statistics from `/ethdev/stats` that is
> available from `DPDK version >= 20.11`. When using `DPDK 20.05 <= version < DPDK 20.11` it is recommended to disable
> querying `/ethdev/stats` by setting corresponding `exclude_commands` configuration option.
> **NOTE:** Since DPDK will most likely run with root privileges, the socket telemetry interface exposed by DPDK
> will also require root access. This means that either access permissions have to be adjusted for socket telemetry
> interface to allow Telegraf to access it, or Telegraf should run with root privileges.
> **NOTE:** The DPDK socket must exist for Telegraf to start successfully. Telegraf will attempt
> to connect to the DPDK socket during the initialization phase.

## Configuration

This plugin offers multiple configuration options, please review examples below for additional usage information.

```toml
# Reads metrics from DPDK applications using v2 telemetry interface.
[[inputs.dpdk]]
  ## Path to DPDK telemetry socket. This shall point to v2 version of DPDK telemetry interface.
  # socket_path = "/var/run/dpdk/rte/dpdk_telemetry.v2"

  ## Duration that defines how long the connected socket client will wait for a response before terminating connection.
  ## This includes both writing to and reading from socket. Since it's local socket access
  ## to a fast packet processing application, the timeout should be sufficient for most users.
  ## Setting the value to 0 disables the timeout (not recommended)
  # socket_access_timeout = "200ms"

  ## Enables telemetry data collection for selected device types.
  ## Adding "ethdev" enables collection of telemetry from DPDK NICs (stats, xstats, link_status).
  ## Adding "rawdev" enables collection of telemetry from DPDK Raw Devices (xstats).
  # device_types = ["ethdev"]

  ## List of custom, application-specific telemetry commands to query
  ## The list of available commands depend on the application deployed. Applications can register their own commands
  ##   via telemetry library API http://doc.dpdk.org/guides/prog_guide/telemetry_lib.html#registering-commands
  ## For e.g. L3 Forwarding with Power Management Sample Application this could be:
  ##   additional_commands = ["/l3fwd-power/stats"]
  # additional_commands = []

  ## Allows turning off collecting data for individual "ethdev" commands.
  ## Remove "/ethdev/link_status" from list to start getting link status metrics.
  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]

  ## When running multiple instances of the plugin it's recommended to add a unique tag to each instance to identify
  ## metrics exposed by an instance of DPDK application. This is useful when multiple DPDK apps run on a single host.
  ##  [inputs.dpdk.tags]
  ##    dpdk_instance = "my-fwd-app"
```

### Example: Minimal Configuration for NIC metrics

This configuration allows getting metrics for all devices reported via `/ethdev/list` command:

* `/ethdev/stats` - basic device statistics (since `DPDK 20.11`)
* `/ethdev/xstats` - extended device statistics
* `/ethdev/link_status` - up/down link status

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]
```

Since this configuration will query `/ethdev/link_status` it's recommended to increase timeout to `socket_access_timeout = "10s"`.

The [plugin collecting interval](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#input-plugins)
should be adjusted accordingly (e.g. `interval = "30s"`).

### Example: Excluding NIC link status from being collected

Checking link status depending on underlying implementation may take more time to complete.
This configuration can be used to exclude this telemetry command to allow faster response for metrics.

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]
```

A separate plugin instance with higher timeout settings can be used to get `/ethdev/link_status` independently.
Consult [Independent NIC link status configuration](#example-independent-nic-link-status-configuration)
and [Getting metrics from multiple DPDK instances running on same host](#example-getting-metrics-from-multiple-dpdk-instances-running-on-same-host)
examples for further details.

### Example: Independent NIC link status configuration

This configuration allows getting `/ethdev/link_status` using separate configuration, with higher timeout.

```toml
[[inputs.dpdk]]
  interval = "30s"
  socket_access_timeout = "10s"
  device_types = ["ethdev"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/stats", "/ethdev/xstats"]
```

### Example: Getting application-specific metrics

This configuration allows reading custom metrics exposed by applications. Example telemetry command obtained from
[L3 Forwarding with Power Management Sample Application](https://doc.dpdk.org/guides/sample_app_ug/l3_forward_power_man.html).

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]
  additional_commands = ["/l3fwd-power/stats"]

  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]
```

Command entries specified in `additional_commands` should match DPDK command format:

* Command entry format: either `command` or `command,params` for commands that expect parameters, where comma (`,`) separates command from params.
* Command entry length (command with params) should be `< 1024` characters.
* Command length (without params) should be `< 56` characters.
* Commands have to start with `/`.

Providing invalid commands will prevent the plugin from starting. Additional commands allow duplicates, but they
will be removed during execution so each command will be executed only once during each metric gathering interval.

### Example: Getting metrics from multiple DPDK instances running on same host

This configuration allows getting metrics from two separate applications exposing their telemetry interfaces
via separate sockets. For each plugin instance a unique tag `[inputs.dpdk.tags]` allows distinguishing between them.

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

# Instance #2 - L2 Forwarding with Intel Cache Allocation Technology (CAT) Application
[[inputs.dpdk]]
  socket_path = "/var/run/dpdk/rte/l2fwd-cat_telemetry.v2"
  device_types = ["ethdev"]

[inputs.dpdk.ethdev]
  exclude_commands = ["/ethdev/link_status"]

  [inputs.dpdk.tags]
    dpdk_instance = "l2fwd-cat"
```

This utilizes Telegraf's standard capability of [adding custom tags](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#input-plugins)
to input plugin's measurements.

## Metrics

The DPDK socket accepts `command,params` requests and returns metric data in JSON format. All metrics from DPDK socket
become flattened using [Telegraf's JSON Flattener](../../parsers/json/README.md) and exposed as fields.
If DPDK response contains no information (is empty or is null) then such response will be discarded.

> **NOTE:**  Since DPDK allows registering custom metrics in its telemetry framework the JSON response from DPDK
> may contain various sets of metrics. While metrics from `/ethdev/stats` should be most stable, the `/ethdev/xstats`
> may contain driver-specific metrics (depending on DPDK application configuration). The application-specific commands
> like `/l3fwd-power/stats` can return their own specific set of metrics.

## Example output

The output consists of plugin name (`dpdk`), and a set of tags that identify querying hierarchy:

```shell
dpdk,host=dpdk-host,dpdk_instance=l3fwd-power,command=/ethdev/stats,params=0 [fields] [timestamp]
```

| Tag | Description |
|-----|-------------|
| `host` | hostname of the machine (consult [Telegraf Agent configuration](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#agent) for additional details) |
| `dpdk_instance` | custom tag from `[inputs.dpdk.tags]` (optional) |
| `command` | executed command (without params) |
| `params` | command parameter, e.g. for `/ethdev/stats` it is the id of NIC as exposed by `/ethdev/list`. For DPDK app that uses 2 NICs the metrics will output e.g. `params=0`, `params=1`. |

When running plugin configuration below...

```toml
[[inputs.dpdk]]
  device_types = ["ethdev"]
  additional_commands = ["/l3fwd-power/stats"]
  [inputs.dpdk.tags]
    dpdk_instance = "l3fwd-power"
```

...expected output for `dpdk` plugin instance running on host named `host=dpdk-host`:

```shell
dpdk,command=/ethdev/stats,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 q_opackets_0=0,q_ipackets_5=0,q_errors_11=0,ierrors=0,q_obytes_5=0,q_obytes_10=0,q_opackets_10=0,q_ipackets_4=0,q_ipackets_7=0,q_ipackets_15=0,q_ibytes_5=0,q_ibytes_6=0,q_ibytes_9=0,obytes=0,q_opackets_1=0,q_opackets_11=0,q_obytes_7=0,q_errors_5=0,q_errors_10=0,q_ibytes_4=0,q_obytes_6=0,q_errors_1=0,q_opackets_5=0,q_errors_3=0,q_errors_12=0,q_ipackets_11=0,q_ipackets_12=0,q_obytes_14=0,q_opackets_15=0,q_obytes_2=0,q_errors_8=0,q_opackets_12=0,q_errors_0=0,q_errors_9=0,q_opackets_14=0,q_ibytes_3=0,q_ibytes_15=0,q_ipackets_13=0,q_ipackets_14=0,q_obytes_3=0,q_errors_13=0,q_opackets_3=0,q_ibytes_0=7092,q_ibytes_2=0,q_ibytes_8=0,q_ipackets_8=0,q_ipackets_10=0,q_obytes_4=0,q_ibytes_10=0,q_ibytes_13=0,q_ibytes_1=0,q_ibytes_12=0,opackets=0,q_obytes_1=0,q_errors_15=0,q_opackets_2=0,oerrors=0,rx_nombuf=0,q_opackets_8=0,q_ibytes_11=0,q_ipackets_3=0,q_obytes_0=0,q_obytes_12=0,q_obytes_11=0,q_obytes_13=0,q_errors_6=0,q_ipackets_1=0,q_ipackets_6=0,q_ipackets_9=0,q_obytes_15=0,q_opackets_7=0,q_ibytes_14=0,ipackets=98,q_ipackets_2=0,q_opackets_6=0,q_ibytes_7=0,imissed=0,q_opackets_4=0,q_opackets_9=0,q_obytes_8=0,q_obytes_9=0,q_errors_4=0,q_errors_14=0,q_opackets_13=0,ibytes=7092,q_ipackets_0=98,q_errors_2=0,q_errors_7=0 1606310780000000000
dpdk,command=/ethdev/stats,dpdk_instance=l3fwd-power,host=dpdk-host,params=1 q_opackets_0=0,q_ipackets_5=0,q_errors_11=0,ierrors=0,q_obytes_5=0,q_obytes_10=0,q_opackets_10=0,q_ipackets_4=0,q_ipackets_7=0,q_ipackets_15=0,q_ibytes_5=0,q_ibytes_6=0,q_ibytes_9=0,obytes=0,q_opackets_1=0,q_opackets_11=0,q_obytes_7=0,q_errors_5=0,q_errors_10=0,q_ibytes_4=0,q_obytes_6=0,q_errors_1=0,q_opackets_5=0,q_errors_3=0,q_errors_12=0,q_ipackets_11=0,q_ipackets_12=0,q_obytes_14=0,q_opackets_15=0,q_obytes_2=0,q_errors_8=0,q_opackets_12=0,q_errors_0=0,q_errors_9=0,q_opackets_14=0,q_ibytes_3=0,q_ibytes_15=0,q_ipackets_13=0,q_ipackets_14=0,q_obytes_3=0,q_errors_13=0,q_opackets_3=0,q_ibytes_0=7092,q_ibytes_2=0,q_ibytes_8=0,q_ipackets_8=0,q_ipackets_10=0,q_obytes_4=0,q_ibytes_10=0,q_ibytes_13=0,q_ibytes_1=0,q_ibytes_12=0,opackets=0,q_obytes_1=0,q_errors_15=0,q_opackets_2=0,oerrors=0,rx_nombuf=0,q_opackets_8=0,q_ibytes_11=0,q_ipackets_3=0,q_obytes_0=0,q_obytes_12=0,q_obytes_11=0,q_obytes_13=0,q_errors_6=0,q_ipackets_1=0,q_ipackets_6=0,q_ipackets_9=0,q_obytes_15=0,q_opackets_7=0,q_ibytes_14=0,ipackets=98,q_ipackets_2=0,q_opackets_6=0,q_ibytes_7=0,imissed=0,q_opackets_4=0,q_opackets_9=0,q_obytes_8=0,q_obytes_9=0,q_errors_4=0,q_errors_14=0,q_opackets_13=0,ibytes=7092,q_ipackets_0=98,q_errors_2=0,q_errors_7=0 1606310780000000000
dpdk,command=/ethdev/xstats,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 out_octets_encrypted=0,rx_fcoe_mbuf_allocation_errors=0,tx_q1packets=0,rx_priority0_xoff_packets=0,rx_priority7_xoff_packets=0,rx_errors=0,mac_remote_errors=0,in_pkts_invalid=0,tx_priority3_xoff_packets=0,tx_errors=0,rx_fcoe_bytes=0,rx_flow_control_xon_packets=0,rx_priority4_xoff_packets=0,tx_priority2_xoff_packets=0,rx_illegal_byte_errors=0,rx_xoff_packets=0,rx_management_packets=0,rx_priority7_dropped=0,rx_priority4_dropped=0,in_pkts_unchecked=0,rx_error_bytes=0,rx_size_256_to_511_packets=0,tx_priority4_xoff_packets=0,rx_priority6_xon_packets=0,tx_priority4_xon_to_xoff_packets=0,in_pkts_delayed=0,rx_priority0_mbuf_allocation_errors=0,out_octets_protected=0,tx_priority7_xon_to_xoff_packets=0,tx_priority1_xon_to_xoff_packets=0,rx_fcoe_no_direct_data_placement_ext_buff=0,tx_priority6_xon_to_xoff_packets=0,flow_director_filter_add_errors=0,rx_total_packets=99,rx_crc_errors=0,flow_director_filter_remove_errors=0,rx_missed_errors=0,tx_size_64_packets=0,rx_priority3_dropped=0,flow_director_matched_filters=0,tx_priority2_xon_to_xoff_packets=0,rx_priority1_xon_packets=0,rx_size_65_to_127_packets=99,rx_fragment_errors=0,in_pkts_notusingsa=0,rx_q0bytes=7162,rx_fcoe_dropped=0,rx_priority1_dropped=0,rx_fcoe_packets=0,rx_priority5_xoff_packets=0,out_pkts_protected=0,tx_total_packets=0,rx_priority2_dropped=0,in_pkts_late=0,tx_q1bytes=0,in_pkts_badtag=0,rx_multicast_packets=99,rx_priority6_xoff_packets=0,tx_flow_control_xoff_packets=0,rx_flow_control_xoff_packets=0,rx_priority0_xon_packets=0,in_pkts_untagged=0,tx_fcoe_packets=0,rx_priority7_mbuf_allocation_errors=0,tx_priority0_xon_to_xoff_packets=0,tx_priority5_xon_to_xoff_packets=0,tx_flow_control_xon_packets=0,tx_q0packets=0,tx_xoff_packets=0,rx_size_512_to_1023_packets=0,rx_priority3_xon_packets=0,rx_q0errors=0,rx_oversize_errors=0,tx_priority4_xon_packets=0,tx_priority5_xoff_packets=0,rx_priority5_xon_packets=0,rx_total_missed_packets=0,rx_priority4_mbuf_allocation_errors=0,tx_priority1_xon_packets=0,tx_management_packets=0,rx_priority5_mbuf_allocation_errors=0,rx_fcoe_no_direct_data_placement=0,rx_undersize_errors=0,tx_priority1_xoff_packets=0,rx_q0packets=99,tx_q2packets=0,tx_priority6_xon_packets=0,rx_good_packets=99,tx_priority5_xon_packets=0,tx_size_256_to_511_packets=0,rx_priority6_dropped=0,rx_broadcast_packets=0,tx_size_512_to_1023_packets=0,tx_priority3_xon_to_xoff_packets=0,in_pkts_unknownsci=0,in_octets_validated=0,tx_priority6_xoff_packets=0,tx_priority7_xoff_packets=0,rx_jabber_errors=0,tx_priority7_xon_packets=0,tx_priority0_xon_packets=0,in_pkts_unusedsa=0,tx_priority0_xoff_packets=0,mac_local_errors=33,rx_total_bytes=7162,in_pkts_notvalid=0,rx_length_errors=0,in_octets_decrypted=0,rx_size_128_to_255_packets=0,rx_good_bytes=7162,tx_size_65_to_127_packets=0,rx_mac_short_packet_dropped=0,tx_size_1024_to_max_packets=0,rx_priority2_mbuf_allocation_errors=0,flow_director_added_filters=0,tx_multicast_packets=0,rx_fcoe_crc_errors=0,rx_priority1_xoff_packets=0,flow_director_missed_filters=0,rx_xon_packets=0,tx_size_128_to_255_packets=0,out_pkts_encrypted=0,rx_priority4_xon_packets=0,rx_priority0_dropped=0,rx_size_1024_to_max_packets=0,tx_good_bytes=0,rx_management_dropped=0,rx_mbuf_allocation_errors=0,tx_xon_packets=0,rx_priority3_xoff_packets=0,tx_good_packets=0,tx_fcoe_bytes=0,rx_priority6_mbuf_allocation_errors=0,rx_priority2_xon_packets=0,tx_broadcast_packets=0,tx_q2bytes=0,rx_priority7_xon_packets=0,out_pkts_untagged=0,rx_priority2_xoff_packets=0,rx_priority1_mbuf_allocation_errors=0,tx_q0bytes=0,rx_size_64_packets=0,rx_priority5_dropped=0,tx_priority2_xon_packets=0,in_pkts_nosci=0,flow_director_removed_filters=0,in_pkts_ok=0,rx_l3_l4_xsum_error=0,rx_priority3_mbuf_allocation_errors=0,tx_priority3_xon_packets=0 1606310780000000000
dpdk,command=/ethdev/xstats,dpdk_instance=l3fwd-power,host=dpdk-host,params=1 tx_priority5_xoff_packets=0,in_pkts_unknownsci=0,tx_q0packets=0,tx_total_packets=0,rx_crc_errors=0,rx_priority4_xoff_packets=0,rx_priority5_dropped=0,tx_size_65_to_127_packets=0,rx_good_packets=98,tx_priority6_xoff_packets=0,tx_fcoe_bytes=0,out_octets_protected=0,out_pkts_encrypted=0,rx_priority1_xon_packets=0,tx_size_128_to_255_packets=0,rx_flow_control_xoff_packets=0,rx_priority7_xoff_packets=0,tx_priority0_xon_to_xoff_packets=0,rx_broadcast_packets=0,tx_priority1_xon_packets=0,rx_xon_packets=0,rx_fragment_errors=0,tx_flow_control_xoff_packets=0,tx_q0bytes=0,out_pkts_untagged=0,rx_priority4_xon_packets=0,tx_priority5_xon_packets=0,rx_priority1_xoff_packets=0,rx_good_bytes=7092,rx_priority4_mbuf_allocation_errors=0,in_octets_decrypted=0,tx_priority2_xon_to_xoff_packets=0,rx_priority3_dropped=0,tx_multicast_packets=0,mac_local_errors=33,in_pkts_ok=0,rx_illegal_byte_errors=0,rx_xoff_packets=0,rx_q0errors=0,flow_director_added_filters=0,rx_size_256_to_511_packets=0,rx_priority3_xon_packets=0,rx_l3_l4_xsum_error=0,rx_priority6_dropped=0,in_pkts_notvalid=0,rx_size_64_packets=0,tx_management_packets=0,rx_length_errors=0,tx_priority7_xon_to_xoff_packets=0,rx_mbuf_allocation_errors=0,rx_missed_errors=0,rx_priority1_mbuf_allocation_errors=0,rx_fcoe_no_direct_data_placement=0,tx_priority3_xoff_packets=0,in_pkts_delayed=0,tx_errors=0,rx_size_512_to_1023_packets=0,tx_priority4_xon_packets=0,rx_q0bytes=7092,in_pkts_unchecked=0,tx_size_512_to_1023_packets=0,rx_fcoe_packets=0,in_pkts_nosci=0,rx_priority6_mbuf_allocation_errors=0,rx_priority1_dropped=0,tx_q2packets=0,rx_priority7_dropped=0,tx_size_1024_to_max_packets=0,rx_management_packets=0,rx_multicast_packets=98,rx_total_bytes=7092,mac_remote_errors=0,tx_priority3_xon_packets=0,rx_priority2_mbuf_allocation_errors=0,rx_priority5_mbuf_allocation_errors=0,tx_q2bytes=0,rx_size_128_to_255_packets=0,in_pkts_badtag=0,out_pkts_protected=0,rx_management_dropped=0,rx_fcoe_bytes=0,flow_director_removed_filters=0,tx_priority2_xoff_packets=0,rx_fcoe_crc_errors=0,rx_priority0_mbuf_allocation_errors=0,rx_priority0_xon_packets=0,rx_fcoe_dropped=0,tx_priority1_xon_to_xoff_packets=0,rx_size_65_to_127_packets=98,rx_q0packets=98,tx_priority0_xoff_packets=0,rx_priority6_xon_packets=0,rx_total_packets=98,rx_undersize_errors=0,flow_director_missed_filters=0,rx_jabber_errors=0,in_pkts_invalid=0,in_pkts_late=0,rx_priority5_xon_packets=0,tx_priority4_xoff_packets=0,out_octets_encrypted=0,tx_q1packets=0,rx_priority5_xoff_packets=0,rx_priority6_xoff_packets=0,rx_errors=0,in_octets_validated=0,rx_priority3_xoff_packets=0,tx_priority4_xon_to_xoff_packets=0,tx_priority5_xon_to_xoff_packets=0,tx_flow_control_xon_packets=0,rx_priority0_dropped=0,flow_director_filter_add_errors=0,tx_q1bytes=0,tx_priority6_xon_to_xoff_packets=0,flow_director_matched_filters=0,tx_priority2_xon_packets=0,rx_fcoe_mbuf_allocation_errors=0,rx_priority2_xoff_packets=0,tx_priority7_xoff_packets=0,rx_priority0_xoff_packets=0,rx_oversize_errors=0,in_pkts_notusingsa=0,tx_size_64_packets=0,rx_size_1024_to_max_packets=0,tx_priority6_xon_packets=0,rx_priority2_dropped=0,rx_priority4_dropped=0,rx_priority7_mbuf_allocation_errors=0,rx_flow_control_xon_packets=0,tx_good_bytes=0,tx_priority3_xon_to_xoff_packets=0,rx_total_missed_packets=0,rx_error_bytes=0,tx_priority7_xon_packets=0,rx_mac_short_packet_dropped=0,tx_priority1_xoff_packets=0,tx_good_packets=0,tx_broadcast_packets=0,tx_xon_packets=0,in_pkts_unusedsa=0,rx_priority2_xon_packets=0,in_pkts_untagged=0,tx_fcoe_packets=0,flow_director_filter_remove_errors=0,rx_priority3_mbuf_allocation_errors=0,tx_priority0_xon_packets=0,rx_priority7_xon_packets=0,rx_fcoe_no_direct_data_placement_ext_buff=0,tx_xoff_packets=0,tx_size_256_to_511_packets=0 1606310780000000000
dpdk,command=/ethdev/link_status,dpdk_instance=l3fwd-power,host=dpdk-host,params=0 status="UP",speed=10000,duplex="full-duplex" 1606310780000000000
dpdk,command=/ethdev/link_status,dpdk_instance=l3fwd-power,host=dpdk-host,params=1 status="UP",speed=10000,duplex="full-duplex" 1606310780000000000
dpdk,command=/l3fwd-power/stats,dpdk_instance=l3fwd-power,host=dpdk-host empty_poll=49506395979901,full_poll=0,busy_percent=0 1606310780000000000
```
