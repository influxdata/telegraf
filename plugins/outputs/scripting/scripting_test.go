package scripting

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type Operations interface{}
type OperationWrite []telegraf.Metric
type OperationConnect struct{}
type OperationClose struct{}

func TestScripting(t *testing.T) {
	timeExample := time.Now()

	tests := map[string]struct {
		script     string
		operations []Operations
	}{
		"dummy output that prints metric to console": {
			script: `
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
			`,
			operations: []Operations{
				OperationConnect{},
				OperationWrite{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, timeExample)},
				OperationClose{},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			out := Scripting{Script: test.script}

			for _, op := range test.operations {
				switch o := (op).(type) {
				case OperationConnect:
					out.Connect()
				case OperationClose:
					out.Close()
				case OperationWrite:
					out.Write(o)
				}
			}
		})
	}
}
