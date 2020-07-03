package timing

import (
	"fmt"
	"testing"
	"time"
)

func TestTiming(t *testing.T) {
	if got := HasRunning(); got {
		t.Errorf("HasRunning() = %v, want %v", got, false)
	}

	if got := Len(); got != 0 {
		t.Errorf("Len() = %v, want %v", got, 0)
	}

	e := NewJobFunc(func() {})
	Add(e, time.Millisecond*100)
	Delete(e)
	Modify(e, time.Millisecond*200)
	time.Sleep(time.Second)

	if got := HasRunning(); !got {
		t.Errorf("HasRunning() = %v, want %v", got, true)
	}

	e1 := NewTimer().WithGoroutine()
	Add(e1, time.Millisecond*150)

	e2 := NewTimer().WithGoroutine()
	Add(e2, time.Millisecond)
	time.Sleep(time.Second)

	// improve couver
	Modify(nil, time.Second)
	Delete(nil)
	Add(nil, time.Second)
}

func ExampleBase_Len() {
	AddJobFunc(func() {
		fmt.Println("1")
	}, time.Millisecond*100)
	AddJobFunc(func() {
		fmt.Println("2")
	}, time.Millisecond*200)
	AddJob(&testJob{}, time.Millisecond*300)
	time.Sleep(time.Second * 2)
	// Output:
	// 1
	// 2
	// job
}
