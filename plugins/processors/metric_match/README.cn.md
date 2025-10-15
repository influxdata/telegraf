# Metric Match Processor（字段过滤与转标签）

用于华为 MDT 指标的字段筛选与将特定字段转换为标签（Prometheus labels）。

## 用法

```toml
[[processors.metric_match]]
  namepass = ["huawei-ifm:ifm/interfaces/interface/ifStatistics"]

  [processors.metric_match.approach]
  approach = "include"  # include: 仅保留白名单；exclude: 仅排除黑名单

  # 将字段转为标签（避免字符串字段被导出为指标）
  [processors.metric_match.tag]
  "huawei-ifm:ifm/interfaces/interface/ifStatistics" = ["node_id_str"]

  # 字段过滤（按字段名后缀匹配，不含点的顶层字段默认保留）
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
```

## 典型组合

- 与 `processors.converter` 配合：把字符串型数值先转为数值，Prometheus
  才会导出为 metric 而非 label。
- 与 `processors.filter` 配合：例如 `fieldexclude = ["current_period"]`
  丢弃头部字段，避免导出 `*_current_period` 指标。
- 与 `processors.rename` 配合：重命名测量和字段，得到更短的指标名
  （如 `huawei_ifm_ifstats_rx_bytes_total`）。