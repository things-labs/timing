package main

import (
	"log"
	"time"

	"github.com/thinkgos/timing"
)

func main() {
	tim := timing.New()
	tim.Start()
	tim.AddCronJob(&myjob{time.Second * 10})
	tim.AddCronJob(&myjob{time.Second * 15})
	tim.AddCronJob(&myjob{time.Second * 5})
	for {
		time.Sleep(time.Minute * 10)
	}
}

type myjob struct {
	timeout time.Duration
}

func (sf myjob) Deploy() (time.Duration, int) {
	return sf.timeout, 0
}

func (sf myjob) Run() bool {
	log.Println(sf.timeout)
	return true
}
