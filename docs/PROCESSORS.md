### Processor Plugins

This section is for developers who want to create a new processor plugin.

### Processor Plugin Guidelines

* A processor must conform to the [telegraf.Processor][] interface.
* Processors should call `processors.Add` in their `init` function to register
  themselves.  See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
  `github.com/influxdata/telegraf/plugins/processors/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
  processor can be configured. This is include in the output of `telegraf
  config`.
- The `SampleConfig` function should return valid toml that describes how the
  plugin can be configured. This is included in `telegraf config`.  Please
  consult the [SampleConfig][] page for the latest style guidelines.
* The `Description` function should say in one line what this processor does.
- Follow the recommended [CodeStyle][].

### Processor Plugin Example

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

### Streaming Processors

Streaming processors are a new processor type available to you. They are
particularly useful to implement processor types that use background processes
or goroutines to process multiple metrics at the same time. Some examples of this
are the execd processor, which pipes metrics out to an external process over stdin
and reads them back over stdout, and the reverse_dns processor, which does reverse
dns lookups on IP addresses in fields. While both of these come with a speed cost,
it would be significantly worse if you had to process one metric completely from
start to finish before handling the next metric, and thus they benefit
significantly from a streaming-pipe approach.

Some differences from classic Processors:

* Streaming processors must conform to the [telegraf.StreamingProcessor][] interface.
* Processors should call `processors.AddStreaming` in their `init` function to register
  themselves.  See below for a quick example.

### Streaming Processor Example

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

func (p *Printer) Init() error {
	return nil
}

func (p *Printer) Start(acc telegraf.Accumulator) error {
}

func (p *Printer) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	// print!
	fmt.Println(metric.String())
	// pass the metric downstream, or metric.Drop() it.
	// Metric will be dropped if this function returns an error.
	acc.AddMetric(metric)

	return nil
}

func (p *Printer) Stop() error {
}

func init() {
	processors.AddStreaming("printer", func() telegraf.StreamingProcessor {
		return &Printer{}
	})
}
```

[SampleConfig]: https://github.com/influxdata/telegraf/wiki/SampleConfig
[CodeStyle]: https://github.com/influxdata/telegraf/wiki/CodeStyle
[telegraf.Processor]: https://godoc.org/github.com/influxdata/telegraf#Processor
[telegraf.StreamingProcessor]: https://godoc.org/github.com/influxdata/telegraf#StreamingProcessor
