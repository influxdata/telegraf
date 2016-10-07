## Steps for Contributing:

1. [Sign the CLA](http://influxdb.com/community/cla.html)
1. Make changes or write plugin (see below for details)
1. Add your plugin to `plugins/inputs/all/all.go` or `plugins/outputs/all/all.go`
1. If your plugin requires a new Go package,
[add it](https://github.com/influxdata/telegraf/blob/master/CONTRIBUTING.md#adding-a-dependency)
1. Write a README for your plugin, if it's an input plugin, it should be structured
like the [input example here](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/EXAMPLE_README.md).
Output plugins READMEs are less structured,
but any information you can provide on how the data will look is appreciated.
See the [OpenTSDB output](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/opentsdb)
for a good example.
1. **Optional:** Help users of your plugin by including example queries for populating dashboards. Include these sample queries in the `README.md` for the plugin.
1. **Optional:** Write a [tickscript](https://docs.influxdata.com/kapacitor/v1.0/tick/syntax/) for your plugin and add it to [Kapacitor](https://github.com/influxdata/kapacitor/tree/master/examples/telegraf). Or mention @jackzampolin in a PR comment with some common queries that you would want to alert on and he will write one for you.

## GoDoc

Public interfaces for inputs, outputs, metrics, and the accumulator can be found
on the GoDoc

[![GoDoc](https://godoc.org/github.com/influxdata/telegraf?status.svg)](https://godoc.org/github.com/influxdata/telegraf)

## Sign the CLA

Before we can merge a pull request, you will need to sign the CLA,
which can be found [on our website](http://influxdb.com/community/cla.html)

## Adding a dependency

Assuming you can already build the project, run these in the telegraf directory:

1. `go get github.com/sparrc/gdm`
1. `gdm restore`
1. `GOOS=linux gdm save`

## Input Plugins

This section is for developers who want to create new collection inputs.
Telegraf is entirely plugin driven. This interface allows for operators to
pick and chose what is gathered and makes it easy for developers
to create new ways of generating metrics.

Plugin authorship is kept as simple as possible to promote people to develop
and submit new inputs.

### Input Plugin Guidelines

* A plugin must conform to the `telegraf.Input` interface.
* Input Plugins should call `inputs.Add` in their `init` function to register themselves.
See below for a quick example.
* Input Plugins must be added to the
`github.com/influxdata/telegraf/plugins/inputs/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
plugin can be configured. This is include in `telegraf -sample-config`.
* The `Description` function should say in one line what this plugin does.

Let's say you've written a plugin that emits metrics about processes on the
current host.

### Input Plugin Example

```go
package simple

// simple.go

import (
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
)

type Simple struct {
    Ok bool
}

func (s *Simple) Description() string {
    return "a demo plugin"
}

func (s *Simple) SampleConfig() string {
    return "ok = true # indicate if everything is fine"
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

## Adding Typed Metrics

In addition the the `AddFields` function, the accumulator also supports an
`AddGauge` and `AddCounter` function. These functions are for adding _typed_
metrics. Metric types are ignored for the InfluxDB output, but can be used
for other outputs, such as [prometheus](https://prometheus.io/docs/concepts/metric_types/).

## Input Plugins Accepting Arbitrary Data Formats

Some input plugins (such as
[exec](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/exec))
accept arbitrary input data formats. An overview of these data formats can
be found
[here](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md).

In order to enable this, you must specify a `SetParser(parser parsers.Parser)`
function on the plugin object (see the exec plugin for an example), as well as
defining `parser` as a field of the object.

You can then utilize the parser internally in your plugin, parsing data as you
see fit. Telegraf's configuration layer will take care of instantiating and
creating the `Parser` object.

You should also add the following to your SampleConfig() return:

```toml
  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

Below is the `Parser` interface.

```go
// Parser is an interface defining functions that a parser plugin must satisfy.
type Parser interface {
    // Parse takes a byte buffer separated by newlines
    // ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
    // and parses it into telegraf metrics
    Parse(buf []byte) ([]telegraf.Metric, error)

    // ParseLine takes a single string metric
    // ie, "cpu.usage.idle 90"
    // and parses it into a telegraf metric.
    ParseLine(line string) (telegraf.Metric, error)
}
```

And you can view the code
[here.](https://github.com/influxdata/telegraf/blob/henrypfhu-master/plugins/parsers/registry.go)

## Service Input Plugins

This section is for developers who want to create new "service" collection
inputs. A service plugin differs from a regular plugin in that it operates
a background service while Telegraf is running. One example would be the `statsd`
plugin, which operates a statsd server.

Service Input Plugins are substantially more complicated than a regular plugin, as they
will require threads and locks to verify data integrity. Service Input Plugins should
be avoided unless there is no way to create their behavior with a regular plugin.

Their interface is quite similar to a regular plugin, with the addition of `Start()`
and `Stop()` methods.

### Service Plugin Guidelines

* Same as the `Plugin` guidelines, except that they must conform to the
`inputs.ServiceInput` interface.

## Output Plugins

This section is for developers who want to create a new output sink. Outputs
are created in a similar manner as collection plugins, and their interface has
similar constructs.

### Output Plugin Guidelines

* An output must conform to the `outputs.Output` interface.
* Outputs should call `outputs.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
`github.com/influxdata/telegraf/plugins/outputs/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
output can be configured. This is include in `telegraf -sample-config`.
* The `Description` function should say in one line what this output does.

### Output Example

```go
package simpleoutput

// simpleoutput.go

import (
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/outputs"
)

type Simple struct {
    Ok bool
}

func (s *Simple) Description() string {
    return "a demo output"
}

func (s *Simple) SampleConfig() string {
    return "url = localhost"
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

## Output Plugins Writing Arbitrary Data Formats

Some output plugins (such as
[file](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/file))
can write arbitrary output data formats. An overview of these data formats can
be found
[here](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md).

In order to enable this, you must specify a
`SetSerializer(serializer serializers.Serializer)`
function on the plugin object (see the file plugin for an example), as well as
defining `serializer` as a field of the object.

You can then utilize the serializer internally in your plugin, serializing data
before it's written. Telegraf's configuration layer will take care of
instantiating and creating the `Serializer` object.

You should also add the following to your SampleConfig() return:

```toml
  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```

## Service Output Plugins

This section is for developers who want to create new "service" output. A
service output differs from a regular output in that it operates a background service
while Telegraf is running. One example would be the `prometheus_client` output,
which operates an HTTP server.

Their interface is quite similar to a regular output, with the addition of `Start()`
and `Stop()` methods.

### Service Output Guidelines

* Same as the `Output` guidelines, except that they must conform to the
`output.ServiceOutput` interface.

## Unit Tests

### Execute short tests

execute `make test-short`

### Execute long tests

As Telegraf collects metrics from several third-party services it becomes a
difficult task to mock each service as some of them have complicated protocols
which would take some time to replicate.

To overcome this situation we've decided to use docker containers to provide a
fast and reproducible environment to test those services which require it.
For other situations
(i.e: https://github.com/influxdata/telegraf/blob/master/plugins/inputs/redis/redis_test.go)
a simple mock will suffice.

To execute Telegraf tests follow these simple steps:

- Install docker following [these](https://docs.docker.com/installation/)
instructions
- execute `make test`

### Unit test troubleshooting

Try cleaning up your test environment by executing `make docker-kill` and
re-running
