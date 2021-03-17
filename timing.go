package timing

import (
	"sync"
	"time"
)

var base = New()
var once sync.Once

func lazyInit() {
	once.Do(func() { base.Run() })
}

// Len the number timer of the base.
func Len() int {
	return base.Len()
}

// HasRunning base running status.
func HasRunning() bool {
	return base.HasRunning()
}

// AddJob add a job
func AddJob(job Job, timeout time.Duration) *Timer {
	tm := NewJob(job)
	Add(tm, timeout)
	return tm
}

// AddJobFunc add a job function
func AddJobFunc(f func(), timeout time.Duration) *Timer { return AddJob(JobFunc(f), timeout) }

// Add add timer to base. and startLocked immediately.
func Add(tm *Timer, timeout time.Duration) {
	lazyInit()
	base.Add(tm, timeout)
}

// Delete Delete timer from base.
func Delete(tm *Timer) {
	lazyInit()
	base.Delete(tm)
}

// Modify modify timer timeout,and restart immediately.
func Modify(tm *Timer, timeout time.Duration) {
	lazyInit()
	base.Modify(tm, timeout)
}
