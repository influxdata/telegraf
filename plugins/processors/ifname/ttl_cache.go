package ifname

import (
	"runtime"
	"time"
)

type TTLValType struct {
	time time.Time // when entry was added
	val  valType
}

type timeFunc func() time.Time

type TTLCache struct {
	validDuration time.Duration
	lru           LRUCache
	now           timeFunc
}

func NewTTLCache(valid time.Duration, capacity uint) TTLCache {
	return TTLCache{
		lru:           NewLRUCache(capacity),
		validDuration: valid,
		now:           time.Now,
	}
}

func (c *TTLCache) Get(key keyType) (valType, bool, time.Duration) {
	v, ok := c.lru.Get(key)
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

	c.lru.Delete(key)
	return valType{}, false, 0
}

func (c *TTLCache) Put(key keyType, value valType) {
	v := TTLValType{
		val:  value,
		time: c.now(),
	}
	c.lru.Put(key, v)
}

func (c *TTLCache) Delete(key keyType) {
	c.lru.Delete(key)
}
