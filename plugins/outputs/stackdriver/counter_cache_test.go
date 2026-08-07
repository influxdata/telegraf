package stackdriver

import (
	"testing"
	"time"

	monpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/stretchr/testify/require"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/influxdata/telegraf/testutil"
)

func TestCreateCounterCacheEntry(t *testing.T) {
	cc := NewCounterCache(testutil.Logger{})
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	endTime := tspb.Now()
	startTime := cc.GetStartTime("key", value, endTime)
	require.Equal(t, endTime.AsTime().Add(-time.Millisecond), startTime.AsTime())
}

func TestUpdateCounterCacheEntry(t *testing.T) {
	cc := NewCounterCache(testutil.Logger{})
	now := time.Now().UTC()
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	endTime := tspb.New(now)
	startTime := cc.GetStartTime("key", value, endTime)
	require.Equal(t, endTime.AsTime().Add(-time.Millisecond), startTime.AsTime())

	// next observation, 1m later
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(2),
		},
	}
	endTime = tspb.New(now.Add(time.Second * 60))
	startTime = cc.GetStartTime("key", value, endTime)
	require.Equal(t, now.Unix(), startTime.GetSeconds())

	obs, ok := cc.get("key")
	require.True(t, ok, "GetStartTime should create a fetchable k/v")
	require.Equal(t, startTime, obs.StartTime)
	require.Equal(t, value, obs.LastValue)
}

func TestCounterCacheEntryReset(t *testing.T) {
	cc := NewCounterCache(testutil.Logger{})
	now := time.Now().UTC()
	backdatedNow := now.Add(-time.Millisecond)
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(2),
		},
	}
	endTime := tspb.New(now)
	startTime := cc.GetStartTime("key", value, endTime)
	require.Equal(t, backdatedNow, startTime.AsTime())

	// next observation, 1m later, but a lower value
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	later := now.Add(time.Second * 60)
	endTime = tspb.New(later)
	startTime = cc.GetStartTime("key", value, endTime)
	require.Equal(t, later.Add(-time.Millisecond), startTime.AsTime())

	obs, ok := cc.get("key")
	require.True(t, ok, "GetStartTime should create a fetchable k/v")
	require.Equal(t, endTime.AsTime().Add(-time.Millisecond), obs.StartTime.AsTime())
	require.Equal(t, value, obs.LastValue)
}

func TestCounterCacheDayRollover(t *testing.T) {
	cc := NewCounterCache(testutil.Logger{})
	now := time.Now().UTC()
	backdatedNow := now.Add(-time.Millisecond)
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	endTime := tspb.New(now)
	startTime := cc.GetStartTime("key", value, endTime)
	require.Equal(t, backdatedNow, startTime.AsTime())

	// next observation, 24h later (within the 86400s window)
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(2),
		},
	}
	later := now.Add(time.Hour * 24)
	endTime = tspb.New(later)
	startTime = cc.GetStartTime("key", value, endTime)
	require.Equal(t, backdatedNow, startTime.AsTime())

	obs, ok := cc.get("key")
	require.True(t, ok, "GetStartTime should create a fetchable k/v")
	require.Equal(t, backdatedNow, obs.StartTime.AsTime())
	require.Equal(t, value, obs.LastValue)

	// next observation, 24h 1s later (exceeds 86400s window, triggers reset)
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(3),
		},
	}
	tomorrow := later.Add(time.Second)
	endTime = tspb.New(tomorrow)
	startTime = cc.GetStartTime("key", value, endTime)
	require.Equal(t, tomorrow.Unix(), startTime.GetSeconds())

	obs, ok = cc.get("key")
	require.True(t, ok, "GetStartTime should create a fetchable k/v")
	require.Equal(t, endTime.AsTime().Add(-time.Millisecond), obs.StartTime.AsTime())
	require.Equal(t, value, obs.LastValue)
}
