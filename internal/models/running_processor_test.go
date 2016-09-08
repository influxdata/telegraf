package models

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

type TestProcessor struct {
}

func (f *TestProcessor) SampleConfig() string { return "" }
func (f *TestProcessor) Description() string  { return "" }

// Apply renames:
//   "foo" to "fuz"
//   "bar" to "baz"
// And it also drops measurements named "dropme"
func (f *TestProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0)
	for _, m := range in {
		switch m.Name() {
		case "foo":
			out = append(out, testutil.TestMetric(1, "fuz"))
		case "bar":
			out = append(out, testutil.TestMetric(1, "baz"))
		case "dropme":
			// drop the metric!
		default:
			out = append(out, m)
		}
	}
	return out
}

func NewTestRunningProcessor() *RunningProcessor {
	out := &RunningProcessor{
		Name:      "test",
		Processor: &TestProcessor{},
		Config:    &ProcessorConfig{Filter: Filter{}},
	}
	return out
}

func TestRunningProcessor(t *testing.T) {
	inmetrics := []telegraf.Metric{
		testutil.TestMetric(1, "foo"),
		testutil.TestMetric(1, "bar"),
		testutil.TestMetric(1, "baz"),
	}

	expectedNames := []string{
		"fuz",
		"baz",
		"baz",
	}
	rfp := NewTestRunningProcessor()
	filteredMetrics := rfp.Apply(inmetrics...)

	actualNames := []string{
		filteredMetrics[0].Name(),
		filteredMetrics[1].Name(),
		filteredMetrics[2].Name(),
	}
	assert.Equal(t, expectedNames, actualNames)
}

func TestRunningProcessor_WithNameDrop(t *testing.T) {
	inmetrics := []telegraf.Metric{
		testutil.TestMetric(1, "foo"),
		testutil.TestMetric(1, "bar"),
		testutil.TestMetric(1, "baz"),
	}

	expectedNames := []string{
		"foo",
		"baz",
		"baz",
	}
	rfp := NewTestRunningProcessor()

	rfp.Config.Filter.NameDrop = []string{"foo"}
	assert.NoError(t, rfp.Config.Filter.Compile())

	filteredMetrics := rfp.Apply(inmetrics...)

	actualNames := []string{
		filteredMetrics[0].Name(),
		filteredMetrics[1].Name(),
		filteredMetrics[2].Name(),
	}
	assert.Equal(t, expectedNames, actualNames)
}

func TestRunningProcessor_DroppedMetric(t *testing.T) {
	inmetrics := []telegraf.Metric{
		testutil.TestMetric(1, "dropme"),
		testutil.TestMetric(1, "foo"),
		testutil.TestMetric(1, "bar"),
	}

	expectedNames := []string{
		"fuz",
		"baz",
	}
	rfp := NewTestRunningProcessor()
	filteredMetrics := rfp.Apply(inmetrics...)

	actualNames := []string{
		filteredMetrics[0].Name(),
		filteredMetrics[1].Name(),
	}
	assert.Equal(t, expectedNames, actualNames)
}
