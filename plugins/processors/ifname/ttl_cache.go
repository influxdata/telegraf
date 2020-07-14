package ifname

import (
	"time"
)

type TTLValType struct {
	expireTime time.Time
	val        valType
}

type timeFunc func() time.Time

type TTLCache struct {
	validDuration time.Duration
	lru           LRUCache
	now           timeFunc
}

func NewTTLCache(expire time.Duration, capacity uint) TTLCache {
	return TTLCache{
		lru:           NewLRUCache(capacity),
		validDuration: expire,
		now:           time.Now,
	}
}

func (c *TTLCache) Get(key keyType) (valType, bool) {
	v, ok := c.lru.Get(key)
	if !ok {
		return valType{}, false
	}
	if c.now().Before(v.expireTime) {
		return v.val, ok
	} else {
		c.lru.Delete(key)
		return valType{}, false
	}
}

func (c *TTLCache) Put(key keyType, value valType) {
	v := TTLValType{
		val:        value,
		expireTime: c.now().Add(c.validDuration),
	}
	c.lru.Put(key, v)
}
