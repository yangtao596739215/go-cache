package pending

import (
	"errors"
	"sync"
	"time"

	"github.com/yangtao596739215/go-cache/lru"
)

type ILogger interface {
	INFO(format string, a ...interface{})
}

var ErrEntityType = errors.New("pendingCache get entity type error")
var ErrTimeOut = errors.New("pendingCache get entity timeout")
var ErrPendingGetWrong = errors.New("pending get wrong result")

const (
	CacheSize     = 20000
	ItemsToPrune  = 500
	ExpireSecends = 10 * time.Second
	TimeoutSecond = 300 * time.Millisecond
)

type Entity struct {
	Res   interface{}
	ready chan struct{}
}

func NewEntity() *Entity {
	return &Entity{
		ready: make(chan struct{}),
	}
}

type PendingCache struct {
	mu    sync.Mutex
	cache lru.ICache //线程安全
}

//自行替换底层lru
func (p *PendingCache) SetCache(c lru.ICache) {
	p.mu.Lock()
	p.cache = c
	p.mu.Unlock()
}

//使用默认cache
func NewPendingCache(size int64, prune uint32) *PendingCache {
	return &PendingCache{
		cache: lru.NewLRUCache(size, prune),
	}
}

//返回isHit，方便业务层知道是否命中了缓存
func (pCache *PendingCache) get(L ILogger, key string, f PendingProducer) (e *Entity, isHit bool, err error) {
	// 1. Lock
	pCache.mu.Lock() //pending的锁和map的锁分开，不加这个锁会导致大量的NotFound进入
	// 2. FindItem
	v, isFound := pCache.cache.Get(key)
	if !isFound { //NotFound
		isHit = false
		e = NewEntity() //这里要存指针
		pCache.cache.Set(key, e, ExpireSecends)
		pCache.mu.Unlock()

		errChan := make(chan error)
		resChan := make(chan interface{})
		go f(resChan, errChan) //异步调用，通过chan传值，实现超时删除，避免pending太多请求
		select {
		case res := <-resChan: //正常收到数据
			e.Res = res
			close(e.ready)
		case err := <-errChan: //收到err   （关闭chan，释放pending） （del key，让下一个进来的请求重新获取）
			L.INFO("pendingCache get entity res failed")
			close(e.ready)
			pCache.cache.Del(key)
			return nil, isHit, err
		case <-time.After(time.Duration(TimeoutSecond)): //超时（关闭chan，释放pending） （del key，让下一个进来的请求重新获取）【超时时间要根据异步rpc的耗时来算，一般设置成两次+buffer】
			L.INFO("pendingCache get entity timeout")
			close(e.ready)
			pCache.cache.Del(key)
			return nil, isHit, ErrTimeOut
		}
	} else { //Found
		isHit = true
		pCache.mu.Unlock()
		ok := false
		e, ok = v.(*Entity)
		if !ok {
			L.INFO("pendingCache get entity type error")
			return nil, isHit, ErrEntityType
		}
		<-e.ready
	}
	return e, isHit, nil
}
