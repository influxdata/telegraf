# Huawei Telemetry Dialout Input (Huawei MDT Dial-out)

This input plugin passively receives Huawei MDT data pushed by devices via gRPC Dial-out.

## Basic Configuration (example)

```toml
[[inputs.huawei_telemetry_dialout]]
  service_address = "0.0.0.0:57000"
  transport = "grpc"
  # max_msg_size = 4194304
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