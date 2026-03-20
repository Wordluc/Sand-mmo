package common

import (
	"sync"
	"time"
)

type Timer struct {
	time     time.Duration
	enable   bool
	running  bool
	callback func()
	*sync.Mutex
}

func NewTimer(time time.Duration, callback func()) Timer {
	return Timer{
		time:     time,
		callback: callback,
		Mutex:    &sync.Mutex{},
	}
}

func (t *Timer) Stop() {
	t.enable = false
}

func (t *Timer) Start() {
	if t.enable {
		return
	}
	t.enable = true
	if !t.running {
		t.running = true
		t.loop()
	}
}

func (t *Timer) loop() {
	t.Lock()
	if !t.enable {
		t.running = false
		t.Unlock()
		return
	}

	go func() {
		t.callback()
		t.Unlock()
	}()

	time.AfterFunc(t.time, t.loop)
}
