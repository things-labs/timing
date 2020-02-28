package main

import (
	"time"

	"github.com/thinkgos/timing"
)

type gp struct{}

func (gp) Submit(job timing.Job) {
	job.Run()
}

func main() {
	p := gp{}
	var tm *timing.Entry
	t := timing.New(timing.WithGoroutine(false, p),
		timing.WithLogger()).Run()
	defer t.Close()
	t.AddPersistJob(timing.JobFunc(func() {
	}), time.Second*2)
	tm = timing.NewEntry(timing.JobFunc(func() {
		t.Start(tm)
	}), timing.OneShot, time.Second*2)
	t.Start(tm)
	//go func() {
	//	for {
	//		time.Sleep(time.Second * 5)
	//		t.Modify(tm, 2*time.Second)
	//	}
	//
	//}()
	select {}

}
