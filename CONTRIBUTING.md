## Sign the CLA

Before we can merge a pull request, you will need to sign the CLA,
which can be found [on our website](http://influxdb.com/community/cla.html)

## Plugins

This section is for developers who want to create new collection plugins.
Telegraf is entirely plugin driven. This interface allows for operators to
pick and chose what is gathered as well as makes it easy for developers
to create new ways of generating metrics.

Plugin authorship is kept as simple as possible to promote people to develop
and submit new plugins.

### Plugin Guidelines

* A plugin must conform to the `plugins.Plugin` interface.
* Each generated metric automatically has the name of the plugin that generated
it prepended. This is to keep plugins honest.
* Plugins should call `plugins.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
`github.com/influxdb/telegraf/plugins/all/all.go` file.
* The `SampleConfig` function should return valid toml that describes how the
plugin can be configured. This is include in `telegraf -sample-config`.
* The `Description` function should say in one line what this plugin does.

### Plugin interface

```go
type Plugin interface {
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
* **measurement**: A string description of the metric. For instance `bytes_read` or `faults`.
* **value**: A value for the metric. This accepts 5 different types of value:
  * **int**: The most common type. All int types are accepted but favor using `int64`
  Useful for counters, etc.
  * **float**: Favor `float64`, useful for gauges, percentages, etc.
  * **bool**: `true` or `false`, useful to indicate the presence of a state. `light_on`, etc.
  * **string**: Typically used to indicate a message, or some kind of freeform information.
  * **time.Time**: Useful for indicating when a state last occurred, for instance `light_on_since`.
* **tags**: This is a map of strings to strings to describe the where or who
about the metric. For instance, the `net` plugin adds a tag named `"interface"`
set to the name of the network interface, like `"eth0"`.

The `AddFieldsWithTime` allows multiple values for a point to be passed. The values
used are the same type profile as **value** above. The **timestamp** argument
allows a point to be registered as having occurred at an arbitrary time.

Let's say you've written a plugin that emits metrics about processes on the current host.

```go

type Process struct {
    CPUTime float64
    MemoryBytes int64
    PID int
}

func Gather(acc plugins.Accumulator) error {
    for _, process := range system.Processes() {
        tags := map[string]string {
            "pid": fmt.Sprintf("%d", process.Pid),
        }

        acc.Add("cpu", process.CPUTime, tags, time.Now())
        acc.Add("memory", process.MemoryBytes, tags, time.Now())
    }
}
```

### Plugin Example

```go
package simple

// simple.go

import "github.com/influxdb/telegraf/plugins"

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
    plugins.Add("simple", func() plugins.Plugin { return &Simple{} })
}
```

## Service Plugins

This section is for developers who want to create new "service" collection
plugins. A service plugin differs from a regular plugin in that it operates
a background service while Telegraf is running. One example would be the `statsd`
plugin, which operates a statsd server.

Service Plugins are substantially more complicated than a regular plugin, as they
will require threads and locks to verify data integrity. Service Plugins should
be avoided unless there is no way to create their behavior with a regular plugin.

Their interface is quite similar to a regular plugin, with the addition of `Start()`
and `Stop()` methods.

### Service Plugin Guidelines

* Same as the `Plugin` guidelines, except that they must conform to the
`plugins.ServicePlugin` interface.

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

## Outputs

This section is for developers who want to create a new output sink. Outputs
are created in a similar manner as collection plugins, and their interface has
similar constructs.

### Output Guidelines

* An output must conform to the `outputs.Output` interface.
* Outputs should call `outputs.Add` in their `init` function to register themselves.
See below for a quick example.
* To be available within Telegraf itself, plugins must add themselves to the
`github.com/influxdb/telegraf/outputs/all/all.go` file.
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
    Write(points []*client.Point) error
}
```

### Output Example

```go
package simpleoutput

// simpleoutput.go

import "github.com/influxdb/telegraf/outputs"

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

func (s *Simple) Write(points []*client.Point) error {
    for _, pt := range points {
        // write `pt` to the output sink here
    }
    return nil
}

func init() {
    outputs.Add("simpleoutput", func() outputs.Output { return &Simple{} })
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
(i.e: https://github.com/influxdb/telegraf/blob/master/plugins/redis/redis_test.go )
a simple mock will suffice.

To execute Telegraf tests follow these simple steps:

- Install docker compose following [these](https://docs.docker.com/compose/install/)
instructions
- execute `make test`

**OSX users**: you will need to install `boot2docker` or `docker-machine`.
The Makefile will assume that you have a `docker-machine` box called `default` to
get the IP address.

### Unit test troubleshooting

Try cleaning up your test environment by executing `make test-cleanup` and
re-running
