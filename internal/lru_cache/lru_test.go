package lrucache

import (
	"testing"
	"time"
)

func TestLRUCache(t *testing.T) {
	lru := NewLRUCache(5)
	lru.Set("hello", "world")
	lru.Set("hello1", "world")
	lru.Set("hello2", "world")
	lru.Set("hello3", "world")
	lru.Set("hello4", "world")
	lru.Set("hello5", "world")

	t.Log(lru.Get("hello") == nil)

	lru.SetWithTTL("hello1", "world", 3)

	t.Log(lru.Get("hello1"))
	time.Sleep(4 * time.Second)
	t.Log(lru.Get("hello1") == nil)

}
