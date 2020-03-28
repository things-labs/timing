package main

import (
	"log"
	"time"

	"github.com/thinkgos/timing/v3"
)

func main() {
	var f func()
	t := timing.New(timing.WithEnableLogger()).Run()
	f = func() {
		log.Println("period")
		t.AddJobFunc(f, time.Second*1)
	}
	t.AddJobFunc(f, time.Second*1)

	t.AddJobFunc(func() {
		log.Println("1")
	}, time.Second*1)
	t.AddJobFunc(func() {
		log.Println("2")
	}, time.Second*2)
	t.AddJobFunc(func() {
		log.Println("3")
	}, time.Second*3)
	t.AddJobFunc(func() {
		log.Println("4")
	}, time.Second*4)

	time.Sleep(time.Second * 30)
	t.Close()
	time.Sleep(time.Second * 5)
}
