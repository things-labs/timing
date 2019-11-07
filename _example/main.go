package main

import (
	"log"
	"time"

	"github.com/thinkgos/timing"
)

func main() {

	tim := timing.New()
	tim.Start()
	tim.AddCronJob(&myjob{timing.NewDuration(time.Second * 10)})
	tim.AddCronJob(&myjob{timing.NewDuration(time.Second * 15)})
	tim.AddCronJob(&myjob{timing.NewDuration(time.Second * 5)})
	for {
		time.Sleep(time.Minute * 10)
	}

}

type myjob struct {
	timeout *timing.Duration
}

func (sf myjob) Deploy() (*timing.Duration, *timing.Int32) {
	return sf.timeout, timing.NewInt32(0)
}

func (sf myjob) Run() bool {
	log.Println(sf.timeout.Load())
	return true
}
