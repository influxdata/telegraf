## Steps for Contributing:

1. [Sign the CLA](http://influxdb.com/community/cla.html)
1. Make changes or write plugin (see below for details)
1. Add your plugin to one of: `plugins/{inputs,outputs,aggregators,processors}/all/all.go`
1. If your plugin requires a new Go package,
[add it](https://github.com/influxdata/telegraf/blob/master/CONTRIBUTING.md#adding-a-dependency)
1. Write a README for your plugin, if it's an input plugin, it should be structured
like the [input example here](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/EXAMPLE_README.md).
Output plugins READMEs are less structured,
but any information you can provide on how the data will look is appreciated.
See the [OpenTSDB output](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/opentsdb)
for a good example.
1. **Optional:** Help users of your plugin by including example queries for populating dashboards. Include these sample queries in the `README.md` for the plugin.
1. **Optional:** Write a [tickscript](https://docs.influxdata.com/kapacitor/v1.0/tick/syntax/) for your plugin and add it to [Kapacitor](https://github.com/influxdata/kapacitor/tree/master/examples/telegraf).

## GoDoc

Public interfaces for inputs, outputs, processors, aggregators, metrics,
and the accumulator can be found on the GoDoc

[![GoDoc](https://godoc.org/github.com/influxdata/telegraf?status.svg)](https://godoc.org/github.com/influxdata/telegraf)

## Sign the CLA

Before we can merge a pull request, you will need to sign the CLA,
which can be found [on our website](http://influxdb.com/community/cla.html)

## Adding a dependency

Assuming you can already build the project, run these in the telegraf directory:

1. `dep ensure -vendor-only`
2. `dep ensure -add github.com/[dependency]/[new-package]`

## Input Plugins

This section is for developers who want to create new collection inputs.
Telegraf is entirely plugin driven. This interface allows for operators to
pick and chose what is gathered and makes it easy for developers
to create new ways of generating metrics.

Plugin authorship is kept as simple as possible to promote people to develop
and submit new inputs.

### Input Plugin Guidelines

* A plugin must conform to the [`telegraf.Input`](https://godoc.org/github.com/influxdata/telegraf#Input) interface.
* Input Plugins should call `inputs.Add` in their `init` function to register themselves.
See below for a quick example.
* Input Plugins must be added to the
`github.com/influxdata/telegraf/plugins/inputs/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
plugin can be configured. This is included in `telegraf config`.  Please
consult the [SampleConfig](https://github.com/influxdata/telegraf/wiki/SampleConfig)
page for the latest style guidelines.
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
    return `
  ## Indicate if everything is fine
  ok = true
`
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

### Input Plugin Development

* Run `make static` followed by `make plugin-[pluginName]` to spin up a docker dev environment
using docker-compose.
* ***[Optional]*** When developing a plugin, add a `dev` directory with a `docker-compose.yml` and `telegraf.conf`
as well as any other supporting files, where sensible.

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
  ## Each data format has its own unique set of configuration options, read
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
[`telegraf.ServiceInput`](https://godoc.org/github.com/influxdata/telegraf#ServiceInput) interface.

## Output Plugins

This section is for developers who want to create a new output sink. Outputs
are created in a similar manner as collection plugins, and their interface has
similar constructs.

### Output Plugin Guidelines

* An output must conform to the [`telegraf.Output`](https://godoc.org/github.com/influxdata/telegraf#Output) interface.
* Outputs should call `outputs.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
`github.com/influxdata/telegraf/plugins/outputs/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
plugin can be configured. This is included in `telegraf config`.  Please
consult the [SampleConfig](https://github.com/influxdata/telegraf/wiki/SampleConfig)
page for the latest style guidelines.
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
    return `
  ok = true
`
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
  ## Each data format has its own unique set of configuration options, read
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

## Processor Plugins

This section is for developers who want to create a new processor plugin.

### Processor Plugin Guidelines

* A processor must conform to the [`telegraf.Processor`](https://godoc.org/github.com/influxdata/telegraf#Processor) interface.
* Processors should call `processors.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
`github.com/influxdata/telegraf/plugins/processors/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
processor can be configured. This is include in the output of `telegraf config`.
* The `Description` function should say in one line what this processor does.

### Processor Example

```go
package printer

// printer.go

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Printer struct {
}

var sampleConfig = `
`

func (p *Printer) SampleConfig() string {
	return sampleConfig
}

func (p *Printer) Description() string {
	return "Print all metrics that pass through this filter."
}

func (p *Printer) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		fmt.Println(metric.String())
	}
	return in
}

func init() {
	processors.Add("printer", func() telegraf.Processor {
		return &Printer{}
	})
}
```

## Aggregator Plugins

This section is for developers who want to create a new aggregator plugin.

### Aggregator Plugin Guidelines

* A aggregator must conform to the [`telegraf.Aggregator`](https://godoc.org/github.com/influxdata/telegraf#Aggregator) interface.
* Aggregators should call `aggregators.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
`github.com/influxdata/telegraf/plugins/aggregators/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
aggregator can be configured. This is include in `telegraf config`.
* The `Description` function should say in one line what this aggregator does.
* The Aggregator plugin will need to keep caches of metrics that have passed
through it. This should be done using the builtin `HashID()` function of each
metric.
* When the `Reset()` function is called, all caches should be cleared.

### Aggregator Example

```go
package min

// min.go

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type Min struct {
	// caches for metric fields, names, and tags
	fieldCache map[uint64]map[string]float64
	nameCache  map[uint64]string
	tagCache   map[uint64]map[string]string
}

func NewMin() telegraf.Aggregator {
	m := &Min{}
	m.Reset()
	return m
}

var sampleConfig = `
  ## period is the flush & clear interval of the aggregator.
  period = "30s"
  ## If true drop_original will drop the original metrics and
  ## only send aggregates.
  drop_original = false
`

func (m *Min) SampleConfig() string {
	return sampleConfig
}

func (m *Min) Description() string {
	return "Keep the aggregate min of each metric passing through."
}

func (m *Min) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.nameCache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		m.nameCache[id] = in.Name()
		m.tagCache[id] = in.Tags()
		m.fieldCache[id] = make(map[string]float64)
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				m.fieldCache[id][k] = fv
			}
		}
	} else {
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				if _, ok := m.fieldCache[id][k]; !ok {
					// hit an uncached field of a cached metric
					m.fieldCache[id][k] = fv
					continue
				}
				if fv < m.fieldCache[id][k] {
                    // set new minimum
					m.fieldCache[id][k] = fv
				}
			}
		}
	}
}

func (m *Min) Push(acc telegraf.Accumulator) {
	for id, _ := range m.nameCache {
		fields := map[string]interface{}{}
		for k, v := range m.fieldCache[id] {
			fields[k+"_min"] = v
		}
		acc.AddFields(m.nameCache[id], fields, m.tagCache[id])
	}
}

func (m *Min) Reset() {
	m.fieldCache = make(map[uint64]map[string]float64)
	m.nameCache = make(map[uint64]string)
	m.tagCache = make(map[uint64]map[string]string)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func init() {
	aggregators.Add("min", func() telegraf.Aggregator {
		return NewMin()
	})
}
```

## Unit Tests

Before opening a pull request you should run the linter checks and
the short tests.

### Execute linter

execute `make lint`

### Execute short tests

execute `make test`

### Execute integration tests

Running the integration tests requires several docker containers to be
running.  You can start the containers with:
```
make docker-run
```

And run the full test suite with:
```
make test-all
```

Use `make docker-kill` to stop the containers.
