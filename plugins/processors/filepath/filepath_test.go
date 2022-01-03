//go:build !windows
// +build !windows

package filepath

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var samplePath = "/my/test//c/../path/file.log"

func TestOptions_Apply(t *testing.T) {
	tests := []testCase{
		{
			name:         "Smoke Test",
			o:            newOptions("/my/test/"),
			inputMetrics: getSmokeTestInputMetrics(samplePath),
			expectedMetrics: []telegraf.Metric{
				testutil.MustMetric(
					smokeMetricName,
					map[string]string{
						"baseTag":  "file.log",
						"dirTag":   "/my/test/path",
						"stemTag":  "file",
						"cleanTag": "/my/test/path/file.log",
						"relTag":   "path/file.log",
						"slashTag": "/my/test//c/../path/file.log",
					},
					map[string]interface{}{
						"baseField":  "file.log",
						"dirField":   "/my/test/path",
						"stemField":  "file",
						"cleanField": "/my/test/path/file.log",
						"relField":   "path/file.log",
						"slashField": "/my/test//c/../path/file.log",
					},
					time.Now()),
			},
		},
		{
			name: "Test Dest Option",
			o: &Options{
				BaseName: []BaseOpts{
					{
						Field: "sourcePath",
						Tag:   "sourcePath",
						Dest:  "basePath",
					},
				}},
			inputMetrics: []telegraf.Metric{
				testutil.MustMetric(
					"testMetric",
					map[string]string{"sourcePath": samplePath},
					map[string]interface{}{"sourcePath": samplePath},
					time.Now()),
			},
			expectedMetrics: []telegraf.Metric{
				testutil.MustMetric(
					"testMetric",
					map[string]string{"sourcePath": samplePath, "basePath": "file.log"},
					map[string]interface{}{"sourcePath": samplePath, "basePath": "file.log"},
					time.Now()),
			},
		},
	}
	runTestOptionsApply(t, tests)
}
