# Secret Store Plugins

This section is for developers who want to create a new secret store plugin.

## Secret Store Plugin Guidelines

* A secret store must conform to the [telegraf.SecretStore][] interface.
* Secret-stores should call `secretstores.Add` in their `init` function to register
  themselves.  See below for a quick example.
* To be available within Telegraf itself, plugins must register themselves
  using a file in `github.com/influxdata/telegraf/plugins/secretstores/all`
  named according to the plugin name. Make sure you also add build-tags to
  conditionally build the plugin.
* Each plugin requires a file called `sample.conf` containing the sample
  configuration  for the plugin in TOML format. Please consult the
  [Sample Config][] page for the latest style guidelines.
* Each plugin `README.md` file should include the `sample.conf` file in a
  section describing the configuration by specifying a `toml` section in the
  form `toml @sample.conf`. The specified file(s) are then injected
  automatically into the Readme.
* Follow the recommended [Code Style][].

[telegraf.SecretStore]: https://pkg.go.dev/github.com/influxdata/telegraf?utm_source=godoc#SecretStore
[Sample Config]: https://github.com/influxdata/telegraf/blob/master/docs/developers/SAMPLE_CONFIG.md
[Code Style]: https://github.com/influxdata/telegraf/blob/master/docs/developers/CODE_STYLE.md

## Secret Store Plugin Example

### Registration

Registration of the plugin on `plugins/secretstores/all/printer.go`:

```go
//go:build !custom || secretstores || secretstores.printer

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/printer" // register plugin
```

The _build-tags_ in the first line allow to selectively include/exclude your
plugin when customizing Telegraf.

### Plugin

```go
//go:generate ../../../tools/readme_config_includer/generator
package main

import (
    _ "embed"
    "errors"

    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type Printer struct {
    Log telegraf.Logger `toml:"-"`

    cache map[string]string
}

func (p *Printer) SampleConfig() string {
    return sampleConfig
}

func (p *Printer) Init() error {
    return nil
}

// Get searches for the given key and return the secret
func (p *Printer) Get(key string) ([]byte, error) {
    v, found := p.cache[key]
    if !found {
        return nil, errors.New("not found")
    }

    return []byte(v), nil
}

// Set sets the given secret for the given key
func (p *Printer) Set(key, value string) error {
    p.cache[key] = value
    return nil
}

// List lists all known secret keys
func (p *Printer) List() ([]string, error) {
    keys := make([]string, 0, len(p.cache))
    for k := range p.cache {
        keys = append(keys, k)
    }
    return keys, nil
}

// GetResolver returns a function to resolve the given key.
func (p *Printer) GetResolver(key string) (telegraf.ResolveFunc, error) {
    resolver := func() ([]byte, bool, error) {
        s, err := p.Get(key)
        return s, false, err
    }
    return resolver, nil
}

// Register the secret-store on load.
func init() {
    secretstores.Add("printer", func(string) telegraf.SecretStore {
        return &Printer{}
    })
}
```
