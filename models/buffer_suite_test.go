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

func (s *BufferSuiteTest) newTestBuffer(capacity int) Buffer {
	s.T().Helper()
	buf, err := NewBuffer("test", "123", "", capacity, s.bufferType, s.bufferPath)
	s.Require().NoError(err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	return buf
}

func (s *BufferSuiteTest) TestBufferLenEmpty() {
	b := s.newTestBuffer(5)

	s.Equal(0, b.Len())
}

func (s *BufferSuiteTest) TestBufferLenOne() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m)

	s.Equal(1, b.Len())
}

func (s *BufferSuiteTest) TestBufferLenFull() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)

	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBufferLenOverfill() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m, m)

	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBufferBatchLenZero() {
	b := s.newTestBuffer(5)
	batch := b.Batch(0)

	s.Empty(batch)
}

func (s *BufferSuiteTest) TestBufferBatchLenBufferEmpty() {
	b := s.newTestBuffer(5)
	batch := b.Batch(2)

	s.Empty(batch)
}

func (s *BufferSuiteTest) TestBufferBatchLenUnderfill() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m)
	batch := b.Batch(2)

	s.Len(batch, 1)
}

func (s *BufferSuiteTest) TestBufferBatchLenFill() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	s.Len(batch, 2)
}

func (s *BufferSuiteTest) TestBufferBatchLenExact() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m)
	batch := b.Batch(2)
	s.Len(batch, 2)
}

func (s *BufferSuiteTest) TestBufferBatchLenLargerThanBuffer() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(6)
	s.Len(batch, 5)
}

func (s *BufferSuiteTest) TestBufferBatchWrap() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	b.Add(m, m)
	batch = b.Batch(5)
	s.Len(batch, 5)
}

func (s *BufferSuiteTest) TestBufferBatchLatest() {
	b := s.newTestBuffer(4)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferBatchLatestWrap() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(4)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferMultipleBatch() {
	b := s.newTestBuffer(10)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	batch := b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
		}, batch)
	b.Accept(batch)
	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)),
		}, batch)
	b.Accept(batch)
}

func (s *BufferSuiteTest) TestBufferRejectWithRoom() {
	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(2)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	b.Reject(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectNothingNewFull() {
	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	batch := b.Batch(2)
	b.Reject(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectNoRoom() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(2)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	b.Reject(batch)

	s.Equal(int64(3), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectRoomExact() {
	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	batch := b.Batch(2)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))

	b.Reject(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectRoomOverwriteOld() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(1)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))

	b.Reject(batch)

	s.Equal(int64(1), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectPartialRoom() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(2)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	b.Reject(batch)

	s.Equal(int64(2), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectNewMetricsWrapped() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(2)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))

	// buffer: 1, 4, 5; batch: 2, 3
	s.Equal(int64(0), b.Stats().MetricsDropped.Get())

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)))

	// buffer: 8, 9, 10, 6, 7; batch: 2, 3
	s.Equal(int64(3), b.Stats().MetricsDropped.Get())

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(13, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(14, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(15, 0)))
	// buffer: 13, 14, 15, 11, 12; batch: 2, 3
	s.Equal(int64(8), b.Stats().MetricsDropped.Get())
	b.Reject(batch)

	s.Equal(int64(10), b.Stats().MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(13, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(14, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(15, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectWrapped() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	batch := b.Batch(3)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)))

	b.Reject(batch)

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferRejectAdjustFirst() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	b := s.newTestBuffer(10)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(3)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	b.Reject(batch)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)))
	batch = b.Batch(3)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)))
	b.Reject(batch)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(13, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(14, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(15, 0)))
	batch = b.Batch(3)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(16, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(17, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(18, 0)))
	b.Reject(batch)

	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(19, 0)))

	batch = b.Batch(10)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(13, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(14, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(15, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(16, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(17, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(18, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(19, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferAddDropsOverwrittenMetrics() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	b.Add(m, m, m, m, m)

	s.Equal(int64(5), b.Stats().MetricsDropped.Get())
	s.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferAcceptRemovesBatch() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	s.Equal(1, b.Len())
}

func (s *BufferSuiteTest) TestBufferRejectLeavesBatch() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	s.Equal(3, b.Len())
}

func (s *BufferSuiteTest) TestBufferAcceptWritesOverwrittenBatch() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Accept(batch)

	s.Equal(int64(0), b.Stats().MetricsDropped.Get())
	s.Equal(int64(5), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferBatchRejectDropsOverwrittenBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Reject(batch)

	s.Equal(int64(5), b.Stats().MetricsDropped.Get())
	s.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferMetricsOverwriteBatchAccept() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Accept(batch)
	s.Equal(int64(0), b.Stats().MetricsDropped.Get(), "dropped")
	s.Equal(int64(3), b.Stats().MetricsWritten.Get(), "written")
}

func (s *BufferSuiteTest) TestBufferMetricsOverwriteBatchReject() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Reject(batch)
	s.Equal(int64(3), b.Stats().MetricsDropped.Get())
	s.Equal(int64(0), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferMetricsBatchAcceptRemoved() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m, m, m)
	b.Accept(batch)
	s.Equal(int64(2), b.Stats().MetricsDropped.Get())
	s.Equal(int64(3), b.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferWrapWithBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)

	b.Add(m, m, m)
	b.Batch(3)
	b.Add(m, m, m, m, m, m)

	s.Equal(int64(1), b.Stats().MetricsDropped.Get())
}

func (s *BufferSuiteTest) TestBufferBatchNotRemoved() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	b.Batch(2)
	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBufferBatchRejectAcceptNoop() {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	b := s.newTestBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	b.Accept(batch)
	s.Equal(5, b.Len())
}

func (s *BufferSuiteTest) TestBufferAddCallsMetricRejectWhenNoBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
		RejectF: func() {
			reject++
		},
	}
	b := s.newTestBuffer(5)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm)
	s.Equal(2, reject)
}

func (s *BufferSuiteTest) TestBufferAddCallsMetricRejectWhenNotInBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
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

func (s *BufferSuiteTest) TestBufferRejectCallsMetricRejectWithOverwritten() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
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

func (s *BufferSuiteTest) TestBufferAddOverwriteAndReject() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
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

func (s *BufferSuiteTest) TestBufferAddOverwriteAndRejectOffset() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	var reject int
	var accept int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
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

func (s *BufferSuiteTest) TestBufferRejectEmptyBatch() {
	b := s.newTestBuffer(5)
	batch := b.Batch(2)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Reject(batch)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	batch = b.Batch(2)
	for _, m := range batch {
		s.NotNil(m)
	}
}

func (s *BufferSuiteTest) TestBufferFlushedPartial() {
	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := b.Batch(2)
	s.Len(batch, 2)

	b.Accept(batch)
	s.Equal(1, b.Len())
}

func (s *BufferSuiteTest) TestBufferFlushedFull() {
	b := s.newTestBuffer(5)
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	b.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	batch := b.Batch(2)
	s.Len(batch, 2)

	b.Accept(batch)
	s.Equal(0, b.Len())
}

type mockMetric struct {
	telegraf.Metric
	AcceptF func()
	RejectF func()
	DropF   func()
}

func (m *mockMetric) Accept() {
	m.AcceptF()
}

func (m *mockMetric) Reject() {
	m.RejectF()
}

func (m *mockMetric) Drop() {
	m.DropF()
}
