package scripting

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type Operations interface{}
type OperationAdd []telegraf.Metric
type OperationPush struct{}
type OperationCheck []telegraf.Metric
type OperationCheckIgnoreTime []telegraf.Metric

func TestScripting(t *testing.T) {
	timeExample := time.Now()

	tests := map[string]struct {
		script     string
		operations []Operations
	}{
		"dummy aggregator that return the metrics Added": {
			script: `
package scripting
import (
  "fmt"
  "time"
	"github.com/influxdata/telegraf"
)

var data []telegraf.Metric

func Push(acc telegraf.Accumulator) {
  // We do not use methods on struct types, so global vars are
  // used to store state
  // Comments outside a function are not allowed
	for _,m := range data {
		acc.AddMetric(m)
	}
}

func Add(in telegraf.Metric) {
	data = append(data, in)
}

func Reset() {
}
			`,
			operations: []Operations{
				OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, timeExample)},
				OperationPush{},
				OperationCheck{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, timeExample)},
			},
		},
		"aggregator that counts the number of metrics seen": {
			script: `
package scripting
import (
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var count int

func Push(acc telegraf.Accumulator) {
	acc.AddMetric(testutil.MustMetric(
		"metrics_seen",
		map[string]string{},
		map[string]interface{}{"count": count},
		time.Now(),
	))
}

func Add(in telegraf.Metric) {
	count++
}

func Reset() {
}
			`,
			operations: []Operations{
				OperationAdd{
					testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, timeExample),
					testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 2}, timeExample),
					testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 3}, timeExample),
				},
				OperationPush{},
				OperationCheckIgnoreTime{testutil.MustMetric("metrics_seen", map[string]string{}, map[string]interface{}{"count": 3}, timeExample)},
			},
		},
		"calculate r_await, w_await from diskio metrics, avg between intervals": {
			script: `
package scripting
import (
	"fmt"
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type data struct {
  count int
  tags map[string]string
  firstReadOperations float64
  firstReadTime float64
  lastReadOperations float64
  lastReadTime float64
  firstWriteOperations float64
  firstWriteTime float64
  lastWriteOperations float64
  lastWriteTime float64
}

var cache map[uint64]data

func Add(in telegraf.Metric) {
  if cache == nil {
		cache = map[uint64]data{}
	}

  id := in.HashID()
	if _, ok := cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		d := data {
			count: 1,
		  tags: in.Tags(),
			firstReadTime: float64(in.GetField("read_time").(int64)),
			firstReadOperations: float64(in.GetField("reads").(int64)),
			firstWriteTime: float64(in.GetField("write_time").(int64)),
			firstWriteOperations: float64(in.GetField("writes").(int64)),
		}

		cache[id] = d
	} else {
		d := cache[id]
		d.count++
		d.lastReadOperations = float64(in.GetField("reads").(int64))
		d.lastReadTime = float64(in.GetField("read_time").(int64))
		d.lastWriteOperations = float64(in.GetField("writes").(int64))
		d.lastWriteTime = float64(in.GetField("write_time").(int64))
		cache[id] = d
	}
}

func Push(acc telegraf.Accumulator) {
	for _,d := range cache {
	  if d.count < 2 {
	    continue
	  }

	  r_await := (d.lastReadTime-d.firstReadTime)/(d.lastReadOperations-d.firstReadOperations)
	  w_await := (d.lastWriteTime-d.firstWriteTime)/(d.lastWriteOperations-d.firstWriteOperations)
	  a_await := (d.lastReadTime+d.lastWriteTime-d.firstReadTime-d.firstWriteTime)/(d.lastReadOperations+d.lastWriteOperations-d.firstReadOperations-d.firstWriteOperations)

		acc.AddMetric(
			testutil.MustMetric(
				"diskio",
				d.tags,
				map[string]interface{}{
				  "r_await": r_await,
				  "w_await": w_await,
				  "a_await": a_await,
				},
				time.Now(),
			),
		)
	}
}

func Reset() {
	cache = map[uint64]data{}
}
			`,
			operations: []Operations{
				OperationAdd{
					testutil.MustMetric(
						"diskio",
						map[string]string{"host": "hostA", "name": "sda"},
						map[string]interface{}{
							"read_time":  5,
							"reads":      5,
							"write_time": 5,
							"writes":     5,
						},
						timeExample,
					),
				},
				OperationPush{},
				// No metric generated if we have seen only one metric
				OperationCheckIgnoreTime{},
				OperationAdd{
					testutil.MustMetric(
						"diskio",
						map[string]string{"host": "hostA", "name": "sda"},
						map[string]interface{}{
							"read_time":  100,
							"reads":      10,
							"write_time": 100,
							"writes":     10,
						},
						timeExample,
					),
					testutil.MustMetric(
						"diskio",
						map[string]string{"host": "hostA", "name": "sda"},
						map[string]interface{}{
							"read_time":  110,
							"reads":      12,
							"write_time": 110,
							"writes":     12,
						},
						timeExample,
					),
					testutil.MustMetric(
						"diskio",
						map[string]string{"host": "hostA", "name": "sda"},
						map[string]interface{}{
							"read_time":  201,
							"reads":      20,
							"write_time": 201,
							"writes":     20,
						},
						timeExample,
					),
					testutil.MustMetric(
						"diskio",
						map[string]string{"host": "hostB", "name": "sdb"},
						map[string]interface{}{
							"read_time":  110,
							"reads":      12,
							"write_time": 110,
							"writes":     12,
						},
						timeExample,
					),
					testutil.MustMetric(
						"diskio",
						map[string]string{"host": "hostB", "name": "sdb"},
						map[string]interface{}{
							"read_time":  201,
							"reads":      29,
							"write_time": 201,
							"writes":     22,
						},
						timeExample,
					),
				},
				OperationPush{},
				// Calculate the average between first and last metric of the period
				OperationCheckIgnoreTime{
					testutil.MustMetric(
						"diskio",
						map[string]string{
							"host": "hostA",
							"name": "sda",
						},
						map[string]interface{}{
							"r_await": 10.1,
							"w_await": 10.1,
							"a_await": 10.1,
						},
						timeExample,
					),
					testutil.MustMetric(
						"diskio",
						map[string]string{
							"host": "hostB",
							"name": "sdb",
						},
						map[string]interface{}{
							"r_await": 5.352941176470588,
							"w_await": 9.1,
							"a_await": 6.7407407407407405,
						},
						timeExample,
					),
				},
			},
		},
		"valuecounter aggregator": {
			script: `
package scripting
import (
	"fmt"
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type aggregate struct {
        name       string
        tags       map[string]string
        fieldCount map[string]int
}

var cache  map[uint64]aggregate
var fieldsConfig  = []string{"status"}

func Push(acc telegraf.Accumulator) {
	for _, agg := range cache {
		/*
		 * This assignation works correctly:
		fields = map[string]interface{}{"foo": 34}
		But this one does not:
		fields["foo"] = 999
		*/

		m := testutil.MustMetric(agg.name, agg.tags, nil, time.Now())
		for fk, fv := range agg.fieldCount {
			m.AddField(fk, fv)
		}

		// acc.AddFields does not work correctly because it does not cast "interface{}" to real values
		acc.AddMetric(m)
	}
}

func Add(in telegraf.Metric) {
  if cache == nil {
		cache = map[uint64]aggregate{}
	}

	id := in.HashID()

	// Check if the cache already has an entry for this metric, if not create it
	if _, ok := cache[id]; !ok {
		a := aggregate{
			name:       in.Name(),
			tags:       in.Tags(),
			fieldCount: make(map[string]int),
		}
		cache[id] = a
	}

	// Check if this metric has fields which we need to count, if so increment
	// the count.
	// Parser broken with "for fk,fv := range in.Fields()"
	for i := 0; i < len(in.FieldList()); i++ {
	  f := in.FieldList()[i]
		fk := f.Key
		fv := f.Value

		// We cannot pass parameters to the aggregator, so we specify them
		// directly in the global vars.
		// In the original fieldcount aggregator "fields" is a config parameter.
		for _, cf := range fieldsConfig {
			if fk == cf {
				fn := fmt.Sprintf("%v_%v", fk, fv)
				// Operator ++ does not work
				cache[id].fieldCount[fn] = cache[id].fieldCount[fn] + 1
			}
		}
	}
}

func Reset() {
	// Using "make" here brokes the parser
	cache = map[uint64]aggregate{}
}
			`,
			operations: []Operations{
				OperationAdd{
					testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status": 200, "foobar": "bar"}, time.Now()),
					testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status": 200, "foobar": "bar"}, time.Now()),
					testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status": "OK", "ignoreme": "string", "andme": true, "boolfield": false}, time.Now()),
				},
				OperationPush{},
				OperationCheckIgnoreTime{
					testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status_200": 2, "status_OK": 1}, time.Now()),
				},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			agg := Scripting{Script: test.script}
			acc := testutil.Accumulator{}

			for _, op := range test.operations {
				switch o := (op).(type) {
				case OperationAdd:
					for _, metric := range o {
						agg.Add(metric)
					}
				case OperationPush:
					agg.Push(&acc)
					agg.Reset() // Telegraf calls Reset() after Push()
				case OperationCheck:
					testutil.RequireMetricsEqual(t, o, acc.GetTelegrafMetrics(), testutil.SortMetrics())
					acc.ClearMetrics()
				case OperationCheckIgnoreTime:
					testutil.RequireMetricsEqual(t, o, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
					acc.ClearMetrics()
				}
			}
		})
	}
}

