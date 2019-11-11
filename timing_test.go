package timing

import (
	"fmt"
	"testing"
	"time"
)

func TestTiming(t *testing.T) {
	tick := time.Millisecond * 100
	interval := time.Second * 10
	tim := NewHashes(WithInterval(interval), WithGranularity(tick), WithGoroutine())
	tim.AddOneShotJobFunc(func() {}, time.Millisecond*100)
	if got := tim.(*Hashes).interval; got != interval {
		t.Errorf("HasRunning() = %v, want %v", got, interval)
	}

	if got := tim.(*Hashes).granularity; got != tick {
		t.Errorf("HasRunning() = %v, want %v", got, tick)
	}

	tim.Run()
	if got := tim.HasRunning(); got != true {
		t.Errorf("HasRunning() = %v, want %v", got, true)
	}
	tim.Run()
	time.Sleep(time.Millisecond * 200)
	_ = tim.Close()
	if got := tim.HasRunning(); got != false {
		t.Errorf("HasRunning() = %v, want %v", got, false)
	}
}

type emptyJob struct{}

func (emptyJob) Run() {}

func TestJob(t *testing.T) {
	tim := NewHashes(WithInterval(time.Second), WithGranularity(time.Minute)).Run()
	e1 := tim.AddPersistJobFunc(func() {})
	tim.AddJobFunc(func() {}, Persist)
	tim.AddPersistJob(&emptyJob{}, time.Second*30)
	tim.AddJob(&emptyJob{}, Persist)
	if got := tim.Len(); got != 4 {
		t.Errorf("HasRunning() = %v, want %v", got, 4)
	}
	tim.Modify(e1, time.Second*2)

	tim.(*Hashes).mu.Lock()
	interval := e1.(*Entry).interval
	tim.(*Hashes).mu.Unlock()
	if interval != time.Second*2 {
		t.Errorf("HasRunning() = %v, want %v", interval, time.Second*2)
	}

	tim.Start(e1)

	tim.Delete(e1)
	if got := tim.Len(); got != 3 {
		t.Errorf("HasRunning() = %v, want %v", got, 3)
	}

	tim.Start(nil)
	tim.Modify(nil, time.Second)

}

type testJob struct{}

func (sf testJob) Run() {
	fmt.Println("job")
}

func ExampleNewHashes() {
	tim := NewHashes().Run()

	tim.AddOneShotJobFunc(func() {
		fmt.Println("1")
	}, time.Millisecond*100)
	tim.AddJobFunc(func() {
		fmt.Println("2")
	}, OneShot, time.Millisecond*200)
	tim.AddOneShotJob(&testJob{}, time.Millisecond*300)
	tim.AddJob(&testJob{}, 2, time.Millisecond*400)
	tim.AddOneShotJobFunc(func() {
		panic("painc happen")
	}, time.Millisecond*100)
	time.Sleep(time.Second * 2)
	// Output:
	// 1
	// 2
	// job
	// job
	// job
}
