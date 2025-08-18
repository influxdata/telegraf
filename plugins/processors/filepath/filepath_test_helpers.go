package filepath

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const smokeMetricName = "testmetric"

type testCase struct {
	name            string
	o               *Filepath
	inputMetrics    []telegraf.Metric
	expectedMetrics []telegraf.Metric
}

func newOptions(basePath string) *Filepath {
	return &Filepath{
		BaseName: []baseOpts{
			{
				Field: "baseField",
				Tag:   "baseTag",
			},
		},
		DirName: []baseOpts{
			{
				Field: "dirField",
				Tag:   "dirTag",
			},
		},
		Stem: []baseOpts{
			{
				Field: "stemField",
				Tag:   "stemTag",
			},
		},
		Clean: []baseOpts{
			{
				Field: "cleanField",
				Tag:   "cleanTag",
			},
		},
		Rel: []relOpts{
			{
				baseOpts: baseOpts{
					Field: "relField",
					Tag:   "relTag",
				},
				BasePath: basePath,
			},
		},
		ToSlash: []baseOpts{
			{
				Field: "slashField",
				Tag:   "slashTag",
			},
		},
	}
}

func getSampleMetricTags(path string) map[string]string {
	return map[string]string{
		"baseTag":  path,
		"dirTag":   path,
		"stemTag":  path,
		"cleanTag": path,
		"relTag":   path,
		"slashTag": path,
	}
}

func getSampleMetricFields(path string) map[string]interface{} {
	return map[string]interface{}{
		"baseField":  path,
		"dirField":   path,
		"stemField":  path,
		"cleanField": path,
		"relField":   path,
		"slashField": path,
	}
}

func getSmokeTestInputMetrics(path string) []telegraf.Metric {
	return []telegraf.Metric{
		testutil.MustMetric(smokeMetricName, getSampleMetricTags(path), getSampleMetricFields(path),
			time.Now()),
	}
}

func runTestOptionsApply(t *testing.T, tests []testCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.o.Apply(tt.inputMetrics...)
			testutil.RequireMetricsEqual(t, tt.expectedMetrics, got, testutil.SortMetrics(), testutil.IgnoreTime())
		})
	}
}
