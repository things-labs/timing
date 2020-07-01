package timing

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

var ErrClosed = errors.New("timing is closed")

// Base keeps track of any number of entries.
type Base struct {
	data    *heapData
	mu      sync.Mutex
	cond    sync.Cond
	running bool
	logger
}

// New new a time with option
func New(opts ...Option) *Base {
	tim := &Base{
		data: &heapData{
			queue: make([]*Timer, 0),
			items: make(map[*Timer]struct{}),
		},
		logger: newLogger("timing: "),
	}
	tim.cond.L = &tim.mu
	for _, opt := range opts {
		opt(tim)
	}
	return tim
}

// Close the time
func (sf *Base) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.running = false
		sf.cond.Broadcast()
	}
	return nil
}

// HasRunning 运行状态
func (sf *Base) HasRunning() bool {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.running
}

func (sf *Base) Len() int {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.data.Len()
}

func (sf *Base) Add(tm *Timer) error {
	if tm == nil {
		return nil
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if !sf.running {
		return ErrClosed
	}
	sf.addTimer(tm)
	return nil
}

// AddJob add a job
func (sf *Base) AddJob(job Job, timeout time.Duration) error {
	return sf.Add(NewJob(job, timeout))
}

// AddJobFunc add a job function
func (sf *Base) AddJobFunc(f JobFunc, timeout time.Duration) error {
	return sf.AddJob(f, timeout)
}

func (sf *Base) Delete(tm *Timer) error {
	if tm == nil {
		return nil
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if !sf.running {
		return ErrClosed
	}
	if sf.data.contains(tm) {
		delete(sf.data.items, tm)
		heap.Remove(sf.data, tm.index)
	}
	return nil
}

// Modify 修改条目的周期时间,重置计数且重新启动定时器
func (sf *Base) Modify(tm *Timer, timeout time.Duration) {
	if tm == nil {
		return
	}

	sf.mu.Lock()
	tm.timeout = timeout
	sf.addTimer(tm)
	sf.mu.Unlock()
}

// Run the timing in its own goroutine, or no-op if already started.
func (sf *Base) Run() *Base {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		return sf
	}
	sf.running = true
	go sf.run()
	return sf
}

func (sf *Base) addTimer(tm *Timer) {
	if sf.data.contains(tm) {
		tm.next = time.Now().Add(tm.timeout)
		heap.Fix(sf.data, tm.index)
	} else {
		tm.next = time.Now().Add(tm.timeout)
		heap.Push(sf.data, tm)
	}
}

func (sf *Base) run() {
	sf.Debug("run start!")

	notice := make(chan time.Duration)
	closed := make(chan struct{})
	// if time
	tm := time.NewTimer(time.Millisecond)
	defer tm.Stop()

	go func() {

		sf.Debug("work start!")
		for {
			select {
			case <-tm.C:
				sf.cond.Broadcast()
			case d := <-notice:
				tm.Reset(d)
			case <-closed:
				sf.Debug("work stop!")
				return
			}
		}
	}()

	for {
		sf.mu.Lock()
		if !sf.running {
			sf.mu.Unlock()
			closed <- struct{}{}
			return
		}
		near := sf.data.peek()
		if near == nil {
			notice <- time.Hour * 365 * 24
			sf.cond.Wait()
			sf.mu.Unlock()
			continue
		} else {
			now := time.Now()
			if near.next.After(now) {
				d := near.next.Sub(now)
				sf.mu.Unlock()
				notice <- d
				continue
			}
		}
		heap.Pop(sf.data)
		sf.mu.Unlock()
		sf.Debug("run: next - %s, entry - %p", near.next, near)
		if near.useGoroutine {
			go near.job.Run()
		} else {
			wrapJob(near.job)
		}
	}
}
