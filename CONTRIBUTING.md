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

## Sign the CLA

Before we can merge a pull request, you will need to sign the CLA,
which can be found [on our website](http://influxdb.com/community/cla.html)

## Adding a dependency

Assuming you can already build the project, run these in the telegraf directory:

1. `go get github.com/sparrc/gdm`
1. `gdm restore`
1. `gdm save`

## Input Plugins

This section is for developers who want to create new collection inputs.
Telegraf is entirely plugin driven. This interface allows for operators to
pick and chose what is gathered as well as makes it easy for developers
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

### Input interface

```go
type Input interface {
    SampleConfig() string
    Description() string
    Gather(Accumulator) error
}

type Accumulator interface {
    Add(measurement string,
        value interface{},
        tags map[string]string,
        timestamp ...time.Time)
    AddFields(measurement string,
        fields map[string]interface{},
        tags map[string]string,
        timestamp ...time.Time)
}
```

### Accumulator

The way that a plugin emits metrics is by interacting with the Accumulator.

The `Add` function takes 3 arguments:
* **measurement**: A string description of the metric. For instance `bytes_read` or `
faults`.
* **value**: A value for the metric. This accepts 5 different types of value:
  * **int**: The most common type. All int types are accepted but favor using `int64`
  Useful for counters, etc.
  * **float**: Favor `float64`, useful for gauges, percentages, etc.
  * **bool**: `true` or `false`, useful to indicate the presence of a state. `light_on`,
  etc.
  * **string**: Typically used to indicate a message, or some kind of freeform
  information.
  * **time.Time**: Useful for indicating when a state last occurred, for instance `
  light_on_since`.
* **tags**: This is a map of strings to strings to describe the where or who
about the metric. For instance, the `net` plugin adds a tag named `"interface"`
set to the name of the network interface, like `"eth0"`.

Let's say you've written a plugin that emits metrics about processes on the current host.

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

func (s *Simple) Gather(acc inputs.Accumulator) error {
    if s.Ok {
        acc.Add("state", "pretty good", nil)
    } else {
        acc.Add("state", "not great", nil)
    }

    return nil
}

func init() {
    inputs.Add("simple", func() telegraf.Input { return &Simple{} })
}
```

## Input Plugins Accepting Arbitrary Data Formats

Some input plugins (such as
[exec](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/exec))
accept arbitrary input data formats. An overview of these data formats can
be found
[here](https://github.com/influxdata/telegraf/blob/master/DATA_FORMATS_INPUT.md).

In order to enable this, you must specify a `SetParser(parser parsers.Parser)`
function on the plugin object (see the exec plugin for an example), as well as
defining `parser` as a field of the object.

You can then utilize the parser internally in your plugin, parsing data as you
see fit. Telegraf's configuration layer will take care of instantiating and
creating the `Parser` object.

You should also add the following to your SampleConfig() return:

```toml
  ### Data format to consume. This can be "json", "influx" or "graphite"
  ### Each data format has it's own unique set of configuration options, read
  ### more about them here:
  ### https://github.com/influxdata/telegraf/blob/master/DATA_FORMATS_INPUT.md
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

### Service Plugin interface

```go
type ServicePlugin interface {
    SampleConfig() string
    Description() string
    Gather(Accumulator) error
    Start() error
    Stop()
}
```

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

### Output interface

```go
type Output interface {
    Connect() error
    Close() error
    Description() string
    SampleConfig() string
    Write(metrics []telegraf.Metric) error
}
```

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
    for _, pt := range points {
        // write `pt` to the output sink here
    }
    return nil
}

func init() {
    outputs.Add("simpleoutput", func() telegraf.Output { return &Simple{} })
}

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

### Service Output interface

```go
type ServiceOutput interface {
    Connect() error
    Close() error
    Description() string
    SampleConfig() string
    Write(metrics []telegraf.Metric) error
    Start() error
    Stop()
}
```

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
(i.e: https://github.com/influxdata/telegraf/blob/master/plugins/redis/redis_test.go)
a simple mock will suffice.

To execute Telegraf tests follow these simple steps:

- Install docker following [these](https://docs.docker.com/installation/)
instructions
- execute `make test`

**OSX users**: you will need to install `boot2docker` or `docker-machine`.
The Makefile will assume that you have a `docker-machine` box called `default` to
get the IP address.

### Unit test troubleshooting

Try cleaning up your test environment by executing `make docker-kill` and
re-running
