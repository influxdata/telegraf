package batch

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

const batchTag = "?internal_batch_idx"

func MakeBatching(batches uint64) *Batch {
	return &Batch{
		BatchTag:   batchTag,
		NumBatches: batches,
	}
}

func MakeXMetrics(count int) []telegraf.Metric {
	ms := make([]telegraf.Metric, 0, count)
	for range count {
		ms = append(ms, testutil.MockMetrics()...)
	}

	return ms
}

func requireMetricInBatch(t *testing.T, m telegraf.Metric, batch string) {
	batchTagValue, ok := m.GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, batch, batchTagValue)
}

func Test_SingleMetricPutInBatch0(t *testing.T) {
	b := MakeBatching(1)
	m := testutil.MockMetricsWithValue(1)
	expectedM := testutil.MockMetricsWithValue(1)
	expectedM[0].AddTag(batchTag, "0")

	res := b.Apply(m...)
	testutil.RequireMetricsEqual(t, expectedM, res)
}

func Test_MetricsSmallerThanBatchSizeAreInDifferentBatches(t *testing.T) {
	b := MakeBatching(3)
	ms := MakeXMetrics(2)
	res := b.Apply(ms...)
	requireMetricInBatch(t, res[0], "0")
	requireMetricInBatch(t, res[1], "1")
}

func Test_MetricsEqualToBatchSizeInDifferentBatches(t *testing.T) {
	b := MakeBatching(3)
	ms := MakeXMetrics(3)
	res := b.Apply(ms...)
	requireMetricInBatch(t, res[0], "0")
	requireMetricInBatch(t, res[1], "1")
	requireMetricInBatch(t, res[2], "2")
}

func Test_MetricsMoreThanBatchSizeInSameBatch(t *testing.T) {
	b := MakeBatching(2)
	ms := MakeXMetrics(3)
	res := b.Apply(ms...)

	requireMetricInBatch(t, res[0], "0")
	requireMetricInBatch(t, res[1], "1")
	requireMetricInBatch(t, res[2], "0")
}
