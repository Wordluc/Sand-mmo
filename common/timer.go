package common

import "time"

type Timer struct {
	time     time.Duration
	enable   bool
	callback func()
}

func NewTimer(time time.Duration, callback func()) Timer {
	return Timer{
		time:     time,
		callback: callback,
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
	t.loop()
}

func (t *Timer) loop() {
	if !t.enable {
		return
	}

	t.callback()

	time.AfterFunc(t.time, t.loop)
}
