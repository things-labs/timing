package timing

import (
	"container/heap"
	"errors"
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
	if sf.running {
		sf.running = false
		sf.cond.Broadcast()
	}
	sf.mu.Unlock()
	return nil
}

// HasRunning base running status.
func (sf *Base) HasRunning() (b bool) {
	sf.mu.Lock()
	b = sf.running
	sf.mu.Unlock()
	return
}

// Len the number timer of the base.
func (sf *Base) Len() (length int) {
	sf.mu.Lock()
	length = sf.data.Len()
	sf.mu.Unlock()
	return
}

// AddJob add a job and start immediately.
func (sf *Base) AddJob(job Job, timeout time.Duration) *Timer {
	tm := NewJob(job)
	sf.Add(tm, timeout)
	return tm
}

// AddJobFunc add a job function and start immediately.
func (sf *Base) AddJobFunc(f JobFunc, timeout time.Duration) *Timer {
	return sf.AddJob(f, timeout)
}

// Add add timer to base and start immediately.
func (sf *Base) Add(tm *Timer, timeout time.Duration) {
	if tm == nil {
		return
	}
	sf.mu.Lock()
	sf.start(tm, timeout)
	sf.mu.Unlock()
}

// Delete delete timer from timer base.
func (sf *Base) Delete(tm *Timer) {
	if tm == nil {
		return
	}
	sf.mu.Lock()
	if sf.data.contains(tm) {
		delete(sf.data.items, tm)
		heap.Remove(sf.data, tm.index)
		sf.cond.Broadcast()
	}
	sf.mu.Unlock()
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

func (sf *Base) start(tm *Timer, timeout time.Duration) {
	tm.next = time.Now().Add(timeout)
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

	go func() {
		var d = time.Hour * 365 * 24

		for {
			tm := time.NewTimer(d)
			select {
			case <-tm.C:
				sf.cond.Broadcast()
			case d = <-notice:
				tm.Stop()
			case <-closed:
				tm.Stop()
				return
			}
		}
	}()

	for {
		item, err := sf.pop(notice)
		if err != nil {
			closed <- struct{}{}
			return
		}
		if item.job != nil {
			if item.useGoroutine {
				go item.job.Run()
			} else {
				wrapJob(item.job)
			}
		}
	}
}

func (sf *Base) pop(notice chan<- time.Duration) (item *Timer, err error) {
	var d time.Duration

	sf.mu.Lock()
	defer sf.mu.Unlock()
	for {
		if !sf.running {
			err = errors.New("base is closed")
			return
		}

		if item = sf.data.peek(); item != nil {
			now := time.Now()
			if item.next.Before(now) {
				heap.Pop(sf.data)
				return
			}
			d = item.next.Sub(now)
		} else {
			d = time.Hour * 365 * 24
		}
		notice <- d
		sf.cond.Wait()
	}
}
