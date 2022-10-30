package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// Func is the type of the function to memoize.
type Func func(key string) (interface{}, error)
type result struct {
	value interface{}
	err   error
}

type entry struct {
	res result
	// ready chan struct{} // closed when res is ready
	exp bool
}

func New(f Func) *Memo {
	return &Memo{f: f, cache: make(map[string]*entry)}
}

type Memo struct {
	f     Func
	mu    sync.Mutex // guards cache
	cache map[string]*entry
}

func (memo *Memo) Get(key string) (value interface{}, err error) {
	memo.mu.Lock()
	defer memo.mu.Unlock()
	e := memo.cache[key]
	if e == nil || e.exp {
		// This is the first request for this key.
		// This goroutine becomes responsible for computing
		// the value and broadcasting the ready condition.
		// e = &entry{ready: make(chan struct{})}
		e = &entry{}
		fmt.Printf("-------rebuild--------%s %t\n", key, e.exp)
		memo.cache[key] = e
		// memo.mu.Unlock()
		e.res.value, e.res.err = memo.f(key)
		// close(e.ready) // broadcast ready condition

	}
	// else {
	// 	// This is a repeat request for this key.
	// 	// memo.mu.Unlock()
	// 	select {
	// 	case <-e.ready: // wait for ready condition
	// 		fmt.Printf("-------hit cache--------")
	// 		// default:
	// 		// 	fmt.Printf("-------MISS--------")
	// 		// 	return []byte("leak"), nil
	// 	}

	// }
	return e.res.value, e.res.err
}

func httpGetBody(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func incomingURLs() []string {
	return []string{
		"https://www.baidu.com/",
		"https://www.cnblogs.com/",
		"https://www.baidu.com/",
		"https://www.taobao.com/",
		"https://www.baidu.com/",
		"https://www.taobao.com/",
	}
}

func (memo *Memo) cacheClean() {
	for {
		time.Sleep(300 * time.Millisecond)
		memo.mu.Lock()
		for _, v := range memo.cache {
			v.exp = true
		}
		memo.mu.Unlock()
	}
}

// // A Memo caches the results of calling a Func.
// type Memo struct {
// 	f     Func
// 	cache map[string]result
// }

// // Func is the type of the function to memoize.
// type Func func(key string) (interface{}, error)
// type result struct {
// 	value interface{}
// 	err   error
// }

// func New(f Func) *Memo {
// 	return &Memo{f: f, cache: make(map[string]result)}
// }

// // NOTE: not concurrency-safe!
// func (memo *Memo) Get(key string) (interface{}, error) {
// 	res, ok := memo.cache[key]
// 	if !ok {
// 		res.value, res.err = memo.f(key)
// 		memo.cache[key] = res
// 	}
// 	return res.value, res.err
// }

// var m = map[string]string{}
// var c chan struct{}
// var n sync.WaitGroup

// func main() {
// 	delete(m, "hello")
// 	// c = make(chan struct{})
// 	// <-c
// 	n.Add(1)
// 	go func() {
// 		for {
// 			time.Sleep(100 * time.Millisecond)
// 			select {
// 			case <-c:
// 				fmt.Println("good")
// 				// default:
// 				// 	fmt.Println("No good")

// 			}

// 		}
// 	}()

// 	go func() {
// 		for {
// 			time.Sleep(1 * time.Second)
// 			c = make(chan struct{})
// 			time.Sleep(1 * time.Second)
// 			close(c)
// 		}
// 	}()

// 	n.Wait()
// }
