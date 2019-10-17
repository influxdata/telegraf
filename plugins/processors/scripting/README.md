# Scripting processor plugin

This processor allows to define in the config which code will process the metrics.

### Configuration
```
[[processors.scripting]]
  ## Go code to process metrics
  script = '''
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
  fmt.Printf("%+v\n", in)
  return in
}
'''
```

### Coding
The main difference between "original" plugins and the same implementation in "scripting" is that method ``Apply`` is "plain", instead of the method on types for the "original" plugins.

[Yaegi](https://github.com/containous/yaegi) has some limitations parsing Go code, this will cause Telegraf to crash at start time.

Some of this limitations:
 - comments are not allowed outside of functions
 - assignation to map[string]interface{} does not work, like ``fields["foo"] = 999``. But this works: ``fields = map[string]interface{}{"foo": 34}``
 - ``:= range in.Fields()`` does not work, do not cast correctly ``interface{}``
 - in general, problems with ``interface{}`` types
 - functions/types should be declared before use
 - some problems defining/setting vars, sometimes ``make(...)`` works, some times ``= type{}``
 - ``++`` operator does not work

### Adding symbols for Yaegi
Yaegi need to ``goexports`` libs used by the scripting.

In directory ``plugins/processors/scripting/telegrafSymbols``.

With go1.13:
```
go run /go/src/github.com/containous/yaegi/cmd/goexports/ github.com/influxdata/telegraf >& go1_13_github.com_influxdata_telegraf.go
go run /go/src/github.com/containous/yaegi/cmd/goexports/ github.com/influxdata/telegraf/metric >& go1_13_github.com_influxdata_telegraf_metric.go
go run /go/src/github.com/containous/yaegi/cmd/goexports/ github.com/influxdata/telegraf/testutil >& go1_13_github.com_influxdata_telegraf_testutil.go
```

With go1.12:
```
go run /go/src/github.com/containous/yaegi/cmd/goexports/ github.com/influxdata/telegraf >& go1_12_github.com_influxdata_telegraf.go
go run /go/src/github.com/containous/yaegi/cmd/goexports/ github.com/influxdata/telegraf/metric >& go1_12_github.com_influxdata_telegraf_metric.go
go run /go/src/github.com/containous/yaegi/cmd/goexports/ github.com/influxdata/telegraf/testutil >& go1_12_github.com_influxdata_telegraf_testutil.go
```

### Examples
Look at the scripting_test.go file for examples.

### Benchmarks

Testing renaming metrics with the ``rename`` original processor against an implementation using "scripting".

100x slower (19.8ns/op VS 3546ns/op)

Original:
goos: linux
goarch: amd64
BenchmarkRename-8       56986599                19.8 ns/op             0 B/op          0 allocs/op

Scripting:
```
goos: linux
goarch: amd64
pkg: github.com/influxdata/telegraf/plugins/processors/scripting
BenchmarkRename-8         625650              3564 ns/op             504 B/op         14 allocs/op
```
