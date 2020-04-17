### Output Plugins

This section is for developers who want to create a new output sink. Outputs
are created in a similar manner as collection plugins, and their interface has
similar constructs.

### Output Plugin Guidelines

- An output must conform to the [telegraf.Output][] interface.
- Outputs should call `outputs.Add` in their `init` function to register
  themselves.  See below for a quick example.
- To be available within Telegraf itself, plugins must add themselves to the
  `github.com/influxdata/telegraf/plugins/outputs/all/all.go` file.
- The `SampleConfig` function should return valid toml that describes how the
  plugin can be configured. This is included in `telegraf config`.  Please
  consult the [SampleConfig][] page for the latest style guidelines.
- The `Description` function should say in one line what this output does.
- Follow the recommended [CodeStyle][].

### Output Plugin Example

```go
package simpleoutput

// simpleoutput.go

import (
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/outputs"
)

type Simple struct {
    Ok bool `toml:"ok"`
}

func (s *Simple) Description() string {
    return "a demo output"
}

func (s *Simple) SampleConfig() string {
    return `
  ok = true
`
}

func (s *Simple) Init() error {
	return nil
}

func (s *Simple) Connect() error {
    // Make a connection to the URL here
    return nil
}

func (s *Simple) Close() error {
    // Close connection to the URL here
    return nil
}

func (s *Simple) Write(metrics []telegraf.Metric) error {
    for _, metric := range metrics {
        // write `metric` to the output sink here
    }
    return nil
}

func init() {
    outputs.Add("simpleoutput", func() telegraf.Output { return &Simple{} })
}

```

## Data Formats

Some output plugins, such as the [file][] plugin, can write in any supported
[output data formats][].

In order to enable this, you must specify a
`SetSerializer(serializer serializers.Serializer)`
function on the plugin object (see the file plugin for an example), as well as
defining `serializer` as a field of the object.

You can then utilize the serializer internally in your plugin, serializing data
before it's written. Telegraf's configuration layer will take care of
instantiating and creating the `Serializer` object.

You should also add the following to your `SampleConfig()`:

```toml
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## Flushing Metrics to Outputs

Metrics are flushed to outputs when any of the following events happen:
- `flush_interval + rand(flush_jitter)` has elapsed since start or the last flush interval
- At least `metric_batch_size` count of metrics are waiting in the buffer
- The telegraf process has received a SIGUSR1 signal

Note that if the flush takes longer than the `agent.interval` to write the metrics
to the output, you'll see a message saying the output `did not complete within its 
flush interval`. This may mean your output is not keeping up with the flow of metrics, 
and you may want to look into enabling compression, reducing the size of your metrics, 
or investigate other reasons why the writes might be taking longer than expected.

[file]: https://github.com/influxdata/telegraf/tree/master/plugins/inputs/file
[output data formats]: https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
[SampleConfig]: https://github.com/influxdata/telegraf/wiki/SampleConfig
[CodeStyle]: https://github.com/influxdata/telegraf/wiki/CodeStyle
[telegraf.Output]: https://godoc.org/github.com/influxdata/telegraf#Output
