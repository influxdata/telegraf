# Metric Match Processor Plugin

`metric_match` filters fields and converts selected fields to tags for Huawei MDT metrics.

⭐ Telegraf v1.37.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
```

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



