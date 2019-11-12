package timing

import (
	"fmt"
	"testing"
	"time"
)

func TestDefaultTiming(t *testing.T) {
	if got := Len(); got != 0 {
		t.Errorf("Len() = %v, want %v", got, 0)
	}
	if got := HasRunning(); got != true {
		t.Errorf("HasRunning() = %v, want %v", got, true)
	}

	AddPersistJobFunc(func() {}, time.Millisecond*100)
	AddPersistJob(&emptyJob{})
	e := NewJobFunc(func() {}, 2, time.Millisecond*100)
	Start(e)
	Delete(e)
	Modify(e, time.Millisecond*200)
	time.Sleep(time.Second)
}

func ExampleAddJob() {
	AddOneShotJobFunc(func() {
		fmt.Println("1")
	}, time.Millisecond*100)
	AddJobFunc(func() {
		fmt.Println("2")
	}, OneShot, time.Millisecond*200)
	AddOneShotJob(&testJob{}, time.Millisecond*300)
	AddJob(&testJob{}, 2, time.Millisecond*400)
	UseGoroutine(true)
	time.Sleep(time.Second * 2)
	// Output:
	// 1
	// 2
	// job
	// job
	// job
}
