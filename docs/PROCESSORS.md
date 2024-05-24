# Processor Plugins

This section is for developers who want to create a new processor plugin.

## Processor Plugin Guidelines

* A processor must conform to the [telegraf.Processor][] interface.
* Processors should call `processors.Add` in their `init` function to register
  themselves.  See below for a quick example.
* To be available within Telegraf itself, plugins must register themselves
  using a file in `github.com/influxdata/telegraf/plugins/processors/all`
  named according to the plugin name. Make sure you also add build-tags to
  conditionally build the plugin.
* Each plugin requires a file called `sample.conf` containing the sample
  configuration  for the plugin in TOML format.
  Please consult the [Sample Config][] page for the latest style guidelines.
* Each plugin `README.md` file should include the `sample.conf` file in a
  section describing the configuration by specifying a `toml` section in the
  form `toml @sample.conf`. The specified file(s) are then injected
  automatically into the Readme.
* Follow the recommended [Code Style][].

[Sample Config]: /docs/developers/SAMPLE_CONFIG.md
[Code Style]: /docs/developers/CODE_STYLE.md
[telegraf.Processor]: https://godoc.org/github.com/influxdata/telegraf#Processor

## Streaming Processors

Streaming processors are a new processor type available to you. They are
particularly useful to implement processor types that use background processes
or goroutines to process multiple metrics at the same time. Some examples of
this are the execd processor, which pipes metrics out to an external process
over stdin and reads them back over stdout, and the reverse_dns processor, which
does reverse dns lookups on IP addresses in fields. While both of these come
with a speed cost, it would be significantly worse if you had to process one
metric completely from start to finish before handling the next metric, and thus
they benefit significantly from a streaming-pipe approach.

Some differences from classic Processors:

* Streaming processors must conform to the [telegraf.StreamingProcessor][] interface.
* Processors should call `processors.AddStreaming` in their `init` function to register
  themselves.  See below for a quick example.

[telegraf.StreamingProcessor]: https://godoc.org/github.com/influxdata/telegraf#StreamingProcessor

## Processor Plugin Example

### Registration

Registration of the plugin on `plugins/processors/all/printer.go`:

```go
//go:build !custom || processors || processors.printer

package all

import _ "github.com/influxdata/telegraf/plugins/processors/printer" // register plugin
```

The _build-tags_ in the first line allow to selectively include/exclude your
plugin when customizing Telegraf.

### Plugin

Content of your plugin file e.g. `printer.go`

```go
//go:generate ../../../tools/readme_config_includer/generator
package printer

import (
    _ "embed"
    "fmt"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Printer struct {
    Log telegraf.Logger `toml:"-"`
}

func (*Printer) SampleConfig() string {
    return sampleConfig
}

// Init is for setup, and validating config.
func (p *Printer) Init() error {
    return nil
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

## Streaming Processor Example

```go
//go:generate ../../../tools/readme_config_includer/generator
package printer

import (
    _ "embed"
    "fmt"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Printer struct {
    Log telegraf.Logger `toml:"-"`
}

func (*Printer) SampleConfig() string {
    return sampleConfig
}

// Init is for setup, and validating config.
func (p *Printer) Init() error {
    return nil
}

// Start is called once when the plugin starts; it is only called once per
// plugin instance, and never in parallel.
// Start should return once it is ready to receive metrics.
// The passed in accumulator is the same as the one passed to Add(), so you
// can choose to save it in the plugin, or use the one received from Add().
func (p *Printer) Start(acc telegraf.Accumulator) error {
}

// Add is called for each metric to be processed. The Add() function does not
// need to wait for the metric to be processed before returning, and it may
// be acceptable to let background goroutine(s) handle the processing if you
// have slow processing you need to do in parallel.
// Keep in mind Add() should not spawn unbounded goroutines, so you may need
// to use a semaphore or pool of workers (eg: reverse_dns plugin does this).
// Metrics you don't want to pass downstream should have metric.Drop() called,
// rather than simply omitting the acc.AddMetric() call
func (p *Printer) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
    // print!
    fmt.Println(metric.String())
    // pass the metric downstream, or metric.Drop() it.
    // Metric will be dropped if this function returns an error.
    acc.AddMetric(metric)

    return nil
}

// Stop gives you an opportunity to gracefully shut down the processor.
// Once Stop() is called, Add() will not be called any more. If you are using
// goroutines, you should wait for any in-progress metrics to be processed
// before returning from Stop().
// When stop returns, you should no longer be writing metrics to the
// accumulator.
func (p *Printer) Stop() error {
}

func init() {
    processors.AddStreaming("printer", func() telegraf.StreamingProcessor {
        return &Printer{}
    })
}
```
