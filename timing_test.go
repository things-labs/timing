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
	tim := New(WithEnableLogger())

	tim.AddJobFunc(func() {}, time.Millisecond*100)
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

	tim.AddJobFunc(func() {}, time.Second)
	tim.Run()
	tim.AddJob(&emptyJob{}, time.Second)
	tim.AddJob(&emptyJob{}, time.Second*30)
	if got := len(tim.Entries()); got != 3 {
		t.Errorf("HasRunning() = %v, want %v", got, 3)
	}
	time.Sleep(time.Second * 2)
	if got := len(tim.Entries()); got != 3 {
		t.Errorf("HasRunning() = %v, want %v", got, 3)
	}

	tim.Location()
}

func ExampleNew() {
	tim := New().Run()

	tim.AddJobFunc(func() {
		fmt.Println("1")
	}, time.Millisecond*100, WithGoroutine())
	tim.AddJobFunc(func() {
		fmt.Println("2")
	}, time.Millisecond*200)
	tim.AddJob(&testJob{}, time.Millisecond*300)
	tim.AddJob(&testJob{}, time.Millisecond*400)
	tim.AddJobFunc(func() {
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
}
