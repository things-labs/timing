package timing

import "time"

// Timer consists of a schedule and the func to execute on that schedule.
type Timer struct {
	// The index of the object's key in the Heap.queue.
	index int
	// next time the job will run, or the zero time if Base has not been
	// started or this entry is unsatisfiable
	next time.Time
	// job is the thing that want to run.
	job Job
	// use goroutine
	useGoroutine bool
}

// NewTimer new timer
func NewTimer() *Timer {
	return &Timer{job: emptyJob{}}
}

// NewJob new timer with job.
func NewJob(job Job) *Timer {
	return NewTimer().WithJob(job)
}

// NewJobFunc new timer with job function.
func NewJobFunc(f func()) *Timer {
	return NewTimer().WithJobFunc(f)
}

// WithGoroutine with goroutine
func (sf *Timer) WithGoroutine() *Timer {
	sf.useGoroutine = true
	return sf
}

// WithJob with job.
func (sf *Timer) WithJob(job Job) *Timer {
	sf.job = job
	return sf
}

// WithJobFunc with job function
func (sf *Timer) WithJobFunc(f func()) *Timer {
	return sf.WithJob(JobFunc(f))
}
