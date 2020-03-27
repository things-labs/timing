package timing

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// num define
const (
	OneShot = 1
	Persist = 0
)

const (
	// DefaultJobChanSize default job chan size
	DefaultJobChanSize = 1024
	// submit job must immediately,time limit timeoutLimit,
	DefaultTimeoutLimit = 50 * time.Millisecond
)

type mdEntry struct {
	entry    *Entry
	interval time.Duration
}

// Timing keeps track of any number of entries.
type Timing struct {
	entries      []*Entry
	stop         chan struct{}
	add          chan *Entry
	remove       chan *Entry
	active       chan mdEntry
	snapshot     chan chan []Entry
	pf           func(err interface{})
	jobs         chan Job
	jobsChanSize int
	timeoutLimit time.Duration
	running      bool
	mu           sync.Mutex
	location     *time.Location
	logger
}

// New new a time with option
func New(opts ...Option) *Timing {
	tim := &Timing{
		entries:      make([]*Entry, 0),
		add:          make(chan *Entry),
		remove:       make(chan *Entry),
		active:       make(chan mdEntry),
		stop:         make(chan struct{}),
		snapshot:     make(chan chan []Entry),
		location:     time.Local,
		pf:           func(err interface{}) {},
		jobsChanSize: DefaultJobChanSize,
		timeoutLimit: DefaultTimeoutLimit,
		logger:       newLogger("timing: "),
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
func (sf *Timing) AddJob(job Job, num uint32, interval time.Duration) *Entry {
	entry := NewEntry(job, num, interval)

	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.add <- entry
	} else {
		sf.entries = append(sf.entries, entry)
	}
	return entry
}

// AddJobFunc add a job function
func (sf *Timing) AddJobFunc(f JobFunc, num uint32, interval time.Duration) *Entry {
	return sf.AddJob(f, num, interval)
}

// AddOneShotJob  add one-shot job
func (sf *Timing) AddOneShotJob(job Job, interval time.Duration) *Entry {
	return sf.AddJob(job, OneShot, interval)
}

// AddOneShotJobFunc add one-shot job function
func (sf *Timing) AddOneShotJobFunc(f JobFunc, interval time.Duration) *Entry {
	return sf.AddJob(f, OneShot, interval)
}

// AddPersistJob add persist job
func (sf *Timing) AddPersistJob(job Job, interval time.Duration) *Entry {
	return sf.AddJob(job, Persist, interval)
}

// AddPersistJobFunc add persist job function
func (sf *Timing) AddPersistJobFunc(f JobFunc, interval time.Duration) *Entry {
	return sf.AddJob(f, Persist, interval)
}

// Start start the entry
func (sf *Timing) Start(e *Entry, newInterval ...time.Duration) {
	var val = time.Duration(-1) // only active this entry

	if e == nil {
		return
	}

	if len(newInterval) > 0 {
		val = newInterval[0]
	}

	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.running {
		sf.active <- mdEntry{e, val}
	} else if val > 0 {
		e.interval = val
		if !sf.hasEntry(e) {
			sf.entries = append(sf.entries, e)
		}
	}
}

// Remove entry form timing
func (sf *Timing) Remove(e *Entry) {
	if e == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.remove <- e
	} else {
		sf.removeEntry(e)
	}
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
			sf.pf(err)
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
		entry.next = now.Add(entry.interval)
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
			// If there are no entries yet, just sleep -
			//it still handles new entries and stop requests.
			timer = time.NewTimer(100000 * time.Hour)
		} else {
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

					e.count++
					if e.number == 0 || e.count < e.number {
						e.prev = e.next
						e.next = now.Add(e.interval)
					} else {
						e.next = time.Time{} // mark it, not work until remove it
					}
				}
			case newEntry := <-sf.add:
				timer.Stop()
				now = Now()
				newEntry.next = now.Add(newEntry.interval)
				sf.entries = append(sf.entries, newEntry)
				sf.Debug("added: now - %s, next - %s, entry - %p",
					now, newEntry.next, newEntry)
			case mdEntry := <-sf.active:
				timer.Stop()
				entry := mdEntry.entry
				if mdEntry.interval > 0 { // if interval < 0 only active entry
					entry.interval = mdEntry.interval
				}
				if !sf.hasEntry(entry) {
					sf.entries = append(sf.entries, entry)
				}
				now = Now()
				entry.next = now.Add(entry.interval)
				sf.Debug("actived: now - %s, next - %s, entry - %p",
					now, entry.next, entry)
			case e := <-sf.remove:
				timer.Stop()
				now = Now()
				sf.removeEntry(e)
				sf.Debug("removed: entry - %p", e)

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

func (sf *Timing) hasEntry(e *Entry) bool {
	for _, v := range sf.entries {
		if e == v {
			return true
		}
	}
	return false
}

func (sf *Timing) removeEntry(e *Entry) {
	entries := make([]*Entry, 0, len(sf.entries))
	for _, v := range sf.entries {
		if e != v {
			entries = append(entries, v)
		}
	}
	sf.entries = entries
}

// entrySnapshot returns a copy of the current cron entry list.
func (sf *Timing) entrySnapshot() []Entry {
	var entries = make([]Entry, len(sf.entries))
	for i, e := range sf.entries {
		entries[i] = *e
	}
	return entries
}
