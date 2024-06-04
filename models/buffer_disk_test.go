package models

import (
	"os"
	"testing"

	models "github.com/influxdata/telegraf/models/mock"
	"github.com/stretchr/testify/require"
)

func newTestDiskBuffer(t testing.TB) Buffer {
	t.Helper()
	path, err := os.MkdirTemp("", "*-buffer-test")
	require.NoError(t, err)
	buf, err := NewBuffer("test", "", 0, "disk", path)
	require.NoError(t, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	return buf
}

func TestBuffer_AddCallsMetricAccept(t *testing.T) {
	var accept int
	mm := &models.MockMetric{
		Metric: Metric(),
		AcceptF: func() {
			accept++
		},
	}
	b := newTestDiskBuffer(t)
	b.Add(mm, mm, mm)
	batch := b.Batch(2)
	b.Accept(batch)
	// all 3 metrics should be accepted as metric Accept() is called on buffer Add()
	require.Equal(t, 3, accept)
}
