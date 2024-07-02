package models

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type MockMetric struct {
	telegraf.Metric
	AcceptF func()
	RejectF func()
	DropF   func()
}

func (m *MockMetric) Accept() {
	m.AcceptF()
}

func (m *MockMetric) Reject() {
	m.RejectF()
}

func (m *MockMetric) Drop() {
	m.DropF()
}

type BufferSuiteTest struct {
	suite.Suite
	bufferType string
	bufferPath string

	hasMaxCapacity bool // whether the buffer type being tested supports a maximum metric capacity
}

func (s *BufferSuiteTest) SetupTest() {
	switch s.bufferType {
	case "", "memory":
		s.hasMaxCapacity = true
	case "disk":
		path, err := os.MkdirTemp("", "*-buffer-test")
		s.Require().NoError(err)
		s.bufferPath = path
		s.hasMaxCapacity = false
	}
}

func (s *BufferSuiteTest) TearDownTest() {
	if s.bufferPath != "" {
		os.RemoveAll(s.bufferPath)
		s.bufferPath = ""
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

func (s *BufferSuiteTest) newTestBuffer(capacity int) Buffer {
	s.T().Helper()
	buf, err := NewBuffer("test", "", capacity, s.bufferType, s.bufferPath)
	s.Require().NoError(err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	return buf
}

func (s *BufferSuiteTest) TestBuffer_LenEmpty() {
	b := s.newTestBuffer(5)

	s.Equal(0, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_LenOne() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m)

	s.Equal(1, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_LenFull() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)

	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_LenOverfill() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m, m)

	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_BatchLenZero() {
	b := s.newTestBuffer(5)
	batch := b.Batch(0)

	s.Empty(batch)
}

func (s *BufferSuiteTest) TestBuffer_BatchLenBufferEmpty() {
	b := s.newTestBuffer(5)
	batch := b.Batch(2)

	s.Empty(batch)
}

func (s *BufferSuiteTest) TestBuffer_BatchLenUnderfill() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m)
	batch := b.Batch(2)

	s.Len(batch, 1)
}

func (s *BufferSuiteTest) TestBuffer_BatchLenFill() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	s.Len(batch, 2)
}

func (s *BufferSuiteTest) TestBuffer_BatchLenExact() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m)
	batch := b.Batch(2)
	s.Len(batch, 2)
}

func (s *BufferSuiteTest) TestBuffer_BatchLenLargerThanBuffer() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(6)
	s.Len(batch, 5)
}

func (s *BufferSuiteTest) TestBuffer_BatchWrap() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	b.Add(m, m)
	batch = b.Batch(5)
	s.Len(batch, 5)
}

