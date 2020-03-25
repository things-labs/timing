package timing

import (
	"fmt"
	"testing"
	"time"
)

type emptyJob struct{}

func (emptyJob) Run() {}

type testJob struct{}

func (sf testJob) Run() {
	fmt.Println("job")
}

func TestHashes(t *testing.T) {
	tim := New()
	tim.AddOneShotJobFunc(func() {}, time.Millisecond*100)

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

func TestHashesJob(t *testing.T) {
	tim := New()
	e1 := tim.AddPersistJobFunc(func() {}, time.Second)
	tim.Start(tim.AddJobFunc(func() {}, Persist, time.Second), time.Second*2)
	tim.Run()
	tim.AddPersistJob(&emptyJob{}, time.Second*30)
	tim.AddJob(&emptyJob{}, Persist, time.Second)
	if got := len(tim.Entries()); got != 4 {
		t.Errorf("HasRunning() = %v, want %v", got, 4)
	}
	tim.Start(e1)

	tim.Remove(e1)
	if got := len(tim.Entries()); got != 3 {
		t.Errorf("HasRunning() = %v, want %v", got, 3)
	}

	tim.Start(e1, time.Second*2)
	tim.Start(nil)
	tim.Start(nil, time.Second)
	tim.Remove(nil)
	tim.Location()
}

func ExampleNew() {
	tim := New().Run()

	tim.AddOneShotJobFunc(func() {
		fmt.Println("1")
	}, time.Millisecond*100).WithGoroutine(true)
	tim.AddJobFunc(func() {
		fmt.Println("2")
	}, OneShot, time.Millisecond*200)
	tim.AddOneShotJob(&testJob{}, time.Millisecond*300)
	tim.AddJob(&testJob{}, 2, time.Millisecond*400)
	tim.AddOneShotJobFunc(func() {
		defer func() {
			_ = recover()
		}()
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
