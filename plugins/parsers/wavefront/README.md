# Wavefront

Wavefront Data Format is metrics are parsed directly into Telegraf metrics.
For more information about the Wavefront Data Format see
[here](https://docs.wavefront.com/wavefront_data_format.html).

### Configuration

There are no additional configuration options for Wavefront Data Format line-protocol.

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "wavefront"
```
