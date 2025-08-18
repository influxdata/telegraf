package ifname

import (
	"runtime"
	"time"
)

type ttlValType struct {
	time time.Time // when entry was added
	val  valType
}

type timeFunc func() time.Time

type ttlCache struct {
	validDuration time.Duration
	lru           lruCache
	now           timeFunc
}

func newTTLCache(valid time.Duration, capacity uint) ttlCache {
	return ttlCache{
		lru:           newLRUCache(capacity),
		validDuration: valid,
		now:           time.Now,
	}
}

func (c *ttlCache) get(key keyType) (valType, bool, time.Duration) {
	v, ok := c.lru.get(key)
	if !ok {
		return valType{}, false, 0
	}

	if runtime.GOOS == "windows" {
		// Sometimes on Windows `c.now().Sub(v.time) == 0` due to clock resolution issues:
		// https://github.com/golang/go/issues/17696
		// https://github.com/golang/go/issues/29485
		// Force clock to refresh:
		time.Sleep(time.Nanosecond)
	}

	age := c.now().Sub(v.time)
	if age < c.validDuration {
		return v.val, ok, age
	}

	c.lru.delete(key)
	return valType{}, false, 0
}

func (c *ttlCache) put(key keyType, value valType) {
	v := ttlValType{
		val:  value,
		time: c.now(),
	}
	c.lru.put(key, v)
}

func (c *ttlCache) delete(key keyType) {
	c.lru.delete(key)
}
