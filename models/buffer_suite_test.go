package models

import (
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	models "github.com/influxdata/telegraf/models/mock"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BufferSuiteTest struct {
	suite.Suite
	bufferType string
	bufferPath string
}

func (suite *BufferSuiteTest) SetupTest() {
	if suite.bufferType == "disk" {
		path, err := os.MkdirTemp("", "*-buffer-test")
		suite.NoError(err)
		suite.bufferPath = path
	}
}

func (suite *BufferSuiteTest) TearDownTest() {
	if suite.bufferPath != "" {
		_ = os.RemoveAll(suite.bufferPath)
		suite.bufferPath = ""
	}
}

func TestMemoryBufferSuite(t *testing.T) {
	suite.Run(t, &BufferSuiteTest{bufferType: "memory"})
}

func TestDiskBufferSuite(t *testing.T) {
	suite.Run(t, &BufferSuiteTest{bufferType: "disk"})
}

func Metric() telegraf.Metric {
	return MetricTime(0)
}

func MetricTime(sec int64) telegraf.Metric {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(sec, 0),
	)
	return m
}

func (suite *BufferSuiteTest) newTestBuffer(capacity int) Buffer {
	suite.T().Helper()
	buf, err := NewBuffer("test", "", capacity, suite.bufferType, suite.bufferPath)
	require.NoError(suite.T(), err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	return buf
}

func (suite *BufferSuiteTest) TestBuffer_LenEmpty() {
	b := suite.newTestBuffer(5)

	suite.Equal(0, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_LenOne() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m)

	suite.Equal(1, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_LenFull() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m, m, m)

	suite.Equal(5, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_LenOverfill() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m, m, m, m)

	suite.Equal(5, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_BatchLenZero() {
	b := suite.newTestBuffer(5)
	batch := b.Batch(0)

	suite.Empty(batch)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLenBufferEmpty() {
	b := suite.newTestBuffer(5)
	batch := b.Batch(2)

	suite.Empty(batch)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLenUnderfill() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m)
	batch := b.Batch(2)

	suite.Len(batch, 1)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLenFill() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	suite.Len(batch, 2)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLenExact() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m)
	batch := b.Batch(2)
	suite.Len(batch, 2)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLenLargerThanBuffer() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(6)
	suite.Len(batch, 5)
}

func (suite *BufferSuiteTest) TestBuffer_BatchWrap() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	b.Add(m, m)
	batch = b.Batch(5)
	suite.Len(batch, 5)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLatest() {
	b := suite.newTestBuffer(4)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_BatchLatestWrap() {
	b := suite.newTestBuffer(4)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(2),
			MetricTime(3),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_MultipleBatch() {
	b := suite.newTestBuffer(10)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	batch := b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
	b.Accept(batch)
	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(6),
		}, batch)
	b.Accept(batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectWithRoom() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Reject(batch)

	suite.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectNothingNewFull() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	batch := b.Batch(2)
	b.Reject(batch)

	suite.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectNoRoom() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))

	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))

	b.Reject(batch)

	suite.Equal(int64(3), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(4),
			MetricTime(5),
			MetricTime(6),
			MetricTime(7),
			MetricTime(8),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectRoomExact() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	batch := b.Batch(2)
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	b.Reject(batch)

	suite.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectRoomOverwriteOld() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(1)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))

	b.Reject(batch)

	suite.Equal(int64(1), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
			MetricTime(6),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectPartialRoom() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))

	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Reject(batch)

	suite.Equal(int64(2), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
			MetricTime(6),
			MetricTime(7),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectNewMetricsWrapped() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	// buffer: 1, 4, 5; batch: 2, 3
	suite.Equal(int64(0), b.Stats().MetricsDropped.Get())

	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	b.Add(MetricTime(9))
	b.Add(MetricTime(10))

	// buffer: 8, 9, 10, 6, 7; batch: 2, 3
	suite.Equal(int64(3), b.Stats().MetricsDropped.Get())

	b.Add(MetricTime(11))
	b.Add(MetricTime(12))
	b.Add(MetricTime(13))
	b.Add(MetricTime(14))
	b.Add(MetricTime(15))
	// buffer: 13, 14, 15, 11, 12; batch: 2, 3
	suite.Equal(int64(8), b.Stats().MetricsDropped.Get())
	b.Reject(batch)

	suite.Equal(int64(10), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(11),
			MetricTime(12),
			MetricTime(13),
			MetricTime(14),
			MetricTime(15),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectWrapped() {
	b := suite.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	batch := b.Batch(3)

	b.Add(MetricTime(9))
	b.Add(MetricTime(10))
	b.Add(MetricTime(11))
	b.Add(MetricTime(12))

	b.Reject(batch)

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(8),
			MetricTime(9),
			MetricTime(10),
			MetricTime(11),
			MetricTime(12),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_RejectAdjustFirst() {
	b := suite.newTestBuffer(10)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(3)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Reject(batch)

	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	b.Add(MetricTime(9))
	batch = b.Batch(3)
	b.Add(MetricTime(10))
	b.Add(MetricTime(11))
	b.Add(MetricTime(12))
	b.Reject(batch)

	b.Add(MetricTime(13))
	b.Add(MetricTime(14))
	b.Add(MetricTime(15))
	batch = b.Batch(3)
	b.Add(MetricTime(16))
	b.Add(MetricTime(17))
	b.Add(MetricTime(18))
	b.Reject(batch)

	b.Add(MetricTime(19))

	batch = b.Batch(10)
	testutil.RequireMetricsEqual(suite.T(),
		[]telegraf.Metric{
			MetricTime(10),
			MetricTime(11),
			MetricTime(12),
			MetricTime(13),
			MetricTime(14),
			MetricTime(15),
			MetricTime(16),
			MetricTime(17),
			MetricTime(18),
			MetricTime(19),
		}, batch)
}

func (suite *BufferSuiteTest) TestBuffer_AddDropsOverwrittenMetrics() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	b.Add(m, m, m, m, m)

	suite.Equal(int64(5), b.Stats().MetricsDropped.Get())
	suite.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (suite *BufferSuiteTest) TestBuffer_AcceptRemovesBatch() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	suite.Equal(1, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_RejectLeavesBatch() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	suite.Equal(3, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_AcceptWritesOverwrittenBatch() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Accept(batch)

	suite.Equal(int64(0), b.Stats().MetricsDropped.Get())
	suite.Equal(int64(5), b.Stats().MetricsWritten.Get())
}

func (suite *BufferSuiteTest) TestBuffer_BatchRejectDropsOverwrittenBatch() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Reject(batch)

	suite.Equal(int64(5), b.Stats().MetricsDropped.Get())
	suite.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (suite *BufferSuiteTest) TestBuffer_MetricsOverwriteBatchAccept() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Accept(batch)
	suite.Equal(int64(0), b.Stats().MetricsDropped.Get(), "dropped")
	suite.Equal(int64(3), b.Stats().MetricsWritten.Get(), "written")
}

func (suite *BufferSuiteTest) TestBuffer_MetricsOverwriteBatchReject() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Reject(batch)
	suite.Equal(int64(3), b.Stats().MetricsDropped.Get())
	suite.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (suite *BufferSuiteTest) TestBuffer_MetricsBatchAcceptRemoved() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m, m, m)
	b.Accept(batch)
	suite.Equal(int64(2), b.Stats().MetricsDropped.Get())
	suite.Equal(int64(3), b.Stats().MetricsWritten.Get())
}

func (suite *BufferSuiteTest) TestBuffer_WrapWithBatch() {
	m := Metric()
	b := suite.newTestBuffer(5)

	b.Add(m, m, m)
	b.Batch(3)
	b.Add(m, m, m, m, m, m)

	suite.Equal(int64(1), b.Stats().MetricsDropped.Get())
}

func (suite *BufferSuiteTest) TestBuffer_BatchNotRemoved() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	b.Batch(2)
	suite.Equal(5, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_BatchRejectAcceptNoop() {
	m := Metric()
	b := suite.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	b.Accept(batch)
	suite.Equal(5, b.Len())
}

func (suite *BufferSuiteTest) TestBuffer_AcceptCallsMetricAccept() {
	var accept int
	mm := &models.MockMetric{
		Metric: Metric(),
		AcceptF: func() {
			accept++
		},
	}
	b := suite.newTestBuffer(5)
	b.Add(mm, mm, mm)
	batch := b.Batch(2)
	b.Accept(batch)
	suite.Equal(2, accept)
}

func (suite *BufferSuiteTest) TestBuffer_AddCallsMetricRejectWhenNoBatch() {
	var reject int
	mm := &models.MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := suite.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm)
	suite.Equal(2, reject)
}

func (suite *BufferSuiteTest) TestBuffer_AddCallsMetricRejectWhenNotInBatch() {
	var reject int
	mm := &models.MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := suite.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(2)
	b.Add(mm, mm, mm, mm)
	suite.Equal(2, reject)
	b.Reject(batch)
	suite.Equal(4, reject)
}

func (suite *BufferSuiteTest) TestBuffer_RejectCallsMetricRejectWithOverwritten() {
	var reject int
	mm := &models.MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := suite.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm)
	suite.Equal(0, reject)
	b.Reject(batch)
	suite.Equal(2, reject)
}

