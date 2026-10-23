package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
)

func TestWriteMetricsFromMeasureRejectsEmptyName(t *testing.T) {
	p := newPostgresql()
	m := metric.New("cpu", map[string]string{}, map[string]any{"value": int64(1)}, time.Unix(0, 0))
	m.SetName("")
	src := NewTableSource(p, "")
	src.AddMetric(m)
	require.Empty(t, src.Name())

	err := p.writeMetricsFromMeasure(context.Background(), nil, src)
	require.ErrorContains(t, err, "empty measurement name")
}
