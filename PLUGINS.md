Tivan is entirely plugin driven. This interface allows for operators to
pick and chose what is gathered as well as makes it easy for developers
to create new ways of generating metrics.

Plugin authorship is kept as simple as possible to promote people to develop
and submit new plugins.

## Guidelines

* A plugin must conform to the `plugins.Plugin` interface.
* Tivan promises to run each plugin's Gather function serially. This means
developers don't have to worry about thread safety within these functions.
* Each generated metric automatically has the name of the plugin that generated
it prepended. This is to keep plugins honest.
* Plugins should call `plugins.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Tivan itself, plugins must add themselves to the `plugins.all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the plugin can be configured. This is include in `tivan -sample-config`.
* The `Description` function should say in one line what this plugin does.

### Plugin interface

```go
type Plugin interface {
	SampleConfig() string
	Description() string
	Gather(Accumulator) error
}
```

### Example

```go

# simple.go

import "github.com/influxdb/tivan/plugins"

type Simple struct {
  Ok bool
}

func (s *Simple) Description() string {
  return "a demo plugin"
}

func (s *Simple) SampleConfig() string {
  return "ok = true # indicate if everything is fine"
}

func (s *Simple) Gather(acc plugins.Accumulator) error {
  if s.Ok {
    acc.Add("state", "pretty good", nil)
  } else {
    acc.Add("state", "not great", nil)
  }

  return nil
}

func init() {
  plugins.Add("simple", func() plugins.Plugin { &Simple{} })
}
```

