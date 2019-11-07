// Package timing 实现定时调度功能,采用最小堆实现,不宜执行任务繁重的任务.
package timing

import (
	"container/heap"
	"sync"
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
	needFix  chan struct{}
	stop     chan struct{}

	running bool
	mu      sync.Mutex
}

// New new a timing
func New() *Timing {
	return &Timing{
		addEntry: make(chan *Entry, 32),
		delEntry: make(chan *Entry, 32),
		stop:     make(chan struct{}),
		needFix:  make(chan struct{}, 1),
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

// HasRunning 是否已运行
func (sf *Timing) HasRunning() bool {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.running
}

// AddJob 添加任务
func (sf *Timing) AddJob(job Job, interval time.Duration, num int32) *Entry {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	entry := &Entry{
		number:   NewInt32(num),
		interval: NewDuration(interval),
		job:      job,
	}
	if sf.running {
		sf.addEntry <- entry
	} else {
		sf.entries.Push(entry)
	}
	return entry
}

// AddPersistJob 添加周期性任务
func (sf *Timing) AddPersistJob(job Job, interval time.Duration) *Entry {
	return sf.AddJob(job, interval, persist)
}

// AddOneShotJob 添加一次性任务
func (sf *Timing) AddOneShotJob(job Job, interval time.Duration) *Entry {
	return sf.AddJob(job, interval, oneShot)
}

// AddJobFunc 添加任务函数
func (sf *Timing) AddJobFunc(f JobFunc, interval time.Duration, num int32) *Entry {
	return sf.AddJob(f, interval, num)
}

// AddPersistJobFunc 添加周期性函数
func (sf *Timing) AddPersistJobFunc(f JobFunc, interval time.Duration) *Entry {
	return sf.AddJob(f, interval, persist)
}

// AddOneShotJobFunc 添加一次性任务函数
func (sf *Timing) AddOneShotJobFunc(f JobFunc, interval time.Duration) *Entry {
	return sf.AddJob(f, interval, oneShot)
}

// Delete 删除指定条目
func (sf *Timing) Delete(e *Entry) {
	if e == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		e.number.Store(del) // 标记删除,当正在执行时,可以即时删除,防止正在执行却
		select {
		case sf.delEntry <- e:
		default: // may not block,because it may be call by job
		}
	} else {
		sf.entries.remove(e)
	}
}

// Modify 修改条目的周期时间
func (sf *Timing) Modify(e *Entry, interval time.Duration) {
	if e == nil {
		return
	}
	e.interval.Store(interval)
	select {
	case sf.needFix <- struct{}{}:
	default:
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
		if e.number.Load() >= 0 {
			e.next = now.Add(e.interval.Load())
		}
	}

	for {
		// 排个序
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
				if e.number.Load() < 0 { // 是标记过删除,时间到,但未及时删除的,删除之
					heap.Pop(&sf.entries)
					continue
				}

				if e.next.After(now) || e.next.IsZero() {
					break
				}

				sf.entries.Swap(0, sf.entries.Len()-1)
				sf.entries.Pop()

				e.job.Run()

				e.count++
				cnt := e.number.Load()
				if cnt < 0 || (cnt > 0 && e.count >= cnt) {
					heap.Init(&sf.entries)
					continue
				}
				e.next = now.Add(e.interval.Load())
				heap.Push(&sf.entries, e)
			}

		case newEntry := <-sf.addEntry:
			tm.Stop()
			if newEntry.number.Load() >= 0 {
				newEntry.next = time.Now().Add(newEntry.interval.Load())
				sf.entries.Push(newEntry)
			}

		case e := <-sf.delEntry:
			tm.Stop()
			sf.entries.remove(e)

		case <-sf.stop:
			tm.Stop()
			return
		}
	}
}
