package timing

import (
	"container/heap"
	"sync"
	"time"
)

// Timing 定时调度
type Timing struct {
	entries  entryByTime
	addEntry chan *Entry
	delEntry chan *Entry

	running bool
	mu      sync.Mutex
	stop    chan struct{}
}

// New new a timing
func New() *Timing {
	return &Timing{
		addEntry: make(chan *Entry, 32),
		delEntry: make(chan *Entry, 32),
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

// AddJob 添加定时任务
func (sf *Timing) AddJob(job Job) *Entry {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	entry := &Entry{
		job: job,
	}
	if !sf.running {
		sf.entries.Push(entry)
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
		}
	}
	heap.Init(&sf.entries)
	for {
		var tm *time.Timer
		// 获得最近超时的时间
		if len(sf.entries) == 0 || sf.entries[0].next.IsZero() {
			tm = time.NewTimer(time.Hour * 10000)
		} else {
			tm = time.NewTimer(time.Until(sf.entries[0].next))
		}
		select {
		case now := <-tm.C:
			for len(sf.entries) > 0 {
				e := sf.entries[0]
				if e.next.After(now) || e.next.IsZero() {
					break
				}
				heap.Pop(&sf.entries)
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
				heap.Push(&sf.entries, e)
			}

		case newEntry := <-sf.addEntry:
			tm.Stop()
			if timeout, cnt := newEntry.job.Deploy(); cnt.Load() >= 0 {
				newEntry.next = time.Now().Add(timeout.Load())
				heap.Push(&sf.entries, newEntry)
			}

		case e := <-sf.delEntry:
			tm.Stop()
			sf.entries.remove(e)
			heap.Init(&sf.entries)
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
		sf.entries.remove(e)
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
