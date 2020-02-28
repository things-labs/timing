package main

import (
	"time"

	"github.com/thinkgos/timing"
)

func main() {
	var tm *timing.Entry
	t := timing.New(timing.WithGoroutine(true),
		timing.WithLogger()).Run()

	tm = t.AddOneShotJob(timing.JobFunc(func() {
		t.Modify(tm, time.Second*2)
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
