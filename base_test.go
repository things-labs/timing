package timing

import (
	"fmt"
	"testing"
	"time"
)

type testJob struct{}

func (sf testJob) Run() {
	fmt.Println("job")
}

func TestBase(t *testing.T) {
	tim := New().Run()

	defer tim.Close()
	if got := tim.Len(); got != 0 {
		t.Errorf("Len() = %v, want %v", got, 0)
	}
	if got := tim.HasRunning(); got != true {
		t.Errorf("HasRunning() = %v, want %v", got, true)
	}

	e := NewJobFunc(func() {})
	tim.Add(e, time.Millisecond*100)
	tim.Delete(e)
	tim.Modify(e, time.Millisecond*200)
	time.Sleep(time.Second)

	e1 := NewTimer().WithGoroutine()
	tim.Add(e1, time.Millisecond*150)

	e2 := NewTimer().WithGoroutine()
	tim.Add(e2, time.Millisecond)
	tim.Add(e2, time.Millisecond*10) //
	time.Sleep(time.Second)

	// improve couver
	tim.Run()
	tim.Modify(nil, time.Second)
	tim.Delete(nil)
	tim.Add(nil, time.Second)
}

func ExampleNew() {
	tim := New().Run()

	tim.AddJobFunc(func() { fmt.Println("1") }, time.Millisecond*100)
	tim.AddJobFunc(func() { fmt.Println("2") }, time.Millisecond*200)
	tim.AddJob(&testJob{}, time.Millisecond*300)
	tim.AddJob(&testJob{}, time.Millisecond*400)
	time.Sleep(time.Second * 2)
	// Output:
	// 1
	// 2
	// job
	// job
}
