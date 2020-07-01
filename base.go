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
	tm := NewJob(job, timeout)
	sf.Add(tm)
	return tm
}

// AddJobFunc add a job function and start immediately.
func (sf *Base) AddJobFunc(f JobFunc, timeout time.Duration) *Timer {
	return sf.AddJob(f, timeout)
}

// Add add timer to base and start immediately.
func (sf *Base) Add(tm *Timer, newTimeout ...time.Duration) {
	if tm == nil {
		return
	}
	sf.mu.Lock()
	sf.start(tm, newTimeout...)
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
	tm := time.NewTimer(time.Hour * 365 * 24)
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
		item, err := sf.pop(closed, notice)
		if err != nil {
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

func (sf *Base) pop(closed chan struct{}, notice chan time.Duration) (item *Timer, err error) {
	var d time.Duration

	sf.mu.Lock()
	defer sf.mu.Unlock()
	for {
		if !sf.running {
			closed <- struct{}{}
			err = errors.New("base is closed")
			return
		}

		if item = sf.data.peek(); item != nil {
			if now := time.Now(); item.next.Before(now) {
				heap.Pop(sf.data)
				return
			} else {
				d = item.next.Sub(now)
			}
		} else {
			d = time.Hour * 365 * 24
		}
		notice <- d
		sf.cond.Wait()
	}
}
