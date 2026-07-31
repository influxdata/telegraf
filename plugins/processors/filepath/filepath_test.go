package filepath

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestOptionsApply(t *testing.T) {
	var base, sample string
	if runtime.GOOS == "windows" {
		base = `c:\my\test\`
		sample = base + `\c\..\path\file.log`
	} else {
		base = `/my/test/`
		sample = base + "/c/../path/file.log"
	}

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
						BasePath: base,
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
						"baseTag":  filepath.ToSlash(sample),
						"dirTag":   filepath.ToSlash(sample),
						"stemTag":  filepath.ToSlash(sample),
						"cleanTag": filepath.ToSlash(sample),
						"relTag":   filepath.ToSlash(sample),
						"slashTag": filepath.ToSlash(sample),
					},
					map[string]interface{}{
						"baseField":  filepath.ToSlash(sample),
						"dirField":   filepath.ToSlash(sample),
						"stemField":  filepath.ToSlash(sample),
						"cleanField": filepath.ToSlash(sample),
						"relField":   filepath.ToSlash(sample),
						"slashField": filepath.ToSlash(sample),
					},
					time.Now()),
			},
			expected: []telegraf.Metric{
				metric.New(
					"testmetric",
					map[string]string{
						"baseTag":  "file.log",
						"dirTag":   filepath.Join(base, "path"),
						"stemTag":  "file",
						"cleanTag": filepath.Join(base, "path", "file.log"),
						"relTag":   filepath.Join("path", "file.log"),
						"slashTag": filepath.ToSlash(sample),
					},
					map[string]interface{}{
						"baseField":  "file.log",
						"dirField":   filepath.Join(base, "path"),
						"stemField":  "file",
						"cleanField": filepath.Join(base, "path", "file.log"),
						"relField":   filepath.Join("path", "file.log"),
						"slashField": filepath.ToSlash(sample),
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
					map[string]string{"sourcePath": filepath.ToSlash(sample)},
					map[string]interface{}{"sourcePath": filepath.ToSlash(sample)},
					time.Now()),
			},
			expected: []telegraf.Metric{
				metric.New(
					"testMetric",
					map[string]string{"sourcePath": filepath.ToSlash(sample), "basePath": "file.log"},
					map[string]interface{}{"sourcePath": filepath.ToSlash(sample), "basePath": "file.log"},
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
	var sample string
	if runtime.GOOS == "windows" {
		sample = `c:\my\test\\c\..\path\file.log`
	} else {
		sample = "/my/test//c/../path/file.log"
	}

	inputRaw := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"sourcePath": filepath.ToSlash(sample)},
			map[string]interface{}{"sourcePath": filepath.ToSlash(sample)},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"sourcePath": filepath.ToSlash(sample), "basePath": "file.log"},
			map[string]interface{}{"sourcePath": filepath.ToSlash(sample), "basePath": "file.log"},
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
