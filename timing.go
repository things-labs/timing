// Package timing 实现定时调度功能,不宜执行任务繁重的任务.
package timing

import (
	"sync"
	"time"
)

// num 定义
const (
	OneShot = 1
	Persist = 0
)

const (
	// DefaultInterval 默认间隔
	DefaultInterval = time.Second
	// DefaultGranularity 默认精度时基
	DefaultGranularity = time.Millisecond * 100
)

// Entry 条目
type Entry struct {
	// next 下一次运行时间  0: 表示未运行,或未启动
	next time.Time
	// 任务已经执行的次数
	count uint32
	//任务需要执行的次数
	number uint32
	// 时间间隔
	interval time.Duration
	// 任务
	job Job
}

// Timing 定时调度,采用map实现
type Timing struct {
	entries      map[*Entry]struct{}
	granularity  time.Duration // 精度
	interval     time.Duration
	mu           sync.Mutex
	stop         chan struct{}
	running      bool
	useGoroutine bool
}

// New new a timing
func New(opt ...Option) *Timing {
	tim := &Timing{
		entries:     make(map[*Entry]struct{}),
		granularity: DefaultGranularity,
		interval:    DefaultInterval,
		stop:        make(chan struct{}),
	}
	for _, opt := range opt {
		opt(tim)
	}
	return tim
}

// Run 运行,不阻塞
func (sf *Timing) Run() *Timing {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		return sf
	}
	sf.running = true
	go sf.runWork()
	return sf
}

// HasRunning 运行状态
func (sf *Timing) HasRunning() bool {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.running
}

// Close 关闭定时
func (sf *Timing) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.stop <- struct{}{}
		sf.running = false
	}
	return nil
}

// Len 条目的个数
func (sf *Timing) Len() int {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return len(sf.entries)
}

// NewJob 新建一个条目,条目未启动定时
func (sf *Timing) NewJob(job Job, num uint32, interval ...time.Duration) *Entry {
	val := sf.interval
	if len(interval) > 0 {
		val = interval[0]
	}
	return &Entry{
		number:   num,
		interval: val,
		job:      job,
	}
}

// NewJobFunc 新建一个条目,条目未启动定时
func (sf *Timing) NewJobFunc(f JobFunc, num uint32, interval ...time.Duration) *Entry {
	return sf.NewJob(f, num, interval...)
}

// AddJob 添加任务
func (sf *Timing) AddJob(job Job, num uint32, interval ...time.Duration) *Entry {
	entry := sf.NewJob(job, num, interval...)
	entry.next = time.Now().Add(entry.interval)
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.entries[entry] = struct{}{}
	return entry
}

// AddOneShotJob 添加一次性任务
func (sf *Timing) AddOneShotJob(job Job, interval ...time.Duration) *Entry {
	return sf.AddJob(job, OneShot, interval...)
}

// AddPersistJob 添加周期性任务
func (sf *Timing) AddPersistJob(job Job, interval ...time.Duration) *Entry {
	return sf.AddJob(job, Persist, interval...)
}

// AddJobFunc 添加任务函数
func (sf *Timing) AddJobFunc(f JobFunc, num uint32, interval ...time.Duration) *Entry {
	return sf.AddJob(f, num, interval...)
}

// AddOneShotJobFunc 添加一次性任务函数
func (sf *Timing) AddOneShotJobFunc(f JobFunc, interval ...time.Duration) *Entry {
	return sf.AddJob(f, OneShot, interval...)
}

// AddPersistJobFunc 添加周期性函数
func (sf *Timing) AddPersistJobFunc(f JobFunc, interval ...time.Duration) *Entry {
	return sf.AddJob(f, Persist, interval...)
}

// Start 启动或重始启动e的计时
func (sf *Timing) Start(e *Entry) *Timing {
	if e == nil {
		return sf
	}
	sf.mu.Lock()
	e.count = 0
	e.next = time.Now().Add(e.interval)
	sf.entries[e] = struct{}{}
	sf.mu.Unlock()
	return sf
}

// Delete 删除指定条目
func (sf *Timing) Delete(e *Entry) *Timing {
	sf.mu.Lock()
	delete(sf.entries, e)
	sf.mu.Unlock()
	return sf
}

// Modify 修改条目的周期时间
func (sf *Timing) Modify(e *Entry, interval time.Duration) *Timing {
	if e == nil {
		return sf
	}
	sf.mu.Lock()
	e.interval = interval
	sf.mu.Unlock()

	return sf
}

func (sf *Timing) runWork() {
	ticker := time.NewTicker(sf.granularity)
	for {
		select {
		case now := <-ticker.C:
			var job []*Entry
			sf.mu.Lock()
			for e := range sf.entries {
				if e.next.After(now) || e.next.IsZero() {
					continue
				}
				job = append(job, e)
				e.count++
				if e.number == 0 || e.count < e.number {
					e.next = now.Add(e.interval)
				} else {
					delete(sf.entries, e)
				}
			}
			sf.mu.Unlock()
			for _, v := range job {
				if sf.useGoroutine {
					go v.job.Run()
				} else {
					wrapJob(v.job)
				}

			}

		case <-sf.stop:
			ticker.Stop()
			return
		}
	}
}
