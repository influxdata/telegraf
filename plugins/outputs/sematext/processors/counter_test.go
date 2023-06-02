package processors

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClearCounterCache(t *testing.T) {
	now := time.Now()

	lastCleared := time.Now().Add(-1 * time.Hour)
	p := &HandleCounter{
		lastCleared:   lastCleared,
		countersCache: make(map[string]*counterCacheEntry),
	}

	p.countersCache["new"] = &counterCacheEntry{
		lastValue:    123,
		lastRecorded: time.Now().Add(-1 * time.Hour),
	}
	p.countersCache["old"] = &counterCacheEntry{
		lastValue:    234,
		lastRecorded: time.Now().Add(-30 * time.Hour),
	}

	p.clearCounterCache()

	// nothing should have changed
	assert.Equal(t, lastCleared.Unix(), p.lastCleared.Unix())
	assert.Equal(t, 2, len(p.countersCache))
	_, exists := p.countersCache["new"]
	assert.Equal(t, true, exists)
	_, exists = p.countersCache["old"]
	assert.Equal(t, true, exists)

	p.lastCleared = time.Now().Add(-25 * time.Hour)
	p.clearCounterCache()

	assert.True(t, p.lastCleared.Unix() >= now.Unix())
	assert.Equal(t, 1, len(p.countersCache))
	_, exists = p.countersCache["new"]
	assert.Equal(t, true, exists)
	_, exists = p.countersCache["old"]
	assert.Equal(t, false, exists)
}

func TestHoursSince(t *testing.T) {
	time1 := time.Date(2021, 04, 06, 10, 0, 0, 0, time.UTC)
	time2 := time.Date(2021, 04, 06, 11, 0, 0, 0, time.UTC)

	assert.Equal(t, 1, int(hoursSince(time2, time1)))
}

func TestGetZeroValue(t *testing.T) {
	assert.Equal(t, "", getZeroValue("string"))
	assert.Equal(t, false, getZeroValue(true))
	assert.Equal(t, false, getZeroValue(false))
	assert.Equal(t, 0.0, getZeroValue(float64(0)))
	assert.Equal(t, 0, getZeroValue(uint64(0)))
	assert.Equal(t, 0, getZeroValue(int64(0)))
}

func TestCalculateDelta(t *testing.T) {
	assert.Equal(t, int64(10), calculateDelta(int64(10), int64(20)))
	assert.Equal(t, float64(10), calculateDelta(float64(10), float64(20)))
	assert.Equal(t, "def", calculateDelta("abc", "def"))
}
