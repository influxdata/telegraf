//go:build !windows

package filepath

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"sourcePath": samplePath},
			map[string]interface{}{"sourcePath": samplePath},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"sourcePath": samplePath, "basePath": "file.log"},
			map[string]interface{}{"sourcePath": samplePath, "basePath": "file.log"},
			time.Unix(0, 0),
		),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	plugin := &Options{
		BaseName: []BaseOpts{
			{
				Field: "sourcePath",
				Tag:   "sourcePath",
				Dest:  "basePath",
			},
		},
	}

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
