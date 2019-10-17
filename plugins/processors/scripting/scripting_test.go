package scripting

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestScripting(t *testing.T) {
	timeExample := time.Now()

	tests := map[string]struct {
		metricsIn  []telegraf.Metric
		script     string
		metricsOut []telegraf.Metric
	}{
		"printer": {
			metricsIn: []telegraf.Metric{testutil.MustMetric("m1", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample)},
			script: `
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
  fmt.Printf("%+v\n", in)
  return in
}`,
			metricsOut: []telegraf.Metric{testutil.MustMetric("m1", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample)},
		},
		"rename": {
			metricsIn: []telegraf.Metric{testutil.MustMetric("m1", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample)},
			script: `
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
	for _,m := range in {
	  m.SetName("m2")
	}
  return in
		}`,
			metricsOut: []telegraf.Metric{testutil.MustMetric("m2", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample)},
		},
		"suffix": {
			metricsIn: []telegraf.Metric{testutil.MustMetric("m1", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample)},
			script: `
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
	for _,m := range in {
	  m.SetName(fmt.Sprintf("%s-%s", m.Name(), "suffix"))
	}
  return in
		}`,
			metricsOut: []telegraf.Metric{testutil.MustMetric("m1-suffix", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample)},
		},
		"clone": {
			metricsIn: []telegraf.Metric{
				testutil.MustMetric("m1", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample),
				testutil.MustMetric("m2", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample),
			},
			script: `
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
	cloned := []telegraf.Metric{}
	for _,m := range in {
		copy := m.Copy()
	  copy.SetName(fmt.Sprintf("%s-%s", m.Name(), "clone"))
	  cloned = append(cloned, copy)
	}
  return append(in, cloned...)
		}`,
			metricsOut: []telegraf.Metric{
				testutil.MustMetric("m1", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample),
				testutil.MustMetric("m1-clone", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample),
				testutil.MustMetric("m2", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample),
				testutil.MustMetric("m2-clone", map[string]string{"host": "hostA"}, map[string]interface{}{"value": 0}, timeExample),
			},
		},
		"disk.used_pct": {
			metricsIn: []telegraf.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "hostA"},
					map[string]interface{}{
						"used":  50.0,
						"total": 100.0,
					},
					timeExample,
				),
			},
			script: `
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
	for _,m := range in {
		used,ok := m.GetField("used")
		if !ok {
			continue
		}
		total,ok := m.GetField("total")
		if !ok {
			continue
		}
		m.AddField("used_percent", 100*used.(float64)/total.(float64))
	}
  return in
		}`,
			metricsOut: []telegraf.Metric{
				testutil.MustMetric(
					"disk",
					map[string]string{"host": "hostA"},
					map[string]interface{}{
						"used":         50.0,
						"total":        100.0,
						"used_percent": 50.0,
					},
					timeExample,
				),
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			s := Scripting{
				Script: test.script,
			}
			out := s.Apply(test.metricsIn...)
			assert.ElementsMatch(t, test.metricsOut, out)
		})
	}
}

func BenchmarkRename(b *testing.B) {
	m1 := testutil.MustMetric("m1", map[string]string{"foo": "bar"}, map[string]interface{}{"status": 200, "foobar": "bar"}, time.Now())

	script := `
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
	for _,m := range in {
	  m.SetName("m2")
	}
  return in
}
	`
	agg := Scripting{Script: script}
	for i := 0; i < b.N; i++ {
		agg.Apply(m1)
	}
}
