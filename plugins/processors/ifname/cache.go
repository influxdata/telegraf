package ifname

// See https://girai.dev/blog/lru-cache-implementation-in-go/

import (
	"container/list"
)

type lruValType = ttlValType

type hashType map[keyType]*list.Element

type lruCache struct {
	cap uint       // capacity
	l   *list.List // doubly linked list
	m   hashType   // hash table for checking if list node exists
}

type pair struct {
	key   keyType
	value lruValType
}

func newLRUCache(capacity uint) lruCache {
	return lruCache{
		cap: capacity,
		l:   new(list.List),
		m:   make(hashType, capacity),
	}
}

func (c *lruCache) get(key keyType) (lruValType, bool) {
	// check if list node exists
	if node, ok := c.m[key]; ok {
		val := node.Value.(*list.Element).Value.(pair).value
		// move node to front
		c.l.MoveToFront(node)
		return val, true
	}
	return lruValType{}, false
}

func (c *lruCache) put(key keyType, value lruValType) {
	// check if list node exists
	if node, ok := c.m[key]; ok {
		// move the node to front
		c.l.MoveToFront(node)
		// update the value of a list node
		node.Value.(*list.Element).Value = pair{key: key, value: value}
	} else {
		// delete the last list node if the list is full
		if uint(c.l.Len()) == c.cap {
			// get the key that we want to delete
			idx := c.l.Back().Value.(*list.Element).Value.(pair).key
			// delete the node pointer in the hash map by key
			delete(c.m, idx)
			// remove the last list node
			c.l.Remove(c.l.Back())
		}
		// initialize a list node
		node := &list.Element{
			Value: pair{
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

func (c *lruCache) delete(key keyType) {
	if node, ok := c.m[key]; ok {
		c.l.Remove(node)
		delete(c.m, key)
	}
}
