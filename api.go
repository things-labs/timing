package timing

import (
	"io"
	"time"
)

// Timer 定时器对象
type Timer interface{}

// Base 定时基础控制器
type Base interface {
	Run() Base
	HasRunning() bool
	Len() int
	NewJob(job Job, num uint32, interval ...time.Duration) Timer
	NewJobFunc(f JobFunc, num uint32, interval ...time.Duration) Timer
	AddJob(job Job, num uint32, interval ...time.Duration) Timer
	AddOneShotJob(job Job, interval ...time.Duration) Timer
	AddPersistJob(job Job, interval ...time.Duration) Timer
	AddJobFunc(f JobFunc, num uint32, interval ...time.Duration) Timer
	AddOneShotJobFunc(f JobFunc, interval ...time.Duration) Timer
	AddPersistJobFunc(f JobFunc, interval ...time.Duration) Timer
	Start(e Timer) Base
	Delete(e Timer) Base
	Modify(e Timer, interval time.Duration) Base
	io.Closer
}
