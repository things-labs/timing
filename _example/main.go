package main

import (
	"log"
	"time"

	"github.com/things-labs/timing"
)

func main() {
	base := timing.New().Run()

	tm := timing.NewTimer()
	tm.WithJobFunc(func() {
		log.Println("hello 7")
		base.Add(tm, time.Second*7)
	})

	tm1 := timing.NewTimer()
	tm1.WithJobFunc(func() {
		log.Println("hello 5")
		base.Add(tm1, time.Second*5)
	})
	base.Add(tm, time.Second*7)
	base.Add(tm1, time.Second*5)
	time.Sleep(time.Second * 60)
}
