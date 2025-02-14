# Aggregator Plugins

This section is for developers who want to create a new aggregator plugin.

## Aggregator Plugin Guidelines

* A aggregator must conform to the [telegraf.Aggregator][] interface.
* Aggregators should call `aggregators.Add` in their `init` function to
  register themselves.  See below for a quick example.
* To be available within Telegraf itself, plugins must register themselves
  using a file in `github.com/influxdata/telegraf/plugins/aggregators/all`
  named according to the plugin name. Make sure you also add build-tags to
  conditionally build the plugin.
* Each plugin requires a file called `sample.conf` containing the sample
  configuration for the plugin in TOML format. Please consult the
  [Sample Config][] page for the latest style guidelines.
* Each plugin `README.md` file should include the `sample.conf` file in a
  section describing the configuration by specifying a `toml` section in the
  form `toml @sample.conf`. The specified file(s) are then injected
  automatically into the Readme.
* The Aggregator plugin will need to keep caches of metrics that have passed
  through it. This should be done using the builtin `HashID()` function of
  each metric.
* When the `Reset()` function is called, all caches should be cleared.
* Follow the recommended [Code Style][].

[telegraf.Aggregator]: https://godoc.org/github.com/influxdata/telegraf#Aggregator
[Sample Config]: /docs/developers/SAMPLE_CONFIG.md
[Code Style]: /docs/developers/CODE_STYLE.md

### Aggregator Plugin Example

### Registration

Registration of the plugin on `plugins/aggregators/all/min.go`:

```go
//go:build !custom || aggregators || aggregators.min

package all

import _ "github.com/influxdata/telegraf/plugins/aggregators/min" // register plugin
```

The _build-tags_ in the first line allow to selectively include/exclude your
plugin when customizing Telegraf.

### Plugin

Content of your plugin file e.g. `min.go`

```go
//go:generate ../../../tools/readme_config_includer/generator
package min

// min.go

import (
    _ "embed"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/aggregators"
)

//go:embed sample.conf
var sampleConfig string

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

func (*Min) SampleConfig() string {
    return sampleConfig
}

func (m *Min) Init() error {
    return nil
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
