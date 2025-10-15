# Huawei Telemetry Dialout Input Plugin

This input plugin passively receives Huawei MDT data pushed by devices via gRPC
Dial-out.

⭐ Telegraf v1.37.0
🏷️ networking
💻 all

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input listening for incoming gRPC connections and
streaming telemetry updates pushed by the device.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
```

### Example configuration

```toml
[[inputs.huawei_telemetry_dialout]]
  service_address = "0.0.0.0:57000"
  transport = "grpc"
  # max_msg_size = 4194304
```

## Metrics

See the Dialin plugin for a list of typical measurements and fields.
Dial-out shares the same data schema as Dial-in.

## Example Output

Example (Influx Line Protocol):

```text
huawei-ifm:ifm/interfaces/interface/ifStatistics,node_id_str=Switch \
interfaces.interface.0.ifStatistics.receiveByte="0" 1760450787711000000
```

## Prometheus Integration (recommended chain)

Dial-out shares the same data schema as Dial-in. Reuse the same processors:

```toml
# Example: only key fragments shown. See Dialin README for the full chain.
[[processors.converter]]
  namepass = ["huawei-ifm:ifm/interfaces/interface/ifStatistics"]
  [processors.converter.fields]
    integer = [
      "interfaces.interface.0.ifStatistics.receiveByte",
      "interfaces.interface.0.ifStatistics.sendByte"
    ]

[[processors.metric_match]]
  namepass = ["huawei-ifm:ifm/interfaces/interface/ifStatistics"]
  [processors.metric_match.approach]
  approach = "include"

[[processors.filter]]
  namepass = [
    "huawei-ifm:ifm/interfaces/interface/ifStatistics",
    "huawei-devm:devm/cpuInfos/cpuInfo",
    "huawei-devm:devm/memoryInfos/memoryInfo"
  ]
  fieldexclude = ["current_period"]

[[outputs.prometheus_client]]
  listen = ":9273"
  path = "/metrics"
  metric_version = 2
  export_timestamp = true
```

## Notes

- Parsers are self-registered in `huawei_grpc_gpb` and `huawei_grpc_json`; no
  manual registration required.
- When adding new business sensors, extend
  `plugins/parsers/huawei_grpc_gpb/telemetry_proto/HuaweiTelemetry.go` with the
  appropriate `ProtoPath → Go type` mapping and rebuild.
