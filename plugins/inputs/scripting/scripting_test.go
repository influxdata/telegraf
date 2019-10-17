package scripting

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type Operations interface{}
type OperationGather struct{}
type OperationCheck []telegraf.Metric
type OperationCheckIgnoreTime []telegraf.Metric

func TestScripting(t *testing.T) {
	timeExample := time.Now()

	tests := map[string]struct {
		script     string
		operations []Operations
	}{
		"dummy input that a constant metric": {
			script: `
package scripting
import (
  "time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
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
			`,
			operations: []Operations{
				OperationGather{},
				OperationCheckIgnoreTime{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, timeExample)},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			agg := Scripting{Script: test.script}
			acc := testutil.Accumulator{}

			for _, op := range test.operations {
				switch o := (op).(type) {
				case OperationGather:
					err := agg.Gather(&acc)
					if err != nil {
						t.Errorf("gather error: %v", err)
					}
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
