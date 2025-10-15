# Huawei Telemetry Dialin Input Plugin

This input plugin subscribes Huawei Model-Driven Telemetry (MDT) data from
devices via gRPC Dial-in.

⭐ Telegraf v1.37.0
🏷️ networking
💻 all

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin runs as a long-lived service that establishes a gRPC session to
the device and receives streaming telemetry updates.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used
to modify metrics, tags, and field or create aliases and configure ordering,
etc. See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Prerequisites

- Protobuf definitions and generated GPB code are present:
  - Interface statistics:
    `plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_ifm/huawei-ifm.proto`
    → `huawei-ifm.pb.go`
  - Device management (CPU/Memory):
    `plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_devm/huawei-devm.proto`
    → `huawei-devm.pb.go`
- Configure ProtoPath mapping in
  `plugins/parsers/huawei_grpc_gpb/telemetry_proto/HuaweiTelemetry.go`, e.g.:
  - `huawei_ifm.Ifm` (1.0) → Go type `huawei_ifm.Ifm`
  - `huawei_devm.Devm` (1.0) → Go type `huawei_devm.Devm` (and nested CPU/Memory)

> Note: The repository already contains the above files and mappings. For new
> sensors, add a `PathKey → []reflect.Type` mapping and rebuild.

## Configuration

```toml @sample.conf
```

### Example configuration

```toml
[[inputs.huawei_telemetry_dialin]]
  [[inputs.huawei_telemetry_dialin.routers]]
    address = "<device-ip>:20000"
    sample_interval = 1000
    encoding = "gpb"
    suppress_redundant = true
    request_id = 1024

    [inputs.huawei_telemetry_dialin.routers.aaa]
      username = "<user>"
      password = "<pass>"

    # Interface statistics
    [[inputs.huawei_telemetry_dialin.routers.Paths]]
      depth = 1
      path = 'huawei-ifm:ifm/interfaces/interface/ifStatistics'

    # CPU
    [[inputs.huawei_telemetry_dialin.routers.Paths]]
      depth = 1
      path = 'huawei-devm:devm/cpuInfos/cpuInfo'

    # Memory
    [[inputs.huawei_telemetry_dialin.routers.Paths]]
      depth = 1
      path = 'huawei-devm:devm/memoryInfos/memoryInfo'
```

## Metrics

The plugin emits measurements derived from Huawei MDT sensors. Typical
measurements include but are not limited to:

- huawei-ifm:ifm/interfaces/interface/ifStatistics
  - tags: node_id_str, interfaces.interface.0.ifName, host
  - fields: receiveByte, receivePacket, rcvUniPacket, rcvMutiPacket,
    rcvBroadPacket, rcvErrorPacket, sendByte, sendPacket,
    in_realtime_bit_rate, out_realtime_bit_rate
- huawei-devm:devm/cpuInfos/cpuInfo
  - tags: node_id_str, host
  - fields: systemCpuUsage
- huawei-devm:devm/memoryInfos/memoryInfo
  - tags: node_id_str, host
  - fields: osMemoryTotal, osMemoryFree, osMemoryUsage

## Example Output

Example (Influx Line Protocol) before any processors are applied:

```text
huawei-ifm:ifm/interfaces/interface/ifStatistics,node_id_str=Switch \
interfaces.interface.0.ifStatistics.receiveByte="0", \
interfaces.interface.0.ifStatistics.sendByte="0" 1760450787711000000
huawei-devm:devm/cpuInfos/cpuInfo,node_id_str=Switch \
cpuInfos.cpuInfo.0.systemCpuUsage=13i 1760450787730000000
huawei-devm:devm/memoryInfos/memoryInfo,node_id_str=Switch \
memoryInfos.memoryInfo.0.osMemoryTotal=3913872i, \
memoryInfos.memoryInfo.0.osMemoryFree=2316916i 1760450787632000000
```

## Prometheus Integration (recommended chain)

To fit Prometheus' scrape model and metric conventions, chain the following
processors before outputs:

```toml
# 1) Convert stringified counters to numeric values (interface statistics)
[[processors.converter]]
  namepass = ["huawei-ifm:ifm/interfaces/interface/ifStatistics"]
  [processors.converter.fields]
    integer = [
      "interfaces.interface.0.ifStatistics.receiveByte",
      "interfaces.interface.0.ifStatistics.receivePacket",
      "interfaces.interface.0.ifStatistics.rcvUniPacket",
      "interfaces.interface.0.ifStatistics.rcvMutiPacket",
      "interfaces.interface.0.ifStatistics.rcvBroadPacket",
      "interfaces.interface.0.ifStatistics.rcvErrorPacket",
      "interfaces.interface.0.ifStatistics.sendByte",
      "interfaces.interface.0.ifStatistics.sendPacket",
      "interfaces.interface.0.ifStatistics.in_realtime_bit_rate",
      "interfaces.interface.0.ifStatistics.out_realtime_bit_rate"
    ]

# 2) Field filtering and field-to-tag conversion (example)
[[processors.metric_match]]
  namepass = ["huawei-ifm:ifm/interfaces/interface/ifStatistics"]
  [processors.metric_match.approach]
  approach = "include"
  [processors.metric_match.tag]
  "huawei-ifm:ifm/interfaces/interface/ifStatistics" = ["node_id_str"]
  [processors.metric_match.field_filter]
  "huawei-ifm:ifm/interfaces/interface/ifStatistics" = [
    "receiveByte",
    "receivePacket",
    "rcvUniPacket",
    "rcvMutiPacket",
    "rcvBroadPacket",
    "rcvErrorPacket",
    "sendByte",
    "sendPacket",
    "in_realtime_bit_rate",
    "out_realtime_bit_rate"
  ]

# 3) Drop header fields to avoid exporting *_current_period
[[processors.filter]]
  namepass = [
    "huawei-ifm:ifm/interfaces/interface/ifStatistics",
    "huawei-devm:devm/cpuInfos/cpuInfo",
    "huawei-devm:devm/memoryInfos/memoryInfo"
  ]
  fieldexclude = ["current_period"]

# 4) Expose Prometheus /metrics
[[outputs.prometheus_client]]
  listen = ":9273"
  path = "/metrics"
  metric_version = 2
  export_timestamp = true
```

> Tip: If you prefer shorter metric names, use `[[processors.rename]]` to rename
> the measurement (e.g., `huawei_ifm_ifstats`) and fields (e.g.,
> `rx_bytes_total`, `tx_bytes_total`).

## Troubleshooting

- Metrics appear as labels on a single time series: the values were strings.
  Use `processors.converter` to convert to numeric.
- `*_current_period` shows up as metrics: use `processors.filter` with
  `fieldexclude = ["current_period"]`.
- Zero-value fields missing: GPB parsing enables `EmitUnpopulated`, but verify
  the device actually reports the fields and filters are not too strict.

### Quick noise reduction (no device changes)

You might see many `data_gpb.row.N.content` fields (raw, pre-decoded header
payload). This is not an error but noisy. Two options:

1. Observe raw fields first: temporarily use a debug config without processors
   to validate counters appear.
2. Or drop the header `data_gpb` in the parser. In
   `plugins/parsers/huawei_grpc_gpb/parser.go`, after creating `headerMap`
   inside `Parse`, add:
   
   ```go
   delete(headerMap, GpbMsgKeyName) // i.e. delete(headerMap, "data_gpb")
   ```
   
   This prevents merging raw `data_gpb.*` into fields.



