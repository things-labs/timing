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

// HasRunning 运行状态
func HasRunning() bool {
	return defaultTimer.HasRunning()
}

// AddJob add a job
func AddJob(job Job, timeout time.Duration) error {
	lazyInit()
	return defaultTimer.AddJob(job, timeout)
}

// AddJobFunc add a job function
func AddJobFunc(f JobFunc, timeout time.Duration) error {
	return AddJob(f, timeout)
}
