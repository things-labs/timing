package main

import (
	"log"
	"time"

	"github.com/thinkgos/timing"
)

func main() {
	tim := timing.New()
	tim.Start()
	e1 := tim.AddJob(&myjob{10}, time.Second*10, 0)
	tim.AddJob(&myjob{5}, time.Second*5, 0)
	tim.AddJob(&myjob{15}, time.Second*15, 0)
	tim.AddOneShotJob(&myjob{1}, time.Second)
	tim.AddJobFunc(func() {

	}, time.Second, 0)
	go func() {
		time.Sleep(time.Second * 20)
		tim.Delete(e1)
	}()
	for {
		time.Sleep(time.Minute * 10)
	}

}

type myjob struct {
	index int
}

func (sf myjob) Run() {
	log.Println(sf.index)
}
