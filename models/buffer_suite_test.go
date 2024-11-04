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
		s.NoError(os.RemoveAll(s.bufferPath))
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
	buf := s.newTestBuffer(5)
	defer buf.Close()

	s.Equal(0, buf.Len())
}

func (s *BufferSuiteTest) TestBufferLenOne() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m)
	s.Equal(1, buf.Len())
}

func (s *BufferSuiteTest) TestBufferLenFull() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	s.Equal(5, buf.Len())
}

func (s *BufferSuiteTest) TestBufferLenOverfill() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m, m)
	s.Equal(5, buf.Len())
}

func (s *BufferSuiteTest) TestBufferBatchLenZero() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	batch := buf.Batch(0)
	s.Empty(batch)
}

func (s *BufferSuiteTest) TestBufferBatchLenBufferEmpty() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	batch := buf.Batch(2)
	s.Empty(batch)
}

func (s *BufferSuiteTest) TestBufferBatchLenUnderfill() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m)
	batch := buf.Batch(2)
	s.Len(batch, 1)
}

func (s *BufferSuiteTest) TestBufferBatchLenFill() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m)
	batch := buf.Batch(2)
	s.Len(batch, 2)
}

func (s *BufferSuiteTest) TestBufferBatchLenExact() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m)
	batch := buf.Batch(2)
	s.Len(batch, 2)
}

func (s *BufferSuiteTest) TestBufferBatchLenLargerThanBuffer() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(6)
	s.Len(batch, 5)
}

func (s *BufferSuiteTest) TestBufferBatchWrap() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(2)
	buf.Accept(batch)
	buf.Add(m, m)
	batch = buf.Batch(5)
	s.Len(batch, 5)
}

func (s *BufferSuiteTest) TestBufferBatchLatest() {
	buf := s.newTestBuffer(4)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(2)

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

	buf := s.newTestBuffer(4)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	batch := buf.Batch(2)

	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
		}, batch)
}

func (s *BufferSuiteTest) TestBufferMultipleBatch() {
	buf := s.newTestBuffer(10)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	batch := buf.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)),
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)),
		}, batch)
	buf.Accept(batch)
	batch = buf.Batch(5)
	testutil.RequireMetricsEqual(s.T(),
		[]telegraf.Metric{
			metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)),
		}, batch)
	buf.Accept(batch)
}

func (s *BufferSuiteTest) TestBufferRejectWithRoom() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(2)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	buf.Reject(batch)

	s.Equal(int64(0), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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
	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	batch := buf.Batch(2)
	buf.Reject(batch)

	s.Equal(int64(0), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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

	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(2)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	buf.Reject(batch)

	s.Equal(int64(3), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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
	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	batch := buf.Batch(2)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))

	buf.Reject(batch)

	s.Equal(int64(0), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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

	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(1)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))

	buf.Reject(batch)

	s.Equal(int64(1), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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

	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(2)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	buf.Reject(batch)

	s.Equal(int64(2), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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

	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(2)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))

	// buffer: 1, 4, 5; batch: 2, 3
	s.Equal(int64(0), buf.Stats().MetricsDropped.Get())

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)))

	// buffer: 8, 9, 10, 6, 7; batch: 2, 3
	s.Equal(int64(3), buf.Stats().MetricsDropped.Get())

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(13, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(14, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(15, 0)))
	// buffer: 13, 14, 15, 11, 12; batch: 2, 3
	s.Equal(int64(8), buf.Stats().MetricsDropped.Get())
	buf.Reject(batch)

	s.Equal(int64(10), buf.Stats().MetricsDropped.Get())

	batch = buf.Batch(5)
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

	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	batch := buf.Batch(3)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)))

	buf.Reject(batch)

	batch = buf.Batch(5)
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

	buf := s.newTestBuffer(10)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(3)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(4, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(5, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(6, 0)))
	buf.Reject(batch)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(7, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(8, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(9, 0)))
	batch = buf.Batch(3)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(10, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(11, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(12, 0)))
	buf.Reject(batch)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(13, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(14, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(15, 0)))
	batch = buf.Batch(3)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(16, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(17, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(18, 0)))
	buf.Reject(batch)

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(19, 0)))

	batch = buf.Batch(10)
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

	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	buf.Add(m, m, m, m, m)

	s.Equal(int64(5), buf.Stats().MetricsDropped.Get())
	s.Equal(int64(0), buf.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferAcceptRemovesBatch() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m)
	batch := buf.Batch(2)
	buf.Accept(batch)
	s.Equal(1, buf.Len())
}

func (s *BufferSuiteTest) TestBufferRejectLeavesBatch() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m)
	batch := buf.Batch(2)
	buf.Reject(batch)
	s.Equal(3, buf.Len())
}

