package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/yangtao596739215/go-cache/pending"
)

var count int

type Log struct{}

func (l Log) INFO(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func mockOk() (interface{}, error) {
	return "xx", nil
}

func mockTimeout() (interface{}, error) {
	time.Sleep(1 * time.Second)
	return "xxx", nil
}

//不返回err，但是值是错误的，在转换的时候报错
func mockWrongValue() (interface{}, error) {
	return nil, nil
}

//第一次调用返回错误,之后返回正确
func mockSecond() (interface{}, error) {
	if count == 0 {
		count++
		return nil, errors.New("xx")
	}
	time.Sleep(1 * time.Second)
	return "xxx", nil
}

func main() {
	p := pending.NewPendingCache(100, 20)
	l := &Log{}

	//正常情况
	print(p.Get(l, "ok", mockOk, 200*time.Microsecond, 1*time.Second))

	//第一个timeout错误(打印not_found) 	第二个设置成功(打印not_found)  第三个及其后面的都成功
	go print(p.Get(l, "timeout", mockTimeout, 800*time.Millisecond, 10*time.Second))
	for i := 0; i <= 20; i++ {
		go print(p.Get(l, "timeout", mockTimeout, 2000*time.Millisecond, 10*time.Second))
	}
	time.Sleep(3 * time.Second)

	//重试一次后成功了(第一个notfound，其他的都find)都返回正确值
	for i := 0; i <= 20; i++ {
		go print(p.GetWithRetry(l, "retry", mockSecond, 2, 2000*time.Millisecond, 10*time.Second))
	}
	time.Sleep(3 * time.Second)

	//第一个结果转换错误(打印not_found) 	第二个及其后面的结果转换错误(打印found)
	go print(p.Get(l, "wrongvalue", mockWrongValue, 800*time.Millisecond, 10*time.Second))
	for i := 0; i <= 20; i++ {
		go print(p.Get(l, "wrongvalue", mockWrongValue, 2000*time.Millisecond, 10*time.Second))
	}
	time.Sleep(3 * time.Second)

}

func print(d interface{}, err error) {
	if err != nil {
		fmt.Printf("err:%+v\n", err)
	} else {
		fmt.Printf("v:%+v\n", d)
	}
}
