package timing

import (
	"sort"
	"sync"
	"time"
)

type Timing struct {
	entries  []*Entry
	addEntry chan *Entry
	delEntry chan EntryID
	nextID   EntryID

	running bool
	mu      sync.Mutex
	stop    chan struct{}
}

func New() *Timing {
	return &Timing{
		entries:  nil,
		addEntry: make(chan *Entry),
		delEntry: make(chan EntryID),
		stop:     make(chan struct{}),
	}
}

func (sf *Timing) Start() {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		return
	}
	sf.running = true
	go sf.run()
}

func (sf *Timing) AddCronJob(job CronJob) EntryID {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.nextID++
	entry := &Entry{
		ID:  sf.nextID,
		job: job,
	}
	if !sf.running {
		sf.entries = append(sf.entries, entry)
	} else {
		sf.addEntry <- entry
	}
	return sf.nextID
}

func (sf *Timing) run() {
	now := time.Now()
	for _, e := range sf.entries {
		timeout, _ := e.job.Next()
		e.Next = now.Add(timeout)
	}

	for {
		// 排个序
		sort.Sort(entryByTime(sf.entries))

		var ts *time.Timer
		now := time.Now()
		// 获得最近超时的时间
		if entry := entryByTime(sf.entries).peak(); entry != nil {
			ts = time.NewTimer(entry.Next.Sub(now))
		} else {
			ts = time.NewTimer(time.Hour * 10000)
		}
		select {
		case now = <-ts.C:
			for _, e := range sf.entries {
				if e.Next.After(now) || e.Next.IsZero() {
					break
				}
				// do job
				e.job.Run()
				e.count++
				// go next timeout
				timeout, cnt := e.job.Next()
				if cnt == 0 || e.count < cnt {
					e.Next = now.Add(timeout)
				} else {
					// remove it ??
				}
			}
		case newEntry := <-sf.addEntry:
			ts.Stop()
			timeout, _ := newEntry.job.Next()
			newEntry.Next = time.Now().Add(timeout)
			sf.entries = append(sf.entries, newEntry)
		case id := <-sf.delEntry:
			ts.Stop()
			sf.Remove(id)
		case <-sf.stop:
			ts.Stop()
			return
		}
	}
}

func (sf *Timing) Remove(id EntryID) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.delEntry <- id
	} else {
		sf.removeEntry(id)
	}
}

func (sf *Timing) removeEntry(id EntryID) {
	entries := sf.entries
	for i, e := range sf.entries {
		if e.ID == id {
			entries = append(sf.entries[:i], sf.entries[i+1:]...)
			break
		}
	}
	sf.entries = entries
}
