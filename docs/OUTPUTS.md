# Output Plugins

This section is for developers who want to create a new output sink. Outputs
are created in a similar manner as collection plugins, and their interface has
similar constructs.

## Output Plugin Guidelines

- An output must conform to the [telegraf.Output][] interface.
- Outputs should call `outputs.Add` in their `init` function to register
  themselves.  See below for a quick example.
- To be available within Telegraf itself, plugins must register themselves
  using a file in `github.com/influxdata/telegraf/plugins/outputs/all` named
  according to the plugin name. Make sure your also add build-tags to
  conditionally build the plugin.
- Each plugin requires a file called `sample.conf` containing the sample
  configuration  for the plugin in TOML format.
  Please consult the [Sample Config][] page for the latest style guidelines.
- Each plugin `README.md` file should include the `sample.conf` file in a section
  describing the configuration by specifying a `toml` section in the form `toml @sample.conf`. The specified file(s) are then injected automatically into the Readme.
- Follow the recommended [Code Style][].

## Output Plugin Example

Content of your plugin file e.g. `simpleoutput.go`

```go
//go:generate ../../../tools/readme_config_includer/generator
package simpleoutput

// simpleoutput.go

import (
    _ "embed"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/outputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Simple struct {
    Ok  bool            `toml:"ok"`
    Log telegraf.Logger `toml:"-"`
}

func (*Simple) SampleConfig() string {
    return sampleConfig
}

// Init is for setup, and validating config.
func (s *Simple) Init() error {
    return nil
}

func (s *Simple) Connect() error {
    // Make any connection required here
    return nil
}

func (s *Simple) Close() error {
    // Close any connections here.
    // Write will not be called once Close is called, so there is no need to synchronize.
    return nil
}

// Write should write immediately to the output, and not buffer writes
// (Telegraf manages the buffer for you). Returning an error will fail this
// batch of writes and the entire batch will be retried automatically.
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

Registration of the plugin on `plugins/outputs/all/simpleoutput.go`:

```go
//go:build !custom || outputs || outputs.simpleoutput

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/simpleoutput" // register plugin

```

The _build-tags_ in the first line allow to selectively include/exclude your
plugin when customizing Telegraf.

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
[Sample Config]: https://github.com/influxdata/telegraf/blob/master/docs/developers/SAMPLE_CONFIG.md
[Code Style]: https://github.com/influxdata/telegraf/blob/master/docs/developers/CODE_STYLE.md
[telegraf.Output]: https://godoc.org/github.com/influxdata/telegraf#Output
