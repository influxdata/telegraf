# Benchmark Input Plugin

The benchmark plugin reads metrics from file and continuously adds them as
quickly as possible.  It can be used for benchmarking Telegraf performance as
well as the output sources.

### Configuration:

```toml
# Generate test data for performance testing
[[inputs.benchmark]]
  ## File containing input data
  filename = "/tmp/testdata"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Metrics:

The metrics in the input file are added without modification.
