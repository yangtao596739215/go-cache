package pending

import (
	"time"
)

//pendingCache内部使用的preducer，不对外暴露
type PendingProducer func(chan interface{}, chan error) //通过通道，把异步执行的结果返回去

//业务传入的producer,定义通用的返回
type BusinessProducer func() (interface{}, error)

//把业务的producer转成内部使用的producer
func commonWrapper(producer BusinessProducer) PendingProducer {
	fn := func(res chan interface{}, errChan chan error) {
		r, err := producer()
		if err != nil {
			errChan <- err
		} else {
			res <- r
		}
	}
	return fn
}

//把业务的producer转成内部使用的producer
func retryWrapper(L ILogger, producer BusinessProducer, reTryTimes int) PendingProducer {
	fn := func(res chan interface{}, errChan chan error) {
		var err error
		for i := 0; i <= reTryTimes; i++ {
			L.INFO("[retryTimes:%d]", i)
			v, err := producer()
			if err != nil {
				continue
			}
			res <- v //获取成功，往通道里面放
			return
		}
		errChan <- err //获取失败，发送err
	}
	return fn
}

func (pCache *PendingCache) Get(L ILogger, key string, producer BusinessProducer, timeout, expired time.Duration) (interface{}, error) {
	entity, isHit, err := pCache.get(L, key, commonWrapper(producer), timeout, expired)
	return handleResult(L, key, entity, isHit, err)
}

//retryTimes 指的是producer执行的retry次数，不是获取的次数，如果该key已经被pending,则不会执行producer，retry也就不生效
func (pCache *PendingCache) GetWithRetry(L ILogger, key string, producer BusinessProducer, retryTimes int, timeout, expired time.Duration) (interface{}, error) {
	entity, isHit, err := pCache.get(L, key, retryWrapper(L, producer, retryTimes), timeout, expired)
	return handleResult(L, key, entity, isHit, err)
}

//返回的错误有四种 producer返回的，超时的，entity类型转换错误的，pending错误的
func handleResult(L ILogger, key string, entity *Entity, isHit bool, err error) (interface{}, error) {
	if err != nil { //这里的err有三种，producer返回的，超时的，entity类型转换错误的
		L.INFO("get pending_cache_err\n")
		return nil, err
	}

	//打印日志，方便统计是否命中
	if isHit {
		L.INFO("get pending_cache_hit key:%s\n", key)
	} else {
		L.INFO("get pending_cache_not_hit key:%s\n", key)
	}

	//同一个key，第一个获取失败以后，其他pending的请求会走到这里
	if entity.Res == nil {
		return nil, ErrPendingGetWrong
	}
	return entity.Res, nil
}
