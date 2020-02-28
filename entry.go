package timing

import (
	"time"
)

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// // job has schedule count
	count uint32
	// job schedule number
	number uint32
	// interval time interval
	interval time.Duration
	// next time the job will run, or the zero time if Timing has not been
	// started or this entry is unsatisfiable
	next time.Time
	// prev is the last time this job was run, or the zero time if never.
	prev time.Time
	// job is the thing that want to run.
	job Job
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].next.IsZero() {
		return false
	}
	if s[j].next.IsZero() {
		return true
	}
	return s[i].next.Before(s[j].next)
}

// NewEntry new entry
func NewEntry(job Job, num uint32, interval time.Duration) *Entry {
	return &Entry{
		number:   num,
		interval: interval,
		job:      job,
	}
}

// NewOneShotEntry new one-shot entry
func NewOneShotEntry(job Job, interval time.Duration) *Entry {
	return NewEntry(job, OneShot, interval)
}

// NewPersistEntry new persist entry
func NewPersistEntry(job Job, interval time.Duration) *Entry {
	return NewEntry(job, Persist, interval)
}

// NewFuncEntry new function entry
func NewFuncEntry(f JobFunc, num uint32, interval time.Duration) *Entry {
	return NewEntry(f, num, interval)
}

// NewOneShotFuncEntry new one-shot function entry
func NewOneShotFuncEntry(f JobFunc, interval time.Duration) *Entry {
	return NewEntry(f, OneShot, interval)
}

// NewPersistFuncEntry new persist function entry
func NewPersistFuncEntry(f JobFunc, interval time.Duration) *Entry {
	return NewEntry(f, Persist, interval)
}
