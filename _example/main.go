package main

import (
	"time"

	"github.com/thinkgos/timing"
)

func main() {
	var tm *timing.Entry
	t := timing.New(timing.WithGoroutine(false),
		timing.WithLogger()).Run()

	tm = t.AddOneShotJob(timing.JobFunc(func() {
		t.Start(tm)
	}), time.Second)

	//go func() {
	//	for {
	//		time.Sleep(time.Second * 5)
	//		t.Modify(tm, 2*time.Second)
	//	}
	//
	//}()
	select {}
	t.Remove(tm)
}
