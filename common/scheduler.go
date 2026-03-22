package common

import (
	"sync"
	"time"
)

type Scheduler struct {
	time     time.Duration
	enable   bool
	running  bool
	callback func()
	desc     string
	*sync.Mutex
}

func NewTimer(time time.Duration, desc string, callback func()) Scheduler {
	return Scheduler{
		time:     time,
		callback: callback,
		Mutex:    &sync.Mutex{},
		desc:     desc,
	}
}

func (t *Scheduler) Stop() {
	t.enable = false
}

func (t *Scheduler) Start() {
	if t.enable {
		return
	}
	t.enable = true
	if !t.running {
		t.running = true
		println("Timer Started: " + t.desc)
		time.AfterFunc(t.time, func() {
			t.loop()

		})
	}
}

func (t *Scheduler) loop() {
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
