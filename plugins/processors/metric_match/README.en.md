# Metric Match Processor (Field filtering and field-to-tag)

`metric_match` filters fields and converts selected fields to tags for Huawei MDT metrics.

## Usage

```toml
[[processors.metric_match]]
  namepass = ["huawei-ifm:ifm/interfaces/interface/ifStatistics"]

  [processors.metric_match.approach]
  approach = "include"  # include: keep whitelist; exclude: drop blacklist

  # Convert fields to tags (avoid exporting string fields as metrics)
  [processors.metric_match.tag]
  "huawei-ifm:ifm/interfaces/interface/ifStatistics" = ["node_id_str"]

  # Field filtering (match by field-name suffix; top-level fields without dots are kept by default)
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

## Typical combinations

- With `processors.converter`: convert stringified numbers to numeric so Prometheus exports them as metrics instead of labels.
- With `processors.filter`: e.g., `fieldexclude = ["current_period"]` to drop header fields and avoid `*_current_period` metrics.
- With `processors.rename`: rename measurement and fields for shorter identifiers (e.g., `huawei_ifm_ifstats_rx_bytes_total`).



