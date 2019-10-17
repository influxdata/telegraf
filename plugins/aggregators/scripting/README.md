# Scripting aggregator plugin

This aggregator allows to define in the config which code will aggregate the metrics.

### Configuration
```
[[aggregators.scripting]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Go code to aggregate metrics
  script = '''
package scripting
import (
	"github.com/influxdata/telegraf"
)

var data []telegraf.Metric

func Push(acc telegraf.Accumulator) {
	for _,m := range data {
		acc.AddMetric(m)
	}
}

func Add(in telegraf.Metric) {
	data = append(data, in)
}

func Reset() {
}
'''
```

### Coding
The main difference between "original" plugins and the same implementation in "scripting" is that methods ``Add`` and ``Push`` are "plain", instead of the methods on types for the "original" plugins.

[Yaegi](https://github.com/containous/yaegi) has some limitations parsing Go code, this will cause Telegraf to crash at start time.

Some of this limitations:
 - comments are not allowed outside of functions
 - adding a new metric with ``acc.AddFields()`` does not work (it does not interpret correctly the ``map[string]interface{}``)
 - assignation to map[string]interface{} does not work, like ``fields["foo"] = 999``. But this works: ``fields = map[string]interface{}{"foo": 34}``
 - ``:= range in.Fields()`` does not work, do not cast correctly ``interface{}``
 - in general, problems with ``interface{}`` types
 - functions/types should be declared before use
 - some problems defining/setting vars, sometimes ``make(...)`` works, some times ``= type{}``
 - ``++`` operator does not work

### Adding symbols for Yaegi
See the same section in ``processors/scripting``. This aggregator uses the symbols in that folder.

### Examples
Look at the scripting_test.go file for examples.

### Benchmarks

Testing the "valuecounter" original aggregator against an implementation using "scripting".

10x slower (1.2ms/op VS 15ms/op)

Original:
```
goos: linux
goarch: amd64
BenchmarkApply-8          956456              1213 ns/op             832 B/op         22 allocs/op
```

Scripting:
```
goos: linux
goarch: amd64
BenchmarkValueCounterApply-8       76390             14944 ns/op            5637 B/op        168 allocs/op
```
