package models

import (
	"sort"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

// MockProcessor is a Processor with an overrideable Apply implementation.
type MockProcessor struct {
	ApplyF func(in ...telegraf.Metric) []telegraf.Metric
}

func (p *MockProcessor) SampleConfig() string {
	return ""
}

func (p *MockProcessor) Description() string {
	return ""
}

func (p *MockProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	return p.ApplyF(in...)
}

// TagProcessor returns a Processor whose Apply function adds the tag and
// value.
func TagProcessor(key, value string) *MockProcessor {
	return &MockProcessor{
		ApplyF: func(in ...telegraf.Metric) []telegraf.Metric {
			for _, m := range in {
				m.AddTag(key, value)
			}
			return in
		},
	}
}

func TestRunningProcessor_Apply(t *testing.T) {
	type args struct {
		Processor telegraf.Processor
		Config    *ProcessorConfig
	}

	tests := []struct {
		name     string
		args     args
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "inactive filter applies metrics",
			args: args{
				Processor: TagProcessor("apply", "true"),
				Config: &ProcessorConfig{
					Filter: Filter{},
				},
			},
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"apply": "true",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "filter applies",
			args: args{
				Processor: TagProcessor("apply", "true"),
				Config: &ProcessorConfig{
					Filter: Filter{
						NamePass: []string{"cpu"},
					},
				},
			},
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"apply": "true",
					},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "filter doesn't apply",
			args: args{
				Processor: TagProcessor("apply", "true"),
				Config: &ProcessorConfig{
					Filter: Filter{
						NameDrop: []string{"cpu"},
					},
				},
			},
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &RunningProcessor{
				Processor: tt.args.Processor,
				Config:    tt.args.Config,
			}
			rp.Config.Filter.Compile()

			actual := rp.Apply(tt.input...)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestRunningProcessor_Order(t *testing.T) {
	rp1 := &RunningProcessor{
		Config: &ProcessorConfig{
			Order: 1,
		},
	}
	rp2 := &RunningProcessor{
		Config: &ProcessorConfig{
			Order: 2,
		},
	}
	rp3 := &RunningProcessor{
		Config: &ProcessorConfig{
			Order: 3,
		},
	}

	procs := RunningProcessors{rp2, rp3, rp1}
	sort.Sort(procs)
	require.Equal(t,
		RunningProcessors{rp1, rp2, rp3},
		procs)
}
