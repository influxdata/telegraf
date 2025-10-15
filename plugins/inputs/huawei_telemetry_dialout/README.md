# Huawei Telemetry Dialout Input（华为 MDT 被动推送）

该输入插件通过 gRPC Dialout 被动接收设备推送的华为 MDT 数据。

## 基本配置（示例）

```toml
[[inputs.huawei_telemetry_dialout]]
  service_address = "0.0.0.0:57000"
  transport = "grpc"
  # max_msg_size = 4194304
```

## 与 Prometheus 集成（推荐处理链）

Dialout 与 Dialin 的数据结构一致，可复用相同的处理器链：

```toml
# 示例，仅展示关键片段（详见 Dialin README 中的完整处理链）
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

## 说明

- 解析器已在 `huawei_grpc_gpb` 与 `huawei_grpc_json` 包内自注册，无需手动注册。
- 若新增业务传感器，请在 `plugins/parsers/huawei_grpc_gpb/telemetry_proto/HuaweiTelemetry.go` 中增加 `ProtoPath → Go 类型` 映射并重新构建。
