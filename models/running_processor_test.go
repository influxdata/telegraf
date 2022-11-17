package models_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/testutil"
)

// MockProcessor is a Processor with an overridable Apply implementation.
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

// MockProcessorToInit is a Processor that needs to be initialized.
type MockProcessorToInit struct {
	HasBeenInit bool
}

func (p *MockProcessorToInit) SampleConfig() string {
	return ""
}

func (p *MockProcessorToInit) Description() string {
	return ""
}

func (p *MockProcessorToInit) Apply(in ...telegraf.Metric) []telegraf.Metric {
	return in
}

func (p *MockProcessorToInit) Init() error {
	p.HasBeenInit = true
	return nil
}

func TestRunningProcessor_Init(t *testing.T) {
	mock := MockProcessorToInit{}
	rp := &models.RunningProcessor{
		Processor: processors.NewStreamingProcessorFromProcessor(&mock),
	}
	err := rp.Init()
	require.NoError(t, err)
	require.True(t, mock.HasBeenInit)
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
		Processor telegraf.StreamingProcessor
		Config    *models.ProcessorConfig
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
				Processor: processors.NewStreamingProcessorFromProcessor(TagProcessor("apply", "true")),
				Config: &models.ProcessorConfig{
					Filter: models.Filter{},
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
				Processor: processors.NewStreamingProcessorFromProcessor(TagProcessor("apply", "true")),
				Config: &models.ProcessorConfig{
					Filter: models.Filter{
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
				Processor: processors.NewStreamingProcessorFromProcessor(TagProcessor("apply", "true")),
				Config: &models.ProcessorConfig{
					Filter: models.Filter{
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
			rp := &models.RunningProcessor{
				Processor: tt.args.Processor,
				Config:    tt.args.Config,
			}
			err := rp.Config.Filter.Compile()
			require.NoError(t, err)

			acc := testutil.Accumulator{}
			err = rp.Start(&acc)
			require.NoError(t, err)
			for _, m := range tt.input {
				err = rp.Add(m, &acc)
				require.NoError(t, err)
			}
			rp.Stop()

			actual := acc.GetTelegrafMetrics()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestRunningProcessor_Order(t *testing.T) {
	rp1 := &models.RunningProcessor{
		Config: &models.ProcessorConfig{
			Order: 1,
		},
	}
	rp2 := &models.RunningProcessor{
		Config: &models.ProcessorConfig{
			Order: 2,
		},
	}
	rp3 := &models.RunningProcessor{
		Config: &models.ProcessorConfig{
			Order: 3,
		},
	}

	procs := models.RunningProcessors{rp2, rp3, rp1}
	sort.Sort(procs)
	require.Equal(t,
		models.RunningProcessors{rp1, rp2, rp3},
		procs)
}
