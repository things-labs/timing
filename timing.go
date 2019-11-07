package timing

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

const (
	oneShot = 1
	persist = 0
	del     = -1
)

// Timing 定时调度
type Timing struct {
	entries  entriesByTime
	addEntry chan *Entry
	delEntry chan *Entry
	modify   chan modInterval
	stop     chan struct{}

	running bool
	mu      sync.Mutex
}

type modInterval struct {
	entry    *Entry
	interval time.Duration
}

// New new a timing
func New() *Timing {
	return &Timing{
		addEntry: make(chan *Entry, 32),
		delEntry: make(chan *Entry, 32),
		modify:   make(chan modInterval, 32),
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
func (sf *Timing) AddJob(job Job, interval time.Duration, num int32) *Entry {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	entry := &Entry{
		number:   num,
		interval: interval,
		job:      job,
	}
	if sf.running {
		sf.addEntry <- entry
	} else {
		sf.entries.Push(entry)
	}
	return entry
}

func (sf *Timing) AddPersistJob(job Job, interval time.Duration) *Entry {
	return sf.AddJob(job, interval, persist)
}

func (sf *Timing) AddOneShotJob(job Job, interval time.Duration) *Entry {
	return sf.AddJob(job, interval, oneShot)
}

// Delete 删除指定id的条目
func (sf *Timing) Delete(e *Entry) {
	if e == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		atomic.StoreInt32(&e.number, del) // 标记删除,当正在执行时,可以即时删除
		sf.delEntry <- e
	} else {
		sf.entries.remove(e)
	}
}

//
func (sf *Timing) Modify(e *Entry, interval time.Duration) {
	if e == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.modify <- modInterval{
			entry:    e,
			interval: interval,
		}
	} else {
		e.interval = interval
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

func (sf *Timing) run() {
	now := time.Now()
	for _, e := range sf.entries {
		if e.number >= 0 {
			e.next = now.Add(e.interval)
		}
	}

	for {
		heap.Init(&sf.entries)

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
				cnt := atomic.LoadInt32(&e.number)
				if cnt < 0 { // 是标记过删除,时间到,但未及时删除的,删除之
					heap.Pop(&sf.entries)
					continue
				}

				if e.next.After(now) || e.next.IsZero() {
					break
				}
				e.count++
				heap.Pop(&sf.entries)

				if !e.job.Run() {
					continue
				}
				if cnt > 0 && e.count >= cnt {
					continue
				}
				e.next = now.Add(e.interval)
				heap.Push(&sf.entries, e)
			}

		case newEntry := <-sf.addEntry:
			tm.Stop()
			if newEntry.number >= 0 {
				newEntry.next = time.Now().Add(newEntry.interval)
				sf.entries.Push(newEntry)
			}

		case e := <-sf.delEntry:
			tm.Stop()
			sf.entries.remove(e)

		case md := <-sf.modify:
			md.entry.interval = md.interval

		case <-sf.stop:
			tm.Stop()
			return
		}
	}
}
