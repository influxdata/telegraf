# Aggregator Plugins

This section is for developers who want to create a new aggregator plugin.

## Aggregator Plugin Guidelines

* A aggregator must conform to the [telegraf.Aggregator][] interface.
* Aggregators should call `aggregators.Add` in their `init` function to
  register themselves.  See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
  `github.com/influxdata/telegraf/plugins/aggregators/all/all.go` file.
* Each plugin requires a file called `<plugin_name>_sample_config.go`, where `<plugin_name>` is replaced with the actual plugin name.
  Copy the [example template](#sample-configuration-template) into this file, also updating `<plugin_name>` were appropriate.
  This file is automatically updated during the build process to include the sample configuration from the `README.md`.
  Please consult the [Sample Config][] page for the latest style guidelines.
* The Aggregator plugin will need to keep caches of metrics that have passed
  through it. This should be done using the builtin `HashID()` function of
  each metric.
* When the `Reset()` function is called, all caches should be cleared.
* Follow the recommended [Code Style][].

### Aggregator Plugin Example

```go
//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
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

### Sample Configuration Template

```go
//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
// DON'T EDIT; This file is used as a template by tools/generate_plugindata
package <plugin_package>

func (k *<plugin_struct>) SampleConfig() string {
    return `{{ .SampleConfig }}`
}
```

[telegraf.Aggregator]: https://godoc.org/github.com/influxdata/telegraf#Aggregator
[Sample Config]: https://github.com/influxdata/telegraf/blob/master/docs/developers/SAMPLE_CONFIG.md
[Code Style]: https://github.com/influxdata/telegraf/blob/master/docs/developers/CODE_STYLE.md
