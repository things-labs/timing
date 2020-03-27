package timing

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultJobChanSize default job chan size
	DefaultJobChanSize = 1024
	// DefaultTimeoutLimit submit job must immediately,time limit timeoutLimit,
	DefaultTimeoutLimit = 50 * time.Millisecond
)

// Timing keeps track of any number of entries.
type Timing struct {
	entries      []*Entry
	stop         chan struct{}
	add          chan *Entry
	snapshot     chan chan []Entry
	panicHandle  func(err interface{})
	jobs         chan Job
	jobsChanSize int
	timeoutLimit time.Duration
	running      bool
	mu           sync.Mutex
	location     *time.Location
	pool         pool
	logger
}

// New new a time with option
func New(opts ...Option) *Timing {
	tim := &Timing{
		entries:      make([]*Entry, 0),
		stop:         make(chan struct{}),
		add:          make(chan *Entry),
		snapshot:     make(chan chan []Entry),
		location:     time.Local,
		panicHandle:  func(err interface{}) {},
		jobsChanSize: DefaultJobChanSize,
		timeoutLimit: DefaultTimeoutLimit,
		logger:       newLogger("timing: "),
		pool:         newPool(),
	}

	for _, opt := range opts {
		opt(tim)
	}
	tim.jobs = make(chan Job, tim.jobsChanSize)
	return tim
}

// Location gets the time zone location
func (sf *Timing) Location() *time.Location {
	return sf.location
}

// Close the time
func (sf *Timing) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.stop <- struct{}{}
		sf.running = false
	}
	return nil
}

// HasRunning 运行状态
func (sf *Timing) HasRunning() bool {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.running
}

// Entries returns a snapshot of the Timing entries.
func (sf *Timing) Entries() []Entry {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		replyChan := make(chan []Entry, 1)
		sf.snapshot <- replyChan
		return <-replyChan
	}
	return sf.entrySnapshot()
}

// AddJob add a job
func (sf *Timing) AddJob(job Job, timeout time.Duration, opts ...EntryOption) *Timing {
	entry := sf.pool.get()

	entry.job = job
	entry.timeout = timeout
	for _, opt := range opts {
		opt(entry)
	}

	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.add <- entry
	} else {
		sf.entries = append(sf.entries, entry)
	}
	return sf
}

// AddJobFunc add a job function
func (sf *Timing) AddJobFunc(f JobFunc, timeout time.Duration, opts ...EntryOption) *Timing {
	return sf.AddJob(f, timeout, opts...)
}

// Run the timing in its own goroutine, or no-op if already started.
func (sf *Timing) Run() *Timing {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		return sf
	}
	sf.running = true
	go sf.run()
	return sf
}

func (sf *Timing) wrapJob(job Job) {
	defer func() {
		if err := recover(); err != nil {
			sf.panicHandle(err)
		}
	}()

	job.Run()
}

func (sf *Timing) run() {
	sf.Debug("run start!")

	Now := func() time.Time { return time.Now().In(sf.location) }

	// Figure out the next activation times for each entry.
	now := Now()
	for _, entry := range sf.entries {
		entry.next = now.Add(entry.timeout)
		sf.Debug("next active: now - %s, next - %s", now, entry.next)
	}
	closed := make(chan struct{})
	go func() {
		sf.Debug("work start!")
		for {
			select {
			case f := <-sf.jobs:
				sf.wrapJob(f)
			case <-closed:
				sf.Debug("work stop!")
				return
			}
		}
	}()
	// if time
	timeout := time.NewTimer(sf.timeoutLimit)
	defer timeout.Stop()
	for {
		// Determine the next entry to run.
		sort.Sort(byTime(sf.entries))
		var timer *time.Timer
		if len(sf.entries) == 0 || sf.entries[0].next.IsZero() {
			for _, v := range sf.entries {
				sf.pool.put(v)
			}
			sf.entries = make([]*Entry, 0)
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			// TODO:
			timer = time.NewTimer(sf.entries[0].next.Sub(now))
		}

	loop:
		for {
			select {
			case now = <-timer.C:
				now = now.In(sf.location)
				sf.Debug("wake up: now - %s", now)

				// Run every entry whose next time was less than now
				for _, e := range sf.entries {
					if e.next.After(now) || e.next.IsZero() {
						break
					}
					sf.Debug("run: now - %s, next - %s, entry - %p", now, e.next, e)

					if atomic.LoadUint32(&e.useGoroutine) == 1 {
						go e.job.Run()
					} else {
						timeout.Reset(sf.timeoutLimit)
						select {
						case sf.jobs <- e.job:
						case <-timeout.C:
							break loop
						}
					}
					e.next = time.Time{} // mark it, not work until remove it
				}
			case newEntry := <-sf.add:
				timer.Stop()
				now = Now()
				newEntry.next = now.Add(newEntry.timeout)
				sf.entries = append(sf.entries, newEntry)
				sf.Debug("added: now - %s, next - %s, entry - %p",
					now, newEntry.next, newEntry)

			case replyChan := <-sf.snapshot:
				replyChan <- sf.entrySnapshot()
				continue

			case <-sf.stop:
				closed <- struct{}{}
				timer.Stop()
				sf.Debug("run stop!")
				return
			}
			break
		}
	}
}

// entrySnapshot returns a copy of the current cron entry list.
func (sf *Timing) entrySnapshot() []Entry {
	var entries = make([]Entry, len(sf.entries))
	for i, e := range sf.entries {
		entries[i] = *e
	}
	return entries
}

// Job job interface
type Job interface {
	Run()
}

// JobFunc job function
type JobFunc func()

// Run implement job interface
func (f JobFunc) Run() { f() }

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// timeout time timeout
	timeout time.Duration
	// next time the job will run, or the zero time if Timing has not been
	// started or this entry is unsatisfiable
	next time.Time
	// job is the thing that want to run.
	job Job
	// use goroutine or not do the job
	useGoroutine uint32
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