func (s *BufferSuiteTest) TestBufferAcceptWritesOverwrittenBatch() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(5)
	buf.Add(m, m, m, m, m)
	buf.Accept(batch)

	s.Equal(int64(0), buf.Stats().MetricsDropped.Get())
	s.Equal(int64(5), buf.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferBatchRejectDropsOverwrittenBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(5)
	buf.Add(m, m, m, m, m)
	buf.Reject(batch)

	s.Equal(int64(5), buf.Stats().MetricsDropped.Get())
	s.Equal(int64(0), buf.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferMetricsOverwriteBatchAccept() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(3)
	buf.Add(m, m, m)
	buf.Accept(batch)
	s.Equal(int64(0), buf.Stats().MetricsDropped.Get(), "dropped")
	s.Equal(int64(3), buf.Stats().MetricsWritten.Get(), "written")
}

func (s *BufferSuiteTest) TestBufferMetricsOverwriteBatchReject() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(3)
	buf.Add(m, m, m)
	buf.Reject(batch)
	s.Equal(int64(3), buf.Stats().MetricsDropped.Get())
	s.Equal(int64(0), buf.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferMetricsBatchAcceptRemoved() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(3)
	buf.Add(m, m, m, m, m)
	buf.Accept(batch)
	s.Equal(int64(2), buf.Stats().MetricsDropped.Get())
	s.Equal(int64(3), buf.Stats().MetricsWritten.Get())
}

func (s *BufferSuiteTest) TestBufferWrapWithBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m)
	buf.Batch(3)
	buf.Add(m, m, m, m, m, m)

	s.Equal(int64(1), buf.Stats().MetricsDropped.Get())
}

func (s *BufferSuiteTest) TestBufferBatchNotRemoved() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	buf.Batch(2)
	s.Equal(5, buf.Len())
}

func (s *BufferSuiteTest) TestBufferBatchRejectAcceptNoop() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	buf.Add(m, m, m, m, m)
	batch := buf.Batch(2)
	buf.Reject(batch)
	buf.Accept(batch)
	s.Equal(5, buf.Len())
}

func (s *BufferSuiteTest) TestBufferAddCallsMetricRejectWhenNoBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
		RejectF: func() {
			reject++
		},
	}
	buf.Add(mm, mm, mm, mm, mm)
	buf.Add(mm, mm)
	s.Equal(2, reject)
}

func (s *BufferSuiteTest) TestBufferAddCallsMetricRejectWhenNotInBatch() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
		RejectF: func() {
			reject++
		},
	}
	buf.Add(mm, mm, mm, mm, mm)
	batch := buf.Batch(2)
	buf.Add(mm, mm, mm, mm)
	s.Equal(2, reject)
	buf.Reject(batch)
	s.Equal(4, reject)
}

func (s *BufferSuiteTest) TestBufferRejectCallsMetricRejectWithOverwritten() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
		RejectF: func() {
			reject++
		},
	}
	buf.Add(mm, mm, mm, mm, mm)
	batch := buf.Batch(5)
	buf.Add(mm, mm)
	s.Equal(0, reject)
	buf.Reject(batch)
	s.Equal(2, reject)
}

func (s *BufferSuiteTest) TestBufferAddOverwriteAndReject() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

	var reject int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
		RejectF: func() {
			reject++
		},
	}
	buf.Add(mm, mm, mm, mm, mm)
	batch := buf.Batch(5)
	buf.Add(mm, mm, mm, mm, mm)
	buf.Add(mm, mm, mm, mm, mm)
	buf.Add(mm, mm, mm, mm, mm)
	buf.Add(mm, mm, mm, mm, mm)
	s.Equal(15, reject)
	buf.Reject(batch)
	s.Equal(20, reject)
}

func (s *BufferSuiteTest) TestBufferAddOverwriteAndRejectOffset() {
	if !s.hasMaxCapacity {
		s.T().Skip("tested buffer does not have a maximum capacity")
	}

	buf := s.newTestBuffer(5)
	defer buf.Close()

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
	buf.Add(mm, mm, mm)
	buf.Add(mm, mm, mm, mm)
	s.Equal(2, reject)
	batch := buf.Batch(5)
	buf.Add(mm, mm, mm, mm)
	s.Equal(2, reject)
	buf.Add(mm, mm, mm, mm)
	s.Equal(5, reject)
	buf.Add(mm, mm, mm, mm)
	s.Equal(9, reject)
	buf.Add(mm, mm, mm, mm)
	s.Equal(13, reject)
	buf.Accept(batch)
	s.Equal(13, reject)
	s.Equal(5, accept)
}

func (s *BufferSuiteTest) TestBufferRejectEmptyBatch() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	batch := buf.Batch(2)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Reject(batch)
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	batch = buf.Batch(2)
	for _, m := range batch {
		s.NotNil(m)
	}
}

func (s *BufferSuiteTest) TestBufferFlushedPartial() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(3, 0)))
	batch := buf.Batch(2)
	s.Len(batch, 2)

	buf.Accept(batch)
	s.Equal(1, buf.Len())
}

func (s *BufferSuiteTest) TestBufferFlushedFull() {
	buf := s.newTestBuffer(5)
	defer buf.Close()

	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(1, 0)))
	buf.Add(metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(2, 0)))
	batch := buf.Batch(2)
	s.Len(batch, 2)

	buf.Accept(batch)
	s.Equal(0, buf.Len())
}

type mockMetric struct {
	telegraf.Metric
	AcceptF func()
	RejectF func()
	DropF   func()
}

func (m *mockMetric) Accept() {
	if m.AcceptF != nil {
		m.AcceptF()
	}
}

func (m *mockMetric) Reject() {
	if m.RejectF != nil {
		m.RejectF()
	}
}

func (m *mockMetric) Drop() {
	if m.DropF != nil {
		m.DropF()
	}
}
