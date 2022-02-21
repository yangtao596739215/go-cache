package lru

import (
	"time"

	"github.com/karlseguin/ccache"
)

type ICache interface {
	Set(key string, value interface{}, duration time.Duration)
	Get(key string) (interface{}, bool) //第二个数代表key是否存在
	Del(key string) bool
	Size() int
}

type LRUCache struct {
	Cache *ccache.Cache //底层自己有锁
}

func (l *LRUCache) Set(key string, value interface{}, duration time.Duration) {
	l.Cache.Set(key, value, duration)
}

func (l *LRUCache) Get(key string) (interface{}, bool) {
	r := l.Cache.Get(key)
	if r != nil && !r.Expired() {
		return r.Value(), true
	}
	return nil, false
}

// Remove the item from the cache, return true if the item was present, false otherwise.
func (l *LRUCache) Del(key string) bool {
	return l.Cache.Delete(key)
}

func (l *LRUCache) Size() int {
	return l.Cache.ItemCount()
}

func NewLRUCache(size int64, prune uint32) *LRUCache {
	cache := &LRUCache{Cache: ccache.New(ccache.Configure().
		MaxSize(size).
		ItemsToPrune(prune))}
	return cache
}
