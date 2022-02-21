package main

import (
	"fmt"

	"github.com/yangtao596739215/go-cache/pending"
)

func main() {

	p := pending.NewPendingCache(100, 20)
	fc := func() (interface{}, error) {
		return getdata()
	}

	result, err := p.Get(nil, "key", fc)
	if err != nil {
		fmt.Println(result)
	}

	fn := func() (interface{}, error) {
		return getdata()
	}

	p.Get(nil, "key_retry", fn)

}

func getdata() (interface{}, error) {
	return "xxx", nil
}
