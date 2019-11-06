package timing

import (
	"sort"
	"sync"
	"time"
)

// Timing 定时调度
type Timing struct {
	entries  []*Entry
	addEntry chan *Entry
	delEntry chan EntryID
	nextID   EntryID

	running bool
	mu      sync.Mutex
	stop    chan struct{}
}

// New new a timing
func New() *Timing {
	return &Timing{
		entries:  nil,
		addEntry: make(chan *Entry, 16),
		delEntry: make(chan EntryID, 16),
		stop:     make(chan struct{}),
	}
}

// Start 启动
func (sf *Timing) Start() {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		return
	}
	sf.running = true
	go sf.run()
}

// AddCronJob 添加定时任务
func (sf *Timing) AddCronJob(job CronJob) EntryID {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.nextID++
	entry := &Entry{
		id:  sf.nextID,
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
		timeout, _ := e.job.Deploy()
		e.next = now.Add(timeout)
	}

	for {
		// 排个序
		sort.Sort(entryByTime(sf.entries))

		var tm *time.Timer
		now := time.Now()
		// 获得最近超时的时间
		if len(sf.entries) == 0 || sf.entries[0].next.IsZero() {
			tm = time.NewTimer(time.Hour * 10000)

		} else {
			tm = time.NewTimer(sf.entries[0].next.Sub(now))
		}
		select {
		case now = <-tm.C:
			var delIDs []EntryID
			for _, e := range sf.entries {
				if e.next.After(now) || e.next.IsZero() {
					break
				}

				e.count++
				if !e.job.Run() {
					delIDs = append(delIDs, e.id)
					continue
				}

				timeout, cnt := e.job.Deploy()
				if cnt < 0 || (cnt > 0 && e.count >= cnt) {
					delIDs = append(delIDs, e.id)
					continue
				}

				e.next = now.Add(timeout)
			}

			for _, id := range delIDs {
				sf.removeEntry(id)
			}

		case newEntry := <-sf.addEntry:
			tm.Stop()
			if timeout, cnt := newEntry.job.Deploy(); cnt >= 0 {
				newEntry.next = time.Now().Add(timeout)
				sf.entries = append(sf.entries, newEntry)
			}

		case id := <-sf.delEntry:
			tm.Stop()
			sf.Remove(id)
		case <-sf.stop:
			tm.Stop()
			return
		}
	}
}

// Remove 删除指定id的条目
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
		if e.id == id {
			entries = append(sf.entries[:i], sf.entries[i+1:]...)
			break
		}
	}
	sf.entries = entries
}

// Close close
func (sf *Timing) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.stop <- struct{}{}
		sf.running = false
	}
	return nil
}
