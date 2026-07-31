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

func TestOptionsApply(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Filepath
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "Smoke Test",
			plugin: &Filepath{
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
						BasePath: "/my/test/",
					},
				},
				ToSlash: []baseOpts{
					{
						Field: "slashField",
						Tag:   "slashTag",
					},
				},
			},
			input: []telegraf.Metric{
				metric.New(
					"testmetric", map[string]string{
						"baseTag":  samplePath,
						"dirTag":   samplePath,
						"stemTag":  samplePath,
						"cleanTag": samplePath,
						"relTag":   samplePath,
						"slashTag": samplePath,
					},
					map[string]interface{}{
						"baseField":  samplePath,
						"dirField":   samplePath,
						"stemField":  samplePath,
						"cleanField": samplePath,
						"relField":   samplePath,
						"slashField": samplePath,
					},
					time.Now()),
			},
			expected: []telegraf.Metric{
				metric.New(
					"testmetric",
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
			plugin: &Filepath{
				BaseName: []baseOpts{
					{
						Field: "sourcePath",
						Tag:   "sourcePath",
						Dest:  "basePath",
					},
				}},
			input: []telegraf.Metric{
				metric.New(
					"testMetric",
					map[string]string{"sourcePath": samplePath},
					map[string]interface{}{"sourcePath": samplePath},
					time.Now()),
			},
			expected: []telegraf.Metric{
				metric.New(
					"testMetric",
					map[string]string{"sourcePath": samplePath, "basePath": "file.log"},
					map[string]interface{}{"sourcePath": samplePath, "basePath": "file.log"},
					time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.SortMetrics(), testutil.IgnoreTime())
		})
	}
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

	plugin := &Filepath{
		BaseName: []baseOpts{
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