func (suite *BufferSuiteTest) TestBuffer_AddOverwriteAndReject() {
	var reject int
	mm := &models.MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := suite.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	suite.Equal(15, reject)
	b.Reject(batch)
	suite.Equal(20, reject)
}

func (suite *BufferSuiteTest) TestBuffer_AddOverwriteAndRejectOffset() {
	var reject int
	var accept int
	mm := &models.MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
		AcceptF: func() {
			accept++
		},
	}
	b := suite.newTestBuffer(5)
	b.Add(mm, mm, mm)
	b.Add(mm, mm, mm, mm)
	suite.Equal(2, reject)
	batch := b.Batch(5)
	b.Add(mm, mm, mm, mm)
	suite.Equal(2, reject)
	b.Add(mm, mm, mm, mm)
	suite.Equal(5, reject)
	b.Add(mm, mm, mm, mm)
	suite.Equal(9, reject)
	b.Add(mm, mm, mm, mm)
	suite.Equal(13, reject)
	b.Accept(batch)
	suite.Equal(13, reject)
	suite.Equal(5, accept)
}

func (suite *BufferSuiteTest) TestBuffer_RejectEmptyBatch() {
	b := suite.newTestBuffer(5)
	batch := b.Batch(2)
	b.Add(MetricTime(1))
	b.Reject(batch)
	b.Add(MetricTime(2))
	batch = b.Batch(2)
	for _, m := range batch {
		suite.NotNil(m)
	}
}

// Benchmark test outside the suite
func BenchmarkMemoryAddMetrics(b *testing.B) {
	buf, err := NewBuffer("test", "", 10000, "memory", "")
	require.NoError(b, err)
	m := Metric()
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}