func (s *BufferSuiteTest) TestBuffer_BatchLatest() {
	b := s.newTestBuffer(4)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_BatchLatestWrap() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(4)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(2),
			MetricTime(3),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_MultipleBatch() {
	b := s.newTestBuffer(10)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	batch := b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
	b.Accept(batch)
	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(6),
		}, batch)
	b.Accept(batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectWithRoom() {
	b := s.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Reject(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectNothingNewFull() {
	b := s.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	batch := b.Batch(2)
	b.Reject(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectNoRoom() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
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

	s.Equal(int64(3), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(4),
			MetricTime(5),
			MetricTime(6),
			MetricTime(7),
			MetricTime(8),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectRoomExact() {
	b := s.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	batch := b.Batch(2)
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	b.Reject(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectRoomOverwriteOld() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(1)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))

	b.Reject(batch)

	s.Equal(int64(1), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(2),
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
			MetricTime(6),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectPartialRoom() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(MetricTime(1))

	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Reject(batch)

	s.Equal(int64(2), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(3),
			MetricTime(4),
			MetricTime(5),
			MetricTime(6),
			MetricTime(7),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectNewMetricsWrapped() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	// buffer: 1, 4, 5; batch: 2, 3
	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	b.Add(MetricTime(9))
	b.Add(MetricTime(10))

	// buffer: 8, 9, 10, 6, 7; batch: 2, 3
	s.Equal(int64(3), b.Stats().MetricsDropped.Get())

	b.Add(MetricTime(11))
	b.Add(MetricTime(12))
	b.Add(MetricTime(13))
	b.Add(MetricTime(14))
	b.Add(MetricTime(15))
	// buffer: 13, 14, 15, 11, 12; batch: 2, 3
	s.Equal(int64(8), b.Stats().MetricsDropped.Get())
	b.Reject(batch)

	s.Equal(int64(10), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(11),
			MetricTime(12),
			MetricTime(13),
			MetricTime(14),
			MetricTime(15),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectWrapped() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
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
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			MetricTime(8),
			MetricTime(9),
			MetricTime(10),
			MetricTime(11),
			MetricTime(12),
		}, batch)
}

func (s *BufferSuiteTest) TestBuffer_RejectAdjustFirst() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(10)
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
	testutil.RequireMetricsEqual(s.T(),
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

func (s *BufferSuiteTest) TestBuffer_AddDropsOverwrittenMetrics() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	b.Add(m, m, m, m, m)

	s.Equal(int64(5), b.Stats().MetricsDropped.Get())
	s.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBuffer_AcceptRemovesBatch() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	s.Equal(1, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_RejectLeavesBatch() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	s.Equal(3, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_AcceptWritesOverwrittenBatch() {
	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Accept(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())
	s.Equal(int64(5), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBuffer_BatchRejectDropsOverwrittenBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Reject(batch)

	s.Equal(int64(5), b.Stats().MetricsDropped.Get())
	s.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBuffer_MetricsOverwriteBatchAccept() {
	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Accept(batch)
	s.Equal(int64(0), b.Stats().MetricsDropped.Get(), "dropped")
	s.Equal(int64(3), b.Stats().MetricsWritten.Get(), "written")
}

func (s *BufferSuiteTest) TestBuffer_MetricsOverwriteBatchReject() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Reject(batch)
	s.Equal(int64(3), b.Stats().MetricsDropped.Get())
	s.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBuffer_MetricsBatchAcceptRemoved() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m, m, m)
	b.Accept(batch)
	s.Equal(int64(2), b.Stats().MetricsDropped.Get())
	s.Equal(int64(3), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBuffer_WrapWithBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := Metric()
	b := s.newTestBuffer(5)

	b.Add(m, m, m)
	b.Batch(3)
	b.Add(m, m, m, m, m, m)

	s.Equal(int64(1), b.Stats().MetricsDropped.Get())
}

func (s *BufferSuiteTest) TestBuffer_BatchNotRemoved() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	b.Batch(2)
	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_BatchRejectAcceptNoop() {
	m := Metric()
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	b.Accept(batch)
	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBuffer_AddCallsMetricRejectWhenNoBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := s.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm)
	s.Equal(2, reject)
}

func (s *BufferSuiteTest) TestBuffer_AddCallsMetricRejectWhenNotInBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := s.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(2)
	b.Add(mm, mm, mm, mm)
	s.Equal(2, reject)
	b.Reject(batch)
	s.Equal(4, reject)
}

func (s *BufferSuiteTest) TestBuffer_RejectCallsMetricRejectWithOverwritten() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := s.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm)
	s.Equal(0, reject)
	b.Reject(batch)
	s.Equal(2, reject)
}

func (s *BufferSuiteTest) TestBuffer_AddOverwriteAndReject() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := s.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	s.Equal(15, reject)
	b.Reject(batch)
	s.Equal(20, reject)
}

func (s *BufferSuiteTest) TestBuffer_AddOverwriteAndRejectOffset() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	var accept int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
		AcceptF: func() {
			accept++
		},
	}
	b := s.newTestBuffer(5)
	b.Add(mm, mm, mm)
	b.Add(mm, mm, mm, mm)
	s.Equal(2, reject)
	batch := b.Batch(5)
	b.Add(mm, mm, mm, mm)
	s.Equal(2, reject)
	b.Add(mm, mm, mm, mm)
	s.Equal(5, reject)
	b.Add(mm, mm, mm, mm)
	s.Equal(9, reject)
	b.Add(mm, mm, mm, mm)
	s.Equal(13, reject)
	b.Accept(batch)
	s.Equal(13, reject)
	s.Equal(5, accept)
}

func (s *BufferSuiteTest) TestBuffer_RejectEmptyBatch() {
	b := s.newTestBuffer(5)
	batch := b.Batch(2)
	b.Add(MetricTime(1))
	b.Reject(batch)
	b.Add(MetricTime(2))
	batch = b.Batch(2)
	for _, m := range batch {
		s.NotNil(m)
	}
}
