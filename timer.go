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

// Close the time
func Close() error {
	return defaultTimer.Close()
}

// HasRunning 运行状态
func HasRunning() bool {
	return defaultTimer.HasRunning()
}

// Entries returns a snapshot of the Timing entries.
func Entries() []Entry {
	return defaultTimer.Entries()
}

// AddJob add a job
func AddJob(job Job, num uint32, interval time.Duration) *Entry {
	lazyInit()
	return defaultTimer.AddJob(job, num, interval)
}

// AddJobFunc add a job function
func AddJobFunc(f JobFunc, num uint32, interval time.Duration) *Entry {
	return AddJob(f, num, interval)
}

// AddOneShotJob  add one-shot job
func AddOneShotJob(job Job, interval time.Duration) *Entry {
	return AddJob(job, OneShot, interval)
}

// AddOneShotJobFunc add one-shot job function
func AddOneShotJobFunc(f JobFunc, interval time.Duration) *Entry {
	return AddJob(f, OneShot, interval)
}

// AddPersistJob add persist job
func AddPersistJob(job Job, interval time.Duration) *Entry {
	return AddJob(job, Persist, interval)
}

// AddPersistJobFunc add persist job function
func AddPersistJobFunc(f JobFunc, interval time.Duration) *Entry {
	return AddJob(f, Persist, interval)
}

// Start start the entry
func Start(e *Entry, newInterval ...time.Duration) {
	lazyInit()
	defaultTimer.Start(e, newInterval...)
}

// Remove entry form timing
func Remove(e *Entry) {
	defaultTimer.Remove(e)
}
