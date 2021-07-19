# Test Input Plugin

The test plugin parses passed metrics in the config file.

**Note:** If you wish to parse a **complete** file then use the [file][] input 
plugin instead.

**Note:** If you wish to parse only newly appended lines use the [tail][] input
plugin instead.

### Configuration:

```toml
[[inputs.test]]
  ## Metrics to parse each interval.
  metrics = [
    'weather,state=ny temperature=81.3',
    'weather,state=ca temperature=75.1'
  ]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

```

[input data format]: /docs/DATA_FORMATS_INPUT.md
[file]: /plugins/inputs/file
[tail]: /plugins/inputs/tail
