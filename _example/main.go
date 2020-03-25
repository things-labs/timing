package main

import (
	"log"
	"time"

	"github.com/thinkgos/timing/v2"
)

func main() {
	var tm *timing.Entry

	t := timing.New(timing.WithEnableLogger()).Run()

	e1 := t.AddPersistJob(timing.JobFunc(func() {
		panic("haha")
	}), time.Second*1)
	// 此条目使用goroutine
	tm = timing.NewEntry(timing.JobFunc(func() {
		log.Println("in goroutine")
		t.Start(tm)
	}), timing.OneShot, time.Second*2).WithGoroutine(true)

	t.Start(tm)

	time.Sleep(time.Second * 30)
	// 如果相应条目不使用,需要使用remove删除
	t.Remove(e1)
	t.Close()
	time.Sleep(time.Second * 5)
}
