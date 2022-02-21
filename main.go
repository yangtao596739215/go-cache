package main

import (
	"fmt"
	"time"

	"github.com/yangtao596739215/go-cache/pending"
)

func main() {

	p := pending.NewPendingCache(100, 20)
	fc := func() (interface{}, error) {
		return getdata()
	}

	//获取的时候传入生产者和超时时间，如果key在cache中不存在，会自动通过生产者获取并设置进去
	result, err := p.Get(nil, "key", fc, 300*time.Microsecond, 10*time.Second)
	if err != nil {
		fmt.Println(result)
	}

	fn := func() (interface{}, error) {
		return getdata()
	}

	p.GetWithRetry(nil, "key_retry", fn, 2, 300*time.Microsecond, 10*time.Second)

}

func getdata() (interface{}, error) {
	return "xxx", nil
}
