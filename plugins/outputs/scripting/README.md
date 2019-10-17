# Scripting output plugin

This output allows to define in the config which code will output the metrics.

### Configuration
```
[[outputs.scripting]]
  ## Go code to output metrics
  script = '''
package scripting
import (
  "fmt"
  "time"
	"github.com/influxdata/telegraf"
)

func Connect() error {
	return nil
}

func Close() error {
	return nil
}

func Write(metrics []telegraf.Metric) error {
	for _,m := range metrics {
		fmt.Printf("%+v\n", m)
	}
	return nil
}
'''
```

### Coding
The main difference between "original" plugins and the same implementation in "scripting" is that methods ``Write``, ``Connect`` and ``Close`` are "plain", instead of the methods on types for the "original" plugins.

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
See the same section in ``processors/scripting``. This aggregator uses the symbols in that folder.

### Examples
Look at the scripting_test.go file for examples.
