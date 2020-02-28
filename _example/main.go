package main

import (
	"fmt"

	"github.com/thinkgos/timing"
)

func main() {
	t := timing.New(timing.WithGoroutine(true),
		timing.WithLogger()).Run()

	t.AddJob(timing.JobFunc(func() {
		fmt.Println("hello world")
	}), timing.Persist)

	select {}
}
