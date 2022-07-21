package lrucache

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	sync.RWMutex
	nodeList *list.List
	cache map[string]*list.Element
	maxSize int
	usedSize int
}

type Item struct {
	Key string
	Value interface{}
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		nodeList: list.New(),
		cache: make(map[string]*list.Element),
		maxSize: cap,
	}
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if elm, ok := c.cache[key]; ok {
		elm.Value = &Item{
			Key: key,
			Value: value,
		}
		c.nodeList.MoveToFront(elm)
	} else {
		p := c.nodeList.PushFront(&Item{Key: key, Value: value})
		c.cache[key] = p
		c.usedSize ++
	}

	for c.usedSize > c.maxSize {
		elm := c.nodeList.Back()
		key := elm.Value.(*Item).Key
		delete(c.cache, key)
		c.nodeList.Remove(elm)
		c.usedSize --
	}
}

func (c *LRUCache) Get(key string) interface{} {
	c.RLock()
	defer c.RUnlock()
	if elm, ok := c.cache[key]; ok {
		c.nodeList.MoveToFront(elm)
		return elm.Value.(*Item).Value
	}
	return nil
}

func (c *LRUCache) Del(key string){
	c.Lock()
	defer c.Unlock()
	if elm, ok := c.cache[key]; ok {
		c.nodeList.Remove(elm)
		delete(c.cache, key)
		c.usedSize --
	}
}