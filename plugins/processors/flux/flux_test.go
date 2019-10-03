package flux

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestFlux(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		script   string
		metrics  []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "simple",
			script: `
		import "telegraf"

		telegraf.from() |> map(fn: (r) => ({r with value: r.value * 2}))`,
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(84),
					},
					now,
				),
			},
		},
		{
			name: "aggregate",
			script: `
		import "telegraf"

		telegraf.from()
		  |> filter(fn: (r) => r.name == "usage")
		  |> mean(column: "value")
		  // aggregates drop '_time', set it to a value to make the test work.
		  |> set(key: "_time", value: "1970-01-01")
		  |> map(fn: (r) => ({r with _time: time(v: r._time)}))`,
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "usage",
					},
					map[string]interface{}{
						"value": int64(1),
					},
					now,
				),
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "usage",
					},
					map[string]interface{}{
						"value": int64(2),
					},
					now,
				),
			},
			// This is partially unexpected, but there is still an empty table for 'idle_time'
			// after filtering, and the mean of an empty table is null.
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					nil,
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "usage",
					},
					map[string]interface{}{
						"value": 1.5,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "change measurement name",
			script: `
		import "telegraf"

		telegraf.from()
		  |> map(fn: (r) => ({r with value: r.value - 1}))
		  |> map(fn: (r) => ({r with _measurement: "changed"}))`,
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("changed",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(41),
					},
					now,
				),
			},
		},
		{
			name: "multiple from",
			script: `
import "telegraf"

telegraf.from() |> yield(name: "1")
telegraf.from() |> yield(name: "2")`,
			metrics: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
				testutil.MustMetric("cpu",
					map[string]string{
						"name": "idle_time",
					},
					map[string]interface{}{
						"value": int64(42),
					},
					now,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := ioutil.TempFile(os.TempDir(), "flux-processor-script")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.WriteString(tt.script); err != nil {
				t.Fatal(err)
			}
			proc := &Flux{
				Path: f.Name(),
				Log:  testutil.Logger{Name: "flux"},
			}
			actual := proc.Apply(tt.metrics...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}
