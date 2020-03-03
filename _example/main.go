package main

import (
	"time"

	"github.com/thinkgos/timing"
)

func main() {
	var tm *timing.Entry
	t := timing.New(timing.WithGoroutine(false),
		timing.WithEnableLogger()).Run()
	defer t.Close()
	t.AddPersistJob(timing.JobFunc(func() {
		panic("haha")
	}), time.Second*1)
	//t.AddPersistJob(timing.JobFunc(func() {
	//}), time.Second*1)
	//t.AddPersistJob(timing.JobFunc(func() {
	//}), time.Second*1)
	//t.AddPersistJob(timing.JobFunc(func() {
	//}), time.Second*1)
	//t.AddPersistJob(timing.JobFunc(func() {
	//}), time.Second*1)
	tm = timing.NewEntry(timing.JobFunc(func() {
		t.Start(tm)
	}), timing.OneShot, time.Second*2)
	t.Start(tm)

	select {}

}
