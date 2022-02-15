package stackdriver

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/models"

	monpb "google.golang.org/genproto/googleapis/monitoring/v3"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateCounterCacheEntry(t *testing.T) {
	cc := NewCounterCache(models.NewLogger("outputs", "stackdriver", "TestCreateCounterCacheEntry"))
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	endTime := tspb.Now()
	startTime := cc.GetStartTime("key", value, endTime)
	if endTime.AsTime().Add(time.Millisecond*-1) != startTime.AsTime() {
		t.Fatal("Start time on a new entry should be 1ms behind the end time")
	}
}

func TestUpdateCounterCacheEntry(t *testing.T) {
	cc := NewCounterCache(models.NewLogger("outputs", "stackdriver", "TestUpdateCounterCacheEntry"))
	now := time.Now().UTC()
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	endTime := tspb.New(now)
	startTime := cc.GetStartTime("key", value, endTime)
	if endTime.AsTime().Add(time.Millisecond*-1) != startTime.AsTime() {
		t.Fatal("Start time on a new entry should be 1ms behind the end time")
	}

	// next observation, 1m later
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(2),
		},
	}
	endTime = tspb.New(now.Add(time.Second * 60))
	startTime = cc.GetStartTime("key", value, endTime)
	// startTime is unchanged
	if startTime.GetSeconds() != now.Unix() {
		t.Fatal("Returned start time on an updated counter on the same day should not change")
	}
	obs, ok := cc.get("key")
	if !ok {
		t.Fatal("GetStartTime should create a fetchable k/v")
	}
	if obs.StartTime != startTime {
		t.Fatal("Start time on fetched observation should match output from GetStartTime()")
	}
	if obs.LastValue != value {
		t.Fatal("Stored value on fetched observation should have been updated.")
	}
}

func TestCounterCounterCacheEntryReset(t *testing.T) {
	cc := NewCounterCache(models.NewLogger("outputs", "stackdriver", "TestCounterCounterCacheEntryReset"))
	now := time.Now().UTC()
	backdatedNow := now.Add(time.Millisecond * -1)
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(2),
		},
	}
	endTime := tspb.New(now)
	startTime := cc.GetStartTime("key", value, endTime)
	if startTime.AsTime() != backdatedNow {
		t.Fatal("Start time on a new entry should be 1ms behind the end time")
	}

	// next observation, 1m later, but a lower value
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	later := now.Add(time.Second * 60)
	endTime = tspb.New(later)
	startTime = cc.GetStartTime("key", value, endTime)
	// startTime should now be the new endTime -1ms
	if startTime.AsTime() != later.Add(time.Millisecond*-1) {
		t.Fatal("Returned start time after a counter reset should equal the end time minus 1ms")
	}
	obs, ok := cc.get("key")
	if !ok {
		t.Fatal("GetStartTime should create a fetchable k/v")
	}
	if obs.StartTime.AsTime() != endTime.AsTime().Add(time.Millisecond*-1) {
		t.Fatal("Start time on fetched observation after a counter reset should equal the end time minus 1ms")
	}
	if obs.LastValue != value {
		t.Fatal("Stored value on fetched observation should have been updated.")
	}
}

func TestCounterCacheDayRollover(t *testing.T) {
	cc := NewCounterCache(models.NewLogger("outputs", "stackdriver", "TestCounterCacheDayRollover"))
	now := time.Now().UTC()
	backdatedNow := now.Add(time.Millisecond * -1)
	value := &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(1),
		},
	}
	endTime := tspb.New(now)
	startTime := cc.GetStartTime("key", value, endTime)
	if startTime.AsTime() != backdatedNow {
		t.Fatal("Start time on a new entry should be 1ms behind the end time")
	}

	// next observation, 24h later
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(2),
		},
	}
	later := now.Add(time.Hour * 24)
	endTime = tspb.New(later)
	startTime = cc.GetStartTime("key", value, endTime)
	if startTime.AsTime() != backdatedNow {
		t.Fatalf("Returned start time %d 1s before a day rollover should equal the end time %d", startTime.GetSeconds(), now.Unix())
	}
	obs, ok := cc.get("key")
	if !ok {
		t.Fatal("GetStartTime should create a fetchable k/v")
	}
	if obs.StartTime.AsTime() != backdatedNow {
		t.Fatal("Start time on an updated counter 1s before a day rollover should be unchanged")
	}
	if obs.LastValue != value {
		t.Fatal("Stored value on an updated counter should have been updated.")
	}

	// next observation, 24h 1s later
	value = &monpb.TypedValue{
		Value: &monpb.TypedValue_Int64Value{
			Int64Value: int64(3),
		},
	}
	tomorrow := later.Add(time.Second * 1)
	endTime = tspb.New(tomorrow)
	startTime = cc.GetStartTime("key", value, endTime)
	// startTime should now be the new endTime
	if startTime.GetSeconds() != tomorrow.Unix() {
		t.Fatalf("Returned start time %d after a day rollover should equal the end time %d", startTime.GetSeconds(), tomorrow.Unix())
	}
	obs, ok = cc.get("key")
	if !ok {
		t.Fatal("GetStartTime should create a fetchable k/v")
	}
	if obs.StartTime.AsTime() != endTime.AsTime().Add(time.Millisecond*-1) {
		t.Fatal("Start time on fetched observation after a day rollover should equal the new end time -1ms")
	}
	if obs.LastValue != value {
		t.Fatal("Stored value on fetched observation should have been updated.")
	}
}
