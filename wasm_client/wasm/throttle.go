package wasm

import (
	"sync"
	"time"
)

func Throttle(t time.Duration, fn func()) func() {
	var mu sync.Mutex
	var running bool

	return func() {
		mu.Lock()
		defer mu.Unlock()
		if running {
			return
		}
		running = true
		go func() {
			time.Sleep(t * time.Millisecond)
			mu.Lock()
			running = false
			mu.Unlock()
		}()
		fn()
	}
}
