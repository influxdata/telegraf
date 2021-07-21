package channel

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestUnbufferedConnection(t *testing.T) {
	src := make(chan telegraf.Metric)
	dst := make(chan telegraf.Metric)
	ch := NewRelay(src, dst)
	ch.Start()
	m := testutil.MustMetric("test", nil, nil, time.Now())
	go func() {
		src <- m
		close(src)
	}()
	m2 := <-dst
	testutil.RequireMetricEqual(t, m, m2)
}

func TestBufferedConnection(t *testing.T) {
	src := make(chan telegraf.Metric, 1)
	dst := make(chan telegraf.Metric, 1)
	ch := NewRelay(src, dst)
	ch.Start()
	m := testutil.MustMetric("test", nil, nil, time.Now())
	src <- m
	close(src)
	m2 := <-dst
	testutil.RequireMetricEqual(t, m, m2)
}

func TestCloseIsRelayed(t *testing.T) {
	src := make(chan telegraf.Metric, 1)
	dst := make(chan telegraf.Metric, 1)
	ch := NewRelay(src, dst)
	ch.Start()
	close(src)
	_, ok := <-dst
	require.False(t, ok)
}

func TestCanReassignDest(t *testing.T) {
	src := make(chan telegraf.Metric, 1)
	dst1 := make(chan telegraf.Metric, 1)
	dst2 := make(chan telegraf.Metric, 1)
	ch := NewRelay(src, dst1)
	ch.Start()
	m1 := testutil.MustMetric("test1", nil, nil, time.Now())
	m2 := testutil.MustMetric("test2", nil, nil, time.Now())
	src <- m1
	m3 := <-dst1
	ch.SetDest(dst2)
	src <- m2
	m4 := <-dst2
	testutil.RequireMetricEqual(t, m1, m3)
	testutil.RequireMetricEqual(t, m2, m4)
}
