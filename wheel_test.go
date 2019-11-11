package timing

import (
	"fmt"
	"time"
)

func ExampleNewWheel() {
	tim := NewWheel().Run()

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
