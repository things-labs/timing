package main

import (
	"fmt"
	"time"

	"github.com/thinkgos/timing/v3"
)

func main() {
	var f func()
	t := timing.New(timing.WithEnableLogger()).Run()
	f = func() {
		fmt.Println("haha")
		//t.AddJobFunc(f, time.Second*1)
	}
	t.AddJobFunc(f, time.Millisecond*500)
	t.AddJobFunc(f, time.Millisecond*600)
	t.AddJobFunc(f, time.Millisecond*700)
	t.AddJobFunc(f, time.Millisecond*800)
	t.AddJobFunc(f, time.Millisecond*900)
	t.AddJobFunc(f, time.Millisecond*1000)

	time.Sleep(time.Second * 30)
	t.Close()
	time.Sleep(time.Second * 5)
}
