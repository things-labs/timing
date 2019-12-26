package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/thinkgos/timing"
)

func main() {
	var periodic timing.Timer

	periodic = timing.NewJobFunc(func() {
		log.Printf("what a fuck")
		timing.Start(periodic, time.Duration(rand.Intn(20))*time.Second+20*time.Second)
	}, timing.OneShot, time.Second*2)
	timing.Start(periodic)
	select {}
}
