package batch

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const batchTag = "?internal_batch_idx"

func Test_SingleMetricPutInBatch0(t *testing.T) {
	b := &Batch{
		BatchTag:   batchTag,
		NumBatches: 1,
	}
	m := testutil.MockMetricsWithValue(1)
	expectedM := testutil.MockMetricsWithValue(1)
	expectedM[0].AddTag(batchTag, "0")

	res := b.Apply(m...)
	testutil.RequireMetricsEqual(t, expectedM, res)
}

func Test_MetricsSmallerThanBatchSizeAreInDifferentBatches(t *testing.T) {
	b := &Batch{
		BatchTag:   batchTag,
		NumBatches: 3,
	}

	ms := make([]telegraf.Metric, 0, 2)
	for range cap(ms) {
		ms = append(ms, testutil.MockMetrics()...)
	}

	res := b.Apply(ms...)

	batchTagValue, ok := res[0].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "0", batchTagValue)

	batchTagValue, ok = res[1].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "1", batchTagValue)
}

func Test_MetricsEqualToBatchSizeInDifferentBatches(t *testing.T) {
	b := &Batch{
		BatchTag:   batchTag,
		NumBatches: 3,
	}

	ms := make([]telegraf.Metric, 0, 3)
	for range cap(ms) {
		ms = append(ms, testutil.MockMetrics()...)
	}

	res := b.Apply(ms...)
	batchTagValue, ok := res[0].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "0", batchTagValue)

	batchTagValue, ok = res[1].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "1", batchTagValue)

	batchTagValue, ok = res[2].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "2", batchTagValue)
}

func Test_MetricsMoreThanBatchSizeInSameBatch(t *testing.T) {
	b := &Batch{
		BatchTag:   batchTag,
		NumBatches: 2,
	}

	ms := make([]telegraf.Metric, 0, 3)
	for range cap(ms) {
		ms = append(ms, testutil.MockMetrics()...)
	}

	res := b.Apply(ms...)
	batchTagValue, ok := res[0].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "0", batchTagValue)

	batchTagValue, ok = res[1].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "1", batchTagValue)

	batchTagValue, ok = res[2].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "0", batchTagValue)
}

func Test_MetricWithExistingTagNotChanged(t *testing.T) {
	b := &Batch{
		BatchTag:     batchTag,
		NumBatches:   2,
		SkipExisting: true,
	}

	m := testutil.MockMetricsWithValue(1)
	m[0].AddTag(batchTag, "4")
	res := b.Apply(m...)
	tv, ok := res[0].GetTag(batchTag)
	require.True(t, ok)
	require.Equal(t, "4", tv)
}
