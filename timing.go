package timing

import (
	"container/heap"
	"sync"
	"time"
)

// Timing 定时调度
type Timing struct {
	entries  []*Entry
	addEntry chan *Entry
	delEntry chan *Entry

	running bool
	mu      sync.Mutex
	stop    chan struct{}
}

// New new a timing
func New() *Timing {
	return &Timing{
		entries:  nil,
		addEntry: make(chan *Entry, 16),
		delEntry: make(chan *Entry, 16),
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
func (sf *Timing) AddCronJob(job CronJob) *Entry {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	entry := &Entry{
		job: job,
	}
	if !sf.running {
		sf.entries = append(sf.entries, entry)
	} else {
		sf.addEntry <- entry
	}
	return entry
}

func (sf *Timing) run() {
	now := time.Now()
	for _, e := range sf.entries {
		if timeout, cnt := e.job.Deploy(); cnt.Load() >= 0 {
			e.next = now.Add(timeout.Load())
			sf.entries = append(sf.entries, e)
		}
	}
	heap.Init((*entryByTime)(&sf.entries))
	for {
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
			for len(sf.entries) > 0 {
				e := sf.entries[0]
				if e.next.After(now) || e.next.IsZero() {
					break
				}
				heap.Pop((*entryByTime)(&sf.entries))
				e.count++
				if !e.job.Run() {
					continue
				}

				timeout, aCnt := e.job.Deploy()
				cnt := aCnt.Load()
				if cnt < 0 || (cnt > 0 && e.count >= cnt) {
					continue
				}
				e.next = now.Add(timeout.Load())
				heap.Push((*entryByTime)(&sf.entries), e)
			}

		case newEntry := <-sf.addEntry:
			tm.Stop()
			if timeout, cnt := newEntry.job.Deploy(); cnt.Load() >= 0 {
				newEntry.next = time.Now().Add(timeout.Load())
				heap.Push((*entryByTime)(&sf.entries), newEntry)
			}

		case e := <-sf.delEntry:
			tm.Stop()
			sf.Remove(e)
		case <-sf.stop:
			tm.Stop()
			return
		}
	}
}

// Remove 删除指定id的条目
func (sf *Timing) Remove(e *Entry) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.delEntry <- e
	} else {
		sf.removeEntry(e)
	}
}

func (sf *Timing) removeEntry(entry *Entry) {
	for i, e := range sf.entries {
		if e == entry {
			heap.Fix((*entryByTime)(&sf.entries), i)
			return
		}
	}
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
