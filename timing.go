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
	// DefaultInterval default interval
	DefaultInterval = time.Second
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
	mod          chan mdEntry
	active       chan *Entry
	snapshot     chan chan []Entry
	running      bool
	useGoroutine uint32
	interval     time.Duration
	mu           sync.Mutex
	location     *time.Location
	logger
}

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

// New new a time with option
func New(opts ...Option) *Timing {
	tim := &Timing{
		entries:  make([]*Entry, 0),
		add:      make(chan *Entry),
		remove:   make(chan *Entry),
		mod:      make(chan mdEntry),
		active:   make(chan *Entry),
		stop:     make(chan struct{}),
		snapshot: make(chan chan []Entry),
		location: time.Local,
		interval: DefaultInterval,
		logger:   newLogger("timing: "),
	}

	for _, opt := range opts {
		opt(tim)
	}

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
	}
	return nil
}

// UseGoroutine use goroutine or callback
func (sf *Timing) UseGoroutine(b bool) {
	if b {
		atomic.StoreUint32(&sf.useGoroutine, 1)
	} else {
		atomic.StoreUint32(&sf.useGoroutine, 0)
	}
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
func (sf *Timing) AddJob(job Job, num uint32, interval ...time.Duration) *Entry {
	entry := &Entry{
		number:   num,
		interval: append(interval, sf.interval)[0],
		job:      job,
	}

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
func (sf *Timing) AddJobFunc(f JobFunc, num uint32, interval ...time.Duration) *Entry {
	return sf.AddJob(f, num, interval...)
}

// AddOneShotJob  add one-shot job
func (sf *Timing) AddOneShotJob(job Job, interval ...time.Duration) *Entry {
	return sf.AddJob(job, OneShot, interval...)
}

// AddOneShotJobFunc add one-shot job function
func (sf *Timing) AddOneShotJobFunc(f JobFunc, interval ...time.Duration) *Entry {
	return sf.AddJob(f, OneShot, interval...)
}

// AddPersistJob add persist job
func (sf *Timing) AddPersistJob(job Job, interval ...time.Duration) *Entry {
	return sf.AddJob(job, Persist, interval...)
}

// AddPersistJobFunc add persist job function
func (sf *Timing) AddPersistJobFunc(f JobFunc, interval ...time.Duration) *Entry {
	return sf.AddJob(f, Persist, interval...)
}

// Start start the entry
func (sf *Timing) Start(e *Entry) {
	if e == nil {
		return
	}

	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.active <- e
	}
}

// Modify entry interval,if entry is not on list ,it will be add it.
// this will start the entry
func (sf *Timing) Modify(e *Entry, interval time.Duration) {
	if e == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.mod <- mdEntry{e, interval}
	} else {
		sf.modifyEntry(e, interval)
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

func (sf *Timing) run() {
	sf.Debug("start")

	Now := func() time.Time { return time.Now().In(sf.location) }

	// Figure out the next activation times for each entry.
	now := Now()
	for _, entry := range sf.entries {
		entry.next = now.Add(entry.interval)
		sf.Debug("next active: now - %s, next - %s", now, entry.next)
	}

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

					go e.job.Run()
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
			case md := <-sf.mod:
				timer.Stop()
				sf.modifyEntry(md.entry, md.interval)
				now = Now()
				md.entry.next = now.Add(md.interval)
				sf.Debug("modify: now - %s, next - %s, entry - %p",
					now, md.entry.next, md.entry)
			case e := <-sf.active:
				timer.Stop()
				now = Now()
				e.next = now.Add(e.interval)
				if !sf.hasEntry(e) {
					sf.entries = append(sf.entries, e)
				}
				sf.Debug("actived: now - %s, next - %s, entry - %p",
					now, e.next, e)
			case e := <-sf.remove:
				timer.Stop()
				now = Now()
				sf.removeEntry(e)
				sf.Debug("removed: entry - %p", e)

			case replyChan := <-sf.snapshot:
				replyChan <- sf.entrySnapshot()
				continue

			case <-sf.stop:
				timer.Stop()
				sf.Debug("stop")
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

func (sf *Timing) modifyEntry(e *Entry, interval time.Duration) {
	e.interval = interval
	if !sf.hasEntry(e) {
		sf.entries = append(sf.entries, e)
	}
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
