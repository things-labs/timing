package timing

import (
	"container/heap"
	"sync"
	"time"
)

// Base keeps track of any number of entries.
type Base struct {
	data    *heapData
	mu      sync.Mutex
	cond    sync.Cond
	running bool
}

// New new a base with option
func New() *Base {
	tim := &Base{
		data: &heapData{
			queue: make([]*Timer, 0),
			items: make(map[*Timer]struct{}),
		},
	}
	tim.cond.L = &tim.mu
	return tim
}

// Close the base
func (sf *Base) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.running {
		sf.running = false
		sf.cond.Broadcast()
	}
	return nil
}

// HasRunning base running status.
func (sf *Base) HasRunning() bool {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.running
}

// Len the number timer of the base.
func (sf *Base) Len() int {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.data.Len()
}

// AddJob add a job
func (sf *Base) AddJob(job Job, timeout time.Duration) *Timer {
	tm := NewJob(job, timeout)
	sf.Add(tm)
	return tm
}

// AddJobFunc add a job function
func (sf *Base) AddJobFunc(f JobFunc, timeout time.Duration) *Timer {
	return sf.AddJob(f, timeout)
}

// Add add timer to base. and start immediately.
func (sf *Base) Add(tm *Timer, newTimeout ...time.Duration) {
	if tm == nil {
		return
	}
	sf.mu.Lock()
	sf.start(tm, newTimeout...)
	sf.mu.Unlock()
}

// Delete Delete timer from base.
func (sf *Base) Delete(tm *Timer) {
	if tm == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.data.contains(tm) {
		delete(sf.data.items, tm)
		heap.Remove(sf.data, tm.index)
		sf.cond.Broadcast()
	}
}

// Modify modify timer timeout,and restart immediately.
func (sf *Base) Modify(tm *Timer, timeout time.Duration) {
	if tm == nil {
		return
	}
	sf.mu.Lock()
	sf.start(tm, timeout)
	sf.mu.Unlock()
}

// Run the base in its own goroutine, or no-op if already started.
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

func (sf *Base) start(tm *Timer, newTimeout ...time.Duration) {
	if len(newTimeout) > 0 {
		tm.timeout = newTimeout[0]
	}
	tm.next = time.Now().Add(tm.timeout)
	if sf.data.contains(tm) {
		heap.Fix(sf.data, tm.index)
	} else {
		heap.Push(sf.data, tm)
	}
	sf.cond.Broadcast()
}

func (sf *Base) run() {
	notice := make(chan time.Duration)
	closed := make(chan struct{})
	// if time
	tm := time.NewTimer(time.Millisecond)
	defer tm.Stop()

	go func() {
		for {
			select {
			case <-tm.C:
				sf.cond.Broadcast()
			case d := <-notice:
				tm.Reset(d)
			case <-closed:
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
		}
		now := time.Now()
		if near.next.After(now) {
			d := near.next.Sub(now)
			sf.mu.Unlock()
			notice <- d
			continue
		}

		heap.Pop(sf.data)
		sf.mu.Unlock()
		if near.job != nil {
			if near.useGoroutine {
				go near.job.Run()
			} else {
				wrapJob(near.job)
			}
		}
	}
}
