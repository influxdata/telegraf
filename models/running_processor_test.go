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

func TestRunningProcessorInit(t *testing.T) {
	mock := mockProcessor{}
	rp := &models.RunningProcessor{
		Processor: processors.NewStreamingProcessorFromProcessor(&mock),
	}
	require.NoError(t, rp.Init())
	require.True(t, mock.hasBeenInit)
}

func TestRunningProcessorApply(t *testing.T) {
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
				Processor: processors.NewStreamingProcessorFromProcessor(
					&mockProcessor{
						applyF: func(in ...telegraf.Metric) []telegraf.Metric {
							for _, m := range in {
								m.AddTag("apply", "true")
							}
							return in
						},
					},
				),
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
				Processor: processors.NewStreamingProcessorFromProcessor(
					&mockProcessor{
						applyF: func(in ...telegraf.Metric) []telegraf.Metric {
							for _, m := range in {
								m.AddTag("apply", "true")
							}
							return in
						},
					},
				),
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
				Processor: processors.NewStreamingProcessorFromProcessor(
					&mockProcessor{
						applyF: func(in ...telegraf.Metric) []telegraf.Metric {
							for _, m := range in {
								m.AddTag("apply", "true")
							}
							return in
						},
					},
				),
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

func TestRunningProcessorOrder(t *testing.T) {
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

// mockProcessor is a processor with an overridable apply implementation.
type mockProcessor struct {
	applyF      func(in ...telegraf.Metric) []telegraf.Metric
	hasBeenInit bool
}

func (*mockProcessor) SampleConfig() string {
	return ""
}

func (p *mockProcessor) Init() error {
	p.hasBeenInit = true
	return nil
}

func (p *mockProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	return p.applyF(in...)
}
