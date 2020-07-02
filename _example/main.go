package main

import (
	"log"
	"time"

	"github.com/thinkgos/timing/v4"
)

func main() {
	base := timing.New().Run()

	tm := timing.NewTimer(time.Second * 7)
	tm.WithJobFunc(func() {
		log.Println("hello 7")
		base.Add(tm)
	})

	tm1 := timing.NewTimer(time.Second * 5)
	tm1.WithJobFunc(func() {
		log.Println("hello 5")
		base.Add(tm1)
	})
	base.Add(tm)
	base.Add(tm1)
	time.Sleep(time.Second * 60)
}
