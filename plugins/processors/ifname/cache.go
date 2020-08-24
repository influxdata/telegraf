package ifname

// See https://girai.dev/blog/lru-cache-implementation-in-go/

import (
	"container/list"
)

type LRUValType = TTLValType

type hashType map[keyType]*list.Element

type LRUCache struct {
	cap uint       // capacity
	l   *list.List // doubly linked list
	m   hashType   // hash table for checking if list node exists
}

// Pair is the value of a list node.
type Pair struct {
	key   keyType
	value LRUValType
}

// initializes a new LRUCache.
func NewLRUCache(capacity uint) LRUCache {
	return LRUCache{
		cap: capacity,
		l:   new(list.List),
		m:   make(hashType, capacity),
	}
}

// Get a list node from the hash map.
func (c *LRUCache) Get(key keyType) (LRUValType, bool) {
	// check if list node exists
	if node, ok := c.m[key]; ok {
		val := node.Value.(*list.Element).Value.(Pair).value
		// move node to front
		c.l.MoveToFront(node)
		return val, true
	}
	return LRUValType{}, false
}

// Put key and value in the LRUCache
func (c *LRUCache) Put(key keyType, value LRUValType) {
	// check if list node exists
	if node, ok := c.m[key]; ok {
		// move the node to front
		c.l.MoveToFront(node)
		// update the value of a list node
		node.Value.(*list.Element).Value = Pair{key: key, value: value}
	} else {
		// delete the last list node if the list is full
		if uint(c.l.Len()) == c.cap {
			// get the key that we want to delete
			idx := c.l.Back().Value.(*list.Element).Value.(Pair).key
			// delete the node pointer in the hash map by key
			delete(c.m, idx)
			// remove the last list node
			c.l.Remove(c.l.Back())
		}
		// initialize a list node
		node := &list.Element{
			Value: Pair{
				key:   key,
				value: value,
			},
		}
		// push the new list node into the list
		ptr := c.l.PushFront(node)
		// save the node pointer in the hash map
		c.m[key] = ptr
	}
}

func (c *LRUCache) Delete(key keyType) {
	if node, ok := c.m[key]; ok {
		c.l.Remove(node)
		delete(c.m, key)
	}
}
