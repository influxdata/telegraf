# Input Plugins

This section is for developers who want to create new collection inputs.
Telegraf is entirely plugin driven. This interface allows for operators to
pick and chose what is gathered and makes it easy for developers
to create new ways of generating metrics.

Plugin authorship is kept as simple as possible to promote people to develop
and submit new inputs.

## Input Plugin Guidelines

- A plugin must conform to the [telegraf.Input][] interface.
- Input Plugins should call `inputs.Add` in their `init` function to register
  themselves.  See below for a quick example.
- To be available within Telegraf itself, plugins must register themselves
  using a file in `github.com/influxdata/telegraf/plugins/inputs/all` named
  according to the plugin name. Make sure you also add build-tags to
  conditionally build the plugin.
- Each plugin requires a file called `sample.conf` containing the sample
  configuration  for the plugin in TOML format.
  Please consult the [Sample Config][] page for the latest style guidelines.
- Each plugin `README.md` file should include the `sample.conf` file in a
  section describing the configuration by specifying a `toml` section in the
  form `toml @sample.conf`. The specified file(s) are then injected
  automatically into the Readme.
- Follow the recommended [Code Style][].

[Sample Config]: /docs/developers/SAMPLE_CONFIG.md
[Code Style]: /docs/developers/CODE_STYLE.md
[telegraf.Input]: https://godoc.org/github.com/influxdata/telegraf#Input

### Typed Metrics

In addition to the `AddFields` function, the accumulator also supports
functions to add typed metrics: `AddGauge`, `AddCounter`, etc.  Metric types
are ignored by the InfluxDB output, but can be used for other outputs, such as
[prometheus][prom metric types].

[prom metric types]: https://prometheus.io/docs/concepts/metric_types/

### Data Formats

Some input plugins, such as the [exec][] plugin, can accept any supported
[input data formats][].

In order to enable this, you must specify a `SetParser(parser parsers.Parser)`
function on the plugin object (see the exec plugin for an example), as well as
defining `parser` as a field of the object.

You can then utilize the parser internally in your plugin, parsing data as you
see fit. Telegraf's configuration layer will take care of instantiating and
creating the `Parser` object.

Add the following to the sample configuration in the README.md:

```toml
  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

[exec]: /plugins/inputs/exec
[input data formats]: /docs/DATA_FORMATS_INPUT.md

### Service Input Plugins

This section is for developers who want to create new "service" collection
inputs. A service plugin differs from a regular plugin in that it operates a
background service while Telegraf is running. One example would be the
`statsd` plugin, which operates a statsd server.

Service Input Plugins are substantially more complicated than a regular
plugin, as they will require threads and locks to verify data integrity.
Service Input Plugins should be avoided unless there is no way to create their
behavior with a regular plugin.

To create a Service Input implement the [telegraf.ServiceInput][] interface.

[telegraf.ServiceInput]: https://godoc.org/github.com/influxdata/telegraf#ServiceInput

### Metric Tracking

Metric Tracking provides a system to be notified when metrics have been
successfully written to their outputs or otherwise discarded.  This allows
inputs to be created that function as reliable queue consumers.

To get started with metric tracking begin by calling `WithTracking` on the
[telegraf.Accumulator][].  Add metrics using the `AddTrackingMetricGroup`
function on the returned [telegraf.TrackingAccumulator][] and store the
`TrackingID`.  The `Delivered()` channel will return a type with information
about the final delivery status of the metric group.

Check the [amqp_consumer][] for an example implementation.

[telegraf.Accumulator]: https://godoc.org/github.com/influxdata/telegraf#Accumulator
[telegraf.TrackingAccumulator]: https://godoc.org/github.com/influxdata/telegraf#Accumulator
[amqp_consumer]: /plugins/inputs/amqp_consumer

### External Services

Plugins that connect or require the use of external services should ensure that
those servers are active. When may depend on the type of input plugin:

For service input plugins, `Init` should be used to check for configuration
issues (e.g. bad option) and for other non-recoverable errors. Then `Start`
is used to create connections or other retry-able operations.

For normal inputs, `Init` should also be used to check for configuration issues
as well as any other dependencies that the plugin will require. For example,
any binaries that must exist for the plugin to function. If making a connection,
this should also take place in `Init`.

Developers may find that they switch to using service input plugins more and
more to take advantage of the error on retry behavior features. This allows
the user to decide what to do on an error, like ignoring the plugin or retrying
constantly.

## Input Plugin Example

Let's say you've written a plugin that emits metrics about processes on the
current host.

### Register Plugin

Registration of the plugin on `plugins/inputs/all/simple.go`:

```go
//go:build !custom || inputs || inputs.simple

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/simple" // register plugin
```

The _build-tags_ in the first line allow to selectively include/exclude your
plugin when customizing Telegraf.

### Plugin

Content of your plugin file e.g. `simple.go`

```go
//go:generate ../../../tools/readme_config_includer/generator
package simple

import (
    _ "embed"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
)

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

func (s *Simple) Gather(acc telegraf.Accumulator) error {
    if s.Ok {
        acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
    } else {
        acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
    }

    return nil
}

func init() {
    inputs.Add("simple", func() telegraf.Input { return &Simple{} })
}
```
