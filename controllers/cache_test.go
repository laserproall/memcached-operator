package controllers

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	m := New(httpGetBody)
	go m.cacheClean()
	var n sync.WaitGroup
	for i := 0; i < 300; i++ {
		// time.Sleep(10 * time.Millisecond)
		for _, url := range incomingURLs() {
			n.Add(1)
			// time.Sleep(10 * time.Millisecond)
			go func(url string) {
				start := time.Now()
				value, err := m.Get(url)
				if err != nil {
					log.Print(err)
				}
				fmt.Printf("%s, %s, %d bytes\n",
					url, time.Since(start), len(value.([]byte)))
				n.Done()
			}(url)
		}
		n.Wait()
	}

}
