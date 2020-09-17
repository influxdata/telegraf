package ifname

import (
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
	age := c.now().Sub(v.time)
	if age < c.validDuration {
		return v.val, ok, age
	} else {
		c.lru.Delete(key)
		return valType{}, false, 0
	}
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
