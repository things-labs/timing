package timing

import (
	"sync"
	"time"
)

var defaultTimer = New()
var once sync.Once

func lazyInit() {
	once.Do(func() {
		defaultTimer.Run()
	})

}

// Location gets the time zone location
func Location() *time.Location {
	return defaultTimer.Location()
}

// HasRunning 运行状态
func HasRunning() bool {
	return defaultTimer.HasRunning()
}

// Count returns a snapshot of the Timing entries.
func Count() int {
	return defaultTimer.Count()
}

// AddJob add a job
func AddJob(job Job, timeout time.Duration) {
	lazyInit()
	defaultTimer.AddJob(job, timeout)
}

// AddJobFunc add a job function
func AddJobFunc(f JobFunc, timeout time.Duration) {
	AddJob(f, timeout)
}