func BenchmarkValueCounterApply(b *testing.B) {
	m1 := testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status": 200, "foobar": "bar"}, time.Now())
	m2 := testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status": "OK", "ignoreme": "string", "andme": true, "boolfield": false}, time.Now())

	script := `
package scripting
import (
	"fmt"
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type aggregate struct {
        name       string
        tags       map[string]string
        fieldCount map[string]int
}

var cache  map[uint64]aggregate
var fieldsConfig  = []string{"status"}

func Push(acc telegraf.Accumulator) {
	for _, agg := range cache {
		/*
		 * This assignation works correctly:
		fields = map[string]interface{}{"foo": 34}
		But this one does not:
		fields["foo"] = 999
		*/

		m := testutil.MustMetric(agg.name, agg.tags, nil, time.Now())
		for fk, fv := range agg.fieldCount {
			m.AddField(fk, fv)
		}

		// acc.AddFields does not work correctly because it does not cast "interface{}" to real values
		acc.AddMetric(m)
	}
}

func Add(in telegraf.Metric) {
  if cache == nil {
		cache = map[uint64]aggregate{}
	}

	id := in.HashID()

	// Check if the cache already has an entry for this metric, if not create it
	if _, ok := cache[id]; !ok {
		a := aggregate{
			name:       in.Name(),
			tags:       in.Tags(),
			fieldCount: make(map[string]int),
		}
		cache[id] = a
	}

	// Check if this metric has fields which we need to count, if so increment
	// the count.
	// Parser broken with "for fk,fv := range in.Fields()"
	for i := 0; i < len(in.FieldList()); i++ {
	  f := in.FieldList()[i]
		fk := f.Key
		fv := f.Value

		// We cannot pass parameters to the aggregator, so we specify them
		// directly in the global vars.
		// In the original fieldcount aggregator "fields" is a config parameter.
		for _, cf := range fieldsConfig {
			if fk == cf {
				fn := fmt.Sprintf("%v_%v", fk, fv)
				// Operator ++ does not work
				cache[id].fieldCount[fn] = cache[id].fieldCount[fn] + 1
			}
		}
	}
}

func Reset() {
	// Using "make" here brokes the parser
	cache = map[uint64]aggregate{}
}
	`
	agg := Scripting{Script: script}
	for i := 0; i < b.N; i++ {
		agg.Add(m1)
		agg.Add(m2)
	}
}
