# Scripting input plugin

This input allows to define in the config which code will gather the metrics.

### Configuration
```
[[inputs.scripting]]
  ## Go code to gather metrics
  script = '''
package scripting
import (
	"github.com/influxdata/telegraf"
)

func Gather(acc telegraf.Accumulator) error {
	acc.AddMetric(testutil.MustMetric(
		"name",
		map[string]string{"host": "hostA", "foo": "bar"},
		map[string]interface{}{"value": 1},
		time.Now(),
	))
	return nil
}
'''
```

### Coding
The main difference between "original" plugins and the same implementation in "scripting" is that method ``Gather`` are "plain", instead of the methods on types for the "original" plugins.

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
See the same section in ``processors/scripting``. This input uses the symbols in that folder.

### Examples
Look at the scripting_test.go file for examples.

```
