# Huawei Telemetry Dialin Input（华为 MDT 主动订阅）

该输入插件通过 gRPC Dialin 主动向设备订阅华为 Model-Driven
Telemetry（MDT）数据。

## 前置要求

- 已放置并生成 GPB 原型：
  - 接口统计：
    `plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_ifm/huawei-ifm.proto`
    → `huawei-ifm.pb.go`
  - 设备管理（CPU/内存）：
    `plugins/parsers/huawei_grpc_gpb/telemetry_proto/huawei_devm/huawei-devm.proto`
    → `huawei-devm.pb.go`
- 在 `plugins/parsers/huawei_grpc_gpb/telemetry_proto/HuaweiTelemetry.go`
  中完成 ProtoPath 映射，例如：
  - `huawei_ifm.Ifm`（1.0）→ `huawei_ifm.Ifm` 类型
  - `huawei_devm.Devm`（1.0）→ `huawei_devm.Devm` 及其子结构（用于 CPU/内存）

> 说明：本仓库已集成上述文件与映射。如新增传感器，请在该文件追加
> `PathKey → []reflect.Type` 映射并重新构建。

## 基本配置（示例）

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

    # 接口统计
    [[inputs.huawei_telemetry_dialin.routers.Paths]]
      depth = 1
      path = 'huawei-ifm:ifm/interfaces/interface/ifStatistics'

    # CPU
    [[inputs.huawei_telemetry_dialin.routers.Paths]]
      depth = 1
      path = 'huawei-devm:devm/cpuInfos/cpuInfo'

    # 内存
    [[inputs.huawei_telemetry_dialin.routers.Paths]]
      depth = 1
      path = 'huawei-devm:devm/memoryInfos/memoryInfo'
```

## 与 Prometheus 集成（推荐处理链）

为适配 Prometheus 的抓取与指标规范，推荐在输出前串联以下处理器：

```toml
# 1) 将接口统计的字符串数值转为数值型
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

# 2) 字段过滤与转标签（示例）
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

# 3) 丢弃头部字段，避免导出 *_current_period
[[processors.filter]]
  namepass = [
    "huawei-ifm:ifm/interfaces/interface/ifStatistics",
    "huawei-devm:devm/cpuInfos/cpuInfo",
    "huawei-devm:devm/memoryInfos/memoryInfo"
  ]
  fieldexclude = ["current_period"]

# 4) 暴露 Prometheus /metrics
[[outputs.prometheus_client]]
  listen = ":9273"
  path = "/metrics"
  metric_version = 2
  export_timestamp = true
```

> 提示：如需简短的 metric 名或字段名，可追加 `[[processors.rename]]`
> 将测量名改为 `huawei_ifm_ifstats` 等，字段改为 `rx_bytes_total`、
> `tx_bytes_total` 等，便于 PromQL。

## 疑难排查

- 指标都挂在一个 metric 上当作 label：通常是数值以字符串形式出现。
  请使用 `processors.converter` 转为数值。
- 出现 `*_current_period` 指标：在 `processors.filter` 中
  `fieldexclude = ["current_period"]`。
- 0 值字段不见了：GPB 解析已启用 `EmitUnpopulated`，若仍缺失请检查
  设备侧是否上报或过滤是否过严。

二、快速去噪（不改设备，仅清理输出）

你现在输出里大量的 `data_gpb.row.N.content` 是解码前的原始内容
（来自头部）。这不是错误，但很吵。两种做法二选一：

1. 配置层面先观察原始字段：沿用 telegraf 调试配置（无处理器），
   确认 counters 是否出现；
2. 或在解析器里丢弃头部 `data_gpb`：在
   `plugins/parsers/huawei_grpc_gpb/parser.go` 的 `Parse` 中，生成
   `headerMap` 后立刻加一行：

   ```go
   delete(headerMap, GpbMsgKeyName) // 即 delete(headerMap, "data_gpb")
   ```

   这样就不会把原始 `data_gpb.*` 合并进字段。