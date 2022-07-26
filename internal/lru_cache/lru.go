package lrucache

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	sync.RWMutex
	nodeList *list.List
	cache    map[string]*list.Element
	maxSize  int
	usedSize int
}

type Item struct {
	Key        string
	Value      interface{}
	ExpireTime time.Time
	exFlag     bool
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		nodeList: list.New(),
		cache:    make(map[string]*list.Element),
		maxSize:  cap,
	}
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.set(key, value, -1)
}

func (c *LRUCache) SetWithTTL(key string, value interface{}, expire int) {
	c.set(key, value, expire)
}

func (c *LRUCache) set(key string, value interface{}, expire int) {
	c.Lock()
	defer c.Unlock()
	flag := false
	if expire != -1 {
		flag = true
	}
	if elm, ok := c.cache[key]; ok {
		elm.Value = &Item{
			Key:        key,
			Value:      value,
			ExpireTime: time.Now().Add(time.Duration(expire) * time.Second),
			exFlag:     flag,
		}
		c.nodeList.MoveToFront(elm)
	} else {
		p := c.nodeList.PushFront(&Item{Key: key, Value: value, ExpireTime: time.Now().Add(time.Duration(expire) * time.Second), exFlag: flag})
		c.cache[key] = p
		c.usedSize++
	}

	for c.usedSize > c.maxSize {
		elm := c.nodeList.Back()
		key := elm.Value.(*Item).Key
		delete(c.cache, key)
		c.nodeList.Remove(elm)
		c.usedSize--
	}
}

func (c *LRUCache) Get(key string) interface{} {

	if elm, ok := c.cache[key]; ok {

		v := elm.Value.(*Item)

		if v.exFlag && v.ExpireTime.Before(time.Now()) {
			c.Del(key)
			return nil
		}
		c.RLock()
		c.nodeList.MoveToFront(elm)
		c.RUnlock()
		return elm.Value.(*Item).Value
	}
	return nil
}

func (c *LRUCache) Del(key string) {
	c.Lock()
	defer c.Unlock()
	if elm, ok := c.cache[key]; ok {
		c.nodeList.Remove(elm)
		delete(c.cache, key)
		c.usedSize--
	}
}
